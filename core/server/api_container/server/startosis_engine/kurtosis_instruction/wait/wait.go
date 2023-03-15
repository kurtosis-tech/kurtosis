package wait

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"time"
)

const (
	WaitBuiltinName = "wait"

	ServiceNameArgName = "service_name"
	RecipeArgName      = "recipe"
	ValueFieldArgName  = "field"
	AssertionArgName   = "assertion"
	TargetArgName      = "target_value"
	IntervalArgName    = "interval"
	TimeoutArgName     = "timeout"

	defaultInterval = 1 * time.Second
	defaultTimeout  = 10 * time.Second
)

func NewWait(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: WaitBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RecipeArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					Validator:         nil,
				},
				{
					Name:              ValueFieldArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              AssertionArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         assert.ValidateAssertionToken,
				},
				{
					Name:              TargetArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Comparable],
					Validator:         nil,
				},
				{
					Name:              IntervalArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              TimeoutArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              ServiceNameArgName,
					IsOptional:        true, //TODO make it non-optional when we remove recipe.service_name, issue pending: https://github.com/kurtosis-tech/kurtosis-private/issues/1128
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ServiceNameArgName)
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &WaitCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,

				serviceName: "",  // populated at interpretation time
				recipe:      nil, // populated at interpretation time
				valueField:  "",  // populated at interpretation time
				assertion:   "",  // populated at interpretation time
				target:      nil, // populated at interpretation time
				backoff:     nil, // populated at interpretation time
				timeout:     0,   // populated at interpretation time
				resultUuid:  "",  // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RecipeArgName:     true,
			ValueFieldArgName: true,
			AssertionArgName:  true,
			TargetArgName:     true,
			IntervalArgName:   false,
			TimeoutArgName:    false,
		},
	}
}

type WaitCapabilities struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	serviceName service.ServiceName
	recipe      recipe.Recipe
	valueField  string
	assertion   string
	target      starlark.Comparable
	backoff     backoff.BackOff
	timeout     time.Duration

	resultUuid string
}

func (builtin *WaitCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	var serviceName service.ServiceName

	if arguments.IsSet(ServiceNameArgName) {
		serviceNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
		}
		serviceName = service.ServiceName(serviceNameArgumentValue.GoString())
	}

	var genericRecipe recipe.Recipe
	httpRecipe, err := builtin_argument.ExtractArgumentValue[*recipe.HttpRequestRecipe](arguments, RecipeArgName)
	if err != nil {
		execRecipe, err := builtin_argument.ExtractArgumentValue[*recipe.ExecRecipe](arguments, RecipeArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RecipeArgName)
		}
		genericRecipe = execRecipe
	} else {
		genericRecipe = httpRecipe
	}

	valueField, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ValueFieldArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ValueFieldArgName)
	}

	assertion, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, AssertionArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", AssertionArgName)
	}

	target, err := builtin_argument.ExtractArgumentValue[starlark.Comparable](arguments, TargetArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", TargetArgName)
	}

	var waitBackoff backoff.BackOff
	if arguments.IsSet(IntervalArgName) {
		interval, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, IntervalArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", IntervalArgName)
		}
		parsedDuration, parseErr := time.ParseDuration(interval.GoString())
		if parseErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(parseErr, "An error occurred when parsing interval '%v'", interval.GoString())
		}
		waitBackoff = backoff.NewConstantBackOff(parsedDuration)
	} else {
		waitBackoff = backoff.NewConstantBackOff(defaultInterval)
	}

	var timeout time.Duration
	if arguments.IsSet(TimeoutArgName) {
		starlarkTimeout, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, TimeoutArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", TimeoutArgName)
		}
		parsedTimeout, parseErr := time.ParseDuration(starlarkTimeout.GoString())
		if parseErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(parseErr, "An error occurred when parsing interval '%v'", starlarkTimeout.GoString())
		}
		timeout = parsedTimeout
	} else {
		timeout = defaultTimeout
	}

	resultUuid, err := builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", WaitBuiltinName)
	}

	returnValue, interpretationErr := genericRecipe.CreateStarlarkReturnValue(resultUuid)
	if interpretationErr != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while creating return value for %v instruction", WaitBuiltinName)
	}

	if _, ok := builtin.target.(starlark.Iterable); (builtin.assertion == assert.InCollectionAssertionToken || builtin.assertion == assert.NotInCollectionAssertionToken) && !ok {
		return nil, startosis_errors.NewInterpretationError("'%v' assertion requires an iterable for target values, got '%v'", builtin.assertion, builtin.target.Type())
	}

	builtin.serviceName = serviceName
	builtin.recipe = genericRecipe
	builtin.valueField = valueField.GoString()
	builtin.assertion = assertion.GoString()
	builtin.target = target
	builtin.backoff = waitBackoff
	builtin.timeout = timeout
	builtin.resultUuid = resultUuid

	return returnValue, nil
}

func (builtin *WaitCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO(vcolombo): Add validation step here
	return nil
}

func (builtin *WaitCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	var requestErr error
	var assertErr error
	tries := 0
	timedOut := false
	lastResult := map[string]starlark.Comparable{}
	startTime := time.Now()
	for {
		tries += 1
		backoffDuration := builtin.backoff.NextBackOff()
		if backoffDuration == backoff.Stop || time.Since(startTime) > builtin.timeout {
			timedOut = true
			break
		}
		lastResult, requestErr = builtin.recipe.Execute(ctx, builtin.serviceNetwork, builtin.runtimeValueStore, builtin.serviceName)
		if requestErr != nil {
			time.Sleep(backoffDuration)
			continue
		}
		builtin.runtimeValueStore.SetValue(builtin.resultUuid, lastResult)
		value, found := lastResult[builtin.valueField]
		if !found {
			return "", stacktrace.NewError("Error extracting value from key '%v'", builtin.valueField)
		}
		assertErr = assert.Assert(value, builtin.assertion, builtin.target)
		if assertErr != nil {
			time.Sleep(backoffDuration)
			continue
		}
		break
	}
	if timedOut {
		return "", stacktrace.NewError("Wait timed-out waiting for the assertion to become valid. Waited for '%v'. Last assertion error was: \n%v", time.Since(startTime), assertErr)
	}
	if requestErr != nil {
		return "", stacktrace.Propagate(requestErr, "Error executing HTTP recipe on '%v'", WaitBuiltinName)
	}
	if assertErr != nil {
		return "", stacktrace.Propagate(assertErr, "Error asserting HTTP recipe on '%v'", WaitBuiltinName)
	}
	instructionResult := fmt.Sprintf("Wait took %d tries (%v in total). Assertion passed with following:\n%s", tries, time.Since(startTime), builtin.recipe.ResultMapToString(lastResult))
	return instructionResult, nil
}
