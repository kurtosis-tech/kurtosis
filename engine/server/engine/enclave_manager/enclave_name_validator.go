package enclave_manager

import (
	enclave_consts "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/enclave"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
)

func validateEnclaveName(enclaveName string) error {
	validEnclaveName, err := regexp.Match(enclave_consts.AllowedEnclaveNameCharsRegexStr, []byte(enclaveName))
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred validating that enclave name '%v' matches allowed enclave name regex '%v'",
			enclaveName,
			enclave_consts.AllowedEnclaveNameCharsRegexStr,
		)
	}
	if !validEnclaveName {
		return stacktrace.NewError(
			"Enclave name '%v' doesn't match allowed enclave name regex '%v'",
			enclaveName,
			enclave_consts.AllowedEnclaveNameCharsRegexStr,
		)
	}
	return nil
}

func isEnclaveNameInUse(newEnclaveName string, allEnclaves map[enclave.EnclaveUUID]*enclave.Enclave) bool {
	for _, enclave := range allEnclaves {
		if enclave.GetName() == newEnclaveName {
			return true
		}
	}
	return false
}
