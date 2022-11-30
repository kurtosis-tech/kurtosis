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

	serviceIdArgName = "service_id"
)

func GenerateRemoveServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceId, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		removeServiceInstruction := NewRemoveServiceInstruction(serviceNetwork, instructionPosition, serviceId)
		*instructionsQueue = append(*instructionsQueue, removeServiceInstruction)
		return starlark.None, nil
	}
}

type RemoveServiceInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position  *kurtosis_instruction.InstructionPosition
	serviceId kurtosis_backend_service.ServiceID
}

func NewRemoveServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID) *RemoveServiceInstruction {
	return &RemoveServiceInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      serviceId,
	}
}

func (instruction *RemoveServiceInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *RemoveServiceInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(starlark.String(instruction.serviceId)), serviceIdArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), RemoveServiceBuiltinName, instruction.String(), args)
}

func (instruction *RemoveServiceInstruction) Execute(ctx context.Context) (*string, error) {
	serviceGUID, err := instruction.serviceNetwork.RemoveService(ctx, instruction.serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed removing service with unexpected error")
	}
	logrus.Infof("Successfully removed service '%v' with guid '%v'", instruction.serviceId, serviceGUID)
	instructionResult := fmt.Sprintf("Service '%s' with internal ID '%s' removed", instruction.serviceId, serviceGUID)
	return &instructionResult, nil
}

func (instruction *RemoveServiceInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(RemoveServiceBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs())
}

func (instruction *RemoveServiceInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return startosis_errors.NewValidationError("There was an error validating remove service as service ID '%v' doesn't exist", instruction.serviceId)
	}
	environment.RemoveServiceId(instruction.serviceId)
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, *startosis_errors.InterpretationError) {
	var serviceIdArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg); err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", RemoveServiceBuiltinName, args, kwargs)
	}

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return "", interpretationErr
	}

	return serviceId, nil
}

func (instruction *RemoveServiceInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{
		serviceIdArgName: starlark.String(instruction.serviceId),
	}
}
