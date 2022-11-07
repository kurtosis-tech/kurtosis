import { ModuleContext, ModuleID, ServiceID } from "kurtosis-sdk"
import log from "loglevel"
import { err, ok, Result } from "neverthrow";

import { createEnclave } from "../../test_helpers/enclave_setup";
import {
    validateDataStoreServiceIsHealthy,
} from "../../test_helpers/test_helpers";

const TEST_NAME = "module"
const IS_PARTITIONING_ENABLED = false

const TEST_MODULE_IMAGE = "kurtosistech/datastore-army-module:0.2.13"

const DATASTORE_ARMY_MODULE_ID:ModuleID = "datastore-army"

const NUM_MODULE_EXECUTE_CALLS = 2

jest.setTimeout(180000)

test("Test module", async () => {
     // ------------------------------------- ENGINE SETUP ----------------------------------------------
     const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

     if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
 
     const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value
 
     try {
        // ------------------------------------- TEST SETUP ----------------------------------------------
        log.info("Loading module...")
        const loadModuleResult = await enclaveContext.loadModule(DATASTORE_ARMY_MODULE_ID, TEST_MODULE_IMAGE, "{}")

        if(loadModuleResult.isErr()) {
            log.error("An error occurred adding the datastore army module")
            throw loadModuleResult.error
        }
        const moduleContext = loadModuleResult.value

        log.info("Module loaded successfully")

        // ------------------------------------- TEST RUN ----------------------------------------------

        const serviceIdsToPortIds = new Map<ServiceID, string>()

        for (let index = 0; index < NUM_MODULE_EXECUTE_CALLS; index++) {
            log.info("Adding two datastore services via the module...")
            const addTwoDatastoreServicesResult = await addTwoDatastoreServices(moduleContext)
            if(addTwoDatastoreServicesResult.isErr()){
                log.error("An error occurred adding two datastore services via the module")
                throw addTwoDatastoreServicesResult.error
            }
            const createdServiceIdsToPortIds = addTwoDatastoreServicesResult.value
            for(const [serviceId, portId] of createdServiceIdsToPortIds) {
                serviceIdsToPortIds.set(serviceId, portId)
            }
            log.info("Successfully added two datastore services via the module")
        }

        // Sanity-check that the datastore services that the module created are functional
        log.info(`Sanity-checking that all ${serviceIdsToPortIds.size} datastore services added via the module work as expected...`)

        for( const [serviceId, portId] of serviceIdsToPortIds) {
            const validationResult = await validateDataStoreServiceIsHealthy(enclaveContext, serviceId, portId);
            if (validationResult.isErr()) {
                throw err(new Error(`Error validating that service '${serviceId}' is healthy`))
            }
        }
        log.info("All services added via the module work as expected")

        log.info(`Unloading module "${DATASTORE_ARMY_MODULE_ID}"...`)

        const unloadModauleResult = await enclaveContext.unloadModule(DATASTORE_ARMY_MODULE_ID)
        if(unloadModauleResult.isErr()){
            log.error(`An error occurred unloading module "${DATASTORE_ARMY_MODULE_ID}"`)
            throw unloadModauleResult.error
        }

        const getModuleContextResult = await enclaveContext.getModuleContext(DATASTORE_ARMY_MODULE_ID)
        if(getModuleContextResult.isOk()){
            log.error(`Getting module context for module "${DATASTORE_ARMY_MODULE_ID}" should throw an error because it should had been unloaded`)
            throw getModuleContextResult.value
        }

        log.info(`Module "${DATASTORE_ARMY_MODULE_ID}" successfully unloaded`)

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
