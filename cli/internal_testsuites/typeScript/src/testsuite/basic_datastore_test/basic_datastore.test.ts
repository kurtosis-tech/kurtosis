import { DatastoreServiceClient, UpsertArgs, ExistsArgs, GetArgs, ExistsResponse, GetResponse } from "example-datastore-server-api-lib";
import { ServiceContext, ServiceID } from "kurtosis-core-api-lib"
import * as grpc from "grpc"
import { Result, ok, err } from "neverthrow"
import log from "loglevel"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

import { createEnclave, CreateEnclaveReturn } from "../../test_helpers/enclave_setup";
import { addDatastoreService } from "../../test_helpers/test_helpers";

const TEST_NAME = "basic-datastore-test";
const IS_PARTITIONING_ENABLED = false;
const DATASTORE_SERVICE_ID: ServiceID = "datastore";
const TEST_KEY = "test-key"
const TEST_VALUE = "test-value";

describe("Test basic data store", () => {
	jest.setTimeout(10000)
	
	let createEnclaveResult: Result<CreateEnclaveReturn, Error>;

	let datastoreService: Result<{
		serviceCtx: ServiceContext;
		client: DatastoreServiceClient;
		clientCloseFunc: () => void;
	}, Error>;

	const existsArgs = new ExistsArgs();
	existsArgs.setKey(TEST_KEY)
	let existsResponse:Promise<Result<ExistsResponse, Error>>;

	const upsertArgs = new UpsertArgs()
	upsertArgs.setKey(TEST_KEY)
	upsertArgs.setValue(TEST_VALUE)
	let upsertResponse:Promise<Result<google_protobuf_empty_pb.Empty, Error>>;

	const getArgs = new GetArgs();
	getArgs.setKey(TEST_KEY)
	let getResponse:Promise<Result<GetResponse, Error>>;

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	
	beforeAll(async() => {
		createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)
	})

	
	test("", async () => {
		if(createEnclaveResult.isErr()) {
			throw createEnclaveResult.error
		}
		
		// ------------------------------------- TEST SETUP ----------------------------------------------
		log.info("Adding datastore service...")
		datastoreService = await addDatastoreService(DATASTORE_SERVICE_ID, createEnclaveResult.value.enclaveContext)
		if(datastoreService.isErr()) {
			throw datastoreService.error
		}
		log.info("Added datastore service")
	

		// ------------------------------------- TEST RUN ----------------------------------------------

		log.info(`Verifying that key ${TEST_KEY} doesn't already exist..."`)
		existsResponse = new Promise((resolve, _unusedReject) => {
			if(datastoreService?.isOk()){
				datastoreService.value.client.exists(existsArgs, (error: grpc.ServiceError | null, response?: ExistsResponse) => {
					if (error === null) {
						if (!response) {
							resolve(err(new Error("Unexpected error from testing if key exists")));
						} else {
							resolve(ok(response));
						}
					} else {
						console.error(error)
						resolve(err(error));
					}
				})
			}
		})
		const existsResponseResult = await existsResponse;
		if(existsResponseResult.isErr()) {
			throw existsResponseResult.error
		}
		if(existsResponseResult.value.getExists()) {
			throw new Error("Test key should not exist")
		}
		log.info(`Confirmed that key ${TEST_KEY} doesn't already exist!`)


		log.info(`Inserting value ${TEST_KEY} at key ${TEST_VALUE}...`)
		upsertResponse = new Promise((resolve, _unusedReject) => {
			if(datastoreService?.isOk()) {
				datastoreService.value.client.upsert(upsertArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
					if (error === null) {
						if (!response) {
							console.error("Unexpected error from upserting the test key")
							resolve(err(new Error()));
						} else {
							resolve(ok(response));
						}
					} else {
						console.error(error)
						resolve(err(error));
					}
				})
			}
		})
		const upsertResponseResult = await upsertResponse;
		if(upsertResponseResult.isErr()){
			throw upsertResponseResult.error
		}
		log.info(`Inserted value successfully`)


		log.info(`Getting the key we just inserted to verify the value...`)
		getResponse = new Promise((resolve, _unusedReject) => {
			if(datastoreService?.isOk()) {
				datastoreService.value.client.get(getArgs, (error: grpc.ServiceError | null, response?: GetResponse) => {
					if (error === null) {
						if (!response) {
							console.error("Unexpected error from getting the test key after upload")
							resolve(err(new Error()));
						} else {
							resolve(ok(response));
						}
					} else {
						console.error(error)
						resolve(err(error));
					}
				})
			}
		})
		const getResponseResult = await getResponse;
		if(getResponseResult.isErr()){
			throw getResponseResult.error
		}
		if(getResponseResult.value.getValue() !== TEST_VALUE) {
			throw new Error(`Returned value ${getResponseResult.value.getValue()} != test value ${TEST_VALUE}`)
		}
		log.info(`Value verified`)
	})

	afterAll(() => {
		if(createEnclaveResult.isOk()) {
			createEnclaveResult.value.stopEnclaveFunction()
		}
		if(datastoreService?.isOk()) {
			datastoreService.value.clientCloseFunc()
		}
	});
})
