package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/stop_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type stopServiceTestCase struct {
	*testing.T
}

func newStopServiceTestCase(t *testing.T) *stopServiceTestCase {
	return &stopServiceTestCase{
		T: t,
	}
}

func (t stopServiceTestCase) GetId() string {
	return stop_service.StopServiceBuiltinName
}

func (t stopServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().StopService(
		mock.Anything,
		string(TestServiceName),
	).Times(1).Return(
		nil,
	)
	return stop_service.NewStopService(serviceNetwork)
}

func (t stopServiceTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", stop_service.StopServiceBuiltinName, stop_service.ServiceNameArgName, TestServiceName)
}

func (t *stopServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t stopServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Service '%s' stopped", TestServiceName)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}
