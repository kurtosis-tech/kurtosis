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
	retriesLeft := retries - 1

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	randomName := nameGenerator.Generate()

	randomEnclaveId := enclave.EnclaveID(randomName)

	validationError := validateEnclaveId(randomEnclaveId)
	if validationError != nil  {
		if retries > 0 {
			return getRandomEnclaveIdWithRetries(allCurrentEnclaves, retriesLeft)
		}
		return autogenerateEnclaveIdKeyword, stacktrace.Propagate(
			validationError,
			"Generating a new random enclave ID has executed all the retries set without success, the last random enclave ID generated '%v' is not valid",
			randomEnclaveId,
		)
	}

	isIdInUse := isEnclaveIdInUse(randomEnclaveId, allCurrentEnclaves)
	if isIdInUse {
		if retries > 0 {
			return getRandomEnclaveIdWithRetries(allCurrentEnclaves, retriesLeft)
		}
		return autogenerateEnclaveIdKeyword, stacktrace.NewError(
			"Generating a new random enclave ID has executed all the retries set without success, the last random enclave ID generated '%v' in in use",
			randomEnclaveId,
		)
	}

	return randomEnclaveId, nil
}
