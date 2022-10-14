package exec

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"strings"
)

const (
	ExecBuiltinName = "exec"

	serviceIdArgName           = "service_id"
	commandArgName             = "command"
	expectedExitCodeArgName    = "expected_exit_code?"
	nonOptionalExitCodeArgName = "expected_exit_code"

	commandSeparator = `", "`
)

func GenerateExecBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceId, commandArgs, expectedExitCode, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		execInstruction := NewExecInstruction(serviceNetwork, kurtosis_instruction.GetPositionFromThread(thread), serviceId, commandArgs, expectedExitCode)
		*instructionsQueue = append(*instructionsQueue, execInstruction)
		return starlark.None, nil
	}
}

type ExecInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position         kurtosis_instruction.InstructionPosition
	serviceId        kurtosis_backend_service.ServiceID
	command          []string
	expectedExitCode int32
}

func NewExecInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, command []string, expectedExitCode int32) *ExecInstruction {
	return &ExecInstruction{
		serviceNetwork:   serviceNetwork,
		position:         position,
		serviceId:        serviceId,
		command:          command,
		expectedExitCode: expectedExitCode,
	}
}

func (instruction *ExecInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *ExecInstruction) GetCanonicalInstruction() string {
	buffer := new(strings.Builder)
	buffer.WriteString(ExecBuiltinName + "(")
	buffer.WriteString(serviceIdArgName + "=\"")
	buffer.WriteString(fmt.Sprintf("%v\", ", instruction.serviceId))
	buffer.WriteString(commandArgName + "=[\"")
	buffer.WriteString(fmt.Sprintf("%v\"], ", strings.Join(instruction.command, commandSeparator)))
	buffer.WriteString(nonOptionalExitCodeArgName + "=")
	buffer.WriteString(fmt.Sprintf("%v)", instruction.expectedExitCode))
	return buffer.String()
}

func (instruction *ExecInstruction) Execute(ctx context.Context) error {
	exitCode, _, err := instruction.serviceNetwork.ExecCommand(ctx, instruction.serviceId, instruction.command)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to execute command '%v' on service '%v'", instruction.command, instruction.serviceId)
	}
	if instruction.expectedExitCode != exitCode {
		return stacktrace.Propagate(err, "The exit code expected '%v' wasn't the exit code received '%v' while running the command", instruction.expectedExitCode, exitCode)
	}
	return nil
}

func (instruction *ExecInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *ExecInstruction) ValidateAndUpdateEnvironment(_ *startosis_validator.ValidatorEnvironment) error {
	// this doesn't do anything but can't return an error as the validator runs this regardless
	// this is a no-op
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, []string, int32, *startosis_errors.InterpretationError) {

	var serviceIdArg starlark.String
	var commandArg *starlark.List
	var expectedExitCodeArg = starlark.MakeInt(0)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, commandArgName, &commandArg, expectedExitCodeArgName, &expectedExitCodeArg); err != nil {
		return "", nil, 0, startosis_errors.NewInterpretationError(err.Error())
	}

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return "", nil, 0, interpretationErr
	}

	command, interpretationErr := kurtosis_instruction.ParseCommand(commandArg)
	if interpretationErr != nil {
		return "", nil, 0, interpretationErr
	}

	expectedExitCode, interpretationErr := kurtosis_instruction.ParseExpectedExitCode(expectedExitCodeArg)
	if interpretationErr != nil {
		return "", nil, 0, interpretationErr
	}
	return serviceId, command, expectedExitCode, nil
}
