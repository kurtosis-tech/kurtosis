import {
    ContainerConfig,
    ContainerConfigBuilder,
    EnclaveContext,
    EnclaveUUID,
    FilesArtifactUUID,
    KurtosisContext,
    LogLineFilter,
    PartitionID,
    PortSpec,
    ServiceContext,
    ServiceUUID,
    ServiceName,
    ServiceLog,
    TransportProtocol
} from "kurtosis-sdk";
import * as datastoreApi from "example-datastore-server-api-lib";
import * as serverApi from "example-api-server-api-lib";
import {err, ok, Result} from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "@grpc/grpc-js";
import log from "loglevel";
import * as fs from 'fs';
import * as path from "path";
import * as os from "os";
import axios from "axios";
import {GetArgs, GetResponse, UpsertArgs} from "example-datastore-server-api-lib";
import {StarlarkRunResult} from "kurtosis-sdk/build/core/lib/enclaves/starlark_run_blocking";
import {Readable} from "stream";
import {receiveExpectedLogLinesFromServiceLogsReadable, ReceivedStreamContent} from "./received_stream_content";

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
    TransportProtocol.TCP,
)
const API_PORT_SPEC = new PortSpec(
    serverApi.LISTEN_PORT,
    TransportProtocol.TCP,
)

const FILE_SERVER_SERVICE_IMAGE = "flashspys/nginx-static"
const FILE_SERVER_PORT_ID = "http"
const FILE_SERVER_PRIVATE_PORT_NUM = 80

const WAIT_FOR_FILE_SERVER_TIMEOUT_MILLISECONDS = 45000
const WAIT_FOR_FILE_SERVER_INTERVAL_MILLISECONDS = 100

const USER_SERVICE_MOUNT_POINT_FOR_TEST_FILES_ARTIFACT = "/static"

const FILE_SERVER_PORT_SPEC = new PortSpec(FILE_SERVER_PRIVATE_PORT_NUM, TransportProtocol.TCP)

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

const WAIT_FOR_GET_AVAILABILITY_STARLARK_SCRIPT = `
def run(plan, args):
	get_recipe = GetHttpRequestRecipe(
		port_id = args["port_id"],
		endpoint = args["endpoint"],
	)
	plan.wait(recipe=get_recipe, field="code", assertion="==", target_value=200, interval=args["interval"], timeout=args["timeout"], service_name=args["service_name"])
`

const DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started";

const ARTIFACT_NAME_PREFIX = "artifact-uploaded-via-helper-"

class StartFileServerResponse {
    fileServerPublicIp: string
    fileServerPublicPortNum: number

    constructor(fileServerPublicIp: string, fileServerPublicPortNum: number) {
        this.fileServerPublicIp = fileServerPublicIp
        this.fileServerPublicPortNum = fileServerPublicPortNum
    }
}

export async function addDatastoreService(serviceName: ServiceName, enclaveContext: EnclaveContext):
    Promise<Result<{
        serviceContext: ServiceContext;
        client: datastoreApi.DatastoreServiceClientNode;
        clientCloseFunction: () => void;
    }, Error>> {

    const containerConfig = getDatastoreContainerConfig();

    const addServiceResult = await enclaveContext.addService(serviceName, containerConfig);
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
    const {client, clientCloseFunction} = createDatastoreClient(publicIp, publicPortNum);

    const waitForHealthyResult = await waitForHealthy(
        client,
        DATASTORE_WAIT_FOR_STARTUP_MAX_POLLS,
        DATASTORE_WAIT_FOR_STARTUP_DELAY_MILLISECONDS
    );

    if (waitForHealthyResult.isErr()) {
        log.error("An error occurred waiting for the datastore service to become available")
        return err(waitForHealthyResult.error);
    }

    return ok({serviceContext, client, clientCloseFunction});
}

export function createDatastoreClient(ipAddr: string, portNum: number): { client: datastoreApi.DatastoreServiceClientNode; clientCloseFunction: () => void } {
    const url = `${ipAddr}:${portNum}`;
    const client = new datastoreApi.DatastoreServiceClientNode(url, grpc.credentials.createInsecure());
    const clientCloseFunction = () => client.close();

    return {client, clientCloseFunction}
}

export async function addAPIService(serviceName: ServiceName, enclaveContext: EnclaveContext, datastoreIPInsideNetwork: string):
    Promise<Result<{
        serviceContext: ServiceContext;
        client: serverApi.ExampleAPIServerServiceClientNode;
        clientCloseFunction: () => void;
    }, Error>> {

    const addAPIServiceToPartitionResult = await addAPIServiceToPartition(
        serviceName,
        enclaveContext,
        datastoreIPInsideNetwork,
        DEFAULT_PARTITION_ID
    );
    if (addAPIServiceToPartitionResult.isErr()) return err(addAPIServiceToPartitionResult.error)
    const {serviceContext, client, clientCloseFunction} = addAPIServiceToPartitionResult.value

    return ok({serviceContext, client, clientCloseFunction})
}

