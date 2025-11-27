package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/kurtosis_print"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	printedValue = 24.675432
)

type printTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestPrint() {
	suite.run(&printTestCase{
		T:                 suite.T(),
		serviceNetwork:    suite.serviceNetwork,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *printTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return kurtosis_print.NewPrint(t.serviceNetwork, t.runtimeValueStore)
}

func (t *printTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%f)", kurtosis_print.PrintBuiltinName, kurtosis_print.PrintArgName, printedValue)
}

func (t *printTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := "24.675432"
	require.Equal(t, &expectedExecutionResult, executionResult)
}

func (t *printTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}
