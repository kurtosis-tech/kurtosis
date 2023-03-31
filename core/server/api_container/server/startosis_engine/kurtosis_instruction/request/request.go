package request

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"net/http"
)

var defaultAcceptableCodes = []int64{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNonAuthoritativeInfo,
	http.StatusNoContent,
	http.StatusResetContent,
	http.StatusPartialContent,
	http.StatusMultiStatus,
	http.StatusAlreadyReported,
	http.StatusIMUsed,
}

const (
	defaultSkipCodeCheck = false
)

const (
	RequestBuiltinName = "request"

	RecipeArgName          = "recipe"
	ServiceNameArgName     = "service_name"
	AcceptableCodesArgName = "acceptable_codes"
	SkipCodeCheckArgName   = "skip_code_check"
)

func NewRequest(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RequestBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ServiceNameArgName)
					},
				},
				{
					Name:              RecipeArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*recipe.HttpRequestRecipe],
					Validator:         nil,
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
			return &RequestCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,

				serviceName:       "",    // populated at interpretation time
				httpRequestRecipe: nil,   // populated at interpretation time
				resultUuid:        "",    // populated at interpretation time
				acceptableCodes:   nil,   // populated at interpretation time
				skipCodeCheck:     false, // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RecipeArgName: true,
		},
	}
}

type RequestCapabilities struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	serviceName       service.ServiceName
	httpRequestRecipe *recipe.HttpRequestRecipe
	resultUuid        string
	acceptableCodes   []int64
	skipCodeCheck     bool
}

func (builtin *RequestCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {

	serviceNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}
	serviceName := service.ServiceName(serviceNameArgumentValue.GoString())

	httpRequestRecipe, err := builtin_argument.ExtractArgumentValue[*recipe.HttpRequestRecipe](arguments, RecipeArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RecipeArgName)
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
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SkipCodeCheckArgName)
		}
		skipCodeCheck = bool(skipCodeCheckArgumentValue)
	}

	resultUuid, err := builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", RequestBuiltinName)
	}

	builtin.serviceName = serviceName
	builtin.httpRequestRecipe = httpRequestRecipe
	builtin.resultUuid = resultUuid
	builtin.acceptableCodes = acceptableCodes
	builtin.skipCodeCheck = skipCodeCheck

	returnValue, interpretationErr := builtin.httpRequestRecipe.CreateStarlarkReturnValue(builtin.resultUuid)
	if interpretationErr != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while creating return value for %v instruction", RequestBuiltinName)
	}
	return returnValue, nil
}

func (builtin *RequestCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, _ *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return nil
}

func (builtin *RequestCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	result, err := builtin.httpRequestRecipe.Execute(ctx, builtin.serviceNetwork, builtin.runtimeValueStore, builtin.serviceName)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error executing http recipe")
	}
	if !builtin.skipCodeCheck && !builtin.isAcceptableCode(result) {
		return "", stacktrace.NewError("Request returned status code '%v' that is not part of the acceptable status codes '%v'", result["code"], builtin.acceptableCodes)
	}
	builtin.runtimeValueStore.SetValue(builtin.resultUuid, result)
	instructionResult := builtin.httpRequestRecipe.ResultMapToString(result)
	return instructionResult, err
}

func (builtin *RequestCapabilities) isAcceptableCode(recipeResult map[string]starlark.Comparable) bool {
	isAcceptableCode := false
	for _, acceptableCode := range builtin.acceptableCodes {
		if recipeResult["code"] == starlark.MakeInt64(acceptableCode) {
			isAcceptableCode = true
			break
		}
	}
	return isAcceptableCode
}
