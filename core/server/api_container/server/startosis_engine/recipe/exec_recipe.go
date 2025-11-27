package recipe

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"golang.org/x/exp/maps"
)

const (
	CommandAttr        = "command"
	ExecRecipeTypeName = "ExecRecipe"

	execOutputKey   = "output"
	execExitCodeKey = "code"
	newlineChar     = "\n"
)

func NewExecRecipeType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ExecRecipeTypeName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              CommandAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						if _, interpretationErr := convertStarlarkListToStringList(value); interpretationErr != nil {
							return interpretationErr
						}
						return nil
					},
				},
				{
					Name:              ExtractAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						_, interpretationErr := convertExtractorsToDict(true, value)
						return interpretationErr
					},
				},
			},
		},
		Instantiate: instantiateExecRecipe,
	}
}

func instantiateExecRecipe(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ExecRecipeTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &ExecRecipe{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type ExecRecipe struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func (recipe *ExecRecipe) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := recipe.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &ExecRecipe{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (recipe *ExecRecipe) Execute(
	ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	service *service.Service,
) (map[string]starlark.Comparable, error) {
	// parse argument
	commandStarlarkList, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		recipe.KurtosisValueTypeDefault, CommandAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Mandatory argument '%s' not found", CommandAttr)
	}
	command, interpretationErr := convertStarlarkListToStringList(commandStarlarkList)
	if interpretationErr != nil {
		// that should never happen as it's being validated at interpretation time
		return nil, stacktrace.NewError("Unexpected '%s' attribute for '%s'. Error was: \n%s",
			CommandAttr, ExecRecipeTypeName, interpretationErr.Error())
	}
	rawExtractors, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		recipe.KurtosisValueTypeDefault, ExtractAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	extractors, interpretationErr := convertExtractorsToDict(found, rawExtractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	// execute recipe
	var commandWithRuntimeValue []string
	for _, subCommand := range command {
		maybeSubCommandWithRuntimeValues, err := magic_string_helper.ReplaceRuntimeValueInString(subCommand, runtimeValueStore)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while replacing runtime values in the command of the exec recipe")
		}
		commandWithRuntimeValue = append(commandWithRuntimeValue, maybeSubCommandWithRuntimeValues)
	}

	serviceNameStr := string(service.GetRegistration().GetName())

	execResult, err := serviceNetwork.RunExec(ctx, serviceNameStr, commandWithRuntimeValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to execute command '%v' on service '%s'", command, serviceNameStr)
	}
	commandOutput := execResult.GetOutput()
	resultDict := map[string]starlark.Comparable{
		execOutputKey:   starlark.String(commandOutput),
		execExitCodeKey: starlark.MakeInt(int(execResult.GetExitCode())),
	}
	extractDict, err := runExtractors([]byte(fmt.Sprintf("%q", commandOutput)), extractors)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while running extractors '%v' on command output '%v'", extractors, commandOutput)
	}
	maps.Copy(resultDict, extractDict)
	return resultDict, nil
}

func (recipe *ExecRecipe) ResultMapToString(resultMap map[string]starlark.Comparable) string {
	exitCode := resultMap[execExitCodeKey]
	rawOutput := resultMap[execOutputKey]
	outputStarlarkStr, ok := rawOutput.(starlark.String)
	if !ok {
		logrus.Errorf("Result of an exec recipe was not a string (was: '%v' of type '%s'). This is not fatal but the object might be malformed in CLI output. It is very unexpected and hides a Kurtosis internal bug. This issue should be reported", rawOutput, reflect.TypeOf(rawOutput))
		outputStarlarkStr = starlark.String(outputStarlarkStr.String())
	}
	outputStr := outputStarlarkStr.GoString()
	if outputStr == "" {
		return fmt.Sprintf("Command returned with exit code '%v' with no output", exitCode)
	}
	if strings.Contains(outputStr, newlineChar) {
		return fmt.Sprintf(`Command returned with exit code '%v' and the following output:
--------------------
%v
--------------------`, exitCode, outputStr)
	}
	return fmt.Sprintf("Command returned with exit code '%v' and the following output: %v", exitCode, outputStr)
}

func (recipe *ExecRecipe) CreateStarlarkReturnValue(resultUuid string) (*starlark.Dict, *startosis_errors.InterpretationError) {
	rawExtractors, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		recipe.KurtosisValueTypeDefault, ExtractAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	extractors, interpretationErr := convertExtractorsToDict(found, rawExtractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	returnValue, _, interpretationErr := createExecRecipeStarlarkReturnValueInternal(resultUuid, extractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return returnValue, nil
}

func (recipe *ExecRecipe) GetStarlarkReturnValuesAsStringList(resultUuid string) ([]string, *startosis_errors.InterpretationError) {
	rawExtractors, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		recipe.KurtosisValueTypeDefault, ExtractAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	extractors, interpretationErr := convertExtractorsToDict(found, rawExtractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	_, returnValueStrings, interpretationErr := createExecRecipeStarlarkReturnValueInternal(resultUuid, extractors)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	return returnValueStrings, nil
}

func createExecRecipeStarlarkReturnValueInternal(resultUuid string, extractors map[string]string) (*starlark.Dict, []string, *startosis_errors.InterpretationError) {
	dict := &starlark.Dict{}
	returnValueStrings := []string{}

	exitCodeValueString := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, execExitCodeKey)
	returnValueStrings = append(returnValueStrings, exitCodeValueString)
	err := dict.SetKey(starlark.String(execExitCodeKey), starlark.String(exitCodeValueString))
	if err != nil {
		return nil, nil, startosis_errors.WrapWithInterpretationError(err, "An error happened while creating exec return value, setting field '%v'", execExitCodeKey)
	}

	outputValueString := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, execOutputKey)
	returnValueStrings = append(returnValueStrings, outputValueString)
	err = dict.SetKey(starlark.String(execOutputKey), starlark.String(outputValueString))
	if err != nil {
		return nil, nil, startosis_errors.WrapWithInterpretationError(err, "An error happened while creating exec return value, setting field '%v'", execOutputKey)
	}

	for extractorKey := range extractors {
		fullExtractorKey := fmt.Sprintf("%v.%v", extractKeyPrefix, extractorKey)
		extractedValueString := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, fullExtractorKey)
		returnValueStrings = append(returnValueStrings, extractedValueString)
		err = dict.SetKey(starlark.String(fullExtractorKey), starlark.String(extractedValueString))
		if err != nil {
			return nil, nil, startosis_errors.NewInterpretationError("An error has occurred when creating return value for request recipe, setting field '%v'", fullExtractorKey)
		}
	}

	dict.Freeze()
	return dict, returnValueStrings, nil
}

func convertStarlarkListToStringList(starlarkValue starlark.Value) ([]string, *startosis_errors.InterpretationError) {
	starlarkList, isList := starlarkValue.(*starlark.List)
	if !isList {
		return nil, startosis_errors.NewInterpretationError("Attribute '%s' is expected to be a list of strings, got '%s'", CommandAttr, reflect.TypeOf(starlarkValue))
	}

	var stringListResult []string
	for idx := 0; idx < starlarkList.Len(); idx++ {
		item := starlarkList.Index(idx)
		itemStr, isString := item.(starlark.String)
		if !isString {
			return nil, startosis_errors.NewInterpretationError("Item number %d in '%s' list was not a string. Expecting '%s' to be a string list", idx, CommandAttr, CommandAttr)
		}
		stringListResult = append(stringListResult, itemStr.GoString())
	}
	return stringListResult, nil
}
