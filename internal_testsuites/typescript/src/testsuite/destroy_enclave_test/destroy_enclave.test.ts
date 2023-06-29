import { ContainerConfig, ContainerConfigBuilder, FilesArtifactUUID, PortSpec, TransportProtocol, ServiceName } from "kurtosis-sdk"
import log from "loglevel";
import { Result, ok, err } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";
import {addServiceViaStarlark} from "../../test_helpers/test_helpers";

const TEST_NAME = "destroy-enclave"
const IS_PARTITIONING_ENABLED = false

const FILE_SERVER_SERVICE_IMAGE = "flashspys/nginx-static"
const FILE_SERVER_SERVICE_NAME: ServiceName = "file-server"
const FILE_SERVER_PORT_ID = "http"
const FILE_SERVER_PRIVATE_PORT_NUM = 80
const FILE_SERVER_PRIVATE_PORT_PROTOCOL = "TCP"

jest.setTimeout(180000)

test("Test destroy enclave", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction, destroyEnclaveFunction } = createEnclaveResult.value

    let shouldStopEnclaveAtTheEnd = true

    try {

        // ------------------------------------- TEST SETUP ----------------------------------------------
        const fileServerContainerConfig = getFileServerContainerConfig()

        const addServiceResult = await addServiceViaStarlark(enclaveContext, FILE_SERVER_SERVICE_NAME, fileServerContainerConfig)
        if(addServiceResult.isErr()){ throw addServiceResult.error }

        const serviceContext = addServiceResult.value
        const publicPort = serviceContext.getPublicPorts().get(FILE_SERVER_PORT_ID)
        if(publicPort === undefined){
            throw new Error(`Expected to find public port for ID "${FILE_SERVER_PORT_ID}", but none was found`)
        }

        const fileServerPublicIp = serviceContext.getMaybePublicIPAddress();
        const fileServerPublicPortNum = publicPort.number

        log.info(`Added file server service with public IP "${fileServerPublicIp}" and port "${fileServerPublicPortNum}"`)

        // ------------------------------------- TEST RUN ----------------------------------------------
        const destroyEnclaveResult = await destroyEnclaveFunction()

        if(destroyEnclaveResult.isErr()) {
            log.error(`An error occurred destroying enclave with ID "${enclaveContext.getEnclaveUuid()}"`)
            throw destroyEnclaveResult.error
        }

        shouldStopEnclaveAtTheEnd = false
    }finally{
        if (shouldStopEnclaveAtTheEnd) {
            stopEnclaveFunction()
        }
    }
    jest.clearAllTimers()
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

function getFileServerContainerConfig(): string {
    const portSpec = `{"${FILE_SERVER_PORT_ID}": PortSpec(number=${FILE_SERVER_PRIVATE_PORT_NUM}, transport_protocol="${FILE_SERVER_PRIVATE_PORT_PROTOCOL}")}`
    return `ServiceConfig(image="${FILE_SERVER_SERVICE_IMAGE}", ports=${portSpec})`
}

