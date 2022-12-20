import log from "loglevel"
import { err, ok, Result } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";
import {
    validateDataStoreServiceIsHealthy,
} from "../../test_helpers/test_helpers";

const TEST_NAME = "module"

const REMOTE_PACKAGE = "github.com/kurtosis-tech/datastore-army-package"
const EXECUTE_PARAMS            = `{"num_datastores": 2}`
const DATASTORE_SERVICE_0_ID     = "datastore-0"
const DATASTORE_SERVICE_1_ID   = "datastore-1"
const DATASTORE_PORT_ID       = "grpc"
const IS_DRY_RUN = false

const IS_PARTITIONING_ENABLED  = false

jest.setTimeout(180000)

test.skip("Test remote Starlark package execution", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value


    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        log.info(`Executing Starlark package: '${REMOTE_PACKAGE}'`)
        const runResult = await enclaveContext.runStarlarkRemotePackageBlocking(REMOTE_PACKAGE, EXECUTE_PARAMS, IS_DRY_RUN)
        if (runResult.isErr()) {
            log.error("An error occurred executing the Starlark Package")
            throw runResult.error
        }

        expect(runResult.value.interpretationError).toBeUndefined()
        expect(runResult.value.validationErrors).toEqual([])
        expect(runResult.value.executionError).toBeUndefined()
        log.info("Successfully ran Starlark Package")

        log.info("Checking that services are all healthy")
        const validationResultDataStore0 = await validateDataStoreServiceIsHealthy(enclaveContext, DATASTORE_SERVICE_0_ID, DATASTORE_PORT_ID);
        if (validationResultDataStore0.isErr()) {
            throw err(new Error(`Error validating that service '${DATASTORE_SERVICE_0_ID}' is healthy`))
        }


        const validationResultDataStore1 = await validateDataStoreServiceIsHealthy(enclaveContext, DATASTORE_SERVICE_1_ID, DATASTORE_PORT_ID);
        if (validationResultDataStore1.isErr()) {
            throw err(new Error(`Error validating that service '${DATASTORE_SERVICE_1_ID}' is healthy`))
        }
        log.info("All services added via the package work as expected")
    }finally{
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
})
