import { DatastoreServiceClient, UpsertArgs, ExistsArgs, GetArgs, ExistsResponse, GetResponse } from "example-datastore-server-api-lib";
import { ServiceContext, ServiceID } from "kurtosis-core-api-lib"
import grpc from "grpc"
import { Result, ok, err } from "neverthrow"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import log from "loglevel"

import { createEnclave, CreateEnclaveReturn } from "../../test_helpers/enclave_setup";
import { addDatastoreService } from "../../test_helpers/test_helpers";

const TEST_NAME = "basic-datastore-test";
const IS_PARTITIONING_ENABLED = false;
const DATASTORE_SERVICE_ID: ServiceID = "datastore";
const TEST_KEY = "test-key"
const TEST_VALUE = "test-value";

describe("Test basic data store", () => {
	jest.setTimeout(30000)
	
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	let enclave:Result<CreateEnclaveReturn, Error>;
	
	beforeAll(async() => {
		enclave = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)
	})

	test("If enclave was created successfully", async () => {
		expect(enclave).toBeTruthy()
		expect(enclave.isOk()).toBe(true)
	})

	afterAll(() => {
		if(enclave.isOk()) enclave.value.stopEnclaveFunction()
	});

	
	// ------------------------------------- TEST SETUP ----------------------------------------------
	log.info("Adding datastore service...")

	let datastoreService: undefined | Result<{
		serviceCtx: ServiceContext;
		client: DatastoreServiceClient;
		clientCloseFunc: () => void;
	}, Error>;

	beforeAll(async () => {
		if(enclave.isOk()) datastoreService = await addDatastoreService(DATASTORE_SERVICE_ID, enclave.value.enclaveContext)
	})

	test("If datastore service added successfully", async () => {
		expect(datastoreService).toBeTruthy()
		expect(datastoreService.isOk()).toBe(true)
		log.info("Added datastore service")
	})

	afterAll(() => {
		if(datastoreService.isOk()) datastoreService.value.clientCloseFunc()
	});
	
	// ------------------------------------- TEST RUN ----------------------------------------------
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

	beforeAll(async () => {
		if(enclave.isOk()){
				existsResponse = new Promise((resolve, _unusedReject) => {
					if(datastoreService.isOk()){
						datastoreService.value.client.exists(existsArgs, (error: grpc.ServiceError | null, response?: ExistsResponse) => {
							if (error === null) {
								if (!response) {
									console.error("Unexpected error from testing if key exists")
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
				upsertResponse = new Promise((resolve, _unusedReject) => {
					if(datastoreService.isOk()) {
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
				getResponse = new Promise((resolve, _unusedReject) => {
					if(datastoreService.isOk()) {
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
		}
	})

	log.info(`Verifying that key ${TEST_KEY} doesn't already exist..."`)
	
	test(`If key ${TEST_KEY} doesn't already exist`, async () => {
		const result = await existsResponse;
		expect(result).toBeTruthy();
		expect(result.isOk()).toBe(true);
		if(result.isOk()) {
			console.log(result.value.getExists())
			expect(result.value.getExists()).toBe(false)
		}
		log.info(`Confirmed that key ${TEST_KEY} doesn't already exist!`)
	})
	
	log.info(`Inserting value ${TEST_KEY} at key ${TEST_VALUE}...`)

	test(`Upserting the test key`, async () => {
		const result = await upsertResponse;
		expect(result).toBeTruthy();
		expect(result.isOk()).toBe(true);
		log.info(`Inserted value successfully`)
	})
	
	log.info(`Getting the key we just inserted to verify the value...`)

	test(`Getting the test key after upload`, async () => {
		const result = await getResponse;
		expect(result).toBeTruthy();
		expect(result.isOk()).toBe(true);
		if(result.isOk()) {
			console.log(result.value.getValue())
			expect(result.value.getValue()).toBe(TEST_VALUE)
		}
		log.info(`Value verified`)
	})

})
