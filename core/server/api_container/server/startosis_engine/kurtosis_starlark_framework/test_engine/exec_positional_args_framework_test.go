package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/benchmark"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

type execWithPositionalArgsTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

// func (suite *KurtosisPlanInstructionTestSuite) TestExecWithPositionalArgs() {
// 	suite.serviceNetwork.EXPECT().RunExec(
// 		mock.Anything,
// 		string(execServiceName),
// 		[]string{"mkdir", "-p", "/tmp/store"},
// 	).Times(1).Return(
// 		exec_result.NewExecResult(0, ""),
// 		nil,
// 	)

// 	suite.run(&execWithPositionalArgsTestCase{
// 		T:                 suite.T(),
// 		serviceNetwork:    suite.serviceNetwork,
// 		runtimeValueStore: suite.runtimeValueStore,
// 	})
// }

func (t *execWithPositionalArgsTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return exec.NewExec(t.serviceNetwork, t.runtimeValueStore, &benchmark.KurtosisPlanInstructionBenchmark{})
}

func (t *execWithPositionalArgsTestCase) GetStarlarkCode() string {
	recipe := `ExecRecipe(command=["mkdir", "-p", "/tmp/store"])`
	return fmt.Sprintf("%s(%q, %s)", exec.ExecBuiltinName, execServiceName, recipe)
}

func (t *execWithPositionalArgsTestCase) GetStarlarkCodeForAssertion() string {
	recipe := `ExecRecipe(command=["mkdir", "-p", "/tmp/store"])`
	return fmt.Sprintf("%s(%s=%q, %s=%s)", exec.ExecBuiltinName, exec.ServiceNameArgName, execServiceName, exec.RecipeArgName, recipe)
}

func (t *execWithPositionalArgsTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResultMap := `{"code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "output": "{{kurtosis:[0-9a-f]{32}:output.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResultMap, interpretationResult.String())

	require.Equal(t, "Command returned with exit code '0' with no output", *executionResult)
}
