package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type removeServiceTestCase struct {
	*testing.T
}

func newRemoveServiceTestCase(t *testing.T) *removeServiceTestCase {
	return &removeServiceTestCase{
		T: t,
	}
}

func (t removeServiceTestCase) GetId() string {
	return remove_service.RemoveServiceBuiltinName
}

func (t removeServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().RemoveService(
		mock.Anything,
		string(TestServiceName),
	).Times(1).Return(
		TestServiceUuid,
		nil,
	)
	return remove_service.NewRemoveService(serviceNetwork)
}

func (t removeServiceTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", remove_service.RemoveServiceBuiltinName, remove_service.ServiceNameArgName, TestServiceName)
}

func (t *removeServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t removeServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Service '%s' with service UUID '%s' removed", TestServiceName, TestServiceUuid)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}
