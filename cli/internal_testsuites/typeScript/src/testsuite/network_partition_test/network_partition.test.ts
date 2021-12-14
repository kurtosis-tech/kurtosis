import { AddPersonArgs, IncrementBooksReadArgs } from "example-api-server-api-lib";
import { EnclaveContext, PartitionID, ServiceID, BlockedPartitionConnection, UnblockedPartitionConnection, PartitionConnections } from "kurtosis-core-api-lib";
import log from "loglevel"
import { Result, ok, err } from "neverthrow"
import * as grpc from "@grpc/grpc-js"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

import { createEnclave } from "../../test_helpers/enclave_setup";
import { addAPIService, addAPIServiceToPartition, addDatastoreService } from "../../test_helpers/test_helpers";
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

            const { client: apiClient1, clientCloseFunction: apiClient1CloseFunction } = addAPIServiceResult.value;

            const addPersonArgs = new AddPersonArgs()
            addPersonArgs.setPersonId(TEST_PERSON_ID)

            const addPerson1ResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                apiClient1.addPerson(addPersonArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
                    apiClient1.incrementBooksRead(incrementBooksReadArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
                const repartitionNetworkResult =  await repartitionNetwork(enclaveContext, true, false)
                if(repartitionNetworkResult.isErr()){
                    log.error("An error occurred repartitioning the network to block access between API <-> datastore")
                    throw repartitionNetworkResult.error
                }
                
                log.info("Repartition complete")

                log.info("Incrementing books read via API 1 while partition is in place, to verify no comms are possible...")

                const newIncrementBooksReadResult = await incrementBooksReadResultPromise;
                if(newIncrementBooksReadResult.isOk()){
                    log.error("Expected the book increment call via API 1 to fail due to the network " +
                    "partition between API and datastore services, but no error was thrown")
                    return err(newIncrementBooksReadResult.value)
                }
                    
                log.info(`Incrementing books read via API 1 threw the following error as expected due to network partition: ${newIncrementBooksReadResult.error}`)
                    
                // Adding another API service while the partition is in place ensures that partitioning works even when you add a node
	            log.info("Adding second API container, to ensure adding a service under partition works...")

                const addAPIServiceToPartitionResult = await addAPIServiceToPartition(API2_SERVICE_ID, enclaveContext, datastoreServiceContext.getPrivateIPAddress(), API_PARTITION_ID)

                if(addAPIServiceToPartitionResult.isErr()){
                    log.error("An error occurred adding the second API service to the network")
                    throw addAPIServiceToPartitionResult.error
                }

                const { client: apiClient2, clientCloseFunction: apiClient2CloseFunction } = addAPIServiceToPartitionResult.value

            	log.info("Second API container added successfully")

                log.info("Incrementing books read via API 2 while partition is in place, to verify no comms are possible...")

                try {

                    const incrementBooksRead2ResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                        apiClient2.incrementBooksRead(incrementBooksReadArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
    
                    const incrementBooksRead2Result = await incrementBooksRead2ResultPromise;
                    if(incrementBooksRead2Result.isOk()) { 
                        log.error("Expected the book increment call via API 2 to fail due to the network " +
                        "partition between API and datastore services, but no error was thrown")
                        throw incrementBooksRead2Result.value
                    }

                    log.info(`Incrementing books read via API 2 threw the following error as expected due to network partition: ${incrementBooksRead2Result.error}`)


                    // Now, open the network back up
                    log.info("Repartitioning to heal partition between API and datastore...")

                    const repartition2NetworkResult = await repartitionNetwork(enclaveContext, false, true)
                    if(repartition2NetworkResult.isErr()){
                        log.error("An error occurred healing the partition")
                        throw repartition2NetworkResult.error
                    }

                    log.info("Partition healed successfully")

                    log.info("Making another call via API 1 to increment books read, to ensure the partition is open...")
                    // Use infinite timeout because we expect the partition healing to fix the issue
                    



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

async function repartitionNetwork( enclaveContext: EnclaveContext,  isConnectionBlocked: boolean, isApi2ServiceAddedYet: boolean ): Promise<Result<null, Error>> {
    
    const apiPartitionServiceIds = new Set<ServiceID>()
    apiPartitionServiceIds.add(API1_SERVICE_ID)

    if(isApi2ServiceAddedYet){
        apiPartitionServiceIds.add(API2_SERVICE_ID)
    }

    const connectionBetweenPartitions: PartitionConnection = isConnectionBlocked ? new BlockedPartitionConnection() : new UnblockedPartitionConnection()
    
    
    const partitionServices = new Map<PartitionID, Set<ServiceID>>()
    const datastoreServiceId = new Set<ServiceID>(DATASTORE_SERVICE_ID)

    partitionServices.set(API_PARTITION_ID, apiPartitionServiceIds)
    partitionServices.set(DATASTORE_PARTITION_ID, datastoreServiceId)

    const partitionConnections = new Map<PartitionID, Map<PartitionID, PartitionConnection>>()
    const datastorePartitionId = new Map<PartitionID, PartitionConnection>([[DATASTORE_PARTITION_ID,connectionBetweenPartitions]]);
    partitionConnections.set(API_PARTITION_ID, datastorePartitionId);

    const defaultPartitionConnection = new UnblockedPartitionConnection()

    const repartitionNetworkResult = await enclaveContext.repartitionNetwork(partitionServices, partitionConnections, defaultPartitionConnection);

    if(repartitionNetworkResult.isErr()){
        log.error(`An error occurred repartitioning the network with isConnectionBlocked = ${isConnectionBlocked}`)
        return err(repartitionNetworkResult.error)
    }

    return ok(null)
}