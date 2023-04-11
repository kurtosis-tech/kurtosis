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
	/*
		We would like to kick an execution right away and after that retry every 'interval' seconds,
		considering time that took the request to complete.
		So we prepend an element to 'tickChan'
	*/
	tickChan := time.NewTicker(interval)
	executionTickChan := make(chan time.Time, 2)
	executionTickChan <- time.Now()
	go func() {
		for tick := range tickChan.C {
			executionTickChan <- tick
		}
	}()
	/*
		By passing 'contextWithDeadline' to recipe execution, we can make sure that when timeout is reached,
		the underlying request is aborted.
		'timeoutChan' serves as an exit signal for the loop repeating the recipe execution
	*/
	contextWithDeadline, cancelContext := context.WithTimeout(ctx, timeout)
	defer cancelContext()
	timeoutChan := time.After(timeout)

	execFunc := func() (map[string]starlark.Comparable, error) {
		lastResult, recipeErr := recipe.Execute(contextWithDeadline, serviceNetwork, runtimeValueStore, serviceName)
		if recipeErr != nil {
			return lastResult, recipeErr
		}
		_, found := lastResult[valueField]
		if !found {
			return lastResult, stacktrace.NewError("Error extracting value from key '%v'. This is a bug in Kurtosis.", valueField)
		}
		return lastResult, nil
	}
	assertFunc := func(lastResult map[string]starlark.Comparable) error {
		assertErr := assert.Assert(lastResult[valueField], assertion, target)
		if assertErr != nil {
			return assertErr
		}
		return nil
	}
	return executeServiceAssertionWithRecipeWithTicker(serviceName, execFunc, assertFunc, executionTickChan, timeoutChan)

}

/*
Executes 'execFunc':
  - If it errors, retry after the next tick from 'executionTickChan'.
  - If it succeeds, executes result with 'assertFunc':
    -- If it succeeds, returns.
    -- If it errors, retry after the next tick from 'executionTickChan'

If a signal is sent to 'timeoutChan', loop will be broken, last value is returned,
alongside if the last error (assert or exec)
*/
func executeServiceAssertionWithRecipeWithTicker(
	serviceName service.ServiceName,
	execFunc func() (map[string]starlark.Comparable, error),
	assertFunc func(map[string]starlark.Comparable) error,
	executionTickChan <-chan time.Time,
	timeoutChan <-chan time.Time,
) (map[string]starlark.Comparable, int, error) {
	var recipeErr error
	var assertErr error
	tries := 0
	lastResult := map[string]starlark.Comparable{}

	for {
		select {
		case <-timeoutChan:
			if recipeErr != nil {
				return lastResult, tries, stacktrace.NewError("Recipe execution timed-out waiting for the recipe execution to become valid on service '%v'. Tried '%v' times. Last recipe execution error was:\n '$%v'\n", serviceName, tries, recipeErr)
			}
			if assertErr != nil {
				return lastResult, tries, stacktrace.NewError("Recipe execution timed-out waiting for the recipe execution to become valid on service '%v'. Tried '%v' times. Last assertion execution error was:\n '$%v'\n", serviceName, tries, assertErr)
			}
			return lastResult, tries, stacktrace.NewError("Recipe execution timed-out but no errors of assert and recipe happened on service '%v'. This is a bug in Kurtosis.", serviceName)
		case <-executionTickChan:
			tries += 1
			lastResult, recipeErr = execFunc()
			if recipeErr != nil {
				continue
			}
			assertErr = assertFunc(lastResult)
			if assertErr != nil {
				continue
			}
			return lastResult, tries, nil
		}
	}
}
