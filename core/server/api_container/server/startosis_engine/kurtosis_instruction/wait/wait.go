package wait

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
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
	defaultTimeout  = 15 * time.Minute
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
					IsOptional:        false,
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
				interval:    0,   // populated at interpretation time
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
	interval    time.Duration
	timeout     time.Duration

	resultUuid string
}

func (builtin *WaitCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {

	serviceNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}
	serviceName := service.ServiceName(serviceNameArgumentValue.GoString())

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

	var interval time.Duration
	if arguments.IsSet(IntervalArgName) {
		intervalStr, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, IntervalArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", IntervalArgName)
		}
		parsedInterval, parseErr := time.ParseDuration(intervalStr.GoString())
		if parseErr != nil {
			return nil, startosis_errors.WrapWithInterpretationError(parseErr, "An error occurred when parsing interval '%v'", intervalStr.GoString())
		}
		interval = parsedInterval
	} else {
		interval = defaultInterval
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
	builtin.interval = interval
	builtin.timeout = timeout
	builtin.resultUuid = resultUuid

	return returnValue, nil
}

func (builtin *WaitCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO(vcolombo): Add validation step here
	return nil
}

func (builtin *WaitCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {

	startTime := time.Now()

	lastResult, tries, err := shared_helpers.ExecuteServiceAssertionWithRecipe(
		ctx,
		builtin.serviceNetwork,
		builtin.runtimeValueStore,
		builtin.serviceName,
		builtin.recipe,
		builtin.valueField,
		builtin.assertion,
		builtin.target,
		builtin.interval,
		builtin.timeout,
	)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred checking if service '%v' is ready.",
			builtin.serviceName,
		)
	}

	instructionResult := fmt.Sprintf(
		"Wait took %d tries (%v in total). Assertion passed with following:\n%s",
		tries,
		time.Since(startTime),
		builtin.recipe.ResultMapToString(lastResult),
	)

	return instructionResult, nil
}
