package request

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

const testUuid = "88a40d8d-8683-439e-ae16-45ea58b635ae"

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

func TestRequestInstruction_StringRepresentationWorks(t *testing.T) {
	extractor := &starlark.Dict{}
	err := extractor.SetKey(starlark.String("key"), starlark.String(".value"))
	require.Nil(t, err)
	testRecipeConfig := starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"service_name": starlark.String("web-server"),
		"port_id":      starlark.String("http-port"),
		"endpoint":     starlark.String("/"),
		"method":       starlark.String("POST"),
		"content_type": starlark.String("text/plain"),
		"body":         starlark.String("post_output"),
		"extract":      extractor,
	})
	starlarkKwargs := starlark.StringDict{
		"recipe": testRecipeConfig,
	}
	starlarkKwargs.Freeze()
	getValueInstruction := NewRequestInstruction(
		emptyServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		nil,
		nil,
		testRecipeConfig,
		testUuid,
		starlarkKwargs,
	)
	expectedStr := `request(recipe=struct(body="post_output", content_type="text/plain", endpoint="/", extract={"key": ".value"}, method="POST", port_id="http-port", service_name="web-server"))`
	require.Equal(t, expectedStr, getValueInstruction.String())
}
