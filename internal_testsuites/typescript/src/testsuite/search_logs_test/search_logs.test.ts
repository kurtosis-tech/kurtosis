import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    EnclaveUUID,
    KurtosisContext,
    LogLineFilter,
    ServiceContext,
    ServiceUUID,
    ServiceName,
    ServiceLog,
} from "kurtosis-sdk";
import {err, ok, Result} from "neverthrow";
import log from "loglevel";
import {addServicesWithLogLines, delay, getLogsResponseAndEvaluateResponse} from "../../test_helpers/test_helpers";

const TEST_NAME = "search-logs";
const IS_PARTITIONING_ENABLED = false;

const EXAMPLE_SERVICE_NAME_PREFIX = "search-logs-";

const SHOULD_FOLLOW_LOGS = true;
const SHOULD_NOT_FOLLOW_LOGS = false;

const SERVICE_1_SERVICE_NAME = EXAMPLE_SERVICE_NAME_PREFIX + "service-1";

const FIRST_FILTER_TEXT = "The data have being loaded";
const SECOND_FILTER_TEXT = "Starting feature";
const THIRD_FILTER_TEXT = "network";
const MATCH_REGEX_FILTER_STR = "Starting.*logs'";

const LOG_LINE_1 = new ServiceLog("Starting feature 'centralized logs'");
const LOG_LINE_2 = new ServiceLog("Starting feature 'network partitioning'");
const LOG_LINE_3 = new ServiceLog("Starting feature 'network soft partitioning'");
const LOG_LINE_4 = new ServiceLog("The data have being loaded");

const EXPECTED_NON_EXISTENCE_SERVICE_UUIDS = new Set<ServiceUUID>;

const SERVICE_1_LOG_LINES = [LOG_LINE_1, LOG_LINE_2, LOG_LINE_3, LOG_LINE_4];

const LOG_LINES_BY_SERVICE = new Map<ServiceName, ServiceLog[]>([
    [SERVICE_1_SERVICE_NAME, SERVICE_1_LOG_LINES],
])

const DOES_CONTAIN_TEXT_FILTER_FOR_FIRST_REQUEST = LogLineFilter.NewDoesContainTextLogLineFilter(FIRST_FILTER_TEXT);
const DOES_CONTAIN_TEXT_FILTER_FOR_SECOND_REQUEST = LogLineFilter.NewDoesContainTextLogLineFilter(SECOND_FILTER_TEXT);
const DOES_NOT_CONTAIN_TEXT_FILTER = LogLineFilter.NewDoesNotContainTextLogLineFilter(THIRD_FILTER_TEXT);
const DOES_CONTAIN_MATCH_REGEX_FILTER = LogLineFilter.NewDoesContainMatchRegexLogLineFilter(MATCH_REGEX_FILTER_STR);
const DOES_NOT_CONTAIN_MATCH_REGEX_FILTER = LogLineFilter.NewDoesNotContainMatchRegexLogLineFilter(MATCH_REGEX_FILTER_STR);

const FILTERS_BY_REQUEST = new Array<LogLineFilter>(
    DOES_CONTAIN_TEXT_FILTER_FOR_FIRST_REQUEST,
    DOES_CONTAIN_TEXT_FILTER_FOR_SECOND_REQUEST,
    DOES_NOT_CONTAIN_TEXT_FILTER,
    DOES_CONTAIN_MATCH_REGEX_FILTER,
    DOES_NOT_CONTAIN_MATCH_REGEX_FILTER,
)

const EXPECTED_LOG_LINES_BY_REQUEST = Array<ServiceLog[]>(
    [
        LOG_LINE_1,
    ],
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

const SHOULD_FOLLOW_LOGS_VALUES_BY_REQUEST: boolean[] = [
    SHOULD_FOLLOW_LOGS,
    SHOULD_NOT_FOLLOW_LOGS,
    SHOULD_NOT_FOLLOW_LOGS,
    SHOULD_NOT_FOLLOW_LOGS,
    SHOULD_NOT_FOLLOW_LOGS,
]

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

        const serviceListResult: Result<Map<ServiceName, ServiceContext>, Error> = await addServicesWithLogLines(enclaveContext, LOG_LINES_BY_SERVICE);

        if (serviceListResult.isErr()) {
            throw new Error(`An error occurred adding the services for the test. Error:\n${serviceListResult.error}`);
        }

        const serviceList: Map<ServiceName, ServiceContext> = serviceListResult.value;

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

        const enclaveUuid: EnclaveUUID = enclaveContext.getEnclaveUuid();

        const userServiceUuids: Set<ServiceUUID> = new Set<ServiceUUID>();

        let serviceUuid: ServiceUUID = "";

        for (let [, serviceCtx] of serviceList) {
            serviceUuid = serviceCtx.getServiceUUID();
            userServiceUuids.add(serviceUuid);
        }

        for (let i = 0; i < FILTERS_BY_REQUEST.length; i++) {
            const filter: LogLineFilter = FILTERS_BY_REQUEST[i];
            const expectedLogLines: ServiceLog[] = EXPECTED_LOG_LINES_BY_REQUEST[i];
            const shouldFollowLogsOption: boolean = SHOULD_FOLLOW_LOGS_VALUES_BY_REQUEST[i];
            const executionResult = await executeGetLogsRequestAndEvaluateResult(
                kurtosisContext,
                enclaveUuid,
                serviceUuid,
                userServiceUuids,
                filter,
                expectedLogLines,
                shouldFollowLogsOption,
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
    enclaveUuid: EnclaveUUID,
    serviceUuid: ServiceUUID,
    userServiceUuids: Set<ServiceUUID>,
    logLineFilter: LogLineFilter,
    expectedLogLines: ServiceLog[],
    shouldFollowLogs: boolean,
): Promise<Result<null, Error>> {

    const serviceUuids: Set<ServiceUUID> = new Set<ServiceUUID>([
        serviceUuid,
    ])

    const expectedLogLinesByService = new Map<ServiceUUID, ServiceLog[]>([
        [serviceUuid, expectedLogLines],
    ])

    const getLogsResponseResult = await getLogsResponseAndEvaluateResponse(
        kurtosisCtx,
        enclaveUuid,
        serviceUuids,
        expectedLogLinesByService,
        EXPECTED_NON_EXISTENCE_SERVICE_UUIDS,
        shouldFollowLogs,
        logLineFilter,
    )

    if (getLogsResponseResult.isErr()) {
        throw getLogsResponseResult.error;
    }

    return ok(null);
}
