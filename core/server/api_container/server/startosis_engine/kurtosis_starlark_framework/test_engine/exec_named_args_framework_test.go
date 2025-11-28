package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

const (
	execServiceName = service.ServiceName("test-service")
)

var execWithNamedArgsTestService *service.Service = getService(execServiceName)

type execWithNamedArgsTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestExecWithNamedArgs() {
	suite.serviceNetwork.EXPECT().GetService(
		mock.Anything,
		string(execServiceName),
	).Times(1).Return(
		execWithNamedArgsTestService,
		nil,
	)

	suite.serviceNetwork.EXPECT().RunExec(
		mock.Anything,
		string(execServiceName),
		[]string{"mkdir", "-p", "/tmp/store"},
	).Times(1).Return(
		exec_result.NewExecResult(0, ""),
		nil,
	)

	suite.run(&execWithNamedArgsTestCase{
		T:                 suite.T(),
		serviceNetwork:    suite.serviceNetwork,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *execWithNamedArgsTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return exec.NewExec(t.serviceNetwork, t.runtimeValueStore)
}

func (t *execWithNamedArgsTestCase) GetStarlarkCode() string {
	recipe := `ExecRecipe(command=["mkdir", "-p", "/tmp/store"])`
	return fmt.Sprintf("%s(%s=%q, %s=%s)", exec.ExecBuiltinName, exec.ServiceNameArgName, execServiceName, exec.RecipeArgName, recipe)
}

func (t *execWithNamedArgsTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *execWithNamedArgsTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResultMap := `{"code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "output": "{{kurtosis:[0-9a-f]{32}:output.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResultMap, interpretationResult.String())

	require.Equal(t, "Command returned with exit code '0' with no output", *executionResult)
}
