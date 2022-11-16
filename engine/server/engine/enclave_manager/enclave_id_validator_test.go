package enclave_manager

import (
	"github.com/stretchr/testify/require"
	"testing"
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

