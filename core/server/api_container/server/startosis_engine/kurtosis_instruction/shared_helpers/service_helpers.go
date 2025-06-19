package shared_helpers

import (
	"context"
	"reflect"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/verify"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	bufferedChannelSize     = 2
	starlarkThreadName      = "starlark-value-serde-for-test-thread"
	configArgName           = "config"
	timeoutExtensionDivider = 4
)

func NewDummyStarlarkValueSerDeForTest() *kurtosis_types.StarlarkValueSerde {
	starlarkThread := &starlark.Thread{
		Name:       starlarkThreadName,
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := starlark.StringDict{}

	serde := kurtosis_types.NewStarlarkValueSerde(starlarkThread, starlarkEnv)

	return serde
}

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
	executionTickChan := make(chan time.Time, bufferedChannelSize)
	executionTickChan <- time.Now()
	go func() {
		for tick := range tickChan.C {
			executionTickChan <- tick
		}
	}()

	// 'timeoutChan' serves as an exit signal for the loop repeating the recipe execution.
	timeoutChan := time.After(timeout)

	// By passing 'contextWithDeadline' to recipe execution, we can make sure that when timeout is reached, the underlying
	// request is aborted.

	// tedi(07-19-25): added a buffer to the context timeout to ensure the exit from timeoutChan is received before the context deadline is exceeded.
	// otherwise, a race condition occurs where the context deadline is exceeded before the timeoutChan is received occurs and the code exits with a context deadline exceeded error
	// not too sure if this is the best way to do it/why the ctxWithDeadline is needed but fixes flaky CI test for now - more context in original PR for this code here:
	// https://github.com/kurtosis-tech/kurtosis/pull/480
	ctxWithDeadline, cancelContext := context.WithTimeout(ctx, timeout+timeout/timeoutExtensionDivider)
	defer cancelContext()

	execFunc := func() (map[string]starlark.Comparable, error) {
		return execRequestAndGetValue(ctxWithDeadline, serviceNetwork, runtimeValueStore, serviceName, recipe, valueField)
	}
	assertFunc := func(currentResult map[string]starlark.Comparable) error {
		return assertResult(currentResult[valueField], assertion, target)
	}
	return executeServiceAssertionWithRecipeWithTicker(serviceName, execFunc, assertFunc, executionTickChan, timeoutChan)

}

func assertResult(currentResult starlark.Comparable, assertion string, target starlark.Comparable) error {
	err := verify.Verify(currentResult, assertion, target)
	if err != nil {
		return err
	}
	return nil
}

func execRequestAndGetValue(ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	serviceName service.ServiceName,
	recipe recipe.Recipe,
	valueField string) (map[string]starlark.Comparable, error) {
	resultMap, err := recipe.Execute(ctx, serviceNetwork, runtimeValueStore, serviceName)
	if err != nil {
		return resultMap, err
	}
	_, found := resultMap[valueField]
	if !found {
		return resultMap, stacktrace.NewError("Error extracting value from key '%v'. This is a bug in Kurtosis.", valueField)
	}
	return resultMap, nil
}

/*
Executes 'execFunc':
  - If it errors, retry after the next tick from 'executionTickChan'.
  - If it succeeds, executes result with 'assertFunc':
    -- If it succeeds, returns.
    -- If it errors, retry after the next tick from 'executionTickChan'

If a signal is sent to 'interruptChan', loop will be broken, last value is returned,
alongside if the last error (assert or exec)

Returns the last output of the exec, number of tries before return and the error (if any)
*/
func executeServiceAssertionWithRecipeWithTicker(
	serviceName service.ServiceName,
	execFunc func() (map[string]starlark.Comparable, error),
	assertFunc func(map[string]starlark.Comparable) error,
	executionTickChan <-chan time.Time,
	interruptChan <-chan time.Time,
) (map[string]starlark.Comparable, int, error) {
	var recipeErr error
	var assertErr error
	tries := 0
	lastResult := map[string]starlark.Comparable{}

	for {
		select {
		case <-interruptChan:
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

func ValidateAndConvertConfigAndReadyCondition(
	serviceNetwork service_network.ServiceNetwork,
	rawConfig starlark.Value,
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
	imageDownloadMode image_download_mode.ImageDownloadMode,
) (*service.ServiceConfig, *service_config.ReadyCondition, *startosis_errors.InterpretationError) {
	config, ok := rawConfig.(*service_config.ServiceConfig)
	if !ok {
		return nil, nil, startosis_errors.NewInterpretationError("The '%s' argument is not a ServiceConfig (was '%s').", configArgName, reflect.TypeOf(rawConfig))
	}
	apiServiceConfig, interpretationErr := config.ToKurtosisType(
		serviceNetwork,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		packageId,
		packageContentProvider,
		packageReplaceOptions,
		imageDownloadMode)
	if interpretationErr != nil {
		return nil, nil, interpretationErr
	}

	readyCondition, interpretationErr := config.GetReadyCondition()
	if interpretationErr != nil {
		return nil, nil, interpretationErr
	}

	return apiServiceConfig, readyCondition, nil
}
