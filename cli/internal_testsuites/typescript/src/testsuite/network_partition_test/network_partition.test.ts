import { AddPersonArgs, IncrementBooksReadArgs } from "example-api-server-api-lib";
import { PartitionConnection } from "kurtosis-core-sdk/build/lib/enclaves/partition_connection";
import { 
    EnclaveContext, 
    PartitionID, 
    ServiceID, 
    BlockedPartitionConnection, 
    UnblockedPartitionConnection, 
} from "kurtosis-core-sdk";
import log from "loglevel"
import { Result, ok, err } from "neverthrow"
import * as grpc from "@grpc/grpc-js"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

import { createEnclave } from "../../test_helpers/enclave_setup";
import { addAPIService, addAPIServiceToPartition, addDatastoreService } from "../../test_helpers/test_helpers";

const TEST_NAME = "network-partition";
const IS_PARTITIONING_ENABLED = true;

const API_PARTITION_ID: PartitionID = "api";
const DATASTORE_PARTITION_ID: PartitionID = "datastore";

const DATASTORE_SERVICE_ID: ServiceID   = "datastore";
const API1_SERVICE_ID: ServiceID   = "api1";
const API2_SERVICE_ID: ServiceID   = "api2";

const TEST_PERSON_ID = "46";

const SECONDS_TO_WAIT_BEFORE_CALLING_GRPC_CALL_TIMED_OUT = 2;

