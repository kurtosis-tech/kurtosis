import {createEnclave} from "../../test_helpers/enclave_setup";
import log from "loglevel";
import { Result } from "neverthrow"
import {ServiceGUID, ServiceID} from "kurtosis-sdk";

const ADD_SERVICE_WITH_EMPTY_PORTS_TEST_NAME = "add-service-empty-ports"
const IS_PARTITIONING_ENABLED = false
const DEFAULT_DRY_RUN = false
const EMPTY_ARGS = "{}"
const SERVICE_ID = "docker-getting-started"
const SERVICE_ID_2 = "docker-getting-started-2"

const STARLARK_SCRIPT_WITH_EMPTY_PORTS = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "${SERVICE_ID}"

def run(plan):
    plan.print("Adding service " + SERVICE_ID + ".")
    
    config = ServiceConfig(
        image = DOCKER_GETTING_STARTED_IMAGE,
        ports = {}
    )
    
    plan.add_service(service_id = SERVICE_ID, config = config)
    plan.print("Service " + SERVICE_ID + " deployed successfully.")`

const STARLARK_SCRIPT_WITHOUT_PORTS = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "${SERVICE_ID_2}"

def run(plan, args):
    plan.print("Adding service " + SERVICE_ID + ".")
    
    config = ServiceConfig(
        image = DOCKER_GETTING_STARTED_IMAGE,
    )
    
    plan.add_service(service_id = SERVICE_ID, config = config)
    plan.print("Service " + SERVICE_ID + " deployed successfully.")`


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
            const runResult = await enclaveContext.runStarlarkScriptBlocking(starlarkScript, EMPTY_ARGS, DEFAULT_DRY_RUN)
            if (runResult.isErr()) {
                log.error("Unexpected error executing Starlark script");
                throw runResult.error;
            }

            const expectedScriptOutputRegexpPattern = `Adding service ${serviceId}.
Service '${serviceId}' added with service GUID '[a-z-0-9]+'
Service ${serviceId} deployed successfully.
`;
            const expectedScriptOutputRegexp = new RegExp(expectedScriptOutputRegexpPattern)

            expect(runResult.value.interpretationError).toBeUndefined();
            expect(runResult.value.validationErrors).toEqual([]);
            expect(runResult.value.executionError).toBeUndefined();
            expect(runResult.value.runOutput).toMatch(expectedScriptOutputRegexp);
            log.info("Successfully ran Starlark script");

            // ------------------------------------- TEST RUN ----------------------------------------------

            // Ensure that the service is listed
            const expectedNumberOfServices: number = i + 1;
            const getServiceIdsPromise: Promise<Result<Map<ServiceID, ServiceGUID>, Error>> = enclaveContext.getServices();
            const getServiceIdsResult = await getServiceIdsPromise;
            if(getServiceIdsResult.isErr()) {
                log.error(`An error occurred getting service IDs`);
                throw getServiceIdsResult.error;
            }

            const serviceIdMap: Map<ServiceID, ServiceGUID> = getServiceIdsResult.value;

            const actualNumberOfServices: number = serviceIdMap.size

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
