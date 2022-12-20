import {createEnclave} from "../../test_helpers/enclave_setup";
import {TransportProtocol } from "kurtosis-sdk";
import log from "loglevel";

const IS_PARTITIONING_ENABLED = false
const DEFAULT_DRY_RUN = false
const EMPTY_ARGS = "{}"
const SERVICE_ID = "docker-getting-started"

const TEST_NAME = "add-service-with-port-spec1"
const PORT_WITHOUT_PROTOCOL = "port1"
const PORT_WITH_PROTOCOL = "port2"

const STARLARK_SCRIPT_WITH_PORT_SPEC_SUCCESS = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "${SERVICE_ID}"

spec = PortSpec(number = 5000, transport_protocol = "UDP")

def run(plan):
    plan.add_service(
        service_id = SERVICE_ID, 
        config = struct(
            image = DOCKER_GETTING_STARTED_IMAGE, 
            ports = {
                "${PORT_WITHOUT_PROTOCOL}": PortSpec(number = 3333),
                "${PORT_WITH_PROTOCOL}": spec
            }
        )
    )
    plan.print("httpd has been added successfully")`

jest.setTimeout(180000)

test("Test add service with optional protocol in port spec", async () => {
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)
    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        const runResult =
            await enclaveContext.runStarlarkScriptBlocking(STARLARK_SCRIPT_WITH_PORT_SPEC_SUCCESS, EMPTY_ARGS, DEFAULT_DRY_RUN)
        if (runResult.isErr()) {
            log.error("Unexpected error executing Starlark script");
            throw runResult.error;
        }

        const result = await enclaveContext.getServiceContext(SERVICE_ID)
        if (result.isErr()) { throw result.error }

        const serviceCtx = result.value
        const ports = serviceCtx.getPrivatePorts()

        const portsWithProtocol = ports.get(PORT_WITH_PROTOCOL)
        expect(portsWithProtocol).toBeDefined()
        if (portsWithProtocol !== undefined) {
            expect(portsWithProtocol.number).toEqual(5000)
            expect(portsWithProtocol.transportProtocol).toEqual(TransportProtocol.UDP)
        }

        const portsWithOutProtocol = ports.get(PORT_WITHOUT_PROTOCOL)
        expect(portsWithOutProtocol).toBeDefined()
        if (portsWithOutProtocol !== undefined) {
            expect(portsWithOutProtocol.number).toEqual(3333)
            expect(portsWithOutProtocol.transportProtocol).toEqual(TransportProtocol.TCP)
        }
    } finally {
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
})
