package exec

import (
	"context"
	"fmt"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/tasks"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

var defaultAcceptableCodes = []int64{
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
	descriptionFormatStr = "Executing command on service '%v'"
)

func NewExec(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ExecBuiltinName,
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
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*recipe.ExecRecipe],
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
			return &ExecCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,

				serviceName:     "",         // will be populated at interpretation time
				execRecipe:      nil,        // will be populated at interpretation time
				resultUuid:      "",         // will be populated at interpretation time
				acceptableCodes: nil,        // will be populated at interpretation time
				skipCodeCheck:   false,      // will be populated at interpretation time
				description:     "",         // populated at interpretation time
				cmdList:         []string{}, // populated at interpretation time
				returnValue:     nil,        // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RecipeArgName:      true,
			ServiceNameArgName: true,
		},
	}
}

type ExecCapabilities struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	serviceName     service.ServiceName
	execRecipe      *recipe.ExecRecipe
	cmdList         []string
	resultUuid      string
	acceptableCodes []int64
	skipCodeCheck   bool
	description     string

	returnValue *starlark.Dict
}

func (builtin *ExecCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {

	serviceNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}
	serviceName := service.ServiceName(serviceNameArgumentValue.GoString())

	execRecipe, err := builtin_argument.ExtractArgumentValue[*recipe.ExecRecipe](arguments, RecipeArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RecipeArgName)
	}
	starlarkCmdList, ok, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](execRecipe.KurtosisValueTypeDefault, recipe.CommandAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !ok {
		return nil, startosis_errors.NewInterpretationError("Unable to extract attribute %v off of exec recipe.", recipe.CommandAttr)
	}
	cmdList, interpretationErr := kurtosis_types.SafeCastToStringSlice(starlarkCmdList, tasks.PythonArgumentsArgName)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred converting Starlark list of passed arguments to Go string slice")
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
		return nil, startosis_errors.NewInterpretationError("An error occurred while generating UUID for future reference for %v instruction", ExecBuiltinName)
	}

	builtin.serviceName = serviceName
	builtin.execRecipe = execRecipe
	builtin.cmdList = cmdList
	builtin.resultUuid = resultUuid
	builtin.acceptableCodes = acceptableCodes
	builtin.skipCodeCheck = skipCodeCheck

	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.serviceName))

	builtin.returnValue, interpretationErr = builtin.execRecipe.CreateStarlarkReturnValue(builtin.resultUuid)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while generating return value for %v instruction", ExecBuiltinName)
	}
	return builtin.returnValue, nil
}

func (builtin *ExecCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, _ *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// TODO: validate recipe
	return nil
}

func (builtin *ExecCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	service, err := builtin.serviceNetwork.GetService(ctx, string(builtin.serviceName))
	if err != nil {
		return "", stacktrace.Propagate(err, "Error getting service '%s'", builtin.serviceName)
	}
	result, err := builtin.execRecipe.Execute(ctx, builtin.serviceNetwork, builtin.runtimeValueStore, service)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error executing exec recipe")
	}
	if !builtin.skipCodeCheck && !builtin.isAcceptableCode(result) {
		errorMessage := fmt.Sprintf("Exec returned exit code '%v' that is not part of the acceptable status codes '%v', with output:", result["code"], builtin.acceptableCodes)
		return "", stacktrace.NewError(formatErrorMessage(errorMessage, result["output"].String()))
	}

	if err := builtin.runtimeValueStore.SetValue(builtin.resultUuid, result); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred setting value '%+v' using key UUID '%s' in the runtime value store", result, builtin.resultUuid)
	}
	instructionResult := builtin.execRecipe.ResultMapToString(result)
	return instructionResult, err
}

func (builtin *ExecCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual && enclaveComponents.HasServiceBeenUpdated(builtin.serviceName) {
		return enclave_structure.InstructionIsUpdate
	} else if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *ExecCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(ExecBuiltinName)
}

func (builtin *ExecCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYamlGenerator) error {
	err := planYaml.AddExec(string(builtin.serviceName), builtin.description, builtin.returnValue, builtin.cmdList, builtin.acceptableCodes)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred updating plan with exec.")
	}
	return nil
}

func (builtin *ExecCapabilities) Description() string {
	return builtin.description
}

func (builtin *ExecCapabilities) isAcceptableCode(recipeResult map[string]starlark.Comparable) bool {
	isAcceptableCode := false
	for _, acceptableCode := range builtin.acceptableCodes {
		if recipeResult["code"] == starlark.MakeInt64(acceptableCode) {
			isAcceptableCode = true
			break
		}
	}
	return isAcceptableCode
}

func formatErrorMessage(errorMessage string, errorFromExec string) string {
	splitErrorMessageNewLine := strings.Split(errorFromExec, "\n")
	reformattedErrorMessage := strings.Join(splitErrorMessageNewLine, "\n  ")
	return fmt.Sprintf("%v\n  %v", errorMessage, reformattedErrorMessage)
}
