package assert

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testRuntimeValue = "{{runtime-value}}"
	testAssertion    = "=="
)

var (
	testTarget             = starlark.MakeInt(0)
	emptyRuntimeValueStore *runtime_value_store.RuntimeValueStore
)

func TestAssertInstruction_StringRepresentationWorks(t *testing.T) {
	starlarkKwargs := starlark.StringDict{
		"value":        starlark.String(testRuntimeValue),
		"assertion":    starlark.String(testAssertion),
		"target_value": testTarget,
	}
	starlarkKwargs.Freeze()
	assertInstruction := NewAssertInstruction(
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		emptyRuntimeValueStore,
		testRuntimeValue,
		testAssertion,
		testTarget,
		starlarkKwargs,
	)
	expectedStr := `assert(assertion="==", target_value=0, value="{{runtime-value}}")`
	require.Equal(t, expectedStr, assertInstruction.String())
}
