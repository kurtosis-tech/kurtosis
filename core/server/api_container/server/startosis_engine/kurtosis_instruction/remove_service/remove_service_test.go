package remove_service

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveService_GetCanonicalInstruction(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(4, 4, "dummyFile")
	removeInstruction := NewRemoveServiceInstruction(
		nil,
		position,
		"dummy-service-id")
	expectedStr := `remove_service(service_id="dummy-service-id")`
	require.Equal(t, expectedStr, removeInstruction.String())

	canonicalInstruction := binding_constructors.NewKurtosisInstruction(
		position.ToAPIType(),
		RemoveServiceBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{
			binding_constructors.NewKurtosisInstructionKwarg(`"dummy-service-id"`, serviceIdArgName, true),
		})
	require.Equal(t, canonicalInstruction, removeInstruction.GetCanonicalInstruction())
}
