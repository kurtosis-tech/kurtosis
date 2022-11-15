package enclave_manager

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

const (
	enclaveIdMaxLength            = uint16(63) // If changing this, change also the regexp allowedEnclaveIdCharsRegexStr below
	allowedEnclaveIdCharsRegexStr = `^[-A-Za-z0-9.]{1,63}$`
)

func validateEnclaveId(enclaveId enclave.EnclaveID) error {
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
