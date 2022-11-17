package enclave_manager

import (
	enclave_consts "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

func validateEnclaveId(enclaveId enclave.EnclaveID) error {
	validEnclaveId, err := regexp.Match(enclave_consts.AllowedEnclaveIdCharsRegexStr, []byte(enclaveId))
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred validating that enclave ID '%v' matches allowed enclave ID regex '%v'",
			enclaveId,
			enclave_consts.AllowedEnclaveIdCharsRegexStr,
		)
	}
	if !validEnclaveId {
		return stacktrace.NewError(
			"Enclave ID '%v' doesn't match allowed enclave ID regex '%v'",
			enclaveId,
			enclave_consts.AllowedEnclaveIdCharsRegexStr,
		)
	}
	return nil
}

func isEnclaveIdInUse(newEnclaveId enclave.EnclaveID, allEnclaves map[enclave.EnclaveID]*enclave.Enclave) bool {

	isInUse := false

	for enclaveId := range allEnclaves {
		if newEnclaveId == enclaveId {
			isInUse = true
		}
	}

	return isInUse
}
