import {createEnclave} from "../../test_helpers/enclave_setup";
import {ContainerConfig, ContainerConfigBuilder} from "kurtosis-core-sdk";
import {err, ok, Result} from "neverthrow";
import log from "loglevel";

const TEST_NAME = "pause-unpause-test"
const IS_PARTITIONING_ENABLED = false
const PAUSE_UNPAUSE_TEST_IMAGE =  "alpine:3.12.4"
const TEST_SERVICE_ID = "test";
const TEST_LOG_FILEPATH = "/test.log"
const DELAY_BETWEEN_COMMANDS_IN_SECONDS = 4
const MINIMUM_GAP_REQUIREMENT_IN_SECONDS = 3

jest.setTimeout(180000)

test("Test service pause", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const containerConfig = getContainerConfig()

        const addServiceResult = await enclaveContext.addService(TEST_SERVICE_ID, containerConfig)

        if(addServiceResult.isErr()) {
            log.error(`An error occurred starting service "${TEST_SERVICE_ID}"`);
            throw addServiceResult.error
        };

        const testServiceContext = addServiceResult.value
        await delay(DELAY_BETWEEN_COMMANDS_IN_SECONDS * 1000)
        log.info("Pausing service.")
        // ------------------------------------- TEST RUN ----------------------------------------------
        const pauseServiceResult = await enclaveContext.pauseService(TEST_SERVICE_ID)
        if(pauseServiceResult.isErr()){
            log.error("An error occurred pausing service.")
            throw(pauseServiceResult.error)

        }

        await delay(DELAY_BETWEEN_COMMANDS_IN_SECONDS * 1000)
        const unpauseServiceResult = await enclaveContext.unpauseService(TEST_SERVICE_ID)
        if(unpauseServiceResult.isErr()){
            log.error("An error occurred unpausing service.")
            throw(unpauseServiceResult.error)

        }
        await delay(DELAY_BETWEEN_COMMANDS_IN_SECONDS * 1000)
        const testLogResults = await testServiceContext.execCommand(["cat", TEST_LOG_FILEPATH])
        if (testLogResults.isErr()) {
            log.error("An error occurred reading test logs")
            throw(testLogResults.error)
        }
        const logString = testLogResults.value[1]
        const logStringArray = logString.split("\n")
        let foundGap = false
        for (let i = 0; i < logStringArray.length; i++) {
            if(i > 0) {
                const logLine = logStringArray[i].trim()
                const currentSeconds = Number(logLine)
                const previousLogLine = logStringArray[i-1].trim()
                const previousSeconds = Number(previousLogLine)
                if(currentSeconds-previousSeconds > MINIMUM_GAP_REQUIREMENT_IN_SECONDS){
                    foundGap = true
                }
                log.info("Seconds: " + currentSeconds)
            }
        }

        if(!foundGap) {
            throw new Error("Failed to find a >" + MINIMUM_GAP_REQUIREMENT_IN_SECONDS + " second gap in second-ticker, which was expected given service should have been paused.")
        }
    } finally{
        stopEnclaveFunction()
    }
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function getContainerConfig(): ContainerConfig {

    // We spam timestamps so that we can measure pausing processes (no more log output) and unpausing (log output resumes)
    const entrypointArgs = ["/bin/sh", "-c"]
    const cmdArgs = ["while sleep 1; do ts=$(date +\"%s\") ; echo \"$ts\" >> " + TEST_LOG_FILEPATH + " ; done"]

    const containerConfig = new ContainerConfigBuilder(PAUSE_UNPAUSE_TEST_IMAGE)
        .withEntrypointOverride(entrypointArgs)
        .withCmdOverride(cmdArgs)
        .build()

    return containerConfig
}

function delay(ms: number) {
    return new Promise( resolve => setTimeout(resolve, ms) );
}