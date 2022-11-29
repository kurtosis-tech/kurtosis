package exec

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	ExecBuiltinName = "exec"

	serviceIdArgName           = "service_id"
	commandArgName             = "command"
	expectedExitCodeArgName    = "expected_exit_code?"
	nonOptionalExitCodeArgName = "expected_exit_code"

	successfulExitCode = 0
)

func GenerateExecBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		execInstruction := newEmptyExecInstruction(serviceNetwork, instructionPosition)
		if interpretationError := execInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, execInstruction)
		return starlark.None, nil
	}
}

type ExecInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceId        kurtosis_backend_service.ServiceID
	command          []string
	expectedExitCode int32
}

func NewExecInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, command []string, expectedExitCode int32, starlarkKwargs starlark.StringDict) *ExecInstruction {
	return &ExecInstruction{
		serviceNetwork:   serviceNetwork,
		position:         position,
		serviceId:        serviceId,
		command:          command,
		expectedExitCode: expectedExitCode,
		starlarkKwargs:   starlarkKwargs,
	}
}

func newEmptyExecInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition) *ExecInstruction {
	return &ExecInstruction{
		serviceNetwork:   serviceNetwork,
		position:         position,
		serviceId:        "",
		command:          nil,
		expectedExitCode: 0,
		starlarkKwargs:   starlark.StringDict{},
	}
}

func (instruction *ExecInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *ExecInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceIdArgName]), serviceIdArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[commandArgName]), commandArgName, kurtosis_instruction.NotRepresentative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[nonOptionalExitCodeArgName]), nonOptionalExitCodeArgName, kurtosis_instruction.NotRepresentative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), ExecBuiltinName, instruction.String(), args)
}

func (instruction *ExecInstruction) Execute(ctx context.Context) (*string, error) {
	exitCode, _, err := instruction.serviceNetwork.ExecCommand(ctx, instruction.serviceId, instruction.command)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to execute command '%v' on service '%v'", instruction.command, instruction.serviceId)
	}
	if instruction.expectedExitCode != exitCode {
		return nil, stacktrace.NewError("The exit code expected '%v' wasn't the exit code received '%v' while running the command", instruction.expectedExitCode, exitCode)
	}
	return nil, nil
}

func (instruction *ExecInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(ExecBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *ExecInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return startosis_errors.NewValidationError("There was an error validating exec with service ID '%v' that does not exist", instruction.serviceId)
	}
	return nil
}

func (instruction *ExecInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {

	var serviceIdArg starlark.String
	var commandArg *starlark.List
	var expectedExitCodeArg = starlark.MakeInt(successfulExitCode)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, commandArgName, &commandArg, expectedExitCodeArgName, &expectedExitCodeArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", ExecBuiltinName, args, kwargs)
	}
	instruction.starlarkKwargs[serviceIdArgName] = serviceIdArg
	instruction.starlarkKwargs[commandArgName] = commandArg
	instruction.starlarkKwargs[nonOptionalExitCodeArgName] = expectedExitCodeArg
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
	instruction.serviceId = serviceId
	instruction.command = command
	instruction.expectedExitCode = expectedExitCode
	return nil
}
