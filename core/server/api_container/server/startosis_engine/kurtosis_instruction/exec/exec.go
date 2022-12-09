package exec

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"strings"
)

const (
	ExecBuiltinName = "exec"

	serviceIdArgName           = "service_id"
	commandArgName             = "command"
	expectedExitCodeArgName    = "expected_exit_code?"
	nonOptionalExitCodeArgName = "expected_exit_code"
	execIdArgName              = "exec_id?"
	nonOptionalExecIdArgName   = "exec_id"

	successfulExitCode = 0

	newlineChar = "\n"

	execOutputSuffix   = "output"
	execExitCodeSuffix = "code"

	emptyStarlarkString = starlark.String("")
)

func GenerateExecBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		execInstruction := newEmptyExecInstruction(serviceNetwork, instructionPosition, runtimeValueStore)
		if interpretationError := execInstruction.parseStartosisArgs(b, args, kwargs, runtimeValueStore); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, execInstruction)
		returnValue := createStarlarkReturnValueForExec(execInstruction.execId)
		return returnValue, nil
	}
}

type ExecInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceId        kurtosis_backend_service.ServiceID
	command          []string
	expectedExitCode int32

	runtimeValueStore *runtime_value_store.RuntimeValueStore
	execId            string
}

func NewExecInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, command []string, expectedExitCode int32, starlarkKwargs starlark.StringDict, runtimeValueStore *runtime_value_store.RuntimeValueStore, execId string) *ExecInstruction {
	return &ExecInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		serviceId:         serviceId,
		command:           command,
		expectedExitCode:  expectedExitCode,
		starlarkKwargs:    starlarkKwargs,
		runtimeValueStore: runtimeValueStore,
		execId:            execId,
	}
}

func newEmptyExecInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, runtimeValueStore *runtime_value_store.RuntimeValueStore) *ExecInstruction {
	return &ExecInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		serviceId:         "",
		command:           nil,
		expectedExitCode:  0,
		starlarkKwargs:    starlark.StringDict{},
		execId:            "",
		runtimeValueStore: runtimeValueStore,
	}
}

func (instruction *ExecInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *ExecInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceIdArgName]), serviceIdArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[commandArgName]), commandArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[nonOptionalExitCodeArgName]), nonOptionalExitCodeArgName, kurtosis_instruction.NotRepresentative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[nonOptionalExecIdArgName]), nonOptionalExecIdArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), ExecBuiltinName, instruction.String(), args)
}

func (instruction *ExecInstruction) Execute(ctx context.Context) (*string, error) {
	exitCode, commandOutput, err := instruction.serviceNetwork.ExecCommand(ctx, instruction.serviceId, instruction.command)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to execute command '%v' on service '%v'", instruction.command, instruction.serviceId)
	}
	if instruction.expectedExitCode != exitCode {
		return nil, stacktrace.NewError("The exit code expected '%v' wasn't the exit code received '%v' while running the command", instruction.expectedExitCode, exitCode)
	}
	instruction.runtimeValueStore.SetValue(instruction.execId, createStarlarkComparable(commandOutput, exitCode))
	instructionResult := formatInstructionOutput(exitCode, commandOutput)
	return &instructionResult, nil
}

func (instruction *ExecInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(ExecBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *ExecInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return startosis_errors.NewValidationError("There was an error validating '%v' with service ID '%v' that does not exist", ExecBuiltinName, instruction.serviceId)
	}
	return nil
}

func (instruction *ExecInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple, runtimeValueStore *runtime_value_store.RuntimeValueStore) *startosis_errors.InterpretationError {

	var serviceIdArg starlark.String
	var commandArg *starlark.List
	var expectedExitCodeArg = starlark.MakeInt(successfulExitCode)
	var execIdArg = emptyStarlarkString
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, commandArgName, &commandArg, expectedExitCodeArgName, &expectedExitCodeArg, execIdArgName, &execIdArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", ExecBuiltinName, args, kwargs)
	}

	if execIdArg == emptyStarlarkString {
		execIdArg = starlark.String(runtimeValueStore.CreateValue())
	}

	instruction.starlarkKwargs[serviceIdArgName] = serviceIdArg
	instruction.starlarkKwargs[commandArgName] = commandArg
	instruction.starlarkKwargs[nonOptionalExitCodeArgName] = expectedExitCodeArg
	instruction.starlarkKwargs[nonOptionalExecIdArgName] = execIdArg
	instruction.starlarkKwargs.Freeze()

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	command, interpretationErr := kurtosis_instruction.ParseCommand(commandArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	expectedExitCode, interpretationErr := kurtosis_instruction.ParseExpectedExitCode(expectedExitCodeArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	execId, interpretationErr := kurtosis_instruction.ParseNonEmptyString(execIdArgName, execIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.serviceId = serviceId
	instruction.command = command
	instruction.expectedExitCode = expectedExitCode
	instruction.execId = execId
	return nil
}

func formatInstructionOutput(exitCode int32, commandOutput string) string {
	trimmedOutput := strings.TrimSpace(commandOutput)
	if trimmedOutput == "" {
		return fmt.Sprintf("Command returned with exit code '%d' with no output", exitCode)
	}
	if strings.Contains(trimmedOutput, newlineChar) {
		return fmt.Sprintf("Command returned with exit code '%d' and the following output: \n%v", exitCode, trimmedOutput)
	}
	return fmt.Sprintf("Command returned with exit code '%d' and the following output: '%s'", exitCode, trimmedOutput)
}

func createStarlarkReturnValueForExec(resultUuid string) *starlarkstruct.Struct {
	return starlarkstruct.FromKeywords(
		starlarkstruct.Default,
		[]starlark.Tuple{
			{starlark.String(execOutputSuffix), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, execOutputSuffix))},
			{starlark.String(execExitCodeSuffix), starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, execExitCodeSuffix))},
		},
	)
}

func createStarlarkComparable(execOutput string, exitCode int32) map[string]starlark.Comparable {
	return map[string]starlark.Comparable{
		execOutputSuffix:   starlark.String(execOutput),
		execExitCodeSuffix: starlark.MakeInt(int(exitCode)),
	}
}
