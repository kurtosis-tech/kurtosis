package request

import (
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

func TestRequestInstruction_StringRepresentationWorks(t *testing.T) {
	extractor := map[string]string{}
	extractor["key"] = ".value"

	valueInstruction := newEmptyGetValueInstruction(
		emptyServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		defaultRuntimeValueStore,
	)

	testRecipeConfig := recipe.NewGetHttpRequestRecipe(
		"web-server",
		"port_id",
		"/",
		extractor,
	)

	starlarkKwargs := starlark.StringDict{
		"recipe": testRecipeConfig,
	}
	starlarkKwargs.Freeze()

	valueInstruction.starlarkKwargs = starlarkKwargs
	expectedStr := `request(recipe=GetHttpRequestRecipe(port_id="port_id", service_name="web-server", endpoint="/", extract={"key": ".value"}))`
	require.Equal(t, expectedStr, valueInstruction.String())
}
