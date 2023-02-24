package recipe

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"reflect"
	"strings"
)

const (
	execOutputKey   = "output"
	execExitCodeKey = "code"
	newlineChar     = "\n"

	commandKey     = "command"
	serviceNameKey = "service_name"
	ExecRecipeName = "ExecRecipe"
)

// TODO: maybe change command to startlark.List once remove backward compatability support
type ExecRecipe struct {
	serviceName service.ServiceName
	command     []string
}

func NewExecRecipe(serviceName service.ServiceName, command []string) *ExecRecipe {
	return &ExecRecipe{
		serviceName: serviceName,
		command:     command,
	}
}

// String the starlark.Value interface
func (recipe *ExecRecipe) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(ExecRecipeName + "(")
	buffer.WriteString(serviceNameKey + "=")
	buffer.WriteString(fmt.Sprintf("%q, ", recipe.serviceName))
	buffer.WriteString(commandKey + "=")

	command := convertListToStarlarkList(recipe.command)
	if command.Len() > 0 {
		buffer.WriteString(fmt.Sprintf("%v)", command))
	} else {
		buffer.WriteString(fmt.Sprintf("%q)", ""))
	}
	return buffer.String()
}

// Type implements the starlark.Value interface
func (recipe *ExecRecipe) Type() string {
	return ExecRecipeName
}

// Freeze implements the starlark.Value interface
func (recipe *ExecRecipe) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (recipe *ExecRecipe) Truth() starlark.Bool {
	return recipe.serviceName != ""
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed, users should use a portId instead
func (recipe *ExecRecipe) Hash() (uint32, error) {
	return 0, startosis_errors.NewInterpretationError("unhashable type: '%v'", ExecRecipeName)
}

// Attr implements the starlark.HasAttrs interface.
func (recipe *ExecRecipe) Attr(name string) (starlark.Value, error) {
	switch name {
	case serviceNameKey:
		return starlark.String(recipe.serviceName), nil
	case commandKey:
		return convertListToStarlarkList(recipe.command), nil
	default:
		return nil, startosis_errors.NewInterpretationError("'%v' has no attribute '%v;", ExecRecipeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (recipe *ExecRecipe) AttrNames() []string {
	return []string{serviceNameKey, commandKey}
}

func (recipe *ExecRecipe) Execute(
	ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	serviceName service.ServiceName,
) (map[string]starlark.Comparable, error) {
	var commandWithIPAddressAndRuntimeValue []string
	for _, subCommand := range recipe.command {
		maybeSubCommandWithIPAddressAndHostname, err := magic_string_helper.ReplaceIPAddressAndHostnameInString(subCommand, serviceNetwork, commandKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while replacing IP address in the command of the exec recipe")
		}
		maybeSubCommandWithRuntimeValuesAndIPAddress, err := magic_string_helper.ReplaceRuntimeValueInString(maybeSubCommandWithIPAddressAndHostname, runtimeValueStore)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while replacing runtime values in the command of the exec recipe")
		}
		commandWithIPAddressAndRuntimeValue = append(commandWithIPAddressAndRuntimeValue, maybeSubCommandWithRuntimeValuesAndIPAddress)
	}
	exitCode, commandOutput, err := serviceNetwork.ExecCommand(ctx, string(serviceName), commandWithIPAddressAndRuntimeValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to execute command '%v' on service '%v'", recipe.command, recipe.serviceName)
	}
	return map[string]starlark.Comparable{
		execOutputKey:   starlark.String(commandOutput),
		execExitCodeKey: starlark.MakeInt(int(exitCode)),
	}, nil
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
	dict := &starlark.Dict{}
	err := dict.SetKey(starlark.String(execExitCodeKey), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, execExitCodeKey)))
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error happened while creating exec return value, setting field '%v'", execExitCodeKey)
	}
	err = dict.SetKey(starlark.String(execOutputKey), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, execOutputKey)))
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error happened while creating exec return value, setting field '%v'", execOutputKey)
	}
	dict.Freeze()
	return dict, nil
}

func MakeExecRequestRecipe(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var serviceNameStr string
	var unpackedCommandList *starlark.List

	if err := starlark.UnpackArgs(builtin.Name(), args, kwargs,
		commandKey, &unpackedCommandList,
		serviceNameKey, &serviceNameStr,
	); err != nil {
		return nil, startosis_errors.NewInterpretationError("%v", err.Error())
	}

	commands, err := kurtosis_types.SafeCastToStringSlice(unpackedCommandList, commandKey)
	if err != nil {
		return nil, err
	}
	serviceName := service.ServiceName(serviceNameStr)
	return NewExecRecipe(serviceName, commands), nil
}

func convertListToStarlarkList(inputList []string) *starlark.List {
	sizeOfCommandArr := len(inputList)
	var elems []starlark.Value
	for i := 0; i < sizeOfCommandArr; i++ {
		elems = append(elems, starlark.String(inputList[i]))
	}
	return starlark.NewList(elems)
}
