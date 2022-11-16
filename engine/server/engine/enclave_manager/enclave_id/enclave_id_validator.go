package enclave_id

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

//WARNING!! if we modify this constant 'allowedEnclaveIdCharsRegexStr' we should upgrade the same one in
//Kurtosis CLI which is inside this file 'cli/cli/user_support_constants/user_support_constants.go'
const allowedEnclaveIdCharsRegexStr = `^[-A-Za-z0-9.]{1,63}$`

func ValidateEnclaveId(enclaveId enclave.EnclaveID) error {
	validEnclaveId, err := regexp.Match(allowedEnclaveIdCharsRegexStr, []byte(enclaveId))
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred validating that enclave ID '%v' matches allowed enclave ID regex '%v'",
			enclaveId,
			allowedEnclaveIdCharsRegexStr,
		)
	}
	if !validEnclaveId {
		return stacktrace.NewError(
			"Enclave ID '%v' doesn't match allowed enclave ID regex '%v'",
			enclaveId,
			allowedEnclaveIdCharsRegexStr,
		)
	}
	return nil
}

func IsEnclaveIdInUse(newEnclaveId enclave.EnclaveID, allEnclaves map[enclave.EnclaveID]*enclave.Enclave) bool {

	isInUse := false

	for enclaveId := range allEnclaves {
		if newEnclaveId == enclaveId {
			isInUse = true
		}
	}

	return isInUse
}
