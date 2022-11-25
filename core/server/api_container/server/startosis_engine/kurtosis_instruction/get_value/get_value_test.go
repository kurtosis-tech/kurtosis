package get_value

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

const (
	testUuid = "88a40d8d-8683-439e-ae16-45ea58b635ae"
)

var (
	testRecipeConfig = starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"service_id":   starlark.String("web-server"),
		"port_id":      starlark.String("http-port"),
		"endpoint":     starlark.String("/"),
		"method":       starlark.String("POST"),
		"content_type": starlark.String("text/plain"),
		"body":         starlark.String("post_output"),
	})
	emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()
)

func TestGetValueInstruction_StringRepresentationWorks(t *testing.T) {
	starlarkKwargs := starlark.StringDict{
		"recipe": testRecipeConfig,
	}
	starlarkKwargs.Freeze()
	getValueInstruction := NewGetValueInstruction(
		emptyServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		nil,
		nil,
		testRecipeConfig,
		testUuid,
		starlarkKwargs,
	)
	expectedStr := `get_value(recipe=struct(body="post_output", content_type="text/plain", endpoint="/", method="POST", port_id="http-port", service_id="web-server"))`
	require.Equal(t, expectedStr, getValueInstruction.String())
}
