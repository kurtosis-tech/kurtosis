import log from "loglevel"
import { err} from "neverthrow";
import * as grpc from "@grpc/grpc-js"

import { createEnclave } from "../../test_helpers/enclave_setup";
import {validateDataStoreServiceIsHealthy} from "../../test_helpers/test_helpers";

const TEST_NAME = "startosis_remove_service_test"
const IS_PARTITIONING_ENABLED = false
const DEFAULT_DRY_RUN = false
const EMPTY_ARGS = "{}"

const SERVICE_ID = "example-datastore-server-1"
const PORT_ID = "grpc"

const STARLARK_SCRIPT = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + SERVICE_ID + `"
DATASTORE_PORT_ID = "` + PORT_ID + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

def run(plan):
	plan.print("Adding service " + DATASTORE_SERVICE_ID + ".")
	
	config = struct(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
		}
	)
	
	plan.add_service(service_id = DATASTORE_SERVICE_ID, config = config)
	plan.print("Service " + DATASTORE_SERVICE_ID + " deployed successfully.")`

const REMOVE_SCRIPT = `
DATASTORE_SERVICE_ID = "` + SERVICE_ID + `"
def run(plan):
	plan.remove_service(DATASTORE_SERVICE_ID)`

jest.setTimeout(180000)

test("Test remove service", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        // Executing Startosis script to first add the datastore service...
        const runResult = await enclaveContext.runStarlarkScriptBlocking(STARLARK_SCRIPT, EMPTY_ARGS, DEFAULT_DRY_RUN)
        if (runResult.isErr()) {
            log.error("Unexpected error executing Starlark script")
            throw runResult.error
        }

        expect(runResult.value.interpretationError).toBeUndefined()
        expect(runResult.value.validationErrors).toEqual([])
        expect(runResult.value.executionError).toBeUndefined()

        const expectedScriptRegex = new RegExp(`Adding service example-datastore-server-1.
Service 'example-datastore-server-1' added with service GUID '[a-z-0-9]+'
Service example-datastore-server-1 deployed successfully.
`)
        expect(runResult.value.runOutput).toMatch(expectedScriptRegex)

        // Checking that services are all healthy
        const validationResult = await validateDataStoreServiceIsHealthy(enclaveContext, SERVICE_ID, PORT_ID);
        if (validationResult.isErr()) {
            throw err(new Error(`Error validating that service '${SERVICE_ID}' is healthy`))
        }

        // ------------------------------------- TEST RUN ----------------------------------------------

        // we run the remove script and see if things still work
        const removeServiceRunResult = await enclaveContext.runStarlarkScriptBlocking(REMOVE_SCRIPT, EMPTY_ARGS, DEFAULT_DRY_RUN)
        if (removeServiceRunResult.isErr()) {
            log.error("Unexpected error executing Starlark script")
            throw removeServiceRunResult.error
        }

        expect(removeServiceRunResult.value.interpretationError).toBeUndefined()
        expect(removeServiceRunResult.value.validationErrors).toEqual([])
        expect(removeServiceRunResult.value.executionError).toBeUndefined()

        const removeServiceExpectedScriptRegex = new RegExp(`Service 'example-datastore-server-1' with service GUID '[a-z-0-9]+' removed
`)
        expect(removeServiceRunResult.value.runOutput).toMatch(removeServiceExpectedScriptRegex)

        // Ensure that service listing is empty
        const serviceInfos = await enclaveContext.getServices()
        if (serviceInfos.isErr()) {
            throw err(new Error(`Error retrieving service infos: ${serviceInfos.error}`))
        }
        expect(serviceInfos.value).toEqual(new Map())
    }
    finally {
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
})
