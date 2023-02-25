package enclave_manager

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/name_generator"
	"sync"
)

const (
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveNameKeyword = ""
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentEnclaveNameGenerator *enclaveNameGenerator
	once                        sync.Once
)

type enclaveNameGenerator struct {
	generateNatureThemeName func() string
}

func GetEnclaveNameGenerator() *enclaveNameGenerator {
	// NOTE: Do not think we still need multiple instances of this struct; though
	// open not having once here as well.
	once.Do(func() {
		currentEnclaveNameGenerator = &enclaveNameGenerator{
			generateNatureThemeName: name_generator.GenerateNatureThemeNameForEngine}
	})
	return currentEnclaveNameGenerator
}

func (enclaveIdGenerator *enclaveNameGenerator) GetRandomEnclaveNameWithRetries(
	allCurrentEnclaves map[enclave.EnclaveUUID]*enclave.Enclave,
	retries uint16,
) string {
	retriesLeft := retries - 1

	randomEnclaveName := enclaveIdGenerator.generateNatureThemeName()
	isIdInUse := isEnclaveNameInUse(randomEnclaveName, allCurrentEnclaves)

	// if the id is unique, return the enclave name
	if !isIdInUse {
		return randomEnclaveName
	}

	// recursive call until no retries are left
	if retries > 0 {
		return enclaveIdGenerator.GetRandomEnclaveNameWithRetries(allCurrentEnclaves, retriesLeft)
	}

	// if no retry is left, then use the last nature theme name with NUMBER at the end
	number := 1
	randomEnclaveNameWithNumber := fmt.Sprintf("%v-%v", randomEnclaveName, number)
	for isEnclaveNameInUse(randomEnclaveNameWithNumber, allCurrentEnclaves) {
		number = number + 1
		randomEnclaveNameWithNumber = fmt.Sprintf("%v-%v", randomEnclaveName, number)
	}
	return randomEnclaveNameWithNumber
}
