import {createEnclave} from "../../test_helpers/enclave_setup";
import log from "loglevel";

const STARLARK_EXEC_TEST = "starlark_exec_test"
const IS_PARTITIONING_ENABLED = false
const DEFAULT_DRY_RUN = false
const EMPTY_ARGS = "{}"

const STARLARK_SCRIPT =`
def run(plan):
	service_config = struct(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(service_id = "web-server", config = service_config)
	response = plan.exec(
	struct(
	    service_id = "web-server",
	    command = ["echo", "hello", "world"]
	))
	plan.assert(response["code"], "==", 0)
	plan.assert(response["output"], "==", "hello world\\n")
`

jest.setTimeout(180000)

test("Test Starlark Exec", TestAddServiceWithEmptyAndWithoutPorts)

async function TestAddServiceWithEmptyAndWithoutPorts() {

    const createEnclaveResult = await createEnclave(STARLARK_EXEC_TEST, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        log.info("Executing Starlark script...")
        const runResult = await enclaveContext.runStarlarkScriptBlocking(STARLARK_SCRIPT, EMPTY_ARGS, DEFAULT_DRY_RUN)
        if (runResult.isErr()) {
            log.error("Unexpected error executing Starlark script");
            throw runResult.error;
        }

        expect(runResult.value.interpretationError).toBeUndefined();
        expect(runResult.value.validationErrors).toEqual([]);
        expect(runResult.value.executionError).toBeUndefined();
        log.info("Successfully ran Starlark script");
    }
    finally {
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
}
