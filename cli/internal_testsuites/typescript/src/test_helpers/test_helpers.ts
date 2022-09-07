import {
    ServiceID,
    EnclaveContext,
    ContainerConfig,
    ContainerConfigBuilder,
    ServiceContext,
    PartitionID,
    PortSpec,
    PortProtocol,
    FilesArtifactUUID,
} from "kurtosis-core-api-lib";
import * as datastoreApi from "example-datastore-server-api-lib";
import * as serverApi from "example-api-server-api-lib";
import { err, ok, Result } from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "@grpc/grpc-js";
import log from "loglevel";
import * as fs from 'fs';
import * as path from "path";
import {create} from "domain";
import * as os from "os";
import exp = require("constants");
import axios from "axios";

const CONFIG_FILENAME = "config.json"
const CONFIG_MOUNTPATH_ON_API_CONTAINER = "/config"

const DATASTORE_IMAGE = "kurtosistech/example-datastore-server";
const API_SERVICE_IMAGE = "kurtosistech/example-api-server";

const DATASTORE_PORT_ID = "rpc";
const API_PORT_ID = "rpc";

const DATASTORE_WAIT_FOR_STARTUP_MAX_POLLS = 10;
const DATASTORE_WAIT_FOR_STARTUP_DELAY_MILLISECONDS = 1000;

const API_WAIT_FOR_STARTUP_MAX_POLLS = 10;
const API_WAIT_FOR_STARTUP_DELAY_MILLISECONDS = 1000;

const DEFAULT_PARTITION_ID = "";

const DATASTORE_PORT_SPEC = new PortSpec(
    datastoreApi.LISTEN_PORT,
    PortProtocol.TCP,
)
const API_PORT_SPEC = new PortSpec(
    serverApi.LISTEN_PORT,
    PortProtocol.TCP,
)
const FILE_SERVER_SERVICE_IMAGE         = "flashspys/nginx-static"
const FILE_SERVER_PORT_ID               = "http"
const FILE_SERVER_SERVICE_ID: ServiceID = "file-server"

const USER_SERVICE_MOUNT_POINT_FOR_TEST_FILES_ARTIFACT  = "/static"

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS = 500
const WAIT_FOR_STARTUP_MAX_RETRIES        = 15
const WAIT_INITIAL_DELAY_MILLISECONDS     = 0
const WAIT_FOR_AVAILABILITY_BODY_TEXT     = ""

const FILE_SERVER_PRIVATE_PORT_NUM      = 80
const FILE_SERVER_PORT_SPEC             = new PortSpec( FILE_SERVER_PRIVATE_PORT_NUM, PortProtocol.TCP )

export class StartFileServerResponse  {
    fileServerPublicIp: string
    fileServerPublicPortNum: number

    constructor(fileServerPublicIp: string, fileServerPublicPortNum: number) {
        this.fileServerPublicIp = fileServerPublicIp
        this.fileServerPublicPortNum = fileServerPublicPortNum
    }
}

export async function addDatastoreService(serviceId: ServiceID, enclaveContext: EnclaveContext):
    Promise<Result<{
        serviceContext: ServiceContext;
        client: datastoreApi.DatastoreServiceClientNode;
        clientCloseFunction: () => void;
    },Error>> {
    
    const containerConfigSupplier = getDatastoreContainerConfigSupplier();

    const addServiceResult = await enclaveContext.addService(serviceId, containerConfigSupplier);
    if (addServiceResult.isErr()) {
        log.error("An error occurred adding the datastore service");
        return err(addServiceResult.error);
    }
    const serviceContext = addServiceResult.value;

    const publicPort: PortSpec | undefined = serviceContext.getPublicPorts().get(DATASTORE_PORT_ID);
    if (publicPort === undefined) {
        return err(new Error(`No datastore public port found for port ID '${DATASTORE_PORT_ID}'`))
    }

    const publicIp = serviceContext.getMaybePublicIPAddress();
    const publicPortNum = publicPort.number;
    const { client, clientCloseFunction } = createDatastoreClient(publicIp, publicPortNum);

    const waitForHealthyResult = await waitForHealthy(
        client,
        DATASTORE_WAIT_FOR_STARTUP_MAX_POLLS,
        DATASTORE_WAIT_FOR_STARTUP_DELAY_MILLISECONDS
    );

    if (waitForHealthyResult.isErr()) {
        log.error("An error occurred waiting for the datastore service to become available")
        return err(waitForHealthyResult.error);
    }

    return ok({ serviceContext, client, clientCloseFunction });
};

