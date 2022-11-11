import {
    ContainerConfig,
    ContainerConfigBuilder,
    EnclaveID,
    KurtosisContext,
    ServiceGUID,
    ServiceID,
    ServiceLog,
    ServiceLogsStreamContent,
} from "kurtosis-sdk";
import log from "loglevel";
import {err} from "neverthrow";
import {createEnclave} from "../../test_helpers/enclave_setup";
import {Readable} from "stream";


const TEST_NAME = "stream-logs"
const IS_PARTITIONING_ENABLED = false

const DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started"
const EXAMPLE_SERVICE_ID: ServiceID = "stream-logs"

const WAIT_FOR_ALL_LOGS_BEING_COLLECTED_IN_SECONDS = 2

const SHOULD_FOLLOW_LOGS = true;
const SHOULD_NOT_FOLLOW_LOGS = false;

const NON_EXISTENT_SERVICE_GUID = "stream-logs-1667939326-non-existent"

class ServiceLogsRequestInfoAndExpectedResults {
    readonly requestedEnclaveID: EnclaveID;
    readonly requestedServiceGuids: Set<ServiceGUID>;
    readonly requestedFollowLogs: boolean;
    readonly expectedLogLines: Array<string>;
    readonly expectedNotFoundServiceGuids: Set<ServiceGUID>

    constructor(
        requestedEnclaveID: EnclaveID,
        requestedServiceGuids: Set<ServiceGUID>,
        requestedFollowLogs: boolean,
        expectedLogLines: Array<string>,
        expectedNotFoundServiceGuids: Set<ServiceGUID>
    ) {
        this.requestedEnclaveID = requestedEnclaveID;
        this.requestedServiceGuids = requestedServiceGuids;
        this.requestedFollowLogs = requestedFollowLogs;
        this.expectedLogLines = expectedLogLines
        this.expectedNotFoundServiceGuids = expectedNotFoundServiceGuids;
    }
}
class ReceivedStreamContent {
    readonly receivedLogLines: Array<ServiceLog>;
    readonly receivedNotFoundServiceGuids: Set<ServiceGUID>;

    constructor(
        receivedLogLines: Array<ServiceLog>,
        receivedNotFoundServiceGuids: Set<ServiceGUID>,
    ) {
        this.receivedLogLines = receivedLogLines;
        this.receivedNotFoundServiceGuids = receivedNotFoundServiceGuids;
    }
}

jest.setTimeout(180000)

test("Test Stream Logs", TestStreamLogs)

