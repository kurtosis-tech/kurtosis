import {
    EnclaveUUID,
    KurtosisContext,
    LogLineFilter,
    ServiceContext,
    ServiceUUID,
    ServiceName,
    ServiceLog,
} from "kurtosis-sdk";
import log from "loglevel";
import {err, Result} from "neverthrow";
import {createEnclave} from "../../test_helpers/enclave_setup";
import {addServicesWithLogLines, getLogsResponseAndEvaluateResponse} from "../../test_helpers/test_helpers";

const TEST_NAME = "stream-logs";
const IS_PARTITIONING_ENABLED = false;

const EXAMPLE_SERVICE_ID: ServiceName = "stream-logs";

const SHOULD_FOLLOW_LOGS = true;
const SHOULD_NOT_FOLLOW_LOGS = false;

const NON_EXISTENT_SERVICE_GUID = "stream-logs-1667939326-non-existent";

const FIRST_LOG_LINE_STR = "kurtosis"
const SECOND_LOG_LINE_STR = "test"
const THIRD_LOG_LINE_STR = "running"
const LAST_LOG_LINE_STR = "successfully"

const FIRST_LOG_LINE = new ServiceLog(FIRST_LOG_LINE_STR)
const SECOND_LOG_LINE = new ServiceLog(SECOND_LOG_LINE_STR)
const THIRD_LOG_LINE = new ServiceLog(THIRD_LOG_LINE_STR)
const LAST_LOG_LINE = new ServiceLog(LAST_LOG_LINE_STR)

const DO_NOT_FILTER_LOG_LINES = undefined;
const DOES_CONTAIN_TEXT_FILTER = LogLineFilter.NewDoesContainTextLogLineFilter(LAST_LOG_LINE_STR);

const EXAMPLE_SERVICE_LOG_LINES = [FIRST_LOG_LINE, SECOND_LOG_LINE, THIRD_LOG_LINE, LAST_LOG_LINE];

const LOG_LINES_BY_SERVICE = new Map<ServiceName, ServiceLog[]>([
    [EXAMPLE_SERVICE_ID, EXAMPLE_SERVICE_LOG_LINES],
])

class ServiceLogsRequestInfoAndExpectedResults {
    readonly requestedEnclaveUUID: EnclaveUUID;
    readonly requestedServiceGuids: Set<ServiceUUID>;
    readonly requestedFollowLogs: boolean;
    readonly expectedLogLines: ServiceLog[];
    readonly expectedNotFoundServiceGuids: Set<ServiceUUID>;
    readonly logLineFilter: LogLineFilter | undefined;

