package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/kurtosis_print"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	printedValue = 24.675432
)

type printTestCase struct {
	*testing.T
}

func newPrintTestCase(t *testing.T) *printTestCase {
	return &printTestCase{
		T: t,
	}
}

func (t *printTestCase) GetId() string {
	return kurtosis_print.PrintBuiltinName
}

func (t *printTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	return kurtosis_print.NewPrint(serviceNetwork, runtimeValueStore)
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
