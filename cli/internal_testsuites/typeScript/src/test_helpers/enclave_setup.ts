import { EnclaveContext, EnclaveID } from "kurtosis-core-api-lib"
import { KurtosisContext,  } from "kurtosis-engine-api-lib"
import {Result, err, ok} from "neverthrow"
import log from "loglevel";

const TEST_SUITE_NAME_ENCLAVE_ID_FRAGMENT = "typescript-engine-server-test";
const MILLISECONDS_IN_SECOND = 1000;

export async function createEnclave(testName:string, isPartitioningEnabled: boolean):
	Promise<Result<{ 
        enclaveContext: EnclaveContext, 
        stopEnclaveFunction: () => void
    }, Error>> {

	const newKurtosisContextResult = KurtosisContext.newKurtosisContextFromLocalEngine();
	if(newKurtosisContextResult.isErr()) {
		return err(new Error(`An error occurred connecting to the Kurtosis engine for running test ${testName}`))
	}
	const kurtosisContext = newKurtosisContextResult.value;
	
	const enclaveId:EnclaveID = `${TEST_SUITE_NAME_ENCLAVE_ID_FRAGMENT}_${testName}_${Math.round(Date.now()/MILLISECONDS_IN_SECOND)}`
	const createEnclaveResult = await kurtosisContext.createEnclave(enclaveId, isPartitioningEnabled);
	
	if(createEnclaveResult.isErr()) {
		return err(new Error(`An error occurred creating enclave ${enclaveId}`))
	}

	const enclaveContext = createEnclaveResult.value;

	const stopEnclaveFunction = async ():Promise<void> => {
		const stopEnclaveResult = await kurtosisContext.stopEnclave(enclaveId)
		if(stopEnclaveResult.isErr()) {
			log.error(`An error occurred stopping enclave ${enclaveId} that we created for this test: ${stopEnclaveResult.error.message}`)
			log.error(`ACTION REQUIRED: You'll need to stop enclave ${enclaveId} manually!!!!`)
		}
	}

	return ok({
		enclaveContext,
		stopEnclaveFunction
	})
}