async function TestStreamLogs() {

    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED);

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error;
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value;

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        const addServiceResult = await enclaveContext.addService(EXAMPLE_SERVICE_ID, containerConfig());

        if (addServiceResult.isErr()) {
            log.error("An error occurred adding the datastore service");
            throw addServiceResult.error;
        }

        // ------------------------------------- TEST RUN ----------------------------------------------

        const newKurtosisContextResult = await KurtosisContext.newKurtosisContextFromLocalEngine();
        if (newKurtosisContextResult.isErr()) {
            log.error(`An error occurred connecting to the Kurtosis engine for running test stream logs`)
            return err(newKurtosisContextResult.error)
        }
        const kurtosisContext = newKurtosisContextResult.value;

        const enclaveID: EnclaveID = enclaveContext.getEnclaveId();

        const getServiceContextResult = await enclaveContext.getServiceContext(EXAMPLE_SERVICE_ID);

        if (getServiceContextResult.isErr()) {
            log.error(`An error occurred getting the service context for service "${EXAMPLE_SERVICE_ID}"`);
            throw getServiceContextResult.error;
        }

        const serviceContext = getServiceContextResult.value;

        const serviceGuid = serviceContext.getServiceGUID();

        const serviceGuids: Set<ServiceGUID> = new Set<ServiceGUID>();
        serviceGuids.add(serviceGuid);

        await delay(WAIT_FOR_ALL_LOGS_BEING_COLLECTED_IN_SECONDS * 1000);

        const serviceLogsRequestInfoAndExpectedResultsList = getServiceLogsRequestInfoAndExpectedResultsList(
            enclaveID,
            serviceGuids,
        )

        for (let serviceLogsRequestInfoAndExpectedResults of serviceLogsRequestInfoAndExpectedResultsList) {

            const requestedEnclaveId = serviceLogsRequestInfoAndExpectedResults.requestedEnclaveID
            const requestedServiceGuids = serviceLogsRequestInfoAndExpectedResults.requestedServiceGuids
            const requestedShouldFollowLogs = serviceLogsRequestInfoAndExpectedResults.requestedFollowLogs
            const expectedLogLines = serviceLogsRequestInfoAndExpectedResults.expectedLogLines
            const expectedNonExistenceServiceGuids = serviceLogsRequestInfoAndExpectedResults.expectedNotFoundServiceGuids

            const streamUserServiceLogsPromise = await kurtosisContext.getServiceLogs(requestedEnclaveId, requestedServiceGuids, requestedShouldFollowLogs);

            if (streamUserServiceLogsPromise.isErr()) {
                throw streamUserServiceLogsPromise.error;
            }

            const serviceLogsReadable: Readable = streamUserServiceLogsPromise.value;

            const receivedStreamContentPromise: Promise<ReceivedStreamContent> = newReceivedStreamContentPromise(
                serviceLogsReadable,
                serviceGuid,
                expectedLogLines,
                expectedNonExistenceServiceGuids,
            )

            const receivedStreamContent: ReceivedStreamContent = await receivedStreamContentPromise
            const receivedLogLines: Array<ServiceLog> = receivedStreamContent.receivedLogLines
            const receivedNotFoundServiceGuids: Set<ServiceGUID> = receivedStreamContent.receivedNotFoundServiceGuids

            if ( expectedLogLines.length === receivedLogLines.length) {
                receivedLogLines.forEach((logLine: ServiceLog, logLineIndex: number) => {
                    if (expectedLogLines[logLineIndex] !== logLine.getContent()) {
                        return err(new Error(`Expected to match the number ${logLineIndex} log line with this value ${expectedLogLines[logLineIndex]} but this one was received instead ${logLine.getContent()}`))
                    }
                })
            } else {
                throw new Error(`Expected to receive ${expectedLogLines.length} of log lines but ${receivedLogLines.length} log lines were received instead`)
            }

            if(!areEqualServiceGuidsSet(expectedNonExistenceServiceGuids, receivedNotFoundServiceGuids)) {
                throw new Error(`Expected to receive a not found service GUIDs set equal to ${expectedNonExistenceServiceGuids} but a different set ${receivedNotFoundServiceGuids} was received instead`)
            }
        }
    } finally {
        stopEnclaveFunction()
    }

    jest.clearAllTimers()

}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function newReceivedStreamContentPromise(
    serviceLogsReadable: Readable,
    serviceGuid: string,
    expectedLogLines: string[],
    expectedNonExistenceServiceGuids: Set<ServiceGUID>,
): Promise<ReceivedStreamContent> {
    const receivedStreamContentPromise: Promise<ReceivedStreamContent> = new Promise<ReceivedStreamContent>((resolve, _unusedReject) => {

        serviceLogsReadable.on('data', (serviceLogsStreamContent: ServiceLogsStreamContent) => {
            const serviceLogsByServiceGuids: Map<ServiceGUID, Array<ServiceLog>> = serviceLogsStreamContent.getServiceLogsByServiceGuids()
            const notFoundServiceGuids: Set<ServiceGUID> = serviceLogsStreamContent.getNotFoundServiceGuids()

            let receivedLogLines: Array<ServiceLog> | undefined = serviceLogsByServiceGuids.get(serviceGuid);

            if (expectedNonExistenceServiceGuids.size > 0) {
                if(receivedLogLines !== undefined){
                    throw new Error(`Expected to receive undefined log lines but these log lines content ${receivedLogLines} was received instead`)
                }
            } else {
                if(receivedLogLines === undefined){
                    throw new Error("Expected to receive log lines content but and undefined value was received instead")
                }
            }

            if(receivedLogLines === undefined) {
                receivedLogLines = new Array<ServiceLog>()
            }

            if (receivedLogLines.length === expectedLogLines.length) {
                const receivedStreamContent: ReceivedStreamContent = new ReceivedStreamContent(
                    receivedLogLines,
                    notFoundServiceGuids,
                )
                serviceLogsReadable.destroy()
                resolve(receivedStreamContent)
            }
        })

        serviceLogsReadable.on('error', function (readableErr: { message: any; }) {
            if (!serviceLogsReadable.destroyed) {
                serviceLogsReadable.destroy()
                throw new Error(`Expected read all user service logs but an error was received from the user service readable object.\n Error: "${readableErr.message}"`)
            }
        })

        serviceLogsReadable.on('end', function () {
            if (!serviceLogsReadable.destroyed) {
                serviceLogsReadable.destroy()
                throw new Error("Expected read all user service logs but the user service readable logs object has finished before reading all the logs.")
            }
        })
    })

    return receivedStreamContentPromise;
}

