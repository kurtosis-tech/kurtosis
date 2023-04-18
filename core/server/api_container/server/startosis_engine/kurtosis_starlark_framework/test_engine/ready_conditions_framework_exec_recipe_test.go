package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
	"time"
)

type readyConditionsExecRecipeTestCase struct {
	*testing.T
}

func newReadyConditionsExecRecipeTestCase(t *testing.T) *readyConditionsExecRecipeTestCase {
	return &readyConditionsExecRecipeTestCase{
		T: t,
	}
}

func (t *readyConditionsExecRecipeTestCase) GetId() string {
	return service_config.ReadyConditionTypeName
}

func (t *readyConditionsExecRecipeTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%s(%s=%s), %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)",
		service_config.ReadyConditionTypeName,
		service_config.RecipeAttr,
		recipe.ExecRecipeTypeName,
		recipe.CommandAttr,
		fmt.Sprintf("[%q, %q]", TestReadyConditionsRecipeCommand[0], TestReadyConditionsRecipeCommand[1]),
		service_config.FieldAttr,
		TestReadyConditionsField,
		service_config.AssertionAttr,
		TestReadyConditionsAssertion,
		service_config.TargetAttr,
		TestReadyConditionsTarget,
		service_config.IntervalAttr,
		TestReadyConditionsInterval,
		service_config.TimeoutAttr,
		TestReadyConditionsTimeout,
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
			require.Equal(t, command.Len(), len(TestReadyConditionsRecipeCommand))
			for i := 0; i < command.Len(); i++ {
				commandPart, ok := command.Index(i).(starlark.String)
				require.True(t, ok)
				require.Equal(t, TestReadyConditionsRecipeCommand[i], commandPart.GoString())
			}
		}
	}

	field, err := receivedReadyConditions.GetField()
	if assert.Nil(t, err) {
		require.Equal(t, TestReadyConditionsField, field)
	}

	assertion, err := receivedReadyConditions.GetAssertion()
	if assert.Nil(t, err) {
		require.Equal(t, TestReadyConditionsAssertion, assertion)
	}

	target, err := receivedReadyConditions.GetTarget()
	if assert.Nil(t, err) {
		require.Equal(t, TestReadyConditionsTarget, target.String())
	}

	interval, err := receivedReadyConditions.GetInterval()
	if assert.Nil(t, err) {
		expectedInterval, err := time.ParseDuration(TestReadyConditionsInterval)
		if assert.Nil(t, err) {
			require.Equal(t, expectedInterval, interval)
		}
	}

	timeout, err := receivedReadyConditions.GetTimeout()
	if assert.Nil(t, err) {
		expectedTimeout, err := time.ParseDuration(TestReadyConditionsTimeout)
		if assert.Nil(t, err) {
			require.Equal(t, expectedTimeout, timeout)
		}
	}

}
