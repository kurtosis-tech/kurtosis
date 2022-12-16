package recipe

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"strings"
)

const (
	execOutputKey   = "output"
	execExitCodeKey = "code"
	newlineChar     = "\n"

	commandKey = "command"
)

type ExecRecipe struct {
	serviceId service.ServiceID
	command   []string
}

func NewExecRecipe(serviceId service.ServiceID, command []string) *ExecRecipe {
	return &ExecRecipe{
		serviceId: serviceId,
		command:   command,
	}
}

func (recipe *ExecRecipe) Execute(ctx context.Context, serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) (map[string]starlark.Comparable, error) {
	var commandWithIPAddressAndRuntimeValue []string
	for _, subCommand := range recipe.command {
		maybeSubCommandWithIPAddress, err := magic_string_helper.ReplaceIPAddressInString(subCommand, serviceNetwork, commandKey)
		if err != nil {
			return nil, stacktrace.Propagate(err,"An error occurred while replacing IP address in the command of the exec recipe")
		}
		maybeSubCommandWithRuntimeValuesAndIPAddress, err := magic_string_helper.ReplaceRuntimeValueInString(maybeSubCommandWithIPAddress, runtimeValueStore)
		if err != nil {
			return nil, stacktrace.Propagate(err,"An error occurred while replacing runtime values in the command of the exec recipe")
		}
		commandWithIPAddressAndRuntimeValue = append(commandWithIPAddressAndRuntimeValue, maybeSubCommandWithRuntimeValuesAndIPAddress)
	}
	exitCode, commandOutput, err := serviceNetwork.ExecCommand(ctx, recipe.serviceId, commandWithIPAddressAndRuntimeValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to execute command '%v' on service '%v'", recipe.command, recipe.serviceId)
	}
	return map[string]starlark.Comparable{
		execOutputKey:   starlark.String(commandOutput),
		execExitCodeKey: starlark.MakeInt(int(exitCode)),
	}, nil
}

func (recipe *ExecRecipe) ResultMapToString(resultMap map[string]starlark.Comparable) string {
	output := resultMap[execOutputKey]
	exitCode := resultMap[execExitCodeKey]
	if output == starlark.String("") {
		return fmt.Sprintf("Command returned with exit code '%v' with no output", exitCode)
	}
	if strings.Contains(output.String(), newlineChar) {
		return fmt.Sprintf("Command returned with exit code '%v' and the following output: \n%v", exitCode, output)
	}
	return fmt.Sprintf("Command returned with exit code '%v' and the following output: %v", exitCode, output)
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
