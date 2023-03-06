package shared_helpers

import (
	"context"
	"github.com/cenkalti/backoff/v4"
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
) error {
	var requestErr error
	var assertErr error
	tries := 0
	timedOut := false
	lastResult := map[string]starlark.Comparable{}
	startTime := time.Now()

	backoffObj := backoff.NewConstantBackOff(interval)

	for {
		tries += 1
		backoffDuration := backoffObj.NextBackOff()
		if backoffDuration == backoff.Stop || time.Since(startTime) > timeout {
			timedOut = true
			break
		}
		lastResult, requestErr = recipe.Execute(ctx, serviceNetwork, runtimeValueStore, serviceName)
		if requestErr != nil {
			time.Sleep(backoffDuration)
			continue
		}
		value, found := lastResult[valueField]
		if !found {
			return stacktrace.NewError("Error extracting value from key '%v'", valueField)
		}
		assertErr = assert.Assert(value, assertion, target)
		if assertErr != nil {
			time.Sleep(backoffDuration)
			continue
		}
		break
	}
	if timedOut {
		return stacktrace.NewError("Wait timed-out waiting for the assertion to become valid on service '%v'. Waited for '%v'. Last assertion error was: \n%v", serviceName, time.Since(startTime), assertErr)
	}
	if requestErr != nil {
		return stacktrace.Propagate(requestErr, "Error executing HTTP recipe on service '%v'", serviceName)
	}
	if assertErr != nil {
		return stacktrace.Propagate(assertErr, "Error asserting HTTP recipe on service '%v'", serviceName)
	}

	return nil
}