export function createDatastoreClient(ipAddr: string, portNum: number): { client: datastoreApi.DatastoreServiceClientNode; clientCloseFunction: () => void } {
    const url = `${ipAddr}:${portNum}`;
    const client = new datastoreApi.DatastoreServiceClientNode(url, grpc.credentials.createInsecure());
    const clientCloseFunction = () => client.close();

    return { client, clientCloseFunction }
};

export async function addAPIService( serviceId: ServiceID, enclaveContext: EnclaveContext, datastoreIPInsideNetwork: string):
    Promise<Result<{
        serviceContext: ServiceContext;
        client: serverApi.ExampleAPIServerServiceClientNode;
        clientCloseFunction: () => void;
    }, Error>> {
  
    const addAPIServiceToPartitionResult = await addAPIServiceToPartition(
        serviceId,
        enclaveContext,
        datastoreIPInsideNetwork,
        DEFAULT_PARTITION_ID
    );
    if(addAPIServiceToPartitionResult.isErr()) return err(addAPIServiceToPartitionResult.error)
    const { serviceContext, client, clientCloseFunction} = addAPIServiceToPartitionResult.value
  
    return ok({ serviceContext, client, clientCloseFunction})
}

export async function addAPIServiceToPartition(
    serviceId: ServiceID,
    enclaveContext: EnclaveContext,
    datastorePrivateIp: string,
    partitionId: PartitionID
): Promise<Result<{
    serviceContext: ServiceContext;
    client: serverApi.ExampleAPIServerServiceClientNode;
    clientCloseFunction: () => void;
}, Error>> {
    const createDatastoreConfigResult = await createApiConfigFile(datastorePrivateIp)
    if (createDatastoreConfigResult.isErr()) {
        return err(createDatastoreConfigResult.error)
    }
    const configFilepath = createDatastoreConfigResult.value;

    const uploadConfigResult = await enclaveContext.uploadFiles(configFilepath)
    if (uploadConfigResult.isErr()) {
        return err(uploadConfigResult.error)
    }
    const datastoreConfigArtifactUuid = uploadConfigResult.value

    const containerConfigSupplier = getApiServiceContainerConfigSupplier(datastoreConfigArtifactUuid)

    const addServiceToPartitionResult = await enclaveContext.addServiceToPartition(serviceId, partitionId, containerConfigSupplier)
    if(addServiceToPartitionResult.isErr()) return err(addServiceToPartitionResult.error)

    const serviceContext = addServiceToPartitionResult.value;

    const publicPort: PortSpec | undefined = serviceContext.getPublicPorts().get(API_PORT_ID);
    if (publicPort === undefined) {
        return err(new Error(`No API service public port found for port ID '${API_PORT_ID}'`));
    }
  
    const url = `${serviceContext.getMaybePublicIPAddress()}:${publicPort.number}`;
    const client = new serverApi.ExampleAPIServerServiceClientNode(url, grpc.credentials.createInsecure());
    const clientCloseFunction = () => client.close();

    const waitForHealthyResult = await waitForHealthy(client, API_WAIT_FOR_STARTUP_MAX_POLLS, API_WAIT_FOR_STARTUP_DELAY_MILLISECONDS)

    if(waitForHealthyResult.isErr()) {
        log.error("An error occurred waiting for the API service to become available")
        return err(waitForHealthyResult.error)
    }
  
    return ok({ serviceContext, client, clientCloseFunction })
};

