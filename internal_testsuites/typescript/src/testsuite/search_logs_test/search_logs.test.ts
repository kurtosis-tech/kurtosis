import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    EnclaveID,
    KurtosisContext,
    LogLineFilter,
    ServiceContext,
    ServiceGUID,
    ServiceID,
    ServiceLog,
} from "kurtosis-sdk";
import {err, ok, Result} from "neverthrow";
import log from "loglevel";
import {getLogsResponseAndEvaluateResponse, addServicesWithLogLines} from "../../test_helpers/test_helpers";

const TEST_NAME = "search-logs";
const IS_PARTITIONING_ENABLED = false;

const EXAMPLE_SERVICE_ID_PREFIX = "search-logs-";

const SHOULD_NOT_FOLLOW_LOGS = false;

const SERVICE_1_SERVICE_ID = EXAMPLE_SERVICE_ID_PREFIX + "service-1";

const FIRST_FILTER_TEXT = "Starting feature";
const SECOND_FILTER_TEXT = "network";
const MATCH_REGEX_FILTER_STR = "Starting.*logs'";

const LOG_LINE_1 = new ServiceLog("Starting feature 'centralized logs'");
const LOG_LINE_2 = new ServiceLog("Starting feature 'network partitioning'");
const LOG_LINE_3 = new ServiceLog("Starting feature 'network soft partitioning'");
const LOG_LINE_4 = new ServiceLog("The data have being loaded");

const EXPECTED_NON_EXISTENCE_SERVICE_GUIDS = new Set<ServiceGUID>;

const SERVICE_1_LOG_LINES = [LOG_LINE_1, LOG_LINE_2, LOG_LINE_3, LOG_LINE_4];

const LOG_LINES_BY_SERVICE = new Map<ServiceID, ServiceLog[]>([
    [SERVICE_1_SERVICE_ID, SERVICE_1_LOG_LINES],
])

const DOES_CONTAIN_TEXT_FILTER = LogLineFilter.NewDoesContainTextLogLineFilter(FIRST_FILTER_TEXT);
const DOES_NOT_CONTAIN_TEXT_FILTER = LogLineFilter.NewDoesNotContainTextLogLineFilter(SECOND_FILTER_TEXT);
const DOES_CONTAIN_MATCH_REGEX_FILTER = LogLineFilter.NewDoesContainMatchRegexLogLineFilter(MATCH_REGEX_FILTER_STR);
const DOES_NOT_CONTAIN_MATCH_REGEX_FILTER = LogLineFilter.NewDoesNotContainMatchRegexLogLineFilter(MATCH_REGEX_FILTER_STR);

const FILTERS_BY_REQUEST = new Array<LogLineFilter>(
    DOES_CONTAIN_TEXT_FILTER,
    DOES_NOT_CONTAIN_TEXT_FILTER,
    DOES_CONTAIN_MATCH_REGEX_FILTER,
    DOES_NOT_CONTAIN_MATCH_REGEX_FILTER,
)

const EXPECTED_LOG_LINES_BY_REQUEST = Array<ServiceLog[]>(
    [
        LOG_LINE_1,
        LOG_LINE_2,
        LOG_LINE_3,
    ],
    [
        LOG_LINE_1,
        LOG_LINE_4,
    ],
    [
        LOG_LINE_1,
    ],
    [
        LOG_LINE_2,
        LOG_LINE_3,
        LOG_LINE_4,
    ],
)

jest.setTimeout(180000);

test("Test Search Logs", TestSearchLogs);

async function TestSearchLogs() {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED);

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error;
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value;

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        const serviceListResult: Result<Map<ServiceID, ServiceContext>, Error> = await addServicesWithLogLines(enclaveContext, LOG_LINES_BY_SERVICE);

        if (serviceListResult.isErr()) {
            throw new Error(`An error occurred adding the services for the test. Error:\n${serviceListResult.error}`);
        }

        const serviceList: Map<ServiceID, ServiceContext> = serviceListResult.value;

        if (LOG_LINES_BY_SERVICE.size != serviceList.size) {
            throw new Error(`Expected number of added services '${LOG_LINES_BY_SERVICE.size}', but the actual number of added services is '${serviceList.size}'`);
        }

        // ------------------------------------- TEST RUN ----------------------------------------------

        const newKurtosisContextResult = await KurtosisContext.newKurtosisContextFromLocalEngine();
        if (newKurtosisContextResult.isErr()) {
            log.error(`An error occurred connecting to the Kurtosis engine for running test search logs`)
            return err(newKurtosisContextResult.error)
        }
        const kurtosisContext = newKurtosisContextResult.value;

        const enclaveId: EnclaveID = enclaveContext.getEnclaveId();

        const userServiceGuids: Set<ServiceGUID> = new Set<ServiceGUID>();

        let serviceGuid: ServiceGUID = "";

        for (let [, serviceCtx] of serviceList) {
            serviceGuid = serviceCtx.getServiceGUID();
            userServiceGuids.add(serviceGuid);
        }

        for (let i = 0; i < FILTERS_BY_REQUEST.length; i++) {
            const filter: LogLineFilter = FILTERS_BY_REQUEST[i];
            const expectedLogLines: ServiceLog[] = EXPECTED_LOG_LINES_BY_REQUEST[i];
            const executionResult = await executeGetLogsRequestAndEvaluateResult(
                kurtosisContext,
                enclaveId,
                serviceGuid,
                userServiceGuids,
                filter,
                expectedLogLines,
            );

            if (executionResult.isErr()) {
                throw executionResult.error;
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
async function executeGetLogsRequestAndEvaluateResult(
    kurtosisCtx: KurtosisContext,
    enclaveId: EnclaveID,
    serviceGuid: ServiceGUID,
    userServiceGuids: Set<ServiceGUID>,
    logLineFilter: LogLineFilter,
    expectedLogLines: ServiceLog[],
): Promise<Result<null, Error>> {

    const serviceGuids: Set<ServiceGUID> = new Set<ServiceGUID>([
        serviceGuid,
    ])

    const expectedLogLinesByService = new Map<ServiceGUID, ServiceLog[]>([
        [serviceGuid, expectedLogLines],
    ])

    const getLogsResponseResult = await getLogsResponseAndEvaluateResponse(
        kurtosisCtx,
        enclaveId,
        serviceGuids,
        expectedLogLinesByService,
        EXPECTED_NON_EXISTENCE_SERVICE_GUIDS,
        SHOULD_NOT_FOLLOW_LOGS,
        logLineFilter,
    )

    if (getLogsResponseResult.isErr()) {
        throw getLogsResponseResult.error;
    }

    return ok(null);
}
