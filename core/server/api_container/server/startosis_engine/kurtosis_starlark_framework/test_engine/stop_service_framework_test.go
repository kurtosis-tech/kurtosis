package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/stop_service"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type stopServiceTestCase struct {
	*testing.T
	serviceNetwork *service_network.MockServiceNetwork
}

func (suite *KurtosisPlanInstructionTestSuite) TestStopService() {
	suite.serviceNetwork.EXPECT().StopService(
		mock.Anything,
		string(testServiceName),
	).Times(1).Return(
		nil,
	)

	suite.run(&stopServiceTestCase{
		T:              suite.T(),
		serviceNetwork: suite.serviceNetwork,
	})
}

func (t *stopServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return stop_service.NewStopService(t.serviceNetwork)
}

func (t *stopServiceTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", stop_service.StopServiceBuiltinName, stop_service.ServiceNameArgName, testServiceName)
}

func (t *stopServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *stopServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Service '%s' stopped", testServiceName)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}
