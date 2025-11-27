package test_engine

import (
	"context"
	"fmt"
	"testing"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type execRecipeTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func (suite *KurtosisTypeConstructorTestSuite) TestExecRecipe() {
	suite.serviceNetwork.EXPECT().GetService(
		mock.Anything,
		string(testServiceName),
	).Times(1).Return(
		getService(testServiceName),
		nil,
	)

	suite.serviceNetwork.EXPECT().RunExec(
		mock.Anything,
		string(testServiceName),
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

	service, err := t.serviceNetwork.GetService(context.Background(), string(testServiceName))
	require.NoError(t, err)

	_, err = execRecipe.Execute(context.Background(), t.serviceNetwork, t.runtimeValueStore, service)
	require.NoError(t, err)

	returnValue, err := execRecipe.CreateStarlarkReturnValue("result-fake-uuid")
	require.Nil(t, err)
	expectedInterpretationResult := `{"code": "{{kurtosis:result-fake-uuid:code.runtime_value}}", "output": "{{kurtosis:result-fake-uuid:output.runtime_value}}"}`
	require.Equal(t, expectedInterpretationResult, returnValue.String())
}
