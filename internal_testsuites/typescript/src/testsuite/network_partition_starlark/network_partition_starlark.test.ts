import {createEnclave} from "../../test_helpers/enclave_setup";
import log from "loglevel";

const STARLARK_SUBNETWORK_TEST_NAME = "starlark-subnetwork"
const IS_PARTITIONING_ENABLED = true
const EXECUTE_NO_DRY_RUN = false
const EMPTY_ARGS = "{}"

const SERVICE_ID_1 = "service_1"
const SERVICE_ID_2 = "service_2"

const SUBNETWORK_1 = "subnetwork_1"
const SUBNETWORK_2 = "subnetwork_2"
const SUBNETWORK_3 = "subnetwork_3"

const SUBNETWORK_IN_STARLARK_SCRIPT = `DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"

SERVICE_ID_1 = "${SERVICE_ID_1}"
SERVICE_ID_2 = "${SERVICE_ID_2}"

SUBNETWORK_1 = "${SUBNETWORK_1}"
SUBNETWORK_2 = "${SUBNETWORK_2}"
SUBNETWORK_3 = "${SUBNETWORK_3}"

CONNECTION_SUCCESS = 0
CONNECTION_FAILURE = 1

def run(plan, args):
	# block all connections by default
	plan.set_connection(kurtosis.connection.BLOCKED)

	# adding 2 services to play with, each in their own subnetwork
	service_1 = plan.add_service(
		service_id=SERVICE_ID_1, 
		config=ServiceConfig(
			image=DOCKER_GETTING_STARTED_IMAGE,
			subnetwork=SUBNETWORK_1,
		)
	)
	service_2 = plan.add_service(
		service_id=SERVICE_ID_2, 
		config=ServiceConfig(
			image=DOCKER_GETTING_STARTED_IMAGE,
			subnetwork=SUBNETWORK_2
		)
	)

	# Validate connection is indeed blocked
	connection_result = plan.exec(recipe=struct(
		service_id=SERVICE_ID_2,
		command=["ping", "-W", "1", "-c", "1", service_1.ip_address],
	))
	plan.assert(connection_result["code"], "==", CONNECTION_FAILURE)

	# Allow connection between 1 and 2
	plan.set_connection((SUBNETWORK_1, SUBNETWORK_2), kurtosis.connection.ALLOWED)

	# Connection now works
	connection_result = plan.exec(recipe=struct(
		service_id=SERVICE_ID_2,
		command=["ping", "-W", "1", "-c", "1", service_1.ip_address],
	))
	plan.assert(connection_result["code"], "==", CONNECTION_SUCCESS)

	# Reset connection to default (which is BLOCKED)
	plan.remove_connection((SUBNETWORK_1, SUBNETWORK_2))

	# Connection is back to BLOCKED
	connection_result = plan.exec(recipe=struct(
		service_id=SERVICE_ID_2,
		command=["ping", "-W", "1", "-c", "1", service_1.ip_address],
	))
	plan.assert(connection_result["code"], "==", CONNECTION_FAILURE)

	# Create a third subnetwork connected to SUBNETWORK_1 and add service2 to it
	plan.set_connection((SUBNETWORK_3, SUBNETWORK_1), ConnectionConfig(packet_loss_percentage=0.0))
	plan.update_service(SERVICE_ID_2, config=UpdateServiceConfig(subnetwork=SUBNETWORK_3))
	
	# Service 2 can now talk to Service 1 again!
	connection_result = plan.exec(recipe=struct(
		service_id=SERVICE_ID_2,
		command=["ping", "-W", "1", "-c", "1", service_1.ip_address],
	))
	plan.assert(connection_result["code"], "==", CONNECTION_SUCCESS)

	plan.print("Test successfully executed")
`

jest.setTimeout(180000)

test("Test subnetwork in Starlark", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(STARLARK_SUBNETWORK_TEST_NAME, IS_PARTITIONING_ENABLED)
    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    // ------------------------------------- TEST RUN ----------------------------------------------
    try {
        const runResult = await enclaveContext.runStarlarkScriptBlocking(SUBNETWORK_IN_STARLARK_SCRIPT, EMPTY_ARGS, EXECUTE_NO_DRY_RUN)

        if (runResult.isErr()) {
            log.error("Unexpected error executing Starlark script");
            throw runResult.error;
        }

        expect(runResult.value.interpretationError).toBeUndefined()
        expect(runResult.value.validationErrors).toEqual([])
        expect(runResult.value.executionError).toBeUndefined()
        expect(runResult.value.instructions).toHaveLength(16)
        expect(runResult.value.runOutput).toContain("Test successfully executed");
    } finally {
        stopEnclaveFunction()
    }
    jest.clearAllTimers()
})
