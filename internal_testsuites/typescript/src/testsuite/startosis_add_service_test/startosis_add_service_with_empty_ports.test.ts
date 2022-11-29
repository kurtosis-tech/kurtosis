import {createEnclave} from "../../test_helpers/enclave_setup";
import log from "loglevel";
import {readStreamContentUntilClosed} from "../../test_helpers/startosis_helpers";
import { Result } from "neverthrow"
import {ServiceID} from "../../../../../api/typescript/src";

const ADD_SERVICE_WITH_EMPTY_PORTS_TEST_NAME = "add-service-empty-ports"
const IS_PARTITIONING_ENABLED = false
const DEFAULT_DRY_RUN = false
const SERVICE_ID = "docker-getting-started"
const SERVICE_ID_2 = "docker-getting-started-2"

const STARLARK_SCRIPT_WITH_EMPTY_PORTS = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "${SERVICE_ID}"

print("Adding service " + SERVICE_ID + ".")

config = struct(
    image = DOCKER_GETTING_STARTED_IMAGE,
	ports = {}
)

add_service(service_id = SERVICE_ID, config = config)
print("Service " + SERVICE_ID + " deployed successfully.")`

const STARLARK_SCRIPT_WITHOUT_PORTS = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "${SERVICE_ID_2}"

print("Adding service " + SERVICE_ID + ".")

config = struct(
    image = DOCKER_GETTING_STARTED_IMAGE,
)

add_service(service_id = SERVICE_ID, config = config)
print("Service " + SERVICE_ID + " deployed successfully.")`


jest.setTimeout(180000)

test("Test add service with empty and without ports test", TestAddServiceWithEmptyAndWithoutPorts)

async function TestAddServiceWithEmptyAndWithoutPorts() {

    const serviceIds: Array<string> = new Array<string>()
    serviceIds.push(SERVICE_ID)
    serviceIds.push(SERVICE_ID_2)

    const starlarkScriptsToRun: Array<string> = new Array<string>()
    starlarkScriptsToRun.push(STARLARK_SCRIPT_WITH_EMPTY_PORTS)
    starlarkScriptsToRun.push(STARLARK_SCRIPT_WITHOUT_PORTS)

    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(ADD_SERVICE_WITH_EMPTY_PORTS_TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        for (let i=0; i < starlarkScriptsToRun.length; i++) {
            const starlarkScript:string = starlarkScriptsToRun[i]
            const serviceId:string = serviceIds[i]
            log.info("Executing Starlark script...");
            log.debug(`Starlark script content: \n%v ${starlarkScript}`);
            const outputStream = await enclaveContext.runStarlarkScript(starlarkScript, DEFAULT_DRY_RUN);
            if (outputStream.isErr()) {
                log.error("Unexpected error executing Starlark script");
                throw outputStream.error;
            }
            const [scriptOutput, _, interpretationError, validationErrors, executionError] = await readStreamContentUntilClosed(outputStream.value);

            const expectedScriptOutput: string = `Adding service ${serviceId}.
Service ${serviceId} deployed successfully.
`;

            expect(interpretationError).toBeUndefined();
            expect(validationErrors).toEqual([]);
            expect(executionError).toBeUndefined();
            expect(expectedScriptOutput).toEqual(scriptOutput);
            log.info("Successfully ran Starlark script");

            // ------------------------------------- TEST RUN ----------------------------------------------

            // Ensure that the service is listed
            const expectedNumberOfServices: number = i + 1;
            const getServiceIdsPromise: Promise<Result<Set<ServiceID>, Error>> = enclaveContext.getServices();
            const getServiceIdsResult = await getServiceIdsPromise;
            if(getServiceIdsResult.isErr()) {
                log.error(`An error occurred getting service IDs`);
                throw getServiceIdsResult.error;
            }

            const servicesIds: Set<string> = getServiceIdsResult.value;

            const actualNumberOfServices: number = servicesIds.size

            if (expectedNumberOfServices !== actualNumberOfServices) {
                throw new Error(`Expected to receive ${expectedNumberOfServices} services from get services, but ${actualNumberOfServices} were received`)
            }
        }
    }
    finally {
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
}
