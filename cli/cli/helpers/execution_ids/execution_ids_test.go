package execution_ids

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateEnclaveId_success(t *testing.T) {
	err := ValidateEnclaveId("valid.enclave.id-1234567")
	require.Nil(t, err)
}

func TestValidateEnclaveId_failureInvalidChar(t *testing.T) {
	err := ValidateEnclaveId("valid.enclave.id-1234567&")
	require.NotNil(t, err)
}

func TestValidateEnclaveId_failureTooLong(t *testing.T) {
	err := ValidateEnclaveId("IAmWayTooLongToBeAnEnclaveIdBecauseIShouldBeLessThan64CharAndIAmNo")
	require.NotNil(t, err)
}
