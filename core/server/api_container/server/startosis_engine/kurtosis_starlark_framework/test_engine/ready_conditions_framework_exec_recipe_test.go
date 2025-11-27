package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
	"time"
)

type readyConditionsExecRecipeTestCase struct {
	*testing.T
}

func (suite *KurtosisTypeConstructorTestSuite) TestReadyConditionsExecRecipe() {
	suite.run(&readyConditionsExecRecipeTestCase{
		T: suite.T(),
	})
}

func (t *readyConditionsExecRecipeTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%s(%s=%s), %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)",
		service_config.ReadyConditionTypeName,
		service_config.RecipeAttr,
		recipe.ExecRecipeTypeName,
		recipe.CommandAttr,
		fmt.Sprintf("[%q, %q]", testReadyConditionsRecipeCommand[0], testReadyConditionsRecipeCommand[1]),
		service_config.FieldAttr,
		testReadyConditionsField,
		service_config.AssertionAttr,
		testReadyConditionsAssertion,
		service_config.TargetAttr,
		testReadyConditionsTarget,
		service_config.IntervalAttr,
		testReadyConditionsInterval,
		service_config.TimeoutAttr,
		testReadyConditionsTimeout,
	)
}

func (t *readyConditionsExecRecipeTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	receivedReadyConditions, ok := typeValue.(*service_config.ReadyCondition)
	require.True(t, ok)

	uncastedRecipe, err := receivedReadyConditions.GetRecipe()
	if assert.Nil(t, err) {
		castedRecipe, ok := uncastedRecipe.(*recipe.ExecRecipe)
		require.True(t, ok)

		commandAttrValue, err := castedRecipe.Attr(recipe.CommandAttr)
		if assert.Nil(t, err) {
			command, ok := commandAttrValue.(*starlark.List)
			require.True(t, ok)
			require.Equal(t, command.Len(), len(testReadyConditionsRecipeCommand))
			for i := 0; i < command.Len(); i++ {
				commandPart, ok := command.Index(i).(starlark.String)
				require.True(t, ok)
				require.Equal(t, testReadyConditionsRecipeCommand[i], commandPart.GoString())
			}
		}
	}

	field, err := receivedReadyConditions.GetField()
	if assert.Nil(t, err) {
		require.Equal(t, testReadyConditionsField, field)
	}

	assertion, err := receivedReadyConditions.GetAssertion()
	if assert.Nil(t, err) {
		require.Equal(t, testReadyConditionsAssertion, assertion)
	}

	target, err := receivedReadyConditions.GetTarget()
	if assert.Nil(t, err) {
		require.Equal(t, testReadyConditionsTarget, target.String())
	}

	interval, err := receivedReadyConditions.GetInterval()
	if assert.Nil(t, err) {
		expectedInterval, err := time.ParseDuration(testReadyConditionsInterval)
		if assert.Nil(t, err) {
			require.Equal(t, expectedInterval, interval)
		}
	}

	timeout, err := receivedReadyConditions.GetTimeout()
	if assert.Nil(t, err) {
		expectedTimeout, err := time.ParseDuration(testReadyConditionsTimeout)
		if assert.Nil(t, err) {
			require.Equal(t, expectedTimeout, timeout)
		}
	}

}
