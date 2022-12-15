import {createEnclave} from "../../test_helpers/enclave_setup";
import log from "loglevel";
import {ContainerConfig, ContainerConfigBuilder, PortSpec} from "kurtosis-sdk";
import {Port} from "kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_pb";
import TransportProtocol = Port.TransportProtocol;

const TEST_NAME = "add-service-with-port-spec"
const IS_PARTITIONING_ENABLED = false
const SERVICE_ID = "service-with-port-spec"

const PORT_WITH_APP_PROTOCOL    = "port1"
const PORT_WITHOUT_APP_PROTOCOL = "port2"
const DOCKER_IMAGE = "docker/getting-started:latest"


jest.setTimeout(180000)

test("Test add service with application protocol", TestAddServiceWithApplicationProtocol)

async function TestAddServiceWithApplicationProtocol () {
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)
    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        let usedPorts = new Map<string, PortSpec>()
        usedPorts.set( PORT_WITH_APP_PROTOCOL, new PortSpec(3333, TransportProtocol.UDP, "https"))
        usedPorts.set( PORT_WITHOUT_APP_PROTOCOL, new PortSpec(4444, TransportProtocol.TCP))

        let containerConfigs = new Map<string, ContainerConfig>()
        containerConfigs.set(SERVICE_ID, new ContainerConfigBuilder(DOCKER_IMAGE).withUsedPorts(usedPorts).build())

        const runResult = await enclaveContext.addServices(containerConfigs);
        if (runResult.isErr()) {
            log.error(`Unexpected error while adding serivce with id ${SERVICE_ID}`);
            throw runResult.error;
        }

        const results = runResult.value
        const serviceContext = results[0].get(SERVICE_ID)

        if (typeof serviceContext === "undefined") {
            log.error(`Service with Id ${SERVICE_ID} not found`)
            const error = results[1].get(SERVICE_ID)
            throw error
        }

        const portSpec = serviceContext.getPrivatePorts();

        const portSpecWithAppProtocol = portSpec.get(PORT_WITH_APP_PROTOCOL)
        expect(portSpecWithAppProtocol).toBeDefined()
        if (portSpecWithAppProtocol !== undefined) {
            expect(portSpecWithAppProtocol.transportProtocol).toEqual(TransportProtocol.UDP)
            expect(portSpecWithAppProtocol.number).toEqual(3333)
            expect(portSpecWithAppProtocol.maybeApplicationProtocol).toEqual("https")
        }

        const portSpecWithoutAppProtocol = portSpec.get(PORT_WITHOUT_APP_PROTOCOL)
        expect(portSpecWithoutAppProtocol).toBeDefined()
        if (portSpecWithoutAppProtocol !== undefined) {
            expect(portSpecWithoutAppProtocol.transportProtocol).toEqual(TransportProtocol.TCP)
            expect(portSpecWithoutAppProtocol.number).toEqual(4444)
            expect(portSpecWithoutAppProtocol.maybeApplicationProtocol).toEqual("")
        }
    } finally {
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
}