export async function addAPIServiceToPartition(
    serviceName: ServiceName,
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
    const artifactName = `${ARTIFACT_NAME_PREFIX}${Date.now()}`
    const uploadConfigResult = await enclaveContext.uploadFiles(configFilepath, artifactName)
    if (uploadConfigResult.isErr()) {
        return err(uploadConfigResult.error)
    }

    const containerConfig = getApiServiceContainerConfig(artifactName)

    const addServiceToPartitionResult = await enclaveContext.addServiceToPartition(serviceName, partitionId, containerConfig)
    if (addServiceToPartitionResult.isErr()) return err(addServiceToPartitionResult.error)

    const serviceContext = addServiceToPartitionResult.value;

    const publicPort: PortSpec | undefined = serviceContext.getPublicPorts().get(API_PORT_ID);
    if (publicPort === undefined) {
        return err(new Error(`No API service public port found for port ID '${API_PORT_ID}'`));
    }

    const url = `${serviceContext.getMaybePublicIPAddress()}:${publicPort.number}`;
    const client = new serverApi.ExampleAPIServerServiceClientNode(url, grpc.credentials.createInsecure());
    const clientCloseFunction = () => client.close();

    const waitForHealthyResult = await waitForHealthy(client, API_WAIT_FOR_STARTUP_MAX_POLLS, API_WAIT_FOR_STARTUP_DELAY_MILLISECONDS)

    if (waitForHealthyResult.isErr()) {
        log.error("An error occurred waiting for the API service to become available")
        return err(waitForHealthyResult.error)
    }

    return ok({serviceContext, client, clientCloseFunction})
}

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

export async function waitForGetAvailabilityStarlarkScript(enclaveContext: EnclaveContext, serviceName: string, portId: string, endpoint: string, interval: number, timeout: number) : Promise<Result<StarlarkRunResult, Error>> {
    return enclaveContext.runStarlarkScriptBlocking(WAIT_FOR_GET_AVAILABILITY_STARLARK_SCRIPT, `{ "service_name": "${serviceName}", "port_id": "${portId}", "endpoint": "/${endpoint}", "interval": "${interval}ms", "timeout": "${timeout}ms"}`, false)
}

export async function startFileServer(fileServerServiceName: ServiceName, filesArtifactUuid: string, pathToCheckOnFileServer: string, enclaveCtx: EnclaveContext): Promise<Result<StartFileServerResponse, Error>> {
    const filesArtifactsMountPoints = new Map<string, FilesArtifactUUID>()
    filesArtifactsMountPoints.set(USER_SERVICE_MOUNT_POINT_FOR_TEST_FILES_ARTIFACT, filesArtifactUuid)

    const fileServerContainerConfig = getFileServerContainerConfig(filesArtifactsMountPoints)
    const addServiceResult = await enclaveCtx.addService(fileServerServiceName, fileServerContainerConfig)
    if (addServiceResult.isErr()) {
        throw addServiceResult.error
    }

    const serviceContext = addServiceResult.value
    const publicPort = serviceContext.getPublicPorts().get(FILE_SERVER_PORT_ID)
    if (publicPort === undefined) {
        throw new Error(`Expected to find public port for ID "${FILE_SERVER_PORT_ID}", but none was found`)
    }

    const fileServerPublicIp = serviceContext.getMaybePublicIPAddress();
    const fileServerPublicPortNum = publicPort.number

    const waitForHttpGetEndpointAvailabilityResult = await waitForGetAvailabilityStarlarkScript(enclaveCtx, fileServerServiceName, FILE_SERVER_PORT_ID, pathToCheckOnFileServer, WAIT_FOR_FILE_SERVER_INTERVAL_MILLISECONDS, WAIT_FOR_FILE_SERVER_TIMEOUT_MILLISECONDS)

    if (waitForHttpGetEndpointAvailabilityResult.isErr()) {
        log.error("An unexpected error has occurred getting endpoint availability using Starlark")
        throw waitForHttpGetEndpointAvailabilityResult.error
    }
    if(waitForHttpGetEndpointAvailabilityResult.value.interpretationError !== undefined){
        log.error("An error has occurred getting endpoint availability during Starlark due to interpretation error")
        throw waitForHttpGetEndpointAvailabilityResult.value.interpretationError
    }
    if(waitForHttpGetEndpointAvailabilityResult.value.validationErrors.length > 0){
        log.error("An error has occurred getting endpoint availability using Starlark due to validation error")
        throw waitForHttpGetEndpointAvailabilityResult.value.validationErrors
    }
    if(waitForHttpGetEndpointAvailabilityResult.value.executionError !== undefined){
        log.error("An error has occurred getting endpoint availability during Starlark due to execution error")
        throw waitForHttpGetEndpointAvailabilityResult.value.executionError
    }
    log.info(`Added file server service with public IP "${fileServerPublicIp}" and port "${fileServerPublicPortNum}"`)

    return ok(new StartFileServerResponse(fileServerPublicIp, fileServerPublicPortNum))
}


