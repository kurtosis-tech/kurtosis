import { EnclaveContext, EnclaveID } from "kurtosis-core-api-lib"
import { KurtosisContext,  } from "kurtosis-engine-api-lib"
import {Result, err, ok} from "neverthrow"
import log from "loglevel";

const TEAST_SUITE_NAME_ENCLAVE_ID_FRAGMENT = "golang-engine-server-test"

export type CreateEnclaveReturn = {
	enclaveContext: EnclaveContext
	stopEnclaveFunction: () => void
}

export async function createEnclave(testName:string, isPartitioningEnabled: boolean):Promise<Result<CreateEnclaveReturn, Error>> {

	const kurtosisContext = KurtosisContext.newKurtosisContextFromLocalEngine();
	if(kurtosisContext.isErr()) {
		console.error(`An error occurred connecting to the Kurtosis engine for running test ${testName}`)
		return err(new Error(`An error occurred connecting to the Kurtosis engine for running test ${testName}`))
	}
	
	const enclaveId:EnclaveID = `${TEAST_SUITE_NAME_ENCLAVE_ID_FRAGMENT}_${testName}_${Date.now()}`
	const enclaveContext = await kurtosisContext.value.createEnclave(enclaveId, isPartitioningEnabled)
	if(enclaveContext.isErr()) {
		console.error(`An error occurred creating enclave ${enclaveId}`)
		return err(new Error(`An error occurred creating enclave ${enclaveId}`))
	}

	const stopEnclaveFunction = async ():Promise<void> => {
		const stopEnclave = await kurtosisContext.value.stopEnclave(enclaveId)
		if(stopEnclave.isErr()) {
			console.error(`An error occurred stopping enclave ${enclaveId} that we created for this test: ${stopEnclave.error.message}`)
			console.error(`ACTION REQUIRED: You'll need to stop enclave ${enclaveId} manually!!!!`)
			log.error(`An error occurred stopping enclave ${enclaveId} that we created for this test: ${stopEnclave.error.message}`)
			log.error(`ACTION REQUIRED: You'll need to stop enclave ${enclaveId} manually!!!!`)
		}
	}

	return ok({
		enclaveContext: enclaveContext.value,
		stopEnclaveFunction
	})
}
