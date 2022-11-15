package enclave_manager

import (
	"github.com/goombaio/namegenerator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const getRandomEnclaveIdRetries = uint16(5)

func getRandomEnclaveIdWithRetries(
	allCurrentEnclaves map[enclave.EnclaveID]*enclave.Enclave,
	retries uint16,
) (enclave.EnclaveID, error) {

	var err error

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	randomName := nameGenerator.Generate()

	randomEnclaveId := enclave.EnclaveID(randomName)
	logrus.Debugf("Genetared new random enclave ID '%v'", randomEnclaveId)

	validationError := validateEnclaveId(randomEnclaveId)

	isIdInUse := isEnclaveIdInUse(randomEnclaveId, allCurrentEnclaves)

	if validationError != nil || isIdInUse {
		if retries > 0 {
			newRetriesValue := retries - 1
			randomEnclaveId, err = getRandomEnclaveIdWithRetries(allCurrentEnclaves, newRetriesValue)
			if err != nil {
				return emptyEnclaveId,
					stacktrace.Propagate(err,
						"An error occurred getting a random enclave ID with all current enclaves value '%+v' and retries '%v'",
						allCurrentEnclaves,
						retries,
					)
			}
		}

		var (
			errMsg = "Generating a new random enclave ID has reached the max allowed retries '%v' without success"
			returnErr = stacktrace.NewError(errMsg, getRandomEnclaveIdRetries)
		)

		if validationError != nil {
			returnErr = stacktrace.Propagate(validationError, errMsg)
		}

		return emptyEnclaveId, returnErr
	}

	return randomEnclaveId, nil
}
