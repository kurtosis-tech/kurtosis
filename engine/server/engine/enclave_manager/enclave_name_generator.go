package enclave_manager

import (
	"github.com/goombaio/namegenerator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
	"time"
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
	generator namegenerator.Generator
}

func GetEnclaveNameGenerator() *enclaveNameGenerator {
	// NOTE: We use a 'once' to initialize the enclaveNameGenerator because it contains a seed,
	// and we don't ever want multiple enclaveNameGenerator instances in existence
	once.Do(func() {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		currentEnclaveNameGenerator = &enclaveNameGenerator{generator: nameGenerator}
	})
	return currentEnclaveNameGenerator
}

func (enclaveIdGenerator *enclaveNameGenerator) GetRandomEnclaveNameWithRetries(
	allCurrentEnclaves map[enclave.EnclaveUUID]*enclave.Enclave,
	retries uint16,
) (string, error) {
	retriesLeft := retries - 1

	randomEnclaveName := enclaveIdGenerator.generator.Generate()

	validationError := validateEnclaveName(randomEnclaveName)
	if validationError != nil {
		if retries > 0 {
			return enclaveIdGenerator.GetRandomEnclaveNameWithRetries(allCurrentEnclaves, retriesLeft)
		}
		return autogenerateEnclaveNameKeyword, stacktrace.Propagate(
			validationError,
			"Generating a new random enclave ID has executed all the retries set without success, the last random enclave ID generated '%v' is not valid",
			randomEnclaveName,
		)
	}

	isIdInUse := isEnclaveNameInUse(randomEnclaveName, allCurrentEnclaves)
	if isIdInUse {
		if retries > 0 {
			return enclaveIdGenerator.GetRandomEnclaveNameWithRetries(allCurrentEnclaves, retriesLeft)
		}
		return autogenerateEnclaveNameKeyword, stacktrace.NewError(
			"Generating a new random enclave ID has executed all the retries set without success, the last random enclave ID generated '%v' in in use",
			randomEnclaveName,
		)
	}

	return randomEnclaveName, nil
}
