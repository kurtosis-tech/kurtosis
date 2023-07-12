package enclave_manager

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveNameKeyword = ""

	// Signifies that this enclave will be created and will be available for further use
	idleEnclaveNamePrefix = "idle-enclave-"
)

func GetRandomEnclaveNameWithRetries(generateNatureThemeName func() string, allCurrentEnclaveNames []string, retries uint16) string {
	retriesLeft := retries - 1

	randomEnclaveName := generateNatureThemeName()
	isIdInUse := isEnclaveNameInUse(randomEnclaveName, allCurrentEnclaveNames)

	// if the id is unique, return the enclave name
	if !isIdInUse {
		return randomEnclaveName
	}

	// recursive call until no retries are left
	if retries > 0 {
		return GetRandomEnclaveNameWithRetries(generateNatureThemeName, allCurrentEnclaveNames, retriesLeft)
	}

	// if no retry is left, then use the last nature theme name with NUMBER at the end
	randomEnclaveNameWithNumber := randomEnclaveName
	for i := 1; isEnclaveNameInUse(randomEnclaveNameWithNumber, allCurrentEnclaveNames); i++ {
		randomEnclaveNameWithNumber = fmt.Sprintf("%v-%v", randomEnclaveName, i)
	}
	return randomEnclaveNameWithNumber
}

func GetRandomIdleEnclaveName() (string, error) {
	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while creating UUID for idle enclave name")
	}

	enclaveName := idleEnclaveNamePrefix + uuid

	return enclaveName, nil
}
