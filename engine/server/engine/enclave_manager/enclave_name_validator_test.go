package enclave_manager

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	firstEnclaveUuidForTest  = "first-enclave-uuid"
	firstEnclaveNameForTest  = "first-enclave-name"
	secondEnclaveUuidForTest = "second-enclave-uuid"
	secondEnclaveNameForTest = "second-enclave-name"
	theirEnclaveUuidForTest  = "their-enclave-uuid"
	theirEnclaveNameForTest  = "their-enclave-name"
	runningEnclaveStatus     = enclave.EnclaveStatus_Running
)

var (
	currentEnclaveNamesForTest = []string{
		firstEnclaveNameForTest,
		secondEnclaveNameForTest,
		theirEnclaveNameForTest,
	}
)

func TestValidateEnclaveId_success(t *testing.T) {
	err := validateEnclaveName("valid.enclave.id-1234567")
	require.Nil(t, err)

	err = validateEnclaveName("go-testsuite.startosis_remove_service_test.1668628850670949")
	require.Nil(t, err)
}

func TestValidateEnclaveId_failureInvalidChar(t *testing.T) {
	err := validateEnclaveName("valid.enclave.id-1234567&")
	require.NotNil(t, err)
}

func TestValidateEnclaveId_failureTooLong(t *testing.T) {
	err := validateEnclaveName("IAmWayTooLongToBeAnEnclaveIdBecauseIShouldBeLessThan60CharsIu")
	require.NotNil(t, err)
}

func TestIsEnclaveIdInUse_isNotInUse(t *testing.T) {
	newEnclaveName := "new-enclave-name"
	isInUse := isEnclaveNameInUse(newEnclaveName, currentEnclaveNamesForTest)
	require.False(t, isInUse)
}

func TestIsEnclaveIdInUse_isInUse(t *testing.T) {
	isInUse := isEnclaveNameInUse(firstEnclaveNameForTest, currentEnclaveNamesForTest)
	require.True(t, isInUse)
}