    constructor(
        requestedEnclaveUUID: EnclaveUUID,
        requestedServiceGuids: Set<ServiceUUID>,
        requestedFollowLogs: boolean,
        expectedLogLines: ServiceLog[],
        expectedNotFoundServiceGuids: Set<ServiceUUID>,
        logLineFilter: LogLineFilter | undefined
    ) {
        this.requestedEnclaveUUID = requestedEnclaveUUID;
        this.requestedServiceGuids = requestedServiceGuids;
        this.requestedFollowLogs = requestedFollowLogs;
        this.expectedLogLines = expectedLogLines;
        this.expectedNotFoundServiceGuids = expectedNotFoundServiceGuids;
        this.logLineFilter = logLineFilter;
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
        const newKurtosisContextResult = await KurtosisContext.newKurtosisContextFromLocalEngine();
        if (newKurtosisContextResult.isErr()) {
            log.error(`An error occurred connecting to the Kurtosis engine for running test stream logs`);
            return err(newKurtosisContextResult.error);
        }
        const kurtosisCtx = newKurtosisContextResult.value;

        const serviceListResult: Result<Map<ServiceName, ServiceContext>, Error> = await addServicesWithLogLines(enclaveContext, LOG_LINES_BY_SERVICE);

        if (serviceListResult.isErr()) {
            throw new Error(`An error occurred adding the services for the test. Error:\n${serviceListResult.error}`);
        }

        const serviceList: Map<ServiceName, ServiceContext> = serviceListResult.value;

        if (LOG_LINES_BY_SERVICE.size != serviceList.size) {
            throw new Error(`Expected number of added services '${LOG_LINES_BY_SERVICE.size}', but the actual number of added services is '${serviceList.size}'`);
        }

        // ------------------------------------- TEST RUN ----------------------------------------------

        const enclaveID: EnclaveUUID = enclaveContext.getEnclaveUuid();

        const serviceGuids: Set<ServiceUUID> = new Set<ServiceUUID>();

        for (let [, serviceCtx] of serviceList) {
            const serviceGuid = serviceCtx.getServiceUUID();
            serviceGuids.add(serviceGuid);
        }

        const serviceLogsRequestInfoAndExpectedResultsList = getServiceLogsRequestInfoAndExpectedResultsList(
            enclaveID,
            serviceGuids,
        )

        for (let serviceLogsRequestInfoAndExpectedResults of serviceLogsRequestInfoAndExpectedResultsList) {

            const requestedEnclaveUUID = serviceLogsRequestInfoAndExpectedResults.requestedEnclaveUUID;
            const requestedServiceGuids = serviceLogsRequestInfoAndExpectedResults.requestedServiceGuids;
            const requestedShouldFollowLogs = serviceLogsRequestInfoAndExpectedResults.requestedFollowLogs;
            const expectedLogLines = serviceLogsRequestInfoAndExpectedResults.expectedLogLines;
            const expectedNonExistenceServiceGuids = serviceLogsRequestInfoAndExpectedResults.expectedNotFoundServiceGuids;
            const filter = serviceLogsRequestInfoAndExpectedResults.logLineFilter;

            let expectedLogLinesByService: Map<ServiceUUID, ServiceLog[]> = new Map<ServiceUUID, ServiceLog[]>;
            for (const userServiceGuid of requestedServiceGuids) {
                expectedLogLinesByService.set(userServiceGuid, expectedLogLines);
            }

            const getLogsResponseResult = await getLogsResponseAndEvaluateResponse(
                kurtosisCtx,
                requestedEnclaveUUID,
                requestedServiceGuids,
                expectedLogLinesByService,
                expectedNonExistenceServiceGuids,
                requestedShouldFollowLogs,
                filter,
            )

            if (getLogsResponseResult.isErr()) {
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
function getServiceLogsRequestInfoAndExpectedResultsList(
    enclaveID: EnclaveUUID,
    serviceGuids: Set<ServiceUUID>,
): Array<ServiceLogsRequestInfoAndExpectedResults> {

    const emptyServiceGuids: Set<ServiceUUID> = new Set<ServiceUUID>();
    const nonExistentServiceGuids: Set<ServiceUUID> = new Set<ServiceUUID>();
    nonExistentServiceGuids.add(NON_EXISTENT_SERVICE_GUID);

    const firstCallRequestInfoAndExpectedResults: ServiceLogsRequestInfoAndExpectedResults = new ServiceLogsRequestInfoAndExpectedResults(
        enclaveID,
        serviceGuids,
        SHOULD_FOLLOW_LOGS,
        [LAST_LOG_LINE],
        emptyServiceGuids,
        DOES_CONTAIN_TEXT_FILTER,
    )

    const secondCallRequestInfoAndExpectedResults: ServiceLogsRequestInfoAndExpectedResults = new ServiceLogsRequestInfoAndExpectedResults(
        enclaveID,
        serviceGuids,
        SHOULD_FOLLOW_LOGS,
        [FIRST_LOG_LINE, SECOND_LOG_LINE, THIRD_LOG_LINE, LAST_LOG_LINE],
        emptyServiceGuids,
        DO_NOT_FILTER_LOG_LINES,
    )

    const thirdCallRequestInfoAndExpectedResults: ServiceLogsRequestInfoAndExpectedResults = new ServiceLogsRequestInfoAndExpectedResults(
        enclaveID,
        serviceGuids,
        SHOULD_NOT_FOLLOW_LOGS,
        [FIRST_LOG_LINE, SECOND_LOG_LINE, THIRD_LOG_LINE, LAST_LOG_LINE],
        emptyServiceGuids,
        DO_NOT_FILTER_LOG_LINES,
    )

    const fourthCallRequestInfoAndExpectedResults: ServiceLogsRequestInfoAndExpectedResults = new ServiceLogsRequestInfoAndExpectedResults(
        enclaveID,
        nonExistentServiceGuids,
        SHOULD_FOLLOW_LOGS,
        [],
        nonExistentServiceGuids,
        DO_NOT_FILTER_LOG_LINES,
    )

    const serviceLogsRequestInfoAndExpectedResultsList: Array<ServiceLogsRequestInfoAndExpectedResults> = new Array<ServiceLogsRequestInfoAndExpectedResults>();
    serviceLogsRequestInfoAndExpectedResultsList.push(firstCallRequestInfoAndExpectedResults)
    serviceLogsRequestInfoAndExpectedResultsList.push(secondCallRequestInfoAndExpectedResults)
    serviceLogsRequestInfoAndExpectedResultsList.push(thirdCallRequestInfoAndExpectedResults)
    serviceLogsRequestInfoAndExpectedResultsList.push(fourthCallRequestInfoAndExpectedResults)

    return serviceLogsRequestInfoAndExpectedResultsList
}
