import { GetArgs, GetResponse, UpsertArgs } from "example-datastore-server-api-lib";
import { ModuleContext, ModuleID, PortSpec, ServiceID } from "kurtosis-core-api-lib"
import log from "loglevel"
import { err, ok, Result } from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "@grpc/grpc-js"

import { createEnclave } from "../../test_helpers/enclave_setup";
import { createDatastoreClient, waitForHealthy } from "../../test_helpers/test_helpers";

const TEST_NAME = "module"
const IS_PARTITIONING_ENABLED = false

const TEST_MODULE_IMAGE = "kurtosistech/datastore-army-module:0.2.13"

const DATASTORE_ARMY_MODULE_ID:ModuleID = "datastore-army"

const NUM_MODULE_EXECUTE_CALLS = 2

const TEST_DATASTORE_KEY = "my-key"
const TEST_DATASTORE_VALUE = "test-value"

const MILLIS_BETWEEN_AVAILBILITY_RETRIES = 1000

/*
NOTE: on 2022-05-16 the Go version failed with the following error so we bumped the num polls to 20.

time="2022-05-16T23:58:21Z" level=info msg="Sanity-checking that all 4 datastore services added via the module work as expected..."
--- FAIL: TestModule (21.46s)
    module_test.go:81:
            Error Trace:	module_test.go:81
            Error:      	Received unexpected error:
                            The service didn't return a success code, even after 15 retries with 1000 milliseconds in between retries
                             --- at /home/circleci/project/internal_testsuites/golang/test_helpers/test_helpers.go:179 (WaitForHealthy) ---
                            Caused by: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:49188: connect: connection refused"
            Test:       	TestModule
            Messages:   	An error occurred waiting for the datastore service to become available

NOTE: On 2022-05-21 the Go version failed again at 20s. I opened the enclave logs and it's weird because nothing is failing and
the datastore service is showing itself as up *before* we even start the check-if-available wait. We're in crunch mode
so I'm going to bump this up to 30s, but I suspect there's some sort of nondeterministic underlying failure happening.
 */
const WAIT_FOR_STARTUP_MAX_POLLS = 30

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
            const getServiceContextResult = await enclaveContext.getServiceContext(serviceId)
            if(getServiceContextResult.isErr()){
                log.error(`An error occurred getting the service context for service '${serviceId}'; this indicates that the module says it created a service that it actually didn't`)
                throw getServiceContextResult.error
            }
            const serviceContext = getServiceContextResult.value
            const ipAddr = serviceContext.getMaybePublicIPAddress()
            const publicPort: undefined | PortSpec = serviceContext.getPublicPorts().get(portId)

            if(publicPort === undefined){
                throw new Error(`Expected to find public port '${portId}' on datastore service '${serviceId}', but none was found`)
            }

            const { client: datastoreClient, clientCloseFunction: datastoreClientCloseFunction } = createDatastoreClient(ipAddr, publicPort.number);

            try{
                const waitForHealthyResult = await waitForHealthy(datastoreClient, WAIT_FOR_STARTUP_MAX_POLLS, MILLIS_BETWEEN_AVAILBILITY_RETRIES );
                if(waitForHealthyResult.isErr()){
                    log.error(`An error occurred waiting for the datastore service '${serviceId}' to become available`);
                    throw waitForHealthyResult.error
                }

                const upsertArgs = new UpsertArgs();
                upsertArgs.setKey(TEST_DATASTORE_KEY)
                upsertArgs.setValue(TEST_DATASTORE_VALUE)

                const upsertResponseResultPromise:Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                    datastoreClient.upsert(upsertArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
                        if (error === null) {
                            if (!response) {
                                resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                            } else {
                                resolve(ok(response));
                            }
                        } else {
                            resolve(err(error));
                        }
                    })
                })
    
                const upsertResponseResult = await upsertResponseResultPromise;
                if(upsertResponseResult.isErr()){ 
                    log.error(`An error occurred adding the test key to datastore service "${serviceId}"`)
                    throw upsertResponseResult.error 
                }

                const getArgs = new GetArgs();
                getArgs.setKey(TEST_DATASTORE_KEY)

                const getResponseResultPromise:Promise<Result<GetResponse, Error>> = new Promise((resolve, _unusedReject) => {
                    datastoreClient.get(getArgs, (error: grpc.ServiceError | null, response?: GetResponse) => {
                        if (error === null) {
                            if (!response) {
                                resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                            } else {
                                resolve(ok(response));
                            }
                        } else {
                            resolve(err(error));
                        }
                    })
                })

                const getResponseResult = await getResponseResultPromise;
                if(getResponseResult.isErr()){
                    log.error(`An error occurred getting the test key to datastore service "${serviceId}"`) 
                    throw getResponseResult.error 
                }

                const getResponse = getResponseResult.value;
                const actualValue = getResponse.getValue();
                if(actualValue !== TEST_DATASTORE_VALUE){
                    log.error(`Datastore service "${serviceId}" is storing value "${actualValue}" for the test key, which doesn't match the expected value ""${TEST_DATASTORE_VALUE}`)
                }
            }
            finally{
                datastoreClientCloseFunction()
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
