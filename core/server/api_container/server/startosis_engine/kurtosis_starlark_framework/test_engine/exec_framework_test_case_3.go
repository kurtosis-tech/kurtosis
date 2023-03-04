package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	execTextCase3ServiceName = service.ServiceName("my-service-for-test-case-2")
)

//For a short period (until we deprecate recipe.service_name) the exec instruction will have a
//dynamic first parameter which will accept the current 'recipe' argument and a new 'service_name' argument
//In the execTestCase1 we test the current behaviour, it means receiving an 'recipe' as the first argument
//In this test case we test that 'service_name' is also accepted as the first parameter, and it is used in the exec call
type execTestCase3 struct {
	*testing.T
}

func newExecTestCase3(t *testing.T) *execTestCase3 {
	return &execTestCase3{
		T: t,
	}
}

func (t execTestCase3) GetId() string {
	return exec.ExecBuiltinName
}

func (t execTestCase3) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	serviceNetwork.EXPECT().ExecCommand(
		mock.Anything,
		string(execTextCase3ServiceName),
		[]string{"mkdir", "-p", "/tmp/store"},
	).Times(1).Return(
		int32(0),
		"",
		nil,
	)

	return exec.NewExec(serviceNetwork, runtimeValueStore)
}

func (t execTestCase3) GetStarlarkCode() string {
	recipe := `ExecRecipe(command=["mkdir", "-p", "/tmp/store"])`
	return fmt.Sprintf("%s(%s=%s, %s=%q)", exec.ExecBuiltinName, exec.RecipeArgName, recipe, exec.ServiceNameArgName, execTextCase3ServiceName)
}

func (t *execTestCase3) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t execTestCase3) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResultMap := `{"code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "output": "{{kurtosis:[0-9a-f]{32}:output.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResultMap, interpretationResult.String())

	require.Equal(t, "Command returned with exit code '0' with no output", *executionResult)
}
