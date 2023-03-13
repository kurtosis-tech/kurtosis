package enclave_manager

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
)

const (
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveNameKeyword = ""
)

func GetRandomEnclaveNameWithRetries(generateNatureThemeName func() string, allCurrentEnclaves map[enclave.EnclaveUUID]*enclave.Enclave, retries uint16) string {
	retriesLeft := retries - 1

	randomEnclaveName := generateNatureThemeName()
	isIdInUse := isEnclaveNameInUse(randomEnclaveName, allCurrentEnclaves)

	// if the id is unique, return the enclave name
	if !isIdInUse {
		return randomEnclaveName
	}

	// recursive call until no retries are left
	if retries > 0 {
		return GetRandomEnclaveNameWithRetries(generateNatureThemeName, allCurrentEnclaves, retriesLeft)
	}

	// if no retry is left, then use the last nature theme name with NUMBER at the end
	randomEnclaveNameWithNumber := randomEnclaveName
	for i := 1; isEnclaveNameInUse(randomEnclaveNameWithNumber, allCurrentEnclaves); i++ {
		randomEnclaveNameWithNumber = fmt.Sprintf("%v-%v", randomEnclaveName, i)
	}
	return randomEnclaveNameWithNumber
}