export async function waitForHealthy(
  client: datastoreApi.DatastoreServiceClientNode | serverApi.ExampleAPIServerServiceClientNode,
  retries: number,
  retriesDelayMilliseconds: number
): Promise<Result<null, Error>> {

    const emptyArgs: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty();

    const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

    const checkClientAvailabilityResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise(async (resolve, _unusedReject) => {
        client.isAvailable(emptyArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
        if (error === null) {
            if (!response) {
                resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
            } else {
                resolve(ok(response));
            }
        } else {
            resolve(err(error));
        }
        });
    });

    for (let i = 0; i < retries; i++) {
        const checkClientAvailabilityResult = await checkClientAvailabilityResultPromise;
        if (checkClientAvailabilityResult.isOk()) {
            return ok(null)
        } else {
            log.debug(checkClientAvailabilityResult.error)
            await sleep(retriesDelayMilliseconds);
        }
    }

    return err(new Error(`The service didn't return a success code, even after ${retries} retries with ${retriesDelayMilliseconds} milliseconds in between retries`));

}

export async function startFileServer(filesArtifactUuid: string, pathToCheckOnFileServer: string, enclaveCtx: EnclaveContext) : Promise<Result<StartFileServerResponse, Error>> {
    const filesArtifactsMountPoints = new Map<FilesArtifactUUID, string>()
    filesArtifactsMountPoints.set(filesArtifactUuid, USER_SERVICE_MOUNT_POINT_FOR_TEST_FILES_ARTIFACT)

    const fileServerContainerConfigSupplier = getFileServerContainerConfigSupplier(filesArtifactsMountPoints)
    const addServiceResult = await enclaveCtx.addService(FILE_SERVER_SERVICE_ID, fileServerContainerConfigSupplier)
    if(addServiceResult.isErr()){ throw addServiceResult.error }

    const serviceContext = addServiceResult.value
    const publicPort = serviceContext.getPublicPorts().get(FILE_SERVER_PORT_ID)
    if(publicPort === undefined){
        throw new Error(`Expected to find public port for ID "${FILE_SERVER_PORT_ID}", but none was found`)
    }

    const fileServerPublicIp = serviceContext.getMaybePublicIPAddress();
    const fileServerPublicPortNum = publicPort.number

    const waitForHttpGetEndpointAvailabilityResult = await enclaveCtx.waitForHttpGetEndpointAvailability(
        FILE_SERVER_SERVICE_ID,
        FILE_SERVER_PRIVATE_PORT_NUM,
        pathToCheckOnFileServer,
        WAIT_INITIAL_DELAY_MILLISECONDS,
        WAIT_FOR_STARTUP_MAX_RETRIES,
        WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS,
        WAIT_FOR_AVAILABILITY_BODY_TEXT
    );

    if(waitForHttpGetEndpointAvailabilityResult.isErr()){
        log.error("An error occurred waiting for the file server service to become available")
        throw waitForHttpGetEndpointAvailabilityResult.error
    }
    log.info(`Added file server service with public IP "${fileServerPublicIp}" and port "${fileServerPublicPortNum}"`)

    return ok(new StartFileServerResponse(fileServerPublicIp, fileServerPublicPortNum))
}


//Test file contents against the ARCHIVE_TEST_CONTENT string.
export async function checkFileContents(ipAddress: string, portNum: number, filename: string, expectedContents: string): Promise<Result<null, Error>> {
    let fileContentResults = await getFileContents(ipAddress, portNum, filename)
    if(fileContentResults.isErr()) { return err(fileContentResults.error)}

    let dataAsString = String(fileContentResults.value)
    if (dataAsString !== expectedContents){
        return err(new Error(`The contents of '${filename}' do not match the test contents of ${expectedContents}.\n`))
    }
    return ok(null)
}

// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================

