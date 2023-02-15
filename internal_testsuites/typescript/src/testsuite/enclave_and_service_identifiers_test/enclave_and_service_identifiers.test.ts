import {createEnclave} from "../../test_helpers/enclave_setup";
import {addDatastoreService} from "../../test_helpers/test_helpers";
import {KurtosisContext} from "kurtosis-sdk";
import log from "loglevel";

const TEST_NAME              = "identifiers-test"
const IS_PARTITIONING_ENABLED = false

const DATASTORE_SERVICE_NAME = "datastore"
const SHORTENED_UUID_LENGTH  = 12

const INVALID_SERVICE_NAME = "invalid-service"
const INVALID_ENCLAVE_NAME = "invalid-enclave"


jest.setTimeout(180000)
test("Test enclave & service identifiers", async() =>  {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const kurtosisCtxResult = await KurtosisContext.newKurtosisContextFromLocalEngine()
    if (kurtosisCtxResult.isErr()) {
        throw kurtosisCtxResult.error
    }
    const kurtosisCtx = kurtosisCtxResult.value

    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, destroyEnclaveFunction: destroyEnclaveFunction } = createEnclaveResult.value

    try {
        const addDatastoreServiceResult = await addDatastoreService(DATASTORE_SERVICE_NAME, enclaveContext)

        if(addDatastoreServiceResult.isErr()) { throw addDatastoreServiceResult.error }

        const {serviceContext, clientCloseFunction: clientCloseFunction} = addDatastoreServiceResult.value

        try {
            const enclaveUuid = enclaveContext.getEnclaveUuid()
            const enclaveName = enclaveContext.getEnclaveName()
            const enclaveShortenedUuid = enclaveUuid.substring(0, SHORTENED_UUID_LENGTH)

            const serviceUuid = serviceContext.getServiceUUID()
            const serviceName = serviceContext.getServiceName()
            const shortenedServiceUuid = serviceUuid.substring(0, SHORTENED_UUID_LENGTH)

            // test with service and enclave alive and running

            const allEnclaveIdentifiersResult = await kurtosisCtx.getExistingAndHistoricalEnclaveIdentifiers()
            if (allEnclaveIdentifiersResult.isErr()) {
                throw allEnclaveIdentifiersResult.error;
            }
            const allEnclaveIdentifiers = allEnclaveIdentifiersResult.value

            let enclaveUuidResult = allEnclaveIdentifiers.getEnclaveUuidForIdentifier(enclaveUuid)
            expect(enclaveUuidResult).toEqual(enclaveUuid)
            let enclave = await kurtosisCtx.getEnclave(enclaveUuid)
            expect(enclave.isErr()).toEqual(false)

            enclaveUuidResult = allEnclaveIdentifiers.getEnclaveUuidForIdentifier(enclaveShortenedUuid)
            expect(enclaveUuidResult).toEqual(enclaveUuid)
            enclave = await kurtosisCtx.getEnclave(enclaveShortenedUuid)
            expect(enclave.isErr()).toEqual(false)


            enclaveUuidResult = allEnclaveIdentifiers.getEnclaveUuidForIdentifier(enclaveName)
            expect(enclaveUuidResult).toEqual(enclaveUuid)
            enclave = await kurtosisCtx.getEnclave(enclaveName)
            expect(enclave.isErr()).toEqual(false)

            expect(function () {
                allEnclaveIdentifiers.getEnclaveUuidForIdentifier(INVALID_ENCLAVE_NAME)
            }).toThrow()

            const allServiceIdentifiersResult = await enclaveContext.getExistingAndHistoricalServiceIdentifiers()
            if (allServiceIdentifiersResult.isErr()) {
                throw allServiceIdentifiersResult.error;
            }
            const allServiceIdentifiers = allServiceIdentifiersResult.value

            let serviceUuidResult = allServiceIdentifiers.getServiceUuidForIdentifier(serviceUuid)
            expect(serviceUuidResult).toEqual(serviceUuid)
            let serviceCtx = await enclaveContext.getServiceContext(serviceUuid)
            expect(serviceCtx.isErr()).toEqual(false)

            serviceUuidResult = allServiceIdentifiers.getServiceUuidForIdentifier(shortenedServiceUuid)
            expect(serviceUuidResult).toEqual(serviceUuid)
            serviceCtx = await enclaveContext.getServiceContext(shortenedServiceUuid)
            expect(serviceCtx.isErr()).toEqual(false)

            serviceUuidResult = allServiceIdentifiers.getServiceUuidForIdentifier(serviceName)
            expect(serviceUuidResult).toEqual(serviceUuid)
            serviceCtx = await enclaveContext.getServiceContext(serviceName)
            expect(serviceCtx.isErr()).toEqual(false)

            expect(function () {
                allServiceIdentifiers.getServiceUuidForIdentifier(INVALID_SERVICE_NAME)
            }).toThrow()
        } finally {
            clientCloseFunction()
        }
    } finally {
        await destroyEnclaveFunction()
    }
})
