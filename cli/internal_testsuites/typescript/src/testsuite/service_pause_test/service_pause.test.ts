import {createEnclave} from "../../test_helpers/enclave_setup";
import {ContainerConfig, ContainerConfigBuilder} from "kurtosis-core-api-lib";
import {err, ok, Result} from "neverthrow";
import log from "loglevel";

const TEST_NAME = "pause-unpause-test"
const IS_PARTITIONING_ENABLED = false
const PAUSE_UNPAUSE_TEST_IMAGE =  "alpine:3.12.4"
const TEST_SERVICE_ID = "test";
const TEST_LOG_FILEPATH = "/test.log"

jest.setTimeout(180000)

test("Test service pause", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const containerConfigSupplier = getContainerConfigSupplier()

        const addServiceResult = await enclaveContext.addService(TEST_SERVICE_ID, containerConfigSupplier)

        if(addServiceResult.isErr()) {
            log.error(`An error occurred starting service "${TEST_SERVICE_ID}"`);
            throw addServiceResult.error
        };

        const testServiceContext = addServiceResult.value
        await delay(4000)
        // ------------------------------------- TEST RUN ----------------------------------------------
        const pauseServiceResult = await enclaveContext.pauseService(TEST_SERVICE_ID)
        if(pauseServiceResult.isErr()){
            log.error("An error occurred pausing service.")
            throw(pauseServiceResult.error)

        }
        // Wait 5 seconds
        await delay(4000)
        const unpauseServiceResult = await enclaveContext.unpauseService(TEST_SERVICE_ID)
        if(unpauseServiceResult.isErr()){
            log.error("An error occurred unpausing service.")
            throw(unpauseServiceResult.error)

        }
        await delay(4000)
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
                if(currentSeconds-previousSeconds > 2){
                    foundGap = true
                }
                log.info("Seconds: " + currentSeconds)
            }
        }
        if(!foundGap) {
            throw new Error("Failed to find a >2 second gap in second-ticker, which was expected given service should have been paused.")
        }
    } finally{
        stopEnclaveFunction()
    }
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function getContainerConfigSupplier(): (ipAddr:string) => Result<ContainerConfig, Error> {

    const containerConfigSupplier = (ipAddr:string): Result<ContainerConfig, Error> => {

        // We spam timestamps so that we can measure pausing processes (no more log output) and unpausing (log output resumes)
        const entrypointArgs = ["/bin/sh", "-c"]
        const cmdArgs = ["while sleep 1; do ts=$(date +\"%s\") ; echo \"$ts\" > " + TEST_LOG_FILEPATH + " ; done"]

        const containerConfig = new ContainerConfigBuilder(PAUSE_UNPAUSE_TEST_IMAGE)
            .withEntrypointOverride(entrypointArgs)
            .withCmdOverride(cmdArgs)
            .build()

        return ok(containerConfig)
    }

    return containerConfigSupplier
}

function delay(ms: number) {
    return new Promise( resolve => setTimeout(resolve, ms) );
}