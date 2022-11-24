import log from "loglevel"
import { err} from "neverthrow";
import * as grpc from "@grpc/grpc-js"

import { createEnclave } from "../../test_helpers/enclave_setup";
import {validateDataStoreServiceIsHealthy} from "../../test_helpers/test_helpers";
import {generateScriptOutput, readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";

const TEST_NAME = "upload-files-test"
const IS_PARTITIONING_ENABLED = false
const DEFAULT_DRY_RUN = false

const SERVICE_ID = "example-datastore-server-1"
const PORT_ID = "grpc"

const PATH_TO_MOUNT_UPLOADED_DIR = "/uploads"
const PATH_TO_CHECK_FOR_UPLOADED_FILE = "/uploads/helpers.star"

const STARTOSIS_SCRIPT = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + SERVICE_ID + `"
DATASTORE_PORT_ID = "` + PORT_ID + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

DIR_TO_UPLOAD = "github.com/kurtosis-tech/datastore-army-module/src"
PATH_TO_MOUNT_UPLOADED_DIR = "` + PATH_TO_MOUNT_UPLOADED_DIR + `"

print("Adding service " + DATASTORE_SERVICE_ID + ".")

uploaded_artifact_id = upload_files(DIR_TO_UPLOAD)
print("Uploaded " + uploaded_artifact_id)


config = struct(
    image = DATASTORE_IMAGE,
    ports = {
        DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
    },
	files = {
		uploaded_artifact_id: PATH_TO_MOUNT_UPLOADED_DIR
	}
)

add_service(service_id = DATASTORE_SERVICE_ID, config = config)`

jest.setTimeout(180000)

test("Test upload files startosis", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        log.info("Loading module...")
        const outputStream = await enclaveContext.executeKurtosisScript(STARTOSIS_SCRIPT, DEFAULT_DRY_RUN)
        if (outputStream.isErr()) {
            log.error("An error occurred executing the Startosis SCript")
            throw outputStream.error
        }
        const [interpretationError, validationErrors, executionError, instructions] = await readStreamContentUntilClosed(outputStream.value);

        const expectedScriptRegexPattern = `Adding service example-datastore-server-1.
Uploaded [a-f0-9-]{36}
`
        const expectedScriptRegex = new RegExp(expectedScriptRegexPattern)

        expect(generateScriptOutput(instructions)).toMatch(expectedScriptRegex)

        expect(interpretationError).toBeUndefined()
        expect(validationErrors).toEqual([])
        expect(executionError).toBeUndefined()
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