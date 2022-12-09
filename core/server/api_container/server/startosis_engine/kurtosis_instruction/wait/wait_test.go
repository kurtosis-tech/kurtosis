package wait

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

const testUuid = "88a40d8d-8683-439e-ae16-45ea58b635ae"

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

const (
	testTargetKey   = "key"
	testAssertion   = "=="
	testTargetValue = "value"
	testDurationStr = "5s"
)

var (
	testTarget             = starlark.MakeInt(0)
	emptyRuntimeValueStore *runtime_value_store.RuntimeValueStore
)

func TestWaitInstruction_StringRepresentationWorks(t *testing.T) {
	extractor := &starlark.Dict{}
	err := extractor.SetKey(starlark.String(testTargetKey), starlark.String(".value"))
	require.Nil(t, err)
	testRecipeConfig := starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"service_id":   starlark.String("web-server"),
		"port_id":      starlark.String("http-port"),
		"endpoint":     starlark.String("/"),
		"method":       starlark.String("POST"),
		"content_type": starlark.String("text/plain"),
		"body":         starlark.String("post_output"),
		"extract":      extractor,
	})
	starlarkKwargs := starlark.StringDict{
		"recipe":       testRecipeConfig,
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
	expectedStr := `wait(assertion="==", interval?="5s", recipe=struct(body="post_output", content_type="text/plain", endpoint="/", extract={"key": ".value"}, method="POST", port_id="http-port", service_id="web-server"), target_value=0, timeout?="5s")`
	require.Equal(t, expectedStr, getValueInstruction.String())
}
