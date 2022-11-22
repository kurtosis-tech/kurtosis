import { ModuleContext, ModuleID, ServiceID } from "kurtosis-sdk"
import log from "loglevel"
import { err, ok, Result } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";
import {
    validateDataStoreServiceIsHealthy,
} from "../../test_helpers/test_helpers";

const TEST_NAME = "module"

const REMOTE_MODULE = "github.com/kurtosis-tech/datastore-army-module"
const EXECUTE_PARAMS            = `{"num_datastores": "2"}`
const DATASTORE_SERVICE_0_ID     = "datastore-0"
const DATASTORE_SERVICE_1_ID   = "datastore-1"
const DATASTORE_PORT_ID       = "grpc"
const IS_DRY_RUN = false

const IS_PARTITIONING_ENABLED  = false

jest.setTimeout(180000)

test("Test remote starlark module execution", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value


    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        log.info(`Executing Startosis module: '${REMOTE_MODULE}'`)
        const executeStartosisRemoteModuleResult = await enclaveContext.executeStartosisRemoteModule(REMOTE_MODULE, EXECUTE_PARAMS, IS_DRY_RUN)
        if (executeStartosisRemoteModuleResult.isErr()) {
            log.error("An error occurred executing the Startosis Module")
            throw executeStartosisRemoteModuleResult.error
        }

        const executeStartosisRemoteModuleValue = executeStartosisRemoteModuleResult.value
        const expectedOutput = `Deploying module datastore_army_module with args:
ModuleInput(num_datastores=2)
Adding service datastore-0
Adding service datastore-1
Module datastore_army_module deployed successfully.
`
        if (expectedOutput !== executeStartosisRemoteModuleValue.getSerializedScriptOutput()) {
            throw err(new Error(`Expected output to be match '${expectedOutput} got '${executeStartosisRemoteModuleValue.getSerializedScriptOutput()}'`))
        }

        if (executeStartosisRemoteModuleValue.getInterpretationError() !== "") {
            throw err(new Error(`Expected Empty Interpretation Error got '${executeStartosisRemoteModuleValue.getInterpretationError()}'`))
        }

        if (executeStartosisRemoteModuleValue.getExecutionError() !== "") {
            throw err(new Error(`Expected Empty Execution Error got '${executeStartosisRemoteModuleValue.getExecutionError()}'`))
        }

        if (executeStartosisRemoteModuleValue.getValidationErrorsList().length != 0) {
            throw err(new Error(`Expected Empty Validation Error got '${executeStartosisRemoteModuleValue.getValidationErrorsList()}'`))
        }
        log.info("Successfully ran Startosis Module")

        log.info("Checking that services are all healthy")
        const validationResultDataStore0 = await validateDataStoreServiceIsHealthy(enclaveContext, DATASTORE_SERVICE_0_ID, DATASTORE_PORT_ID);
        if (validationResultDataStore0.isErr()) {
            throw err(new Error(`Error validating that service '${DATASTORE_SERVICE_0_ID}' is healthy`))
        }


        const validationResultDataStore1 = await validateDataStoreServiceIsHealthy(enclaveContext, DATASTORE_SERVICE_1_ID, DATASTORE_PORT_ID);
        if (validationResultDataStore1.isErr()) {
            throw err(new Error(`Error validating that service '${DATASTORE_SERVICE_1_ID}' is healthy`))
        }

        log.info("All services added via the module work as expected")

    }finally{
        stopEnclaveFunction()
    }

    jest.clearAllTimers()
})

async function addTwoDatastoreServices(moduleContext: ModuleContext):Promise<Result<Map<ServiceID,string>, Error>>{
    const paramsJsonStr = `{"numDatastores": 2}`

    const executeResult = await moduleContext.execute(paramsJsonStr);
    if(executeResult.isErr()) {
        log.error("An error occurred executing the datastore army module")
        return err(executeResult.error)
    }
    const respJsonStr = executeResult.value

    let parsedResult:{
        createdServiceIdsToPortIds: {
            [property:string]: string
        }
    }
    try {
        parsedResult = JSON.parse(respJsonStr)
    }catch(error){
        log.error("An error occurred deserializing the module response")
        if(error instanceof Error){
            return err(error)
        }else{
            return err(new Error("Encountered error while writing the file, but the error wasn't of type Error"))
        }
    }

    const result = new Map<ServiceID,string>()

    const createdServiceIdsToPortIds = parsedResult.createdServiceIdsToPortIds

    for (const createdServiceIdStr in createdServiceIdsToPortIds) {
        const createdServicePortId = createdServiceIdsToPortIds[createdServiceIdStr]
        result.set(createdServiceIdStr, createdServicePortId)
    }

    return ok(result)
}
