package wait

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testTargetKey   = "key"
	testAssertion   = "=="
	testTargetValue = "value"
	testDurationStr = "5s"
	testUuid        = "88a40d8d-8683-439e-ae16-45ea58b635ae"
)

var (
	testTarget          = starlark.MakeInt(0)
	emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()
)

func TestWaitInstruction_StringRepresentationWorks(t *testing.T) {
	extractor := map[string]string{}
	extractor[testTargetKey] = ".value"
	testRecipe := recipe.NewPostHttpRequestRecipe(
		"web-server",
		"http-port",
		"text/plain",
		"/",
		"post_body",
		extractor,
	)

	starlarkKwargs := starlark.StringDict{
		"recipe":       testRecipe,
		"assertion":    starlark.String(testAssertion),
		"target_value": testTarget,
		"timeout?":     starlark.String(testDurationStr),
		"interval?":    starlark.String(testDurationStr),
	}
	starlarkKwargs.Freeze()
	getValueInstruction := newWaitInstructionInstructionForTest(
		emptyServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		nil,
		nil,
		testUuid,
		testTargetKey,
		testAssertion,
		starlark.String(testTargetValue),
		starlarkKwargs,
	)
	expectedStr := `wait(assertion="==", interval?="5s", recipe=PostHttpRequestRecipe(port_id="http-port", service_name="web-server", endpoint="/", body="post_body", content_type="text/plain", extract={"key": ".value"}), target_value=0, timeout?="5s")`
	require.Equal(t, expectedStr, getValueInstruction.String())
}
