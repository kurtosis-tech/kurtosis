package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/start_service"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type startServiceTestCase struct {
	*testing.T
	serviceNetwork *service_network.MockServiceNetwork
}

func (suite *KurtosisPlanInstructionTestSuite) TestStartService() {
	suite.serviceNetwork.EXPECT().StartService(
		mock.Anything,
		string(testServiceName),
	).Times(1).Return(
		nil,
	)

	suite.run(&startServiceTestCase{
		T:              suite.T(),
		serviceNetwork: suite.serviceNetwork,
	})
}

func (t *startServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return start_service.NewStartService(t.serviceNetwork)
}

func (t *startServiceTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", start_service.StartServiceBuiltinName, start_service.ServiceNameArgName, testServiceName)
}

func (t *startServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *startServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Service '%s' started", testServiceName)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}
