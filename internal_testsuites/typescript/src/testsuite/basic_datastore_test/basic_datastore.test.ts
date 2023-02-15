import { UpsertArgs, ExistsArgs, GetArgs, ExistsResponse, GetResponse } from "example-datastore-server-api-lib";
import { ServiceName } from "kurtosis-sdk"
import * as grpc from "@grpc/grpc-js"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import log from "loglevel"
import { Result, ok, err } from "neverthrow"

import { createEnclave } from "../../test_helpers/enclave_setup";
import { addDatastoreService } from "../../test_helpers/test_helpers";

const TEST_NAME = "basic-datastore";
const IS_PARTITIONING_ENABLED = false;
const DATASTORE_SERVICE_NAME: ServiceName = "datastore";
const TEST_KEY = "test-key"
const TEST_VALUE = "test-value";

jest.setTimeout(180000)

test("Test basic datastore test", async () => {
        
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)
    
    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }
    
    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        log.info("Adding datastore service...")

        const addDatastoreServiceResult = await addDatastoreService(DATASTORE_SERVICE_NAME, enclaveContext)

        if(addDatastoreServiceResult.isErr()) { throw addDatastoreServiceResult.error }

        const { client:datastoreClient, clientCloseFunction } = addDatastoreServiceResult.value

        log.info("Added datastore service")

        try {
            // ------------------------------------- TEST RUN ----------------------------------------------

            log.info(`Verifying that key ${TEST_KEY} doesn't already exist..."`)

            const existsArgs = new ExistsArgs();
            existsArgs.setKey(TEST_KEY)

            const existsResponseResultPromise:Promise<Result<ExistsResponse, Error>> = new Promise((resolve, _unusedReject) => {
                datastoreClient.exists(existsArgs, (error: grpc.ServiceError | null, response?: ExistsResponse) => {
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

            const existsResponseResult = await existsResponseResultPromise;
            if(existsResponseResult.isErr()) { 
                log.error("An error occurred checking if the test key exists")
                throw existsResponseResult.error 
            }
            const existsResponse = existsResponseResult.value
            if(existsResponse.getExists()) { throw new Error("Test key should not exist yet") }

            log.info(`Confirmed that key ${TEST_KEY} doesn't already exist`)


            log.info(`Inserting value ${TEST_VALUE} at key ${TEST_KEY}...`)

            const upsertArgs = new UpsertArgs()
            upsertArgs.setKey(TEST_KEY)
            upsertArgs.setValue(TEST_VALUE)

            const upsertResponseResultPromise:Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                datastoreClient.upsert(upsertArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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

            const upsertResponseResult = await upsertResponseResultPromise;
            if(upsertResponseResult.isErr()){ 
                log.error("An error occurred upserting the test key")
                throw upsertResponseResult.error 
            }

            log.info(`Inserted value successfully`)


            log.info(`Getting the key we just inserted to verify the value...`)

            const getArgs = new GetArgs();
            getArgs.setKey(TEST_KEY)

            const getResponseResultPromise:Promise<Result<GetResponse, Error>> = new Promise((resolve, _unusedReject) => {
                datastoreClient.get(getArgs, (error: grpc.ServiceError | null, response?: GetResponse) => {
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

            const getResponseResult = await getResponseResultPromise;
            if(getResponseResult.isErr()){
                log.error("An error occurred getting the test key after upload") 
                throw getResponseResult.error 
            }

            const getResponse = getResponseResult.value
            if(getResponse.getValue() !== TEST_VALUE) {throw new Error(`Returned value ${getResponse.getValue()} != test value ${TEST_VALUE}`) }

            log.info(`Value verified`)
        
        }finally{
            clientCloseFunction()
        }

    }finally{
        stopEnclaveFunction()
    }

})
