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

type readyConditionsHttpRecipeTestCase struct {
	*testing.T
}

func (suite *KurtosisTypeConstructorTestSuite) TestReadyConditionsHttpRecipe() {
	suite.run(&readyConditionsHttpRecipeTestCase{
		T: suite.T(),
	})
}

func (t *readyConditionsHttpRecipeTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%s(%s=%q, %s=%q, %s=%s), %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)",
		service_config.ReadyConditionTypeName,
		service_config.RecipeAttr,
		recipe.GetHttpRecipeTypeName,
		recipe.PortIdAttr,
		testReadyConditionsRecipePortId,
		recipe.EndpointAttr,
		testReadyConditionsRecipeEndpoint,
		recipe.ExtractAttr,
		testReadyConditionsRecipeExtract,
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

func (t *readyConditionsHttpRecipeTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	receivedReadyConditions, ok := typeValue.(*service_config.ReadyCondition)
	require.True(t, ok)

	uncastedRecipe, err := receivedReadyConditions.GetRecipe()
	if assert.Nil(t, err) {
		castedRecipe, ok := uncastedRecipe.(recipe.HttpRequestRecipe)
		require.True(t, ok)

		portIdAttrValue, err := castedRecipe.Attr(recipe.PortIdAttr)
		if assert.Nil(t, err) {
			portId, ok := portIdAttrValue.(starlark.String)
			require.True(t, ok)
			require.Equal(t, testReadyConditionsRecipePortId, portId.GoString())
		}

		endpointAttrValue, err := castedRecipe.Attr(recipe.EndpointAttr)
		if assert.Nil(t, err) {
			endpoint, ok := endpointAttrValue.(starlark.String)
			require.True(t, ok)
			require.Equal(t, testReadyConditionsRecipeEndpoint, endpoint.GoString())
		}

		extractAttrValue, err := castedRecipe.Attr(recipe.ExtractAttr)
		if assert.Nil(t, err) {
			extract, ok := extractAttrValue.(*starlark.Dict)
			require.True(t, ok)
			expectedExtractLen := 0
			require.Equal(t, expectedExtractLen, extract.Len())
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
