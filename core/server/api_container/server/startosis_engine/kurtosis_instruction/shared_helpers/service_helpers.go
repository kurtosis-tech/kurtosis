package shared_helpers

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"time"
)

func ExecuteServiceAssertionWithRecipe(
	ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	serviceName service.ServiceName,
	recipe recipe.Recipe,
	valueField string,
	assertion string,
	target starlark.Comparable,
	interval time.Duration,
	timeout time.Duration,
) (map[string]starlark.Comparable, int, error) {
	var recipeErr error
	var assertErr error
	tries := 0
	lastResult := map[string]starlark.Comparable{}
	deadline := time.Now().Add(timeout)
	contextWithDeadline, cancelContext := context.WithDeadline(ctx, deadline)
	defer cancelContext()
	timeoutChan := time.After(timeout)
	tickChan := time.Tick(interval)
	instantFirstRunChan := make(chan bool, 2)
	instantFirstRunChan <- true

	//TODO check if we can refactor this portion in order to use the time.Ticker(backoffDuration) pattern here:
	for {
		tries += 1
		select {
		case <-timeoutChan:
			if recipeErr != nil {
				return lastResult, tries, stacktrace.NewError("Recipe execution timed-out waiting for the recipe execution to become valid on service '%v'. Tried '%v' times. Last recipe execution error was:\n '$%v'\n", serviceName, tries, recipeErr)
			}
			if assertErr != nil {
				return lastResult, tries, stacktrace.NewError("Recipe execution timed-out waiting for the recipe execution to become valid on service '%v'. Tried '%v' times. Last assertion execution error was:\n '$%v'\n", serviceName, tries, assertErr)
			}
			return lastResult, tries, stacktrace.NewError("Recipe execution timed-out waiting for recipe execution on service '%v'. Tries '%v'", serviceName, tries)
		case <-instantFirstRunChan:
			lastResult, recipeErr, assertErr = runRecipe(contextWithDeadline, serviceNetwork, runtimeValueStore, serviceName, recipe, valueField, assertion, target)
			if recipeErr == nil && assertErr == nil {
				return lastResult, tries, nil
			}
		case <-tickChan:
			lastResult, recipeErr, assertErr = runRecipe(contextWithDeadline, serviceNetwork, runtimeValueStore, serviceName, recipe, valueField, assertion, target)
			if recipeErr == nil && assertErr == nil {
				return lastResult, tries, nil
			}
		}
	}
}

func runRecipe(ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	serviceName service.ServiceName,
	recipe recipe.Recipe,
	valueField string,
	assertion string,
	target starlark.Comparable) (map[string]starlark.Comparable, error, error) {
	lastResult, recipeErr := recipe.Execute(ctx, serviceNetwork, runtimeValueStore, serviceName)
	if recipeErr != nil {
		return lastResult, recipeErr, nil
	}
	value, found := lastResult[valueField]
	if !found {
		return lastResult, stacktrace.NewError("Error extracting value from key '%v'. This is a bug in Kurtosis.", valueField), nil
	}
	assertErr := assert.Assert(value, assertion, target)
	if assertErr != nil {
		return lastResult, nil, assertErr
	}
	return lastResult, nil, nil
}
