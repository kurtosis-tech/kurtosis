import {JEST_TIMEOUT_MS} from "../startosis_module_test/shared_constants";
import {createEnclave} from "../../test_helpers/enclave_setup";
import log from "loglevel";
import {generateScriptOutput, readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";
import { Result } from "neverthrow"
import {ServiceID} from "../../../../../api/typescript/src";

const ADD_SERVICE_WITH_EMPTY_PORTS_TEST_NAME = "add-service-empty-ports"
const IS_PARTITIONING_ENABLED = false
const DEFAULT_DRY_RUN = false
const SERVICE_ID = "docker-getting-started"

const STARTOSIS_SCRIPT = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "${SERVICE_ID}"

print("Adding service " + SERVICE_ID + ".")

config = struct(
    image = DOCKER_GETTING_STARTED_IMAGE,
	ports = {}
)

add_service(service_id = SERVICE_ID, config = config)
print("Service " + SERVICE_ID + " deployed successfully.")`

jest.setTimeout(JEST_TIMEOUT_MS)

test("Test add service with empty ports test", TestAddServiceWithEmptyPorts)

async function TestAddServiceWithEmptyPorts() {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(ADD_SERVICE_WITH_EMPTY_PORTS_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        log.info("Executing Starlark script...");
        log.debug(`Starlark script content: \n%v ${STARTOSIS_SCRIPT}`);
        const outputStream = await enclaveContext.executeKurtosisScript(STARTOSIS_SCRIPT, DEFAULT_DRY_RUN);
        if (outputStream.isErr()) {
            log.error("Unexpected error executing Starlark script");
            throw outputStream.error;
        }
        const [interpretationError, validationErrors, executionError, instructions] = await readStreamContentUntilClosed(outputStream.value);

        const expectedScriptOutput: string = `Adding service docker-getting-started.
Service docker-getting-started deployed successfully.
`;

        expect(interpretationError).toBeUndefined();
        expect(validationErrors).toEqual([]);
        expect(executionError).toBeUndefined();
        expect(generateScriptOutput(instructions)).toEqual(expectedScriptOutput);
        log.info("Successfully ran Starlark script");

        // ------------------------------------- TEST RUN ----------------------------------------------

        // Ensure that the service is listed
        const expectedAmountOfServices: number = 1;
        const getServiceIdsPromise: Promise<Result<Set<ServiceID>, Error>> = enclaveContext.getServices();
        const getServiceIdsResult = await getServiceIdsPromise;
        if(getServiceIdsResult.isErr()) {
            log.error(`An error occurred getting service IDs`);
            throw getServiceIdsResult.error;
        }

        const servicesIds: Set<string> = getServiceIdsResult.value;

        const actualAmountOfServices: number = servicesIds.size

        if (expectedAmountOfServices !== actualAmountOfServices) {
            throw new Error(`Expected to receive ${expectedAmountOfServices} services from get services, but ${actualAmountOfServices} were received`)
        }
    }
    finally {
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
}
