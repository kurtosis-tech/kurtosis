import {
    ContainerConfig,
    ContainerConfigBuilder,
    EnclaveID,
    KurtosisContext,
    ServiceGUID,
    ServiceID,
    ServiceLog,
} from "kurtosis-sdk";
import log from "loglevel";
import {err} from "neverthrow";
import {createEnclave} from "../../test_helpers/enclave_setup";
import {Readable} from "stream";
import {newReceivedStreamContentPromise, ReceivedStreamContent} from "../../test_helpers/received_stream_content";
import {areEqualServiceGuidsSet, delay, getLogsResponseAndEvaluateResponse} from "../../test_helpers/test_helpers";


const TEST_NAME = "stream-logs";
const IS_PARTITIONING_ENABLED = false;

const DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started";
const EXAMPLE_SERVICE_ID: ServiceID = "stream-logs";

const SHOULD_FOLLOW_LOGS = true;
const SHOULD_NOT_FOLLOW_LOGS = false;

const GET_LOGS_MAX_RETRIES = 5;
const GET_LOGS_TIME_BETWEEN_RETRIES_MILLISECONDS  = 1000;

const NON_EXISTENT_SERVICE_GUID = "stream-logs-1667939326-non-existent";

class ServiceLogsRequestInfoAndExpectedResults {
    readonly requestedEnclaveID: EnclaveID;
    readonly requestedServiceGuids: Set<ServiceGUID>;
    readonly requestedFollowLogs: boolean;
    readonly expectedLogLines: ServiceLog[];
    readonly expectedNotFoundServiceGuids: Set<ServiceGUID>

    constructor(
        requestedEnclaveID: EnclaveID,
        requestedServiceGuids: Set<ServiceGUID>,
        requestedFollowLogs: boolean,
        expectedLogLines: ServiceLog[],
        expectedNotFoundServiceGuids: Set<ServiceGUID>
    ) {
        this.requestedEnclaveID = requestedEnclaveID;
        this.requestedServiceGuids = requestedServiceGuids;
        this.requestedFollowLogs = requestedFollowLogs;
        this.expectedLogLines = expectedLogLines
        this.expectedNotFoundServiceGuids = expectedNotFoundServiceGuids;
    }
}

jest.setTimeout(180000);

test("Test Stream Logs", TestStreamLogs);

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
            log.error(`An error occurred connecting to the Kurtosis engine for running test stream logs`);
            return err(newKurtosisContextResult.error);
        }
        const kurtosisCtx = newKurtosisContextResult.value;

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

        const serviceLogsRequestInfoAndExpectedResultsList = getServiceLogsRequestInfoAndExpectedResultsList(
            enclaveID,
            serviceGuids,
        )

        for (let serviceLogsRequestInfoAndExpectedResults of serviceLogsRequestInfoAndExpectedResultsList) {

            const requestedEnclaveId = serviceLogsRequestInfoAndExpectedResults.requestedEnclaveID;
            const requestedServiceGuids = serviceLogsRequestInfoAndExpectedResults.requestedServiceGuids;
            const requestedShouldFollowLogs = serviceLogsRequestInfoAndExpectedResults.requestedFollowLogs;
            const expectedLogLines = serviceLogsRequestInfoAndExpectedResults.expectedLogLines;
            const expectedNonExistenceServiceGuids = serviceLogsRequestInfoAndExpectedResults.expectedNotFoundServiceGuids;

            const requestedLogLineFilter = undefined;

            let expectedLogLinesByService: Map<ServiceGUID, ServiceLog[]> = new Map<ServiceGUID, ServiceLog[]>;
            for (const userServiceGuid of requestedServiceGuids) {
                expectedLogLinesByService.set(userServiceGuid, expectedLogLines);
            };

            const getLogsResponseResult = await getLogsResponseAndEvaluateResponse(
                kurtosisCtx,
                requestedEnclaveId,
                requestedServiceGuids,
                expectedLogLinesByService,
                expectedNonExistenceServiceGuids,
                requestedShouldFollowLogs,
                requestedLogLineFilter,
                GET_LOGS_MAX_RETRIES,
                GET_LOGS_TIME_BETWEEN_RETRIES_MILLISECONDS,
            )

            if (getLogsResponseResult.isErr()){
                throw getLogsResponseResult.error
            }
        }
    } finally {
        stopEnclaveFunction();
    }

    jest.clearAllTimers();

}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function containerConfig(): ContainerConfig {

    const entrypointArgs = ["/bin/sh", "-c"]
    const cmdArgs = ["for i in kurtosis test running successfully; do echo \"$i\"; done;"]

    const containerConfig = new ContainerConfigBuilder(DOCKER_GETTING_STARTED_IMAGE)
        .withEntrypointOverride(entrypointArgs)
        .withCmdOverride(cmdArgs)
        .build()

    return containerConfig
}


function getServiceLogsRequestInfoAndExpectedResultsList(
    enclaveID: EnclaveID,
    serviceGuids: Set<ServiceGUID>,
): Array<ServiceLogsRequestInfoAndExpectedResults> {
    const expectedEmptyLogLineValues: ServiceLog[] = [];
    const emptyServiceGuids: Set<ServiceGUID> = new Set<ServiceGUID>();

    const nonExistentServiceGuids: Set<ServiceGUID> = new Set<ServiceGUID>();
    nonExistentServiceGuids.add(NON_EXISTENT_SERVICE_GUID);

    const expectedLogLineValues: ServiceLog[] = [
        new ServiceLog("kurtosis"),
        new ServiceLog("test"),
        new ServiceLog("running"),
        new ServiceLog("successfully"),
    ];

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

