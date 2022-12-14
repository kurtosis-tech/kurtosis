import {
    ServiceID,
    EnclaveContext,
    ContainerConfig,
    ContainerConfigBuilder,
    ServiceContext,
    PartitionID,
    PortSpec,
    PortProtocol,
    FilesArtifactUUID, ServiceGUID,
} from "kurtosis-sdk";
import * as datastoreApi from "example-datastore-server-api-lib";
import * as serverApi from "example-api-server-api-lib";
import { err, ok, Result } from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "@grpc/grpc-js";
import log from "loglevel";
import * as fs from 'fs';
import * as path from "path";
import * as os from "os";
import axios from "axios";
import {GetArgs, GetResponse, UpsertArgs} from "example-datastore-server-api-lib";

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
const FILE_SERVER_PRIVATE_PORT_NUM      = 80

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS = 500
const WAIT_FOR_STARTUP_MAX_RETRIES        = 15
const WAIT_INITIAL_DELAY_MILLISECONDS     = 0
const WAIT_FOR_AVAILABILITY_BODY_TEXT     = ""

const USER_SERVICE_MOUNT_POINT_FOR_TEST_FILES_ARTIFACT  = "/static"

const FILE_SERVER_PORT_SPEC = new PortSpec( FILE_SERVER_PRIVATE_PORT_NUM, PortProtocol.TCP )

// for validating data store is healthy
/*
NOTE: on 2022-05-16 the Go version failed with the following error so we bumped the num polls to 20.

time="2022-05-16T23:58:21Z" level=info msg="Sanity-checking that all 4 datastore services added via the module work as expected..."
--- FAIL: TestModule (21.46s)
    module_test.go:81:
            Error Trace:	module_test.go:81
            Error:      	Received unexpected error:
                            The service didn't return a success code, even after 15 retries with 1000 milliseconds in between retries
                             --- at /home/circleci/project/internal_testsuites/golang/test_helpers/test_helpers.go:179 (WaitForHealthy) ---
                            Caused by: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:49188: connect: connection refused"
            Test:       	TestModule
            Messages:   	An error occurred waiting for the datastore service to become available

NOTE: On 2022-05-21 the Go version failed again at 20s. I opened the enclave logs and it's weird because nothing is failing and
the datastore service is showing itself as up *before* we even start the check-if-available wait. We're in crunch mode
so I'm going to bump this up to 30s, but I suspect there's some sort of nondeterministic underlying failure happening.
 */
const WAIT_FOR_STARTUP_MAX_POLLS = 30
const MILLIS_BETWEEN_AVAILABILITY_RETRIES = 1000
const TEST_DATASTORE_KEY = "my-key"
const TEST_DATASTORE_VALUE = "test-value"


class StartFileServerResponse  {
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
    
    const containerConfig = getDatastoreContainerConfig();

    const addServiceResult = await enclaveContext.addService(serviceId, containerConfig);
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

    const containerConfig = getApiServiceContainerConfig(datastoreConfigArtifactUuid)

    const addServiceToPartitionResult = await enclaveContext.addServiceToPartition(serviceId, partitionId, containerConfig)
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