function getDatastoreContainerConfigSupplier(): ( ipAddr: string) => Result<ContainerConfig, Error> {

    const containerConfigSupplier = ( ipAddr: string): Result<ContainerConfig, Error> => {
        const usedPorts = new Map<string, PortSpec>();
        usedPorts.set(DATASTORE_PORT_ID, DATASTORE_PORT_SPEC);

        const containerConfig = new ContainerConfigBuilder(DATASTORE_IMAGE).withUsedPorts(usedPorts).build();

        return ok(containerConfig);
    };

    return containerConfigSupplier;
}

function getApiServiceContainerConfigSupplier(
    apiConfigArtifactUuid: FilesArtifactUUID,
): (ipAddr:string) => Result<ContainerConfig, Error> {

    const containerConfigSupplier = (ipAddr: string): Result<ContainerConfig, Error> => {

        const usedPorts = new Map<string, PortSpec>();
        usedPorts.set(API_PORT_ID, API_PORT_SPEC);
        const startCmd: string[] = [
            "./example-api-server.bin",
            "--config",
            path.join(CONFIG_MOUNTPATH_ON_API_CONTAINER, CONFIG_FILENAME),
        ]

        const filesArtifactMountpoints = new Map<FilesArtifactUUID, string>()
        filesArtifactMountpoints.set(apiConfigArtifactUuid, CONFIG_MOUNTPATH_ON_API_CONTAINER)

        const containerConfig = new ContainerConfigBuilder(API_SERVICE_IMAGE)
            .withUsedPorts(usedPorts)
            .withFiles(filesArtifactMountpoints)
            .withCmdOverride(startCmd)
            .build()

        return ok(containerConfig)
    }

    return containerConfigSupplier;

}

async function createApiConfigFile(datastoreIP: string): Promise<Result<string, Error>> {
    const mkdirResult = await fs.promises.mkdtemp(
        `${os.tmpdir()}${path.sep}`,
    ).then(
        (result) => ok(result),
    ).catch(
        (mkdirErr) => err(mkdirErr),
    )
    if (mkdirResult.isErr()) {
        return err(mkdirResult.error);
    }
    const tempDirpath = mkdirResult.value;
    const tempFilepath = path.join(tempDirpath, CONFIG_FILENAME)

    const config = {
        datastoreIp: datastoreIP,
        datastorePort: datastoreApi.LISTEN_PORT,
    };

    const configJSONStringified = JSON.stringify(config);

    log.debug(`API config JSON: ${configJSONStringified}`)

    try {
        fs.writeFileSync(tempFilepath, configJSONStringified);
    } catch (error) {
        log.error("An error occurred writing the serialized config JSON to file")
        if (error instanceof Error) {
            return err(error)
        } else {
            return err(new Error("Encountered error while writing the file, but the error wasn't of type Error"))
        }
    }

    return ok(tempFilepath);

}

function getFileServerContainerConfigSupplier(filesArtifactMountPoints: Map<FilesArtifactUUID, string>): (ipAddr: string) => Result<ContainerConfig, Error> {
    const containerConfigSupplier = (ipAddr:string): Result<ContainerConfig, Error> => {
        const usedPorts = new Map<string, PortSpec>()
        usedPorts.set(FILE_SERVER_PORT_ID, FILE_SERVER_PORT_SPEC)

        const containerConfig = new ContainerConfigBuilder(FILE_SERVER_SERVICE_IMAGE)
            .withUsedPorts(usedPorts)
            .withFiles(filesArtifactMountPoints)
            .build()

        return ok(containerConfig)
    }
    return containerConfigSupplier
}

async function getFileContents(ipAddress: string, portNum: number, relativeFilepath: string): Promise<Result<any, Error>> {
    let response;
    try {
        response = await axios(`http://${ipAddress}:${portNum}/${relativeFilepath}`)
    } catch(error){
        log.error(`An error occurred getting the contents of file "${relativeFilepath}"`)
        if(error instanceof Error){
            return err(error)
        }else{
            return err(new Error("An error occurred getting the contents of file, but the error wasn't of type Error"))
        }
    }

    const bodyStr = String(response.data)
    return ok(bodyStr)
}