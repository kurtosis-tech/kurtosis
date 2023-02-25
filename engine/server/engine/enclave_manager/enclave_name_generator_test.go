package enclave_manager

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetRandomEnclaveIdWithRetriesSuccess(t *testing.T) {
	retries := uint16(3)

	currentEnclavePresent := map[enclave.EnclaveUUID]*enclave.Enclave{
		"123": enclave.NewEnclave("123", "name-exists", enclave.EnclaveStatus_Empty, nil),
		"456": enclave.NewEnclave("456", "not-unique", enclave.EnclaveStatus_Empty, nil),
	}

	timesCalled := 0
	mockNatureThemeName := func() string {
		timesCalled++
		// found unique on the last try
		if timesCalled == 4 {
			return "unique-name"
		}

		if timesCalled%2 == 0 {
			return "name-exists"
		}

		return "not-unique"
	}

	enclaveIdGeneratorObj := enclaveNameGenerator{generateNatureThemeName: mockNatureThemeName}
	randomEnclaveId := enclaveIdGeneratorObj.GetRandomEnclaveNameWithRetries(currentEnclavePresent, retries)
	require.NotEmpty(t, randomEnclaveId)
	require.Equal(t, "unique-name", randomEnclaveId)
}

func TestGetRandomEnclaveIdWithRetriesUniqueNameNotFound(t *testing.T) {
	retries := uint16(3)

	currentEnclavePresent := map[enclave.EnclaveUUID]*enclave.Enclave{
		"123": enclave.NewEnclave("123", "name-exists", enclave.EnclaveStatus_Empty, nil),
		"456": enclave.NewEnclave("456", "last-name", enclave.EnclaveStatus_Empty, nil),
		"789": enclave.NewEnclave("789", "last-name-1", enclave.EnclaveStatus_Empty, nil),
	}

	timesCalled := 0
	mockNatureThemeName := func() string {
		timesCalled++

		if timesCalled == 4 {
			return "last-name"
		}

		return "name-exists"
	}

	enclaveIdGeneratorObj := enclaveNameGenerator{generateNatureThemeName: mockNatureThemeName}
	randomEnclaveId := enclaveIdGeneratorObj.GetRandomEnclaveNameWithRetries(currentEnclavePresent, retries)
	require.NotEmpty(t, randomEnclaveId)
	require.Equal(t, "last-name-2", randomEnclaveId)
}