jest.setTimeout(180000)
test("Test network partition", async () => {
     // ------------------------------------- ENGINE SETUP ----------------------------------------------
     const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

     if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
 
     const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value
 
     try {

        // ------------------------------------- TEST SETUP ----------------------------------------------

        const addDatastoreServiceResult = await addDatastoreService(DATASTORE_SERVICE_ID, enclaveContext)

        if(addDatastoreServiceResult.isErr()) { 
            log.error("An error occurred adding the datastore service")
            throw addDatastoreServiceResult.error
         }

        const { 
            serviceContext: datastoreServiceContext, 
            clientCloseFunction:datastoreClientCloseFunction 
        } = addDatastoreServiceResult.value

        try {

            const addAPIServiceResult = await addAPIService(API1_SERVICE_ID, enclaveContext, datastoreServiceContext.getPrivateIPAddress())
            if(addAPIServiceResult.isErr()){
                log.error("An error occurred adding the first API service") 
                throw addAPIServiceResult.error 
            }

            const { client: apiClient1, clientCloseFunction: apiClient1CloseFunction } = addAPIServiceResult.value;

            try {

                const addPersonArgs = new AddPersonArgs()
                addPersonArgs.setPersonId(TEST_PERSON_ID)

                const addPerson1ResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                    apiClient1.addPerson(addPersonArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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

                const addPerson1Result = await addPerson1ResultPromise;
                if(addPerson1Result.isErr()) {
                    log.error(`An error occurred adding test person with ID ${TEST_PERSON_ID}`)
                    throw addPerson1Result.error;
                }

                const incrementBooksReadArgs = new IncrementBooksReadArgs()
                incrementBooksReadArgs.setPersonId(TEST_PERSON_ID)
            
                const incrementBooksReadResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                    apiClient1.incrementBooksRead(incrementBooksReadArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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

                const incrementBooksReadResult = await incrementBooksReadResultPromise;
                if(incrementBooksReadResult.isErr()) { 
                    log.error("An error occurred test person's books read in preparation for the test")
                    throw incrementBooksReadResult.error;
                }

                // ------------------------------------- TEST RUN ----------------------------------------------
                
                log.info("Partitioning API and datastore services off from each other...")
                const repartitionNetworkResult =  await repartitionNetwork(enclaveContext, true, false)
                if(repartitionNetworkResult.isErr()){
                    log.error("An error occurred repartitioning the network to block access between API <-> datastore")
                    throw repartitionNetworkResult.error
                }
                
                log.info("Repartition complete")

                log.info("Incrementing books read via API 1 while partition is in place, to verify no comms are possible...")

                {
                    const incrementBooksReadResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                        const incrementBooksDeadline = new Date();
                        incrementBooksDeadline.setSeconds(incrementBooksDeadline.getSeconds() + SECONDS_TO_WAIT_BEFORE_CALLING_GRPC_CALL_TIMED_OUT)
                        apiClient1.incrementBooksRead(incrementBooksReadArgs, { deadline: incrementBooksDeadline }, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
                    const incrementBooksReadResult = await incrementBooksReadResultPromise;
                    if(incrementBooksReadResult.isOk()){
                        log.error("Expected the book increment call via API 1 to fail due to the network " +
                        "partition between API and datastore services, but no error was thrown")
                        throw incrementBooksReadResult.value
                    }
                    log.info(`Incrementing books read via API 1 threw the following error as expected due to network partition: ${incrementBooksReadResult.error}`)
                }
                    
                // Adding another API service while the partition is in place ensures that partitioning works even when you add a node
                log.info("Adding second API container, to ensure adding a service under partition works...")

                const addAPIServiceToPartitionResult = await addAPIServiceToPartition(API2_SERVICE_ID, enclaveContext, datastoreServiceContext.getPrivateIPAddress(), API_PARTITION_ID)

                if(addAPIServiceToPartitionResult.isErr()){
                    log.error("An error occurred adding the second API service to the network")
                    throw addAPIServiceToPartitionResult.error
                }

                const { client: apiClient2, clientCloseFunction: apiClient2CloseFunction } = addAPIServiceToPartitionResult.value

                try {

                    log.info("Second API container added successfully")

                    log.info("Incrementing books read via API 2 while partition is in place, to verify no comms are possible...")

                    const incrementBooksReadResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                        const incrementBooksDeadline = new Date();
                        incrementBooksDeadline.setSeconds(incrementBooksDeadline.getSeconds() + SECONDS_TO_WAIT_BEFORE_CALLING_GRPC_CALL_TIMED_OUT)
                        apiClient2.incrementBooksRead(incrementBooksReadArgs, { deadline: incrementBooksDeadline },  (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
                    
                    const incrementBooksReadResult = await incrementBooksReadResultPromise;
                    if(incrementBooksReadResult.isOk()) { 
                        log.error("Expected the book increment call via API 2 to fail due to the network " +
                        "partition between API and datastore services, but no error was thrown")
                        throw incrementBooksReadResult.value
                    }
                    
                    log.info(`Incrementing books read via API 2 threw the following error as expected due to network partition: ${incrementBooksReadResult.error}`)

                    // Now, open the network back up
                    log.info("Repartitioning to heal partition between API and datastore...")

                    const repartitionNetworkResult = await repartitionNetwork(enclaveContext, false, true)
                    if(repartitionNetworkResult.isErr()){
                        log.error("An error occurred healing the partition")
                        throw repartitionNetworkResult.error
                    }

                    log.info("Partition healed successfully")

                    log.info("Making another call via API 1 to increment books read, to ensure the partition is open...")
                    // Use infinite timeout because we expect the partition healing to fix the issue
                    
                    {
                        const incrementBooksReadResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                            apiClient1.incrementBooksRead(incrementBooksReadArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
                        
                        const incrementBooksReadResult = await incrementBooksReadResultPromise;
                        if(incrementBooksReadResult.isErr()){
                            log.error("An error occurred incrementing the number of books read via API 1, even though the partition should have been "+
                            "healed by the goroutine")
                            throw incrementBooksReadResult.error
                        }
                    }
                    log.info("Successfully incremented books read via API 1, indicating that the partition has healed successfully!")

                    log.info("Making another call via API 2 to increment books read, to ensure the partition is open...")
                    // Use infinite timeout because we expect the partition healing to fix the issue

                    {
                        const incrementBooksReadResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                            apiClient2.incrementBooksRead(incrementBooksReadArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
                        
                        const incrementBooksReadResult = await incrementBooksReadResultPromise;
                        if(incrementBooksReadResult.isErr()){
                            log.error("An error occurred incrementing the number of books read via API 2, even though the partition should have been "+
                            "healed by the goroutine")
                            throw incrementBooksReadResult.error
                        }
                    }

                    log.info("Successfully incremented books read via API 2, indicating that the partition has healed successfully!")

                }finally{
                    apiClient2CloseFunction()
                }
            }finally{
                apiClient1CloseFunction()
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

async function repartitionNetwork( 
    enclaveContext: EnclaveContext, 
    isConnectionBlocked: boolean, 
    isApi2ServiceAddedYet: boolean 
    ): Promise<Result<null, Error>> {
    
    const apiPartitionServiceIds = new Set<ServiceID>()
    apiPartitionServiceIds.add(API1_SERVICE_ID)

    if(isApi2ServiceAddedYet){ apiPartitionServiceIds.add(API2_SERVICE_ID) }

    const connectionBetweenPartitions: PartitionConnection = 
        isConnectionBlocked ? new BlockedPartitionConnection() : new UnblockedPartitionConnection();
    
    const partitionServices = new Map<PartitionID, Set<ServiceID>>();
    const datastorePartitionServiceIds = new Set<ServiceID>();

    datastorePartitionServiceIds.add(DATASTORE_SERVICE_ID)

    partitionServices.set(API_PARTITION_ID, apiPartitionServiceIds)
    partitionServices.set(DATASTORE_PARTITION_ID, datastorePartitionServiceIds)

    const partitionConnections = new Map<PartitionID, Map<PartitionID, PartitionConnection>>();
    const apiPartitionConnections = new Map<PartitionID, PartitionConnection>();

    apiPartitionConnections.set(DATASTORE_PARTITION_ID,connectionBetweenPartitions);
    
    partitionConnections.set(API_PARTITION_ID, apiPartitionConnections);

    const defaultPartitionConnection = new UnblockedPartitionConnection()

    const repartitionNetworkResult = await enclaveContext.repartitionNetwork(partitionServices, partitionConnections, defaultPartitionConnection);

    if(repartitionNetworkResult.isErr()){
        log.error(`An error occurred repartitioning the network with isConnectionBlocked = ${isConnectionBlocked}`)
        return err(repartitionNetworkResult.error)
    }

    return ok(null)
}