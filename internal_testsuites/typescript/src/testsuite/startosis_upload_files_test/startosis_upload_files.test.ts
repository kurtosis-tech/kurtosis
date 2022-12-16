import log from "loglevel"
import { err} from "neverthrow";
import * as grpc from "@grpc/grpc-js"

import { createEnclave } from "../../test_helpers/enclave_setup";
import {validateDataStoreServiceIsHealthy} from "../../test_helpers/test_helpers";

const TEST_NAME = "upload-files-test"
const IS_PARTITIONING_ENABLED = false
const DEFAULT_DRY_RUN = false
const EMPTY_ARGS = "{}"

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

DIR_TO_UPLOAD = "github.com/kurtosis-tech/datastore-army-package/src"
PATH_TO_MOUNT_UPLOADED_DIR = "` + PATH_TO_MOUNT_UPLOADED_DIR + `"

def run(args):
    print("Adding service " + DATASTORE_SERVICE_ID + ".")
    
    uploaded_artifact_id = upload_files(DIR_TO_UPLOAD)
    print("Uploaded " + uploaded_artifact_id)
    
    
    config = struct(
        image = DATASTORE_IMAGE,
        ports = {
            DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
        },
        files = {
            PATH_TO_MOUNT_UPLOADED_DIR: uploaded_artifact_id
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
        const runResult = await enclaveContext.runStarlarkScriptBlocking(STARTOSIS_SCRIPT, EMPTY_ARGS, DEFAULT_DRY_RUN)
        if (runResult.isErr()) {
            log.error("An error occurred executing the Startosis SCript")
            throw runResult.error
        }

        expect(runResult.value.interpretationError).toBeUndefined()
        expect(runResult.value.validationErrors).toEqual([])
        expect(runResult.value.executionError).toBeUndefined()

        const expectedScriptRegexPattern = `Adding service example-datastore-server-1.
Files uploaded with artifact ID '[a-f0-9-]{36}'
Uploaded [a-f0-9-]{36}
Service 'example-datastore-server-1' added with service GUID '[a-z-0-9]+'
`
        const expectedScriptRegex = new RegExp(expectedScriptRegexPattern)
        expect(runResult.value.runOutput).toMatch(expectedScriptRegex)

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