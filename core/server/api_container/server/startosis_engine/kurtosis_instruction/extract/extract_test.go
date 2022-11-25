package extract

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testUuid           = "88a40d8d-8683-439e-ae16-45ea58b635ae"
	testRuntimeValue   = "{{value}}"
	testFieldExtractor = ".test.test"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

func TestGetValueInstruction_StringRepresentationWorks(t *testing.T) {
	starlarkKwargs := starlark.StringDict{
		"extractor": starlark.String(testFieldExtractor),
		"input":     starlark.String(testRuntimeValue),
	}
	starlarkKwargs.Freeze()
	getValueInstruction := NewExtractInstruction(
		emptyServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		nil,
		testUuid,
		testRuntimeValue,
		recipe.NewExtractRecipe(testFieldExtractor),
		starlarkKwargs,
	)
	expectedStr := `extract(extractor=".test.test", input="{{value}}")`
	require.Equal(t, expectedStr, getValueInstruction.String())
}
