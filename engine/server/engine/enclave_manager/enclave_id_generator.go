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
	autogenerateEnclaveIdKeyword = enclave.EnclaveID("")
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentEnclaveIdGenerator *enclaveIdGenerator
	once sync.Once
)

type enclaveIdGenerator struct{
	generator namegenerator.Generator
}

func GetEnclaveIdGenerator() *enclaveIdGenerator {
	// NOTE: We use a 'once' to initialize the enclaveIdGenerator because it contains a seed,
	// and we don't ever want multiple enclaveIdGenerator instances in existence
	once.Do(func() {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		currentEnclaveIdGenerator = &enclaveIdGenerator{generator: nameGenerator}
	})
	return currentEnclaveIdGenerator
}

func (enclaveIdGenerator *enclaveIdGenerator) GetRandomEnclaveIdWithRetries(
	allCurrentEnclaves map[enclave.EnclaveID]*enclave.Enclave,
	retries uint16,
) (enclave.EnclaveID, error) {
	retriesLeft := retries - 1

	randomName := enclaveIdGenerator.generator.Generate()

	randomEnclaveId := enclave.EnclaveID(randomName)

	validationError := validateEnclaveId(randomEnclaveId)
	if validationError != nil  {
		if retries > 0 {
			return enclaveIdGenerator.GetRandomEnclaveIdWithRetries(allCurrentEnclaves, retriesLeft)
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
			return enclaveIdGenerator.GetRandomEnclaveIdWithRetries(allCurrentEnclaves, retriesLeft)
		}
		return autogenerateEnclaveIdKeyword, stacktrace.NewError(
			"Generating a new random enclave ID has executed all the retries set without success, the last random enclave ID generated '%v' in in use",
			randomEnclaveId,
		)
	}

	return randomEnclaveId, nil
}