// Compare the file contents on the server against expectedContent and see if they match.
export async function checkFileContents(ipAddress: string, portNum: number, filename: string, expectedContents: string): Promise<Result<null, Error>> {
    let fileContentResults = await getFileContents(ipAddress, portNum, filename)
    if (fileContentResults.isErr()) {
        return err(fileContentResults.error)
    }

    let dataAsString = String(fileContentResults.value)
    if (dataAsString !== expectedContents) {
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
    artifactName: string,
): ContainerConfig {
    const usedPorts = new Map<string, PortSpec>();
    usedPorts.set(API_PORT_ID, API_PORT_SPEC);
    const startCmd: string[] = [
        "./example-api-server.bin",
        "--config",
        path.join(CONFIG_MOUNTPATH_ON_API_CONTAINER, CONFIG_FILENAME),
    ]

    const filesArtifactMountpoints = new Map<string, FilesArtifactUUID>()
    filesArtifactMountpoints.set(CONFIG_MOUNTPATH_ON_API_CONTAINER, artifactName)

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

function getFileServerContainerConfig(filesArtifactMountPoints: Map<string, FilesArtifactUUID>): ContainerConfig {
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
    } catch (error) {
        log.error(`An error occurred getting the contents of file "${relativeFilepath}"`)
        if (error instanceof Error) {
            return err(error)
        } else {
            return err(new Error("An error occurred getting the contents of file, but the error wasn't of type Error"))
        }
    }

    const bodyStr = String(response.data)
    return ok(bodyStr)
}

export async function validateDataStoreServiceIsHealthy(enclaveContext: EnclaveContext, serviceName: string, portId: string): Promise<Result<null, Error>> {
    const getServiceContextResult = await enclaveContext.getServiceContext(serviceName)
    if (getServiceContextResult.isErr()) {
        log.error(`An error occurred getting the service context for service '${serviceName}'; this indicates that the module says it created a service that it actually didn't`)
        throw getServiceContextResult.error
    }
    const serviceContext = getServiceContextResult.value
    const ipAddr = serviceContext.getMaybePublicIPAddress()
    const publicPort: undefined | PortSpec = serviceContext.getPublicPorts().get(portId)

    if (publicPort === undefined) {
        throw new Error(`Expected to find public port '${portId}' on datastore service '${serviceName}', but none was found`)
    }

    const {
        client: datastoreClient,
        clientCloseFunction: datastoreClientCloseFunction
    } = createDatastoreClient(ipAddr, publicPort.number);

    try {
        const waitForHealthyResult = await waitForHealthy(datastoreClient, WAIT_FOR_STARTUP_MAX_POLLS, MILLIS_BETWEEN_AVAILABILITY_RETRIES);
        if (waitForHealthyResult.isErr()) {
            log.error(`An error occurred waiting for the datastore service '${serviceName}' to become available`);
            throw waitForHealthyResult.error
        }

        const upsertArgs = new UpsertArgs();
        upsertArgs.setKey(TEST_DATASTORE_KEY)
        upsertArgs.setValue(TEST_DATASTORE_VALUE)

        const upsertResponseResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
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
        if (upsertResponseResult.isErr()) {
            log.error(`An error occurred adding the test key to datastore service "${serviceName}"`)
            throw upsertResponseResult.error
        }

        const getArgs = new GetArgs();
        getArgs.setKey(TEST_DATASTORE_KEY)

        const getResponseResultPromise: Promise<Result<GetResponse, Error>> = new Promise((resolve, _unusedReject) => {
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
        if (getResponseResult.isErr()) {
            log.error(`An error occurred getting the test key to datastore service "${serviceName}"`)
            throw getResponseResult.error
        }

        const getResponse = getResponseResult.value;
        const actualValue = getResponse.getValue();
        if (actualValue !== TEST_DATASTORE_VALUE) {
            log.error(`Datastore service "${serviceName}" is storing value "${actualValue}" for the test key, which doesn't match the expected value ""${TEST_DATASTORE_VALUE}`)
        }

    } finally {
        datastoreClientCloseFunction()
    }

    return ok(null)
}

export function areEqualServiceUuidsSet(firstSet: Set<ServiceUUID>, secondSet: Set<ServiceUUID>): boolean {
    const haveEqualSize: boolean = firstSet.size === secondSet.size;
    const haveEqualContent: boolean = [...firstSet].every((x) => secondSet.has(x));

    const areEqual: boolean = haveEqualSize && haveEqualContent;

    return areEqual
}

export function delay(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export async function addServicesWithLogLines(
    enclaveContext: EnclaveContext,
    logLinesByServiceID: Map<ServiceName, ServiceLog[]>,
): Promise<Result<Map<ServiceName, ServiceContext>, Error>> {

    const servicesAdded: Map<ServiceName, ServiceContext> = new Map<ServiceName, ServiceContext>();

    for (let [serviceName, logLines] of logLinesByServiceID) {
        const containerConf: ContainerConfig = getServiceWithLogLinesConfig(logLines);
        const addServiceResult = await enclaveContext.addService(serviceName, containerConf);

        if (addServiceResult.isErr()) {
            return err(new Error(`An error occurred adding service '${serviceName}'`));
        }

        const serviceContext: ServiceContext = addServiceResult.value;

        servicesAdded.set(serviceName, serviceContext)
    }

    return ok(servicesAdded)
}

export async function getLogsResponseAndEvaluateResponse(
    kurtosisCtx: KurtosisContext,
    enclaveUuid: EnclaveUUID,
    serviceUuids: Set<ServiceUUID>,
    expectedLogLinesByService: Map<ServiceUUID, ServiceLog[]>,
    expectedNonExistenceServiceUuids: Set<ServiceUUID>,
    shouldFollowLogs: boolean,
    logLineFilter: LogLineFilter | undefined,
): Promise<Result<null, Error>> {

    let receivedLogLinesByService: Map<ServiceUUID, Array<ServiceLog>> = new Map<ServiceUUID, Array<ServiceLog>>;
    let receivedNotFoundServiceUuids: Set<ServiceUUID> = new Set<ServiceUUID>();

    const streamUserServiceLogsPromise = await kurtosisCtx.getServiceLogs(enclaveUuid, serviceUuids, shouldFollowLogs, logLineFilter);

    if (streamUserServiceLogsPromise.isErr()) {
        return err(streamUserServiceLogsPromise.error);
    }

    const serviceLogsReadable: Readable = streamUserServiceLogsPromise.value;

    const receivedStreamContentPromise: Promise<ReceivedStreamContent> = receiveExpectedLogLinesFromServiceLogsReadable(
        serviceLogsReadable,
        expectedLogLinesByService,
    )

    const receivedStreamContent: ReceivedStreamContent = await receivedStreamContentPromise;
    receivedLogLinesByService = receivedStreamContent.receivedLogLinesByService;
    receivedNotFoundServiceUuids = receivedStreamContent.receivedNotFoundServiceUuids;


    receivedLogLinesByService.forEach((receivedLogLines, serviceUuid) => {
        const expectedLogLines = expectedLogLinesByService.get(serviceUuid);
        if (expectedLogLines === undefined) {
            return err(new Error(`No expected log lines for service with UUID '${serviceUuid}' was found in the expectedLogLinesByService map'${expectedLogLinesByService}'`))
        }
        if (expectedLogLines.length === receivedLogLines.length) {
            receivedLogLines.forEach((logLine: ServiceLog, logLineIndex: number) => {
                if (expectedLogLines[logLineIndex].getContent() !== logLine.getContent()) {
                    return err(new Error(`Expected to match the number ${logLineIndex} log line with this value ${expectedLogLines[logLineIndex]} but this one was received instead ${logLine.getContent()}`));
                }
            })
        } else {
            throw new Error(`Expected to receive ${expectedLogLines.length} of log lines but ${receivedLogLines.length} log lines were received instead`);
        }
    })

    if (!areEqualServiceUuidsSet(expectedNonExistenceServiceUuids, receivedNotFoundServiceUuids)) {
        throw new Error(`Expected to receive a not found service UUIDs set equal to '${[...expectedNonExistenceServiceUuids.entries()]}' but a different set '${[...receivedNotFoundServiceUuids.entries()]}' was received instead`);
    }

    return ok(null)
}

function getServiceWithLogLinesConfig(logLines: ServiceLog[]): ContainerConfig {

    const entrypointArgs = ["/bin/sh", "-c"];

    const logLinesWithQuotes: Array<string> = new Array<string>();

    for (let logLine of logLines) {
        const logLineWithQuotes: string = `"${logLine.getContent()}"`;
        logLinesWithQuotes.push(logLineWithQuotes);
    }

    const logLineSeparator: string = " ";
    const logLinesStr: string = logLinesWithQuotes.join(logLineSeparator);
    const echoLogLinesLoopCmdStr: string = `for i in ${logLinesStr}; do echo "$i"; done;`

    const cmdArgs = [echoLogLinesLoopCmdStr]

    const containerConfig = new ContainerConfigBuilder(DOCKER_GETTING_STARTED_IMAGE)
        .withEntrypointOverride(entrypointArgs)
        .withCmdOverride(cmdArgs)
        .build()

    return containerConfig
}
