package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
	"time"
)

type readyConditionsTestCase struct {
	*testing.T
}

func newReadyConditionsTestCase(t *testing.T) *readyConditionsTestCase {
	return &readyConditionsTestCase{
		T: t,
	}
}

func (t *readyConditionsTestCase) GetId() string {
	return service_config.ReadyConditionsTypeName
}

func (t *readyConditionsTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return service_config.NewReadyConditionsType()
}

func (t *readyConditionsTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%s(%s=%q, %s=%q, %s=%s), %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)",
		service_config.ReadyConditionsTypeName,
		service_config.RecipeAttr,
		recipe.GetHttpRecipeTypeName,
		recipe.PortIdAttr,
		TestReadyConditionsRecipePortId,
		recipe.EndpointAttr,
		TestReadyConditionsRecipeEndpoint,
		recipe.ExtractKeyPrefix,
		TestReadyConditionsRecipeExtract,
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

func (t *readyConditionsTestCase) Assert(typeValue starlark.Value) {
	receivedReadyConditions, ok := typeValue.(*service_config.ReadyConditions)
	require.True(t, ok)

	uncastedRecipe, err := receivedReadyConditions.GetRecipe()
	if assert.Nil(t, err) {
		castedRecipe, ok := uncastedRecipe.(*recipe.HttpRequestRecipe)
		require.True(t, ok)

		portIdAttrValue, err := castedRecipe.Attr(recipe.PortIdAttr)
		if assert.Nil(t, err) {
			portId, ok := portIdAttrValue.(starlark.String)
			require.True(t, ok)
			require.Equal(t, TestReadyConditionsRecipePortId, portId.GoString())
		}

		endpointAttrValue, err := castedRecipe.Attr(recipe.EndpointAttr)
		if assert.Nil(t, err) {
			endpoint, ok := endpointAttrValue.(starlark.String)
			require.True(t, ok)
			require.Equal(t, TestReadyConditionsRecipeEndpoint, endpoint.GoString())
		}

		extractAttrValue, err := castedRecipe.Attr(recipe.ExtractKeyPrefix)
		if assert.Nil(t, err) {
			extract, ok := extractAttrValue.(*starlark.Dict)
			require.True(t, ok)
			expectedExtractLen := 0
			require.Equal(t, expectedExtractLen, extract.Len())
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
