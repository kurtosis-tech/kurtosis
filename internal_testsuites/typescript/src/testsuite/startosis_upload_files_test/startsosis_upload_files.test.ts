import {EnclaveContext, PortSpec} from "kurtosis-sdk"
import log from "loglevel"
import { err, ok, Result } from "neverthrow";
import * as grpc from "@grpc/grpc-js"

import { createEnclave } from "../../test_helpers/enclave_setup";
import { createDatastoreClient, waitForHealthy } from "../../test_helpers/test_helpers";

const TEST_NAME = "upload-files-test"
const IS_PARTITIONING_ENABLED = false

const SERVICE_ID = "example-datastore-server-1"
const PORT_ID = "grpc"

const PATH_TO_MOUNT_UPLOADED_DIR = "/uploads"
const PATH_TO_CHECK_FOR_UPLOADED_FILE = "/uploads/lib.star"

const STARTOSIS_SCRIPT = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + SERVICE_ID + `"
DATASTORE_PORT_ID = "` + PORT_ID + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

DIR_TO_UPLOAD = "github.com/kurtosis-tech/datastore-army-module-demo/lib"
PATH_TO_MOUNT_UPLOADED_DIR = "` + PATH_TO_MOUNT_UPLOADED_DIR + `"

print("Adding service " + DATASTORE_SERVICE_ID + ".")

uploaded_artifact_uuid = upload_files(DIR_TO_UPLOAD)
print("Uploaded " + uploaded_artifact_uuid)


service_config = struct(
    container_image_name = DATASTORE_IMAGE,
    used_ports = {
        DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
    },
	files_artifact_mount_dirpaths = {
		uploaded_artifact_uuid: PATH_TO_MOUNT_UPLOADED_DIR
	}
)

add_service(service_id = DATASTORE_SERVICE_ID, service_config = service_config)`


const WAIT_FOR_STARTUP_MAX_POLLS = 30
const MILLIS_BETWEEN_AVAILABILITY_RETRIES = 1000

jest.setTimeout(180000)

test("Test upload files startosis", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        log.info("Loading module...")
        const executeStartosisScriptResult = await enclaveContext.executeStartosisScript(STARTOSIS_SCRIPT)

        if (executeStartosisScriptResult.isErr()) {
            log.error("An error occurred executing the Startosis SCript")
            throw executeStartosisScriptResult.error
        }
        const executeStartosisScriptValue = executeStartosisScriptResult.value
        const expectedScriptOutput = `Adding service example-datastore-server-1.
Uploaded {{kurtosis:FILENAME_NOT_USED-13:38.artifact_uuid}}
`

        if (expectedScriptOutput !== executeStartosisScriptValue.getSerializedScriptOutput()) {
            throw err(new Error(`Expected output to be '${expectedScriptOutput} got '${executeStartosisScriptValue.getSerializedScriptOutput()}'`))
        }

        if (executeStartosisScriptValue.getInterpretationError() !== "") {
            throw err(new Error(`Expected Empty Interpretation Error got '${executeStartosisScriptValue.getInterpretationError()}'`))
        }

        if (executeStartosisScriptValue.getExecutionError() !== "") {
            throw err(new Error(`Expected Empty Execution Error got '${executeStartosisScriptValue.getExecutionError()}'`))
        }

        if (executeStartosisScriptValue.getValidationErrorsList().length != 0) {
            throw err(new Error(`Expected Empty Validation Error got '${executeStartosisScriptValue.getValidationErrorsList()}'`))
        }
        log.info("Script Executed Successfully")

        // ------------------------------------- TEST RUN ----------------------------------------------

        log.info("Checking that services are all healthy")
        const validationResult = await validateDataStoreServiceIsHealthy(enclaveContext, SERVICE_ID, PORT_ID);
        if (validationResult.isErr()) {
            throw err(new Error(`Error validating that service '${SERVICE_ID}' is healthy`))
        }

        log.info("Validated that all services are healthy")

        // Check that the file got uploaded on the service
        log.info("Checking that the file got uploaded on " + SERVICE_ID)
        const serviceCtxResult = await enclaveContext.getServiceContext(SERVICE_ID)
        if (serviceCtxResult.isErr()) {
            throw err(new Error("Unexpected Error Creating Service Context"))
        }
        const serviceCtx = serviceCtxResult.value
        const execResult =  await serviceCtx.execCommand(["ls", PATH_TO_CHECK_FOR_UPLOADED_FILE])
        if (execResult.isErr()) {
            throw err(new Error(`Unexpected err running verification on upload file on "${SERVICE_ID}"`))
        }
        const  [exitCode, _] = execResult.value
        if (exitCode !== 0) {
            throw err(new Error(`Expected exit code to be 0 got ${exitCode}`))
        }

    }
    finally {
            stopEnclaveFunction()
    }

    jest.clearAllTimers()
})

async function validateDataStoreServiceIsHealthy(enclaveContext : EnclaveContext, serviceId: string, portId: string): Promise<Result<null, Error>> {
    const getServiceContextResult = await enclaveContext.getServiceContext(serviceId)
    if (getServiceContextResult.isErr()) {
        log.error(`An error occurred getting the service context for service '${serviceId}'; this indicates that the module says it created a service that it actually didn't`)
        throw getServiceContextResult.error
    }
    const serviceContext = getServiceContextResult.value
    const ipAddr = serviceContext.getMaybePublicIPAddress()
    const publicPort: undefined | PortSpec = serviceContext.getPublicPorts().get(portId)

    if (publicPort === undefined) {
        throw new Error(`Expected to find public port '${portId}' on datastore service '${serviceId}', but none was found`)
    }

    const {
        client: datastoreClient,
        clientCloseFunction: datastoreClientCloseFunction
    } = createDatastoreClient(ipAddr, publicPort.number);

    try {
        const waitForHealthyResult = await waitForHealthy(datastoreClient, WAIT_FOR_STARTUP_MAX_POLLS, MILLIS_BETWEEN_AVAILABILITY_RETRIES);
        if (waitForHealthyResult.isErr()) {
            log.error(`An error occurred waiting for the datastore service '${serviceId}' to become available`);
            throw waitForHealthyResult.error
        }
    } finally {
        datastoreClientCloseFunction()
    }

    return ok(null)
}