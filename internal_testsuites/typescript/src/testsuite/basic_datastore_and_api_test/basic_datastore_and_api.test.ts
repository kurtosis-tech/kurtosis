import { ServiceContext } from "kurtosis-sdk"
import * as apiServerApi from "example-api-server-api-lib";
import { Result, ok, err } from "neverthrow"
import log from "loglevel"
import * as grpc from "@grpc/grpc-js"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

import { createEnclave } from "../../test_helpers/enclave_setup";
import { addAPIService, addDatastoreService } from "../../test_helpers/test_helpers";

const TEST_NAME = "basic-datastore-and-api";
const IS_PARTITIONING_ENABLED = false;
const DATASTORE_SERVICE_NAME = "datastore";
const API_SERVICE_NAME = "api";
const TEST_PERSON_ID = "23";
const TEST_NUM_BOOKS_READ = 3;

jest.setTimeout(180000)

test("Test basic data store and API", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        log.info("Adding datastore service...")

        const addDatastoreServiceResult = await addDatastoreService(DATASTORE_SERVICE_NAME, enclaveContext)

        if(addDatastoreServiceResult.isErr()) { throw addDatastoreServiceResult.error }

        const { 
            serviceContext: datastoreServiceContext, 
            clientCloseFunction:datastoreClientCloseFunction 
        } = addDatastoreServiceResult.value

        log.info("Added datastore service")

        try {

            log.info("Adding API service...")
            const apiClientServiceResult: Result<{
                serviceContext: ServiceContext;
                client: apiServerApi.ExampleAPIServerServiceClientNode;
                clientCloseFunction: () => void;
            }, Error> = await addAPIService(API_SERVICE_NAME, enclaveContext, datastoreServiceContext.getPrivateIPAddress())
            
            if(apiClientServiceResult.isErr()){ throw apiClientServiceResult.error }

            const { 
                client: apiClient, 
                clientCloseFunction: apiClientCloseFunction  
            } = apiClientServiceResult.value

                log.info("Added API service")
            
            try {
                // ------------------------------------- TEST RUN ----------------------------------------------
                log.info(`Verifying that person with test ID ${TEST_PERSON_ID} doesn't already exist...`);
                
                const getPersonArgs = new apiServerApi.GetPersonArgs()
                getPersonArgs.setPersonId(TEST_PERSON_ID)

                const getPersonResultPromise: Promise<Result<apiServerApi.GetPersonResponse, Error>> = new Promise((resolve, _unusedReject) => {
                        apiClient.getPerson(getPersonArgs, (error: grpc.ServiceError | null, response?: apiServerApi.GetPersonResponse) => {
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

                const getPersonResult = await getPersonResultPromise;
                if(getPersonResult.isOk()) { 
                    throw new Error("Expected an error trying to get a person who doesn't exist yet, but didn't receive one")
                }
                log.info("Verified that test person doesn't already exist")
                
                log.info(`Adding test person with ID ${TEST_PERSON_ID}...`)

                const addPersonArgs = new apiServerApi.AddPersonArgs()
                addPersonArgs.setPersonId(TEST_PERSON_ID)

                const addPersonResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                    apiClient.addPerson(addPersonArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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

                const addPersonResult = await addPersonResultPromise;
                if(addPersonResult.isErr()) {
                    log.error(`An error occurred adding test person with ID ${TEST_PERSON_ID}`)
                    throw addPersonResult.error;
                }
                log.info("Test person added")
                
                log.info(`Incrementing test person's number of books read by ${TEST_NUM_BOOKS_READ}...`)
                
                const incrementBooksReadArgs = new apiServerApi.IncrementBooksReadArgs()
                incrementBooksReadArgs.setPersonId(TEST_PERSON_ID)
            
                for (let i = 0; i < TEST_NUM_BOOKS_READ; i++) {
                    const incrementBooksReadResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                        apiClient.incrementBooksRead(incrementBooksReadArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
                        log.error("An error occurred incrementing the number of books read")
                        throw incrementBooksReadResult.error;
                    }
                }
                
                log.info("Incremented number of books read")
                
                log.info("Retrieving test person to verify number of books read...")
                
                const getPersonAfterUpdatingResultPromise: Promise<Result<apiServerApi.GetPersonResponse, Error>> = new Promise((resolve, _unusedReject) => {
                    apiClient.getPerson(getPersonArgs, (error: grpc.ServiceError | null, response?: apiServerApi.GetPersonResponse) => {
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
                
                const  getPersonAfterUpdatingResult = await getPersonAfterUpdatingResultPromise
                if(getPersonAfterUpdatingResult.isErr()){
                    log.error("An error occurred getting the test person to verify the number of books read")
                    throw getPersonAfterUpdatingResult.error
                }
                log.info("Retrieved test person")
                
                const newPersonBooksRead = getPersonAfterUpdatingResult.value.getBooksRead()
                
                if(TEST_NUM_BOOKS_READ !== newPersonBooksRead){
                    throw new Error(`Expected number of book read ${TEST_NUM_BOOKS_READ} != actual number of books read ${newPersonBooksRead}`)
                }
            }finally{
                apiClientCloseFunction()
            }
        }finally{
            datastoreClientCloseFunction()
        }
    }finally{
        stopEnclaveFunction()
    }
});