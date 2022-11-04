package remove_service

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveService_StringRepresentation(t *testing.T) {
	removeInstruction := NewRemoveServiceInstruction(nil, *kurtosis_instruction.NewInstructionPosition(4, 4, "dummyFile"), "dummy-service-id")
	expectedStrRep := `remove_service(service_id="dummy-service-id")`
	require.Equal(t, expectedStrRep, removeInstruction.String())
	require.Equal(t, expectedStrRep, removeInstruction.GetCanonicalInstruction())
}
