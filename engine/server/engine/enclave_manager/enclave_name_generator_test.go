package enclave_manager

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	uniqueNameFoundOnLastTry        = "unique-name"
	nonUniqueName                   = "non-unique-name"
	nameAlreadyExists1              = "name-exists"
	nameAlreadyExists2              = "name-exists-1"
	expectedUniqueNameWithRandomNum = "name-exists-2"
)

func TestGetRandomEnclaveIdWithRetriesSuccess(t *testing.T) {
	retries := uint16(3)

	currentEnclaveName := []string{
		nonUniqueName,
		nameAlreadyExists1,
	}

	timesCalled := 0
	mockNatureThemeName := func() string {
		timesCalled++
		// found unique on the last try
		if timesCalled == 4 {
			return uniqueNameFoundOnLastTry
		}

		if timesCalled == 2 {
			return nonUniqueName
		}

		return nameAlreadyExists1
	}

	randomEnclaveId := GetRandomEnclaveNameWithRetries(mockNatureThemeName, currentEnclaveName, retries)
	require.NotEmpty(t, randomEnclaveId)
	require.Equal(t, uniqueNameFoundOnLastTry, randomEnclaveId)
}

func TestGetRandomEnclaveIdWithRetriesUniqueNameNotFound(t *testing.T) {
	retries := uint16(3)

	currentEnclaveName := []string{
		nonUniqueName,
		nameAlreadyExists1,
		nameAlreadyExists2,
	}

	timesCalled := 0
	mockNatureThemeName := func() string {
		timesCalled++

		if timesCalled == 4 {
			return nameAlreadyExists1
		}

		return nonUniqueName
	}

	randomEnclaveId := GetRandomEnclaveNameWithRetries(mockNatureThemeName, currentEnclaveName, retries)
	require.NotEmpty(t, randomEnclaveId)
	require.Equal(t, expectedUniqueNameWithRandomNum, randomEnclaveId)
}
