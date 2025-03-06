package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config/kubernetes"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
)

func getDefaultReadyConditionsScriptPart() string {
	return getCustomReadyConditionsScripPart(
		testReadyConditionsRecipePortId,
		testReadyConditionsRecipeEndpoint,
		testReadyConditionsRecipeExtract,
		testReadyConditionsField,
		testReadyConditionsAssertion,
		testReadyConditionsTarget,
		testReadyConditionsInterval,
		testReadyConditionsTimeout,
	)
}

func getCustomReadyConditionsScripPart(
	portStr string,
	endpointStr string,
	extractStr string,
	fieldStr string,
	assertionStr string,
	targetStr string,
	intervalStr string,
	timeoutStr string,
) string {
	return fmt.Sprintf("%s(%s=%s(%s=%q, %s=%q, %s=%s), %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)",
		service_config.ReadyConditionTypeName,
		service_config.RecipeAttr,
		recipe.GetHttpRecipeTypeName,
		recipe.PortIdAttr,
		portStr,
		recipe.EndpointAttr,
		endpointStr,
		recipe.ExtractAttr,
		extractStr,
		service_config.FieldAttr,
		fieldStr,
		service_config.AssertionAttr,
		assertionStr,
		kubernetes.TargetAttr,
		targetStr,
		service_config.IntervalAttr,
		intervalStr,
		service_config.TimeoutAttr,
		timeoutStr,
	)
}
