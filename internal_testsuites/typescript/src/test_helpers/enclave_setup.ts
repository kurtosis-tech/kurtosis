import { EnclaveContext, EnclaveUUID, KurtosisContext } from "kurtosis-sdk"
import {Result, err, ok} from "neverthrow"
import log from "loglevel";

const TEST_SUITE_NAME_ENCLAVE_UUID_FRAGMENT = "ts-testsuite";
const MILLISECONDS_IN_SECOND = 1000;

export async function createEnclave(testName:string, isPartitioningEnabled: boolean):
	Promise<Result<{
		enclaveContext: EnclaveContext,
		stopEnclaveFunction: () => void
		destroyEnclaveFunction: () => Promise<Result<null, Error>>,
	}, Error>> {

	const newKurtosisContextResult = await KurtosisContext.newKurtosisContextFromLocalEngine()
	if(newKurtosisContextResult.isErr()) {
		log.error(`An error occurred connecting to the Kurtosis engine for running test ${testName}`)
		return err(newKurtosisContextResult.error)
	}
	const kurtosisContext = newKurtosisContextResult.value;

	const enclaveId:EnclaveUUID = `${TEST_SUITE_NAME_ENCLAVE_UUID_FRAGMENT}.${testName}.${Math.round(Date.now()/MILLISECONDS_IN_SECOND)}`
	const createEnclaveResult = await kurtosisContext.createEnclave(enclaveId, isPartitioningEnabled);

	if(createEnclaveResult.isErr()) {
		log.error(`An error occurred creating enclave ${enclaveId}`)
		return err(createEnclaveResult.error)
	}

	const enclaveContext = createEnclaveResult.value;

	const stopEnclaveFunction = async ():Promise<void> => {
		const stopEnclaveResult = await kurtosisContext.stopEnclave(enclaveId)
		if(stopEnclaveResult.isErr()) {
			log.error(`An error occurred stopping enclave ${enclaveId} that we created for this test: ${stopEnclaveResult.error.message}`)
			log.error(`ACTION REQUIRED: You'll need to stop enclave ${enclaveId} manually!!!!`)
		}
	}

	const destroyEnclaveFunction = async ():Promise<Result<null, Error>> => {
		const destroyEnclaveResult = await kurtosisContext.destroyEnclave(enclaveId)
		if(destroyEnclaveResult.isErr()) {
			const errMsg = `An error occurred destroying enclave ${enclaveId} that we created for this test: ${destroyEnclaveResult.error.message}`
			log.error(errMsg)
			log.error(`ACTION REQUIRED: You'll need to destroy enclave ${enclaveId} manually!!!!`)
			return err(new Error(errMsg))
		}
		return ok(null)
	}

	return ok({ enclaveContext, stopEnclaveFunction, destroyEnclaveFunction })
}
