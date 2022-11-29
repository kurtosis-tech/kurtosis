package wait

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testServiceId = "example-service-id"
	testFactName  = "example-fact-name"
)

var (
	emptyFactsEngine *facts_engine.FactsEngine = nil
)

func TestWaitInstruction_GetCanonicalizedInstruction(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile")
	waitInstruction := newEmptyWaitInstruction(emptyFactsEngine, position)
	waitInstruction.starlarkKwargs = starlark.StringDict{}
	waitInstruction.starlarkKwargs[serviceIdArgName] = starlark.String(testServiceId)
	waitInstruction.starlarkKwargs[factNameArgName] = starlark.String(testFactName)

	expectedFormatStr := `wait(fact_name="%v", service_id="%v")`
	expectedStr := fmt.Sprintf(expectedFormatStr, testFactName, testServiceId)
	require.Equal(t, expectedStr, waitInstruction.String())

	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		position.ToAPIType(),
		WaitBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(`"`+string(testServiceId)+`"`, serviceIdArgName, true),
			binding_constructors.NewStarlarkInstructionKwarg(`"`+testFactName+`"`, factNameArgName, true),
		})
	require.Equal(t, canonicalInstruction, waitInstruction.GetCanonicalInstruction())
}