    const clientAvailability: () => Promise<Result<google_protobuf_empty_pb.Empty, Error>> = () => new Promise(async (resolve, _unusedReject) => {
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
        const availabilityResult = await clientAvailability();
        if (availabilityResult.isOk()) {
            return ok(null)
        } else {
            log.debug(availabilityResult.error)
            await sleep(retriesDelayMilliseconds);
        }
    }

    return err(new Error(`The service didn't return a success code, even after ${retries} retries with ${retriesDelayMilliseconds} milliseconds in between retries`));

}

export async function startFileServer(fileServerServiceId: ServiceID, filesArtifactUuid: string, pathToCheckOnFileServer: string, enclaveCtx: EnclaveContext) : Promise<Result<StartFileServerResponse, Error>> {
    const filesArtifactsMountPoints = new Map<FilesArtifactUUID, string>()
    filesArtifactsMountPoints.set(filesArtifactUuid, USER_SERVICE_MOUNT_POINT_FOR_TEST_FILES_ARTIFACT)

    const fileServerContainerConfig = getFileServerContainerConfig(filesArtifactsMountPoints)
    const addServiceResult = await enclaveCtx.addService(fileServerServiceId, fileServerContainerConfig)
    if(addServiceResult.isErr()){ throw addServiceResult.error }

    const serviceContext = addServiceResult.value
    const publicPort = serviceContext.getPublicPorts().get(FILE_SERVER_PORT_ID)
    if(publicPort === undefined){
        throw new Error(`Expected to find public port for ID "${FILE_SERVER_PORT_ID}", but none was found`)
    }

    const fileServerPublicIp = serviceContext.getMaybePublicIPAddress();
    const fileServerPublicPortNum = publicPort.number

    const waitForHttpGetEndpointAvailabilityResult = await enclaveCtx.waitForHttpGetEndpointAvailability(
        fileServerServiceId,
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


// Compare the file contents on the server against expectedContent and see if they match.
export async function checkFileContents(ipAddress: string, portNum: number, filename: string, expectedContents: string): Promise<Result<null, Error>> {
    let fileContentResults = await getFileContents(ipAddress, portNum, filename)
    if(fileContentResults.isErr()) { return err(fileContentResults.error)}

    let dataAsString = String(fileContentResults.value)
    if (dataAsString !== expectedContents){
        return err(new Error(`The contents of '${filename}' do not match the expected content '${expectedContents}'.`))
    }
    return ok(null)
}

// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================

function getDatastoreContainerConfig(): ContainerConfig {

    const usedPorts = new Map<string, PortSpec>();
    usedPorts.set(DATASTORE_PORT_ID, DATASTORE_PORT_SPEC);

    const containerConfig = new ContainerConfigBuilder(DATASTORE_IMAGE).withUsedPorts(usedPorts).build();

    return containerConfig;
}

function getApiServiceContainerConfig(
    apiConfigArtifactUuid: FilesArtifactUUID,
): ContainerConfig {
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

    return containerConfig
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

function getFileServerContainerConfig(filesArtifactMountPoints: Map<FilesArtifactUUID, string>): ContainerConfig {
    const usedPorts = new Map<string, PortSpec>()
    usedPorts.set(FILE_SERVER_PORT_ID, FILE_SERVER_PORT_SPEC)

    const containerConfig = new ContainerConfigBuilder(FILE_SERVER_SERVICE_IMAGE)
        .withUsedPorts(usedPorts)
        .withFiles(filesArtifactMountPoints)
        .build()

   return containerConfig
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

export async function validateDataStoreServiceIsHealthy(enclaveContext : EnclaveContext, serviceId: string, portId: string): Promise<Result<null, Error>> {
    const getServiceContextResult = await enclaveContext.getServiceContext(serviceId)
    if (getServiceContextResult.isErr()) {
        log.error(`An error occurred getting the service context for service '${serviceId}'; this indicates that the module says it created a service that it actually didn't`)
        throw getServiceContextResult.error
    }
    const serviceContext = getServiceContextResult.value
    const ipAddr = serviceContext.getMaybePublicIPAddress()
    const publicPort: undefined | PortSpec = serviceContext.getPublicPorts().get(portId)

    if (publicPort === undefined) {
        throw new Error(`Expected to find public port '${portId}' on datastore service '${serviceId}', but none was found`)
    }

    const {
        client: datastoreClient,
        clientCloseFunction: datastoreClientCloseFunction
    } = createDatastoreClient(ipAddr, publicPort.number);

    try {
        const waitForHealthyResult = await waitForHealthy(datastoreClient, WAIT_FOR_STARTUP_MAX_POLLS, MILLIS_BETWEEN_AVAILABILITY_RETRIES);
        if (waitForHealthyResult.isErr()) {
            log.error(`An error occurred waiting for the datastore service '${serviceId}' to become available`);
            throw waitForHealthyResult.error
        }

        const upsertArgs = new UpsertArgs();
        upsertArgs.setKey(TEST_DATASTORE_KEY)
        upsertArgs.setValue(TEST_DATASTORE_VALUE)

        const upsertResponseResultPromise:Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            datastoreClient.upsert(upsertArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                    } else {
                        resolve(ok(response));
                    }
                } else {
                    resolve(err(error));
                }
            })
        })

        const upsertResponseResult = await upsertResponseResultPromise;
        if(upsertResponseResult.isErr()){
            log.error(`An error occurred adding the test key to datastore service "${serviceId}"`)
            throw upsertResponseResult.error
        }

        const getArgs = new GetArgs();
        getArgs.setKey(TEST_DATASTORE_KEY)

        const getResponseResultPromise:Promise<Result<GetResponse, Error>> = new Promise((resolve, _unusedReject) => {
            datastoreClient.get(getArgs, (error: grpc.ServiceError | null, response?: GetResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                    } else {
                        resolve(ok(response));
                    }
                } else {
                    resolve(err(error));
                }
            })
        })

        const getResponseResult = await getResponseResultPromise;
        if(getResponseResult.isErr()){
            log.error(`An error occurred getting the test key to datastore service "${serviceId}"`)
            throw getResponseResult.error
        }

        const getResponse = getResponseResult.value;
        const actualValue = getResponse.getValue();
        if(actualValue !== TEST_DATASTORE_VALUE){
            log.error(`Datastore service "${serviceId}" is storing value "${actualValue}" for the test key, which doesn't match the expected value ""${TEST_DATASTORE_VALUE}`)
        }

    } finally {
        datastoreClientCloseFunction()
    }

    return ok(null)
}

export function areEqualServiceGuidsSet(firstSet: Set<ServiceGUID>, secondSet: Set<ServiceGUID>): boolean {
    const haveEqualSize: boolean = firstSet.size === secondSet.size;
    const haveEqualContent: boolean = [...firstSet].every((x) => secondSet.has(x));

    const areEqual: boolean = haveEqualSize && haveEqualContent;

    return areEqual
}

export function delay(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
