package define_fact

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

const (
	testServiceId = "example-service-id"
	testFactName  = "example-fact-name"
)

var (
	emptyFactsEngine *facts_engine.FactsEngine = nil
)

func TestDefineFactInstruction_GetCanonicalizedInstruction(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile")
	defineFactInstruction := newEmptyDefineFactInstruction(
		emptyFactsEngine,
		position,
	)
	defineFactInstruction.starlarkKwargs = starlark.StringDict{}
	defineFactInstruction.starlarkKwargs[serviceIdArgName] = starlark.String(testServiceId)
	defineFactInstruction.starlarkKwargs[factNameArgName] = starlark.String(testFactName)
	defineFactInstruction.starlarkKwargs[recipeArgName] = starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"method":   starlark.String("GET"),
		"endpoint": starlark.String("my_endpoint"),
	})

	expectedFormatStr := `define_fact(fact_name="%v", fact_recipe=struct(endpoint="my_endpoint", method="GET"), service_id="%v")`
	expectedStr := fmt.Sprintf(expectedFormatStr, testFactName, testServiceId)
	require.Equal(t, expectedStr, defineFactInstruction.String())

	canonicalInstruction := binding_constructors.NewKurtosisInstruction(
		position.ToAPIType(),
		DefineFactBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{
			binding_constructors.NewKurtosisInstructionKwarg(`"example-service-id"`, serviceIdArgName, true),
			binding_constructors.NewKurtosisInstructionKwarg(`"example-fact-name"`, factNameArgName, true),
			binding_constructors.NewKurtosisInstructionKwarg(`struct(endpoint="my_endpoint", method="GET")`, recipeArgName, false),
		})
	require.Equal(t, canonicalInstruction, defineFactInstruction.GetCanonicalInstruction())
}
