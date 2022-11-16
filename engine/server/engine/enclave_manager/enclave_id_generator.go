package enclave_manager

import (
	"github.com/goombaio/namegenerator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"time"
)

const (
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveIdKeyword = enclave.EnclaveID("")
)

func getRandomEnclaveIdWithRetries(
	allCurrentEnclaves map[enclave.EnclaveID]*enclave.Enclave,
	retries uint16,
) (enclave.EnclaveID, error) {

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	randomName := nameGenerator.Generate()

	randomEnclaveId := enclave.EnclaveID(randomName)

	validationError := validateEnclaveId(randomEnclaveId)

	isIdInUse := IsEnclaveIdInUse(randomEnclaveId, allCurrentEnclaves)

	if validationError != nil || isIdInUse {
		if retries > 0 {
			newRetriesValue := retries - 1
			return getRandomEnclaveIdWithRetries(allCurrentEnclaves, newRetriesValue)
		}

		var (
			errMsg = "Generating a new random enclave ID has executed all the retries set without success"
			returnErr = stacktrace.NewError(errMsg)
		)

		if validationError != nil {
			returnErr = stacktrace.Propagate(validationError, errMsg)
		}

		return autogenerateEnclaveIdKeyword, returnErr
	}

	return randomEnclaveId, nil
}
