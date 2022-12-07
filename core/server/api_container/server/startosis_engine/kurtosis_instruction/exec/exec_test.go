package exec

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()
var defaultRuntimeValueStore *runtime_value_store.RuntimeValueStore = nil

const testExecId = "aefd992"

func TestExecInstruction_StringRepresentationWorks(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile")
	execInstruction := newEmptyExecInstruction(emptyServiceNetwork, position, defaultRuntimeValueStore)
	execInstruction.starlarkKwargs = starlark.StringDict{}
	execInstruction.starlarkKwargs[serviceIdArgName] = starlark.String("example-service-id")
	execInstruction.starlarkKwargs[commandArgName] = starlark.NewList([]starlark.Value{
		starlark.String("mkdir"),
		starlark.String("-p"),
		starlark.String("/tmp/store"),
	})
	execInstruction.starlarkKwargs[nonOptionalExitCodeArgName] = starlark.MakeInt(0)
	execInstruction.starlarkKwargs[nonOptionalExecIdArgName] = starlark.String(testExecId)

	expectedStr := `exec(command=["mkdir", "-p", "/tmp/store"], exec_id="` + testExecId + `", expected_exit_code=0, service_id="example-service-id")`
	require.Equal(t, expectedStr, execInstruction.String())

	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		position.ToAPIType(),
		ExecBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(`"example-service-id"`, serviceIdArgName, true),
			binding_constructors.NewStarlarkInstructionKwarg(`["mkdir", "-p", "/tmp/store"]`, commandArgName, true),
			binding_constructors.NewStarlarkInstructionKwarg(`0`, nonOptionalExitCodeArgName, false),
			binding_constructors.NewStarlarkInstructionKwarg(`"`+testExecId+`"`, nonOptionalExecIdArgName, true),
		})
	require.Equal(t, canonicalInstruction, execInstruction.GetCanonicalInstruction())
}
