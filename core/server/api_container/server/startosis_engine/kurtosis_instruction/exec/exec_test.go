package exec

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()
var defaultRuntimeValueStore *runtime_value_store.RuntimeValueStore = nil

func TestExecInstruction_StringRepresentationWorks(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile")
	execInstruction := newEmptyExecInstruction(emptyServiceNetwork, position, defaultRuntimeValueStore)
	execInstruction.starlarkKwargs = starlark.StringDict{}

	commands := []string{"mkdir", "-p", "/tmp/store"}
	execRecipe := recipe.NewExecRecipe("example-service-name", commands)

	execInstruction.starlarkKwargs[recipeArgName] = execRecipe
	expectedStr := `exec(recipe=ExecRecipe(service_name="example-service-name", command=["mkdir", "-p", "/tmp/store"]))`
	require.Equal(t, expectedStr, execInstruction.String())

	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		position.ToAPIType(),
		ExecBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(`ExecRecipe(service_name="example-service-name", command=["mkdir", "-p", "/tmp/store"])`, recipeArgName, true),
		})
	require.Equal(t, canonicalInstruction, execInstruction.GetCanonicalInstruction())
}
