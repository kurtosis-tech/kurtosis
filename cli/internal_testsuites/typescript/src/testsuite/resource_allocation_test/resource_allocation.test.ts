import {createEnclave} from "../../test_helpers/enclave_setup";
import {ContainerConfig, ContainerConfigBuilder} from "kurtosis-core-api-lib";
import {err, ok, Result} from "neverthrow";
import log from "loglevel";

const TEST_NAME = "resource-allocation-test"
const IS_PARTITIONING_ENABLED = false
const RESOURCE_ALLOC_TEST_IMAGE =  "flashspys/nginx-static"
const TEST_SERVICE_ID = "test"
const TEST_MEMORY_ALLOC_MEGABYTES = 1000 // 10000 megabytes = 1 GB
const TEST_CPU_ALLOC_MILLICPUS = 1000 // 1000 millicpus = 1 CPU
const TEST_INVALID_MEMORY_ALLOC_MEGABYTES = 4 // 6 megabytes is Dockers min, so this should throw error

jest.setTimeout(180000)

test("Test setting resource allocation fields adds service with no error", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const containerConfig = getContainerConfigWithCPUAndMemory()

        const addServiceResult = await enclaveContext.addService(TEST_SERVICE_ID, containerConfig)

        if(addServiceResult.isErr()) {
            log.error(`An error occurred adding the file server service with the cpuAllocationMillicpus=${TEST_CPU_ALLOC_MILLICPUS} and memoryAllocationMegabytes=${TEST_MEMORY_ALLOC_MEGABYTES}`);
            throw addServiceResult.error
        };
    } finally{
        stopEnclaveFunction()
    }
})

test("Test setting invalid memory allocation megabytes returns error", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        const containerConfig = getContainerConfigWithInvalidMemory()

        const addServiceResult = await enclaveContext.addService(TEST_SERVICE_ID, containerConfig)

        if(!addServiceResult.isErr()) {
            log.error(`An error should have occurred with the following invalid memory allocation: ${TEST_INVALID_MEMORY_ALLOC_MEGABYTES}`);
        };
    } finally{
        stopEnclaveFunction()
    }
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
function getContainerConfigWithCPUAndMemory(): ContainerConfig {

    const containerConfig = new ContainerConfigBuilder(RESOURCE_ALLOC_TEST_IMAGE)
        .withCpuAllocationMillicpus(TEST_CPU_ALLOC_MILLICPUS)
        .withMemoryAllocationMegabytes(TEST_MEMORY_ALLOC_MEGABYTES)
        .build()

    return containerConfig
}

function getContainerConfigWithInvalidMemory(): ContainerConfig {
    const containerConfig = new ContainerConfigBuilder(RESOURCE_ALLOC_TEST_IMAGE)
        .withMemoryAllocationMegabytes(TEST_INVALID_MEMORY_ALLOC_MEGABYTES)
        .build()
    return containerConfig
}

function delay(ms: number) {
    return new Promise( resolve => setTimeout(resolve, ms) );
}