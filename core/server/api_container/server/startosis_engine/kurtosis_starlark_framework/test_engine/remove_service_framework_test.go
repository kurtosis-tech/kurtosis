package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type removeServiceTestCase struct {
	*testing.T
	serviceNetwork          *service_network.MockServiceNetwork
	interpretationTimeStore *interpretation_time_value_store.InterpretationTimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestRemoveService() {
	suite.serviceNetwork.EXPECT().RemoveService(
		mock.Anything,
		string(testServiceName),
	).Times(1).Return(
		testServiceUuid,
		nil,
	)

	suite.run(&removeServiceTestCase{
		T:                       suite.T(),
		serviceNetwork:          suite.serviceNetwork,
		interpretationTimeStore: suite.interpretationTimeValueStore,
	})
}

func (t *removeServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return remove_service.NewRemoveService(t.serviceNetwork, t.interpretationTimeStore)
}

func (t *removeServiceTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", remove_service.RemoveServiceBuiltinName, remove_service.ServiceNameArgName, testServiceName)
}

func (t *removeServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *removeServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Service '%s' with service UUID '%s' removed", testServiceName, testServiceUuid)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}
