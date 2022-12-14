import {createEnclave} from "../../test_helpers/enclave_setup";
import {
    ContainerConfig,
    ContainerConfigBuilder,
    EnclaveContext,
    EnclaveID, KurtosisContext,
    ServiceContext, ServiceGUID,
    ServiceID,
    LogLineFilter, ServiceLog,
} from "kurtosis-sdk";
import {err, ok, Result} from "neverthrow";
import {Readable} from "stream";
import log from "loglevel";
import {newReceivedStreamContentPromise, ReceivedStreamContent} from "../../test_helpers/received_stream_content";
import {areEqualServiceGuidsSet, delay} from "../../test_helpers/test_helpers";

const TEST_NAME = "search-logs";
const IS_PARTITIONING_ENABLED = false;

const DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started";

const EXAMPLE_SERVICE_ID_PREFIX = "search-logs-";

const SHOULD_NOT_FOLLOW_LOGS  = false;

const WAIT_FOR_ALL_LOGS_BEING_COLLECTED_IN_SECONDS = 3

const SERVICE_1_SERVICE_ID = EXAMPLE_SERVICE_ID_PREFIX + "service-1"

const FIRST_FILTER_TEXT      = "Starting feature"
const SECOND_FILTER_TEXT     = "network"
const MATCH_REGEX_FILTER_STR  = "Starting.*logs'"

const LOG_LINE_1 = "Starting feature 'centralized logs'"
const LOG_LINE_2 = "Starting feature 'network partitioning'"
const LOG_LINE_3 = "Starting feature 'network soft partitioning'"
const LOG_LINE_4 = "The data have being loaded"

const EXPECTED_NON_EXISTENCE_SERVICE_GUIDS = new Set<ServiceGUID>;

const SERVICE_1_LOG_LINES = [LOG_LINE_1, LOG_LINE_2, LOG_LINE_3, LOG_LINE_4]

const LOG_LINES_BY_SERVICE = new Map<ServiceID, Array<string>>([
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

const EXPECTED_LOG_LINES_BY_REQUEST = Array<Array<string>>(
    Array<string>(
        LOG_LINE_1,
        LOG_LINE_2,
        LOG_LINE_3,
    ),
    Array<string>(
        LOG_LINE_1,
        LOG_LINE_4,
    ),
    Array<string>(
        LOG_LINE_1,
    ),
    Array<string>(
        LOG_LINE_2,
        LOG_LINE_3,
        LOG_LINE_4,
    ),
)

jest.setTimeout(180000)

test("Test Search Logs", TestSearchLogs)

async function TestSearchLogs() {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED);

    if (createEnclaveResult.isErr()) {
        throw createEnclaveResult.error;
    }

    const {enclaveContext, stopEnclaveFunction} = createEnclaveResult.value;

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        const serviceListResult: Result<Map<ServiceID, ServiceContext>, Error> = await addServices(enclaveContext, LOG_LINES_BY_SERVICE);

        if(serviceListResult.isErr()){
            throw new Error(`An error occurred adding the services for the test. Error:\n${serviceListResult.error}`);
        }

        const serviceList: Map<ServiceID, ServiceContext> = serviceListResult.value;

        if (LOG_LINES_BY_SERVICE.size != serviceList.size){
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

        for (let [serviceId, serviceCtx] of serviceList) {
            serviceGuid = serviceCtx.getServiceGUID();
            userServiceGuids.add(serviceGuid);
        }

        await delay(WAIT_FOR_ALL_LOGS_BEING_COLLECTED_IN_SECONDS * 1000);

        for (let i=0; i<FILTERS_BY_REQUEST.length; i++){
            const filter: LogLineFilter = FILTERS_BY_REQUEST[i];
            const expectedLogLines: Array<string> = EXPECTED_LOG_LINES_BY_REQUEST[i];
            const executionResult = await executeGetLogsRequestAndEvaluateResult(
                kurtosisContext,
                enclaveId,
                serviceGuid,
                userServiceGuids,
                filter,
                expectedLogLines,
            );

            if(executionResult.isErr()){
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
    expectedLogLines: Array<string>,
): Promise<Result<null, Error>> {

    const streamUserServiceLogsResult = await kurtosisCtx.getServiceLogs(enclaveId, userServiceGuids, SHOULD_NOT_FOLLOW_LOGS, logLineFilter);

    if (streamUserServiceLogsResult.isErr()) {
        return err(streamUserServiceLogsResult.error);
    }

    const serviceLogsReadable: Readable = streamUserServiceLogsResult.value;

    const receivedStreamContentPromise: Promise<ReceivedStreamContent> = newReceivedStreamContentPromise(
        serviceLogsReadable,
        serviceGuid,
        expectedLogLines,
        EXPECTED_NON_EXISTENCE_SERVICE_GUIDS,
    )

    const receivedStreamContent: ReceivedStreamContent = await receivedStreamContentPromise
    const receivedLogLines: Array<ServiceLog> = receivedStreamContent.receivedLogLines
    const receivedNotFoundServiceGuids: Set<ServiceGUID> = receivedStreamContent.receivedNotFoundServiceGuids

    if ( expectedLogLines.length === receivedLogLines.length) {
        receivedLogLines.forEach((logLine: ServiceLog, logLineIndex: number) => {
            if (expectedLogLines[logLineIndex] !== logLine.getContent()) {
                return err(new Error(`Expected to match the number ${logLineIndex} log line with this value ${expectedLogLines[logLineIndex]} but this one was received instead ${logLine.getContent()}`));
            }
        })
    } else {
        return err( new Error(`Expected to receive ${expectedLogLines.length} of log lines but ${receivedLogLines.length} log lines were received instead`));
    }

    if(!areEqualServiceGuidsSet(EXPECTED_NON_EXISTENCE_SERVICE_GUIDS, receivedNotFoundServiceGuids)) {
        return err(new Error(`Expected to receive a not found service GUIDs set equal to ${EXPECTED_NON_EXISTENCE_SERVICE_GUIDS} but a different set ${receivedNotFoundServiceGuids} was received instead`));
    }

    return ok(null);
}

async function addServices(
    enclaveContext: EnclaveContext,
    logLinesByServiceID: Map<ServiceID, Array<string>>,
): Promise<Result<Map<ServiceID, ServiceContext>, Error>> {

    const servicesAdded: Map<ServiceID, ServiceContext> = new Map<ServiceID, ServiceContext>();

    for (let [serviceId, logLines] of logLinesByServiceID){
        const containerConf: ContainerConfig = containerConfig(logLines);
        const addServiceResult = await enclaveContext.addService(serviceId, containerConf);

        if (addServiceResult.isErr()) {
            return err(new Error(`An error occurred adding service '${serviceId}'`));
        }

        const serviceContext: ServiceContext = addServiceResult.value;

        servicesAdded.set(serviceId, serviceContext)
    }
    return ok(servicesAdded)
}

function containerConfig(logLines: Array<string>): ContainerConfig {

    const entrypointArgs = ["/bin/sh", "-c"];

    const logLinesWithQuotes: Array<string> = new Array<string>();

    for (let logLine of logLines) {
        const logLineWithQuotes: string = `"${logLine}"`;
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
