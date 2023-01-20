package remove_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	RemoveServiceBuiltinName = "remove_service"

	serviceNameArgName = "service_name"
)

func GenerateRemoveServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceName, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		removeServiceInstruction := NewRemoveServiceInstruction(serviceNetwork, instructionPosition, serviceName)
		*instructionsQueue = append(*instructionsQueue, removeServiceInstruction)
		return starlark.None, nil
	}
}

type RemoveServiceInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position    *kurtosis_instruction.InstructionPosition
	serviceName kurtosis_backend_service.ServiceName
}

func NewRemoveServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceName kurtosis_backend_service.ServiceName) *RemoveServiceInstruction {
	return &RemoveServiceInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceName:    serviceName,
	}
}

func (instruction *RemoveServiceInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *RemoveServiceInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(starlark.String(instruction.serviceName)), serviceNameArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), RemoveServiceBuiltinName, instruction.String(), args)
}

func (instruction *RemoveServiceInstruction) Execute(ctx context.Context) (*string, error) {
	serviceUUID, err := instruction.serviceNetwork.RemoveService(ctx, string(instruction.serviceName))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed removing service with unexpected error")
	}
	logrus.Infof("Successfully removed service '%v' with guid '%v'", instruction.serviceName, serviceUUID)
	instructionResult := fmt.Sprintf("Service '%s' with service UUID '%s' removed", instruction.serviceName, serviceUUID)
	return &instructionResult, nil
}

func (instruction *RemoveServiceInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(RemoveServiceBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs())
}

func (instruction *RemoveServiceInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceNameExist(instruction.serviceName) {
		return startosis_errors.NewValidationError("There was an error validating '%v' as service name '%v' doesn't exist", RemoveServiceBuiltinName, instruction.serviceName)
	}
	environment.RemoveServiceName(instruction.serviceName)
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceName, *startosis_errors.InterpretationError) {
	var serviceNameArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceNameArgName, &serviceNameArg); err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", RemoveServiceBuiltinName, args, kwargs)
	}

	serviceName, interpretationErr := kurtosis_instruction.ParseServiceName(serviceNameArg)
	if interpretationErr != nil {
		return "", interpretationErr
	}

	return serviceName, nil
}

func (instruction *RemoveServiceInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{
		serviceNameArgName: starlark.String(instruction.serviceName),
	}
}