function containerConfig(): ContainerConfig {

    const entrypointArgs = ["/bin/sh", "-c"]
    const cmdArgs = ["for i in kurtosis test running successfully; do echo \"$i\"; done;"]

    const containerConfig = new ContainerConfigBuilder(DOCKER_GETTING_STARTED_IMAGE)
        .withEntrypointOverride(entrypointArgs)
        .withCmdOverride(cmdArgs)
        .build()

    return containerConfig
}

function delay(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

function getServiceLogsRequestInfoAndExpectedResultsList(
    enclaveID: EnclaveID,
    serviceGuids: Set<ServiceGUID>,
): Array<ServiceLogsRequestInfoAndExpectedResults> {
    const expectedEmptyLogLineValues: Array<string> = [];
    const emptyServiceGuids: Set<ServiceGUID> = new Set<ServiceGUID>();

    const nonExistentServiceGuids: Set<ServiceGUID> = new Set<ServiceGUID>();
    nonExistentServiceGuids.add(NON_EXISTENT_SERVICE_GUID);

    const expectedLogLineValues = ["kurtosis", "test", "running", "successfully"];

    const firstCallRequestInfoAndExpectedResults: ServiceLogsRequestInfoAndExpectedResults = new ServiceLogsRequestInfoAndExpectedResults(
        enclaveID,
        serviceGuids,
        SHOULD_NOT_FOLLOW_LOGS,
        expectedLogLineValues,
        emptyServiceGuids,
    )

    const secondCallRequestInfoAndExpectedResults: ServiceLogsRequestInfoAndExpectedResults = new ServiceLogsRequestInfoAndExpectedResults(
        enclaveID,
        serviceGuids,
        SHOULD_FOLLOW_LOGS,
        expectedLogLineValues,
        emptyServiceGuids,
    )

    const thirdCallRequestInfoAndExpectedResults: ServiceLogsRequestInfoAndExpectedResults = new ServiceLogsRequestInfoAndExpectedResults(
        enclaveID,
        nonExistentServiceGuids,
        SHOULD_FOLLOW_LOGS,
        expectedEmptyLogLineValues,
        nonExistentServiceGuids,
    )

    const serviceLogsRequestInfoAndExpectedResultsList: Array<ServiceLogsRequestInfoAndExpectedResults> = new Array<ServiceLogsRequestInfoAndExpectedResults>();
    serviceLogsRequestInfoAndExpectedResultsList.push(firstCallRequestInfoAndExpectedResults)
    serviceLogsRequestInfoAndExpectedResultsList.push(secondCallRequestInfoAndExpectedResults)
    serviceLogsRequestInfoAndExpectedResultsList.push(thirdCallRequestInfoAndExpectedResults)

    return serviceLogsRequestInfoAndExpectedResultsList
}

function areEqualServiceGuidsSet(firstSet: Set<ServiceGUID>, secondSet: Set<ServiceGUID>): boolean {
    const haveEqualSize: boolean = firstSet.size === secondSet.size;
    const haveEqualContent: boolean = [...firstSet].every((x) => secondSet.has(x));

    const areEqual: boolean = haveEqualSize && haveEqualContent;

    return areEqual
}
