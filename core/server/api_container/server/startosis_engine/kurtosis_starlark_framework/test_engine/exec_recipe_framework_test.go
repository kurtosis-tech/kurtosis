package test_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type execRecipeTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func (suite *KurtosisTypeConstructorTestSuite) TestExecRecipe() {
	suite.serviceNetwork.EXPECT().RunExec(
		mock.Anything,
		string(TestServiceName),
		[]string{"echo", "run"},
	).Times(1).Return(
		exec_result.NewExecResult(0, "run"),
		nil,
	)

	suite.run(&execRecipeTestCase{
		T:                 suite.T(),
		serviceNetwork:    suite.serviceNetwork,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *execRecipeTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return recipe.NewExecRecipeType()
}

func (t *execRecipeTestCase) GetStarlarkCode() string {
	command := fmt.Sprintf("[%q, %q]", "echo", "run")
	return fmt.Sprintf("%s(%s=%s)", recipe.ExecRecipeTypeName, recipe.CommandAttr, command)
}

func (t *execRecipeTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	execRecipe, ok := typeValue.(*recipe.ExecRecipe)
	require.True(t, ok)

	_, err := execRecipe.Execute(context.Background(), t.serviceNetwork, t.runtimeValueStore, TestServiceName)
	require.NoError(t, err)

	returnValue, err := execRecipe.CreateStarlarkReturnValue("result-fake-uuid")
	require.Nil(t, err)
	expectedInterpretationResult := `{"code": "{{kurtosis:result-fake-uuid:code.runtime_value}}", "output": "{{kurtosis:result-fake-uuid:output.runtime_value}}"}`
	require.Equal(t, expectedInterpretationResult, returnValue.String())
}
