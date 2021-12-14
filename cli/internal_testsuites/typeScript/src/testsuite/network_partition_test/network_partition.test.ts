import { AddPersonArgs, IncrementBooksReadArgs } from "example-api-server-api-lib";
import { EnclaveContext, PartitionID, ServiceID, BlockedPartitionConnection, UnblockedPartitionConnection } from "kurtosis-core-api-lib";
import log from "loglevel"
import { Result, ok, err } from "neverthrow"
import * as grpc from "@grpc/grpc-js"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

import { createEnclave } from "../../test_helpers/enclave_setup";
import { addAPIService, addDatastoreService } from "../../test_helpers/test_helpers";
import { PartitionConnection } from "kurtosis-core-api-lib/build/lib/enclaves/partition_connection";

const TEST_NAME = "network-partition-test";
const IS_PARTITIONING_ENABLED = true;

const API_PARTITION_ID: PartitionID = "api";
const DATASTORE_PARTITION_ID: PartitionID = "datastore";

const DATASTORE_SERVICE_ID: ServiceID   = "datastore";
const API1_SERVICE_ID: ServiceID   = "api1";
const API2_SERVICE_ID: ServiceID   = "api2";

const TEST_PERSON_ID = "46";

const MILLISECONDS_IN_A_SECOND = 1000;
const CONTEXT_TIME_OUT = 2 * MILLISECONDS_IN_A_SECOND;

jest.setTimeout(30000)

test("Test network partition", async () => {
     // ------------------------------------- ENGINE SETUP ----------------------------------------------
     const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

     if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
 
     const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value
 
     try {

        // ------------------------------------- TEST SETUP ----------------------------------------------

        const addDatastoreServiceResult = await addDatastoreService(DATASTORE_SERVICE_ID, enclaveContext)

        if(addDatastoreServiceResult.isErr()) { throw addDatastoreServiceResult.error }

        const { 
            serviceContext: datastoreServiceContext, 
            clientCloseFunction:datastoreClientCloseFunction 
        } = addDatastoreServiceResult.value

        try {

            const addAPIServiceResult = await addAPIService(API1_SERVICE_ID, enclaveContext, datastoreServiceContext.getPrivateIPAddress())
            if(addAPIServiceResult.isErr()){ throw addAPIServiceResult.error }

            const { client: api1Client, clientCloseFunction: api1ClientCloseFunction } = addAPIServiceResult.value;

            const addPersonArgs = new AddPersonArgs()
            addPersonArgs.setPersonId(TEST_PERSON_ID)

            const addPerson1ResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                api1Client.addPerson(addPersonArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
                    if (error === null) {
                        if (!response) {
                            resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                        } else {
                            resolve(ok(response!));
                        }
                    } else {
                        resolve(err(error));
                    }
                })
            })

            try {
                const addPerson1Result = await addPerson1ResultPromise;
                if(addPerson1Result.isErr()) {
                    log.error(`An error occurred adding test person with ID ${TEST_PERSON_ID}`)
                    throw addPerson1Result.error;
                }

                const incrementBooksReadArgs = new IncrementBooksReadArgs()
                incrementBooksReadArgs.setPersonId(TEST_PERSON_ID)
            
                const incrementBooksReadResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                    api1Client.incrementBooksRead(incrementBooksReadArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
                        if (error === null) {
                            if (!response) {
                                resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                            } else {
                                resolve(ok(response!));
                            }
                        } else {
                            resolve(err(error));
                        }
                    })
                })

                const incrementBooksReadResult = await incrementBooksReadResultPromise;
                if(incrementBooksReadResult.isErr()) { 
                    log.error("An error occurred test person's books read in preparation for the test")
                    throw incrementBooksReadResult.error;
                }

                // ------------------------------------- TEST RUN ----------------------------------------------
                
                log.info("Partitioning API and datastore services off from each other...")
                

            }finally{
                api1ClientCloseFunction()
            }
        }finally{
            datastoreClientCloseFunction()
        }
     }finally{
         stopEnclaveFunction()
     }

     jest.clearAllTimers()
})

/*
Creates a repartitioner that will partition the network between the API & datastore services, with the connection between them configurable
*/

function repartitionNetwork( enclaveContext: EnclaveContext,  isConnectionBlocked: boolean, isApi2ServiceAddedYet: boolean ): Result<null, Error> {
    
    const apiPartitionServiceIds = new Map<ServiceID, boolean>()
    apiPartitionServiceIds.set(API1_SERVICE_ID, true)

    if(isApi2ServiceAddedYet){
        apiPartitionServiceIds.set(API2_SERVICE_ID, true)
    }

    const connectionBetweenPartitions: PartitionConnection = isConnectionBlocked ? new BlockedPartitionConnection() : new UnblockedPartitionConnection()
    
    const datastoreServices = new Map<ServiceID, boolean>()
    datastoreServices.set(DATASTORE_SERVICE_ID, true)

    const partitionServices = new Map<PartitionID, Map<ServiceID, boolean>>()
    partitionServices.set(API_PARTITION_ID, apiPartitionServiceIds)
    partitionServices.set(DATASTORE_PARTITION_ID, datastoreServices)

    

    return ok(null)
}