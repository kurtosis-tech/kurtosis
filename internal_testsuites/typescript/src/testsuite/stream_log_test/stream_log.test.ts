import {
    ContainerConfig,
    ContainerConfigBuilder,
    KurtosisContext,
    PortProtocol,
    PortSpec,
    ServiceGUID,
    ServiceID
} from "kurtosis-sdk";
import log from "loglevel";
import {err} from "neverthrow";
import {Readable} from "stream";
import {createEnclave} from "../../test_helpers/enclave_setup";


const TEST_NAME = "stream-logs"
const IS_PARTITIONING_ENABLED = false

const DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started"
const EXAMPLE_SERVICE_ID:ServiceID = "stream-logs"

const EXAMPLE_SERVICE_PORT_ID = "http"
const EXAMPLE_SERVICE_PRIVATE_PORT_NUM = 80

const WAIT_FOR_ALL_LOGS_BEING_COLLECTED_IN_SECONDS = 2

const exampleServicePrivatePortSpec = new PortSpec(EXAMPLE_SERVICE_PRIVATE_PORT_NUM, PortProtocol.TCP)

jest.setTimeout(180000)

test("Test Stream Logs", TestStreamLogs)

async function TestStreamLogs() {

    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED);

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value;

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        const addServiceResult = await enclaveContext.addService(EXAMPLE_SERVICE_ID, containerConfig());

        if(addServiceResult.isErr()){
            log.error("An error occurred adding the datastore service")
            throw addServiceResult.error
        }

        // ------------------------------------- TEST RUN ----------------------------------------------

        const newKurtosisContextResult = await KurtosisContext.newKurtosisContextFromLocalEngine();
        if(newKurtosisContextResult.isErr()) {
            log.error(`An error occurred connecting to the Kurtosis engine for running test stream logs`)
            return err(newKurtosisContextResult.error)
        }
        const kurtosisContext = newKurtosisContextResult.value;

        const enclaveID: string = enclaveContext.getEnclaveId();

        const getServiceContextResult = await enclaveContext.getServiceContext(EXAMPLE_SERVICE_ID)

        if (getServiceContextResult.isErr()) {
            log.error(`An error occurred getting the service context for service "${EXAMPLE_SERVICE_ID}"`)
            throw getServiceContextResult.error
        }

        const serviceContext = getServiceContextResult.value

        const userServiceGuid = serviceContext.getServiceGUID()

        const userServiceGUIDs: Set<ServiceGUID> = new Set<ServiceGUID>();
        userServiceGUIDs.add(userServiceGuid);

        await delay(WAIT_FOR_ALL_LOGS_BEING_COLLECTED_IN_SECONDS * 1000)

        const streamUserServiceLogsPromise = await kurtosisContext.streamUserServiceLogs(enclaveID, userServiceGUIDs);

        if(streamUserServiceLogsPromise.isErr()) {
            throw streamUserServiceLogsPromise.error
        }

        const streamUserServiceLogs: Map<ServiceGUID, Readable> = streamUserServiceLogsPromise.value;

        const expectedLogLines = ["kurtosis", "test", "running", "successfully"];

        const expectedAmountOfLogLines = 4;

        const receivedLogLines: string[] = []

        streamUserServiceLogs.forEach((userServiceReadable) => {
            if (userServiceReadable !== undefined) {
                userServiceReadable.on('data', function(logline) {
                    receivedLogLines.push(logline.toString())
                    if (receivedLogLines.length === expectedAmountOfLogLines) {
                        receivedLogLines.forEach((logline, loglineIndex) => {
                            if (expectedLogLines[loglineIndex] !=  logline) {
                                return err( new Error(`Expected to match the ${loglineIndex}º log line with this value ${expectedLogLines[loglineIndex]} but this one was received instead ${logline}`))
                            }
                        })
                        if(!userServiceReadable.destroyed) {
                            userServiceReadable.destroy()
                        }
                    }
                })
                userServiceReadable.on('error', function(readableErr) {
                    if(!userServiceReadable.destroyed) {
                        userServiceReadable.destroy()
                    }
                    throw new Error(`Expected read all user service logs but an error was received from the user service readable object.\n Error: "${readableErr.message}"`)
                })
                userServiceReadable.on('end', function() {
                    if(!userServiceReadable.destroyed) {
                        userServiceReadable.destroy()
                        throw new Error(`Expected read all user service logs but the user service readable logs object has finished before reading all the logs.`)
                    }
                })
            }
        })

    }finally{
       stopEnclaveFunction()
    }

    jest.clearAllTimers()

}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function containerConfig(): ContainerConfig {

    const entrypointArgs = ["/bin/sh", "-c"]
    const cmdArgs = ["for i in kurtosis test running successfully; do echo \"$i\"; if [ \"$i\" == \"successfully\" ]; then sleep 300; fi; done;"]

    const exampleServicePort = new Map<string, PortSpec>()
    exampleServicePort.set(EXAMPLE_SERVICE_PORT_ID, exampleServicePrivatePortSpec)

    const containerConfig = new ContainerConfigBuilder(DOCKER_GETTING_STARTED_IMAGE)
        .withEntrypointOverride(entrypointArgs)
        .withCmdOverride(cmdArgs)
        .withUsedPorts(exampleServicePort)
        .build()

    return containerConfig
}

function delay(ms: number) {
    return new Promise( resolve => setTimeout(resolve, ms) );
}