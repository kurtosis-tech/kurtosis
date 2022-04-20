import { 
    ContainerConfig, 
    ContainerConfigBuilder, 
    PortProtocol, 
    PortSpec, 
    ServiceID, 
} from "kurtosis-core-api-lib"
import log from "loglevel"
import { ok, Result } from "neverthrow"

import { createEnclave } from "../../test_helpers/enclave_setup"

const TEST_NAME = "wait-for-endpoint-availability"
const IS_PARTITIONING_ENABLED = false

const DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started"
const EXAMPLE_SERVICE_ID:ServiceID = "docker-getting-started"
const EXAMPLE_SERVICE_PORT_ID = "http"
const EXAMPLE_SERVICE_PRIVATE_PORT_NUM = 80
const HEALTH_CHECK_URL_SLUG = ""
const HEALTHY_VALUE = ""

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS = 1
const WAIT_FOR_STARTUP_MAX_POLLS = 15
const WAIT_INITIAL_DELAY_MILLISECONDS = 500

const exampleServicePrivatePortSpec = new PortSpec(EXAMPLE_SERVICE_PRIVATE_PORT_NUM, PortProtocol.TCP)

jest.setTimeout(180000)

test("Test wait for endpoint availability", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction, kurtosisContext } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        const addServiceResult = await enclaveContext.addService(EXAMPLE_SERVICE_ID, containerConfigSupplier)

        if(addServiceResult.isErr()){
            log.error("An error occurred adding the datastore service")
            throw addServiceResult.error
        }

        // ------------------------------------- TEST RUN ----------------------------------------------

        const waitWaitForHttpGetEndpointAvailabilityResult = 
            await enclaveContext.waitForHttpGetEndpointAvailability(
                    EXAMPLE_SERVICE_ID, 
                    EXAMPLE_SERVICE_PRIVATE_PORT_NUM, 
                    HEALTH_CHECK_URL_SLUG, 
                    WAIT_INITIAL_DELAY_MILLISECONDS, 
                    WAIT_FOR_STARTUP_MAX_POLLS, 
                    WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS, 
                    HEALTHY_VALUE
            );
        
        if(waitWaitForHttpGetEndpointAvailabilityResult.isErr()){
            log.error("An error occurred waiting for the datastore service to become available")
            throw waitWaitForHttpGetEndpointAvailabilityResult.error
        }

        log.info(`Service: ${EXAMPLE_SERVICE_ID} is available`)

    }finally{
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
})


// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

function containerConfigSupplier(): Result<ContainerConfig, Error> {
    const exampleServicePort = new Map<string, PortSpec>()
    exampleServicePort.set(EXAMPLE_SERVICE_PORT_ID, exampleServicePrivatePortSpec)
    
    const containerConfig = new ContainerConfigBuilder(DOCKER_GETTING_STARTED_IMAGE)
        .withUsedPorts(exampleServicePort)
        .build()

    return ok(containerConfig)
}