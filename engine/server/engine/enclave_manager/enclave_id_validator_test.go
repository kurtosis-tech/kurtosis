package enclave_manager

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	firstEnclaveIdForTest = enclave.EnclaveID("first-enclave-id")
	secondEnclaveIdForTest = enclave.EnclaveID("second-enclave-id")
	theirEnclaveIdForTest = enclave.EnclaveID("their-enclave-id")
	runningEnclaveStatus = enclave.EnclaveStatus_Running

)

var (
	creationTime             = time.Now()
	firstEnclaveForTest      = enclave.NewEnclave(firstEnclaveIdForTest, runningEnclaveStatus, &creationTime)
	secondEnclaveForTest      = enclave.NewEnclave(secondEnclaveIdForTest, runningEnclaveStatus, &creationTime)
	theirEnclaveForTest      = enclave.NewEnclave(theirEnclaveIdForTest, runningEnclaveStatus, &creationTime)
	currentEnclaveIdsForTest = map[enclave.EnclaveID]*enclave.Enclave{
		firstEnclaveIdForTest: firstEnclaveForTest,
		secondEnclaveIdForTest: secondEnclaveForTest,
		theirEnclaveIdForTest: theirEnclaveForTest,
	}
)

func TestValidateEnclaveId_success(t *testing.T) {
	err := validateEnclaveId("valid.enclave.id-1234567")
	require.Nil(t, err)

	err = validateEnclaveId("go-testsuite.startosis_remove_service_test.1668628850670949")
	require.Nil(t, err)
}

func TestValidateEnclaveId_failureInvalidChar(t *testing.T) {
	err := validateEnclaveId("valid.enclave.id-1234567&")
	require.NotNil(t, err)
}

func TestValidateEnclaveId_failureTooLong(t *testing.T) {
	err := validateEnclaveId("IAmWayTooLongToBeAnEnclaveIdBecauseIShouldBeLessThan64CharAndIAmNo")
	require.NotNil(t, err)
}

func TestIsEnclaveIdInUse_isNotInUse(t *testing.T) {
	newEnclaveId := enclave.EnclaveID("new-enclave-id")
	isInUse := isEnclaveIdInUse(newEnclaveId, currentEnclaveIdsForTest)
	require.False(t, isInUse)
}

func TestIsEnclaveIdInUse_isInUse(t *testing.T) {
	isInUse := isEnclaveIdInUse(firstEnclaveIdForTest, currentEnclaveIdsForTest)
	require.True(t, isInUse)
}
