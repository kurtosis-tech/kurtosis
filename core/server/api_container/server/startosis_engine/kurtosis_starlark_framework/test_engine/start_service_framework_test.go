package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/start_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type startServiceTestCase struct {
	*testing.T
}

func newStartServiceTestCase(t *testing.T) *startServiceTestCase {
	return &startServiceTestCase{
		T: t,
	}
}

func (t startServiceTestCase) GetId() string {
	return start_service.StartServiceBuiltinName
}

func (t startServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().StartService(
		mock.Anything,
		string(TestServiceName),
	).Times(1).Return(
		nil,
	)
	return start_service.NewStartService(serviceNetwork)
}

func (t startServiceTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", start_service.StartServiceBuiltinName, start_service.ServiceNameArgName, TestServiceName)
}

func (t *startServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t startServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Service '%s' started", TestServiceName)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}
