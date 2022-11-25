package extract

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testUuid           = "88a40d8d-8683-439e-ae16-45ea58b635ae"
	testRuntimeValue   = "{{value}}"
	testFieldExtractor = ".test.test"
)

var emptyServiceNetwork = service_network.NewEmptyMockServiceNetwork()

func TestGetValueInstruction_StringRepresentationWorks(t *testing.T) {
	getValueInstruction := NewExtractInstruction(
		emptyServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		nil,
		testUuid,
		testRuntimeValue,
		testFieldExtractor,
		nil,
	)
	expectedStr := `extract(extractor=".test.test", input="{{value}}")`
	require.Equal(t, expectedStr, getValueInstruction.String())
}
