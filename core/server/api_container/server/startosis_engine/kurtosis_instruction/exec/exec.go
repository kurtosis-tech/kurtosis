package exec

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

var defaultAcceptableCodes = []int{
	0, // EXIT_SUCCESS
}

const (
	ExecBuiltinName = "exec"

	RecipeArgName          = "recipe"
	ServiceNameArgName     = "service_name"
	AcceptableCodesArgName = "acceptable_codes"
	SkipCodeCheckArgName   = "skip_code_check"
)

const (
	defaultSkipCodeCheck = false
)

func NewExec(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ExecBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RecipeArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*recipe.ExecRecipe],
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
				{
					Name:              AcceptableCodesArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator:         nil,
				},
				{
					Name:              SkipCodeCheckArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Bool],
					Validator:         nil,
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &ExecCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,

				serviceName:     "",    // will be populated at interpretation time
				execRecipe:      nil,   // will be populated at interpretation time
				resultUuid:      "",    // will be populated at interpretation time
				acceptableCodes: nil,   // will be populated at interpretation time
				skipCodeCheck:   false, // will be populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RecipeArgName: true,
		},
	}
}

type ExecCapabilities struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	serviceName     service.ServiceName
	execRecipe      *recipe.ExecRecipe
	resultUuid      string
	acceptableCodes []int
	skipCodeCheck   bool
}

func (builtin *ExecCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	execRecipe, err := builtin_argument.ExtractArgumentValue[*recipe.ExecRecipe](arguments, RecipeArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RecipeArgName)
	}

	var serviceName service.ServiceName
	if arguments.IsSet(ServiceNameArgName) {
		serviceNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
		}
		serviceName = service.ServiceName(serviceNameArgumentValue.GoString())
	} else if execRecipe.GetServiceName() != shared_helpers.EmptyServiceName {
		serviceName = execRecipe.GetServiceName()
		logrus.Warnf("The recipe.service_name field will be deprecated soon, users will have to pass the service name value direclty to the 'exec', 'request' and 'wait' instructions")
	} else {
		return nil, startosis_errors.NewInterpretationError("Service name is not set, either as an exec instruction's argument or as a recipe field. You can fix it passing the 'service_name' argument in the 'exec' call")
	}

	acceptableCodes := defaultAcceptableCodes
	if arguments.IsSet(AcceptableCodesArgName) {
		acceptableCodesValue, err := builtin_argument.ExtractArgumentValue[*starlark.List](arguments, AcceptableCodesArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%v' argument", acceptableCodes)
		}
		acceptableCodes, err = kurtosis_types.SafeCastToIntegerSlice(acceptableCodesValue)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to parse '%v' argument", acceptableCodes)
		}
	}

	skipCodeCheck := defaultSkipCodeCheck
	if arguments.IsSet(SkipCodeCheckArgName) {
		skipCodeCheckArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.Bool](arguments, SkipCodeCheckArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
		}
		skipCodeCheck = bool(skipCodeCheckArgumentValue)
	}

	resultUuid, err := builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", ExecBuiltinName)
	}

	builtin.serviceName = serviceName
	builtin.execRecipe = execRecipe
	builtin.resultUuid = resultUuid
	builtin.acceptableCodes = acceptableCodes
	builtin.skipCodeCheck = skipCodeCheck

	returnValue, interpretationErr := builtin.execRecipe.CreateStarlarkReturnValue(builtin.resultUuid)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while generating return value for %v instruction", ExecBuiltinName)
	}
	return returnValue, nil
}

func (builtin *ExecCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, _ *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO: validate recipe
	return nil
}

func (builtin *ExecCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	result, err := builtin.execRecipe.Execute(ctx, builtin.serviceNetwork, builtin.runtimeValueStore, builtin.serviceName)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error executing exec recipe")
	}
	if !builtin.skipCodeCheck && !builtin.isAcceptableCode(result) {
		return "", stacktrace.NewError("Exec returned status code '%v' that is not part of the acceptable status codes '%v'", result["code"], builtin.acceptableCodes)
	}
	builtin.runtimeValueStore.SetValue(builtin.resultUuid, result)
	instructionResult := builtin.execRecipe.ResultMapToString(result)
	return instructionResult, err
}

func (builtin *ExecCapabilities) isAcceptableCode(recipeResult map[string]starlark.Comparable) bool {
	isAcceptableCode := false
	for _, acceptableCode := range builtin.acceptableCodes {
		if recipeResult["code"] == starlark.MakeInt(acceptableCode) {
			isAcceptableCode = true
			break
		}
	}
	return isAcceptableCode
}
