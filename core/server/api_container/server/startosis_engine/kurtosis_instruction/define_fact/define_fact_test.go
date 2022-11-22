package define_fact

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testServiceId = "example-service-id"
	testFactName  = "example-fact-name"
)

var (
	emptyFactsEngine *facts_engine.FactsEngine                  = nil
	emptyFactsRecipe *kurtosis_core_rpc_api_bindings.FactRecipe = nil
)

func TestDefineFactInstruction_GetCanonicalizedInstruction(t *testing.T) {
	execInstruction := NewDefineFactInstruction(
		emptyFactsEngine,
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
		testServiceId,
		testFactName,
		emptyFactsRecipe,
	)
	expectedMultiLineFormatStr := `# from: dummyFile[1:1]
define_fact(
	fact_name="%v",
	service_id="%v"
)`
	expectedMultiLineStr := fmt.Sprintf(expectedMultiLineFormatStr, testFactName, testServiceId)
	require.Equal(t, expectedMultiLineStr, execInstruction.GetCanonicalInstruction())

	expectedSingleLineFormatStr := `define_fact(fact_name="%v", service_id="%v")`
	expectedSingleLineStr := fmt.Sprintf(expectedSingleLineFormatStr, testFactName, testServiceId)
	require.Equal(t, expectedSingleLineStr, execInstruction.String())
}
