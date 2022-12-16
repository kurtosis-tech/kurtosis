package update_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	UpdateServiceBuiltinName = "update_service"

	serviceIdArgName = "service_id"

	updateServiceConfigArgName = "config"
)

func GenerateUpdateServiceBuiltin(
	instructionsQueue *[]kurtosis_instruction.KurtosisInstruction,
	serviceNetwork service_network.ServiceNetwork,
) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		newInstruction := newEmptyUpdateServiceInstruction(serviceNetwork, instructionPosition)
		if interpretationError := newInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, newInstruction)
		return starlark.None, nil
	}
}

type UpdateServiceInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceId           kurtosis_backend_service.ServiceID
	updateServiceConfig *kurtosis_core_rpc_api_bindings.UpdateServiceConfig
}

func newEmptyUpdateServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition) *UpdateServiceInstruction {
	return &UpdateServiceInstruction{
		serviceNetwork:      serviceNetwork,
		position:            position,
		starlarkKwargs:      starlark.StringDict{},
		serviceId:           "",
		updateServiceConfig: nil,
	}
}

func NewUpdateServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, updateServiceConfig *kurtosis_core_rpc_api_bindings.UpdateServiceConfig, starlarkKwargs starlark.StringDict) *UpdateServiceInstruction {
	return &UpdateServiceInstruction{
		serviceNetwork:      serviceNetwork,
		position:            position,
		serviceId:           serviceId,
		updateServiceConfig: updateServiceConfig,
		starlarkKwargs:      starlarkKwargs,
	}
}

func (instruction *UpdateServiceInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *UpdateServiceInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceIdArgName]), serviceIdArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[updateServiceConfigArgName]), updateServiceConfigArgName, kurtosis_instruction.NotRepresentative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), UpdateServiceBuiltinName, instruction.String(), args)
}

func (instruction *UpdateServiceInstruction) Execute(ctx context.Context) (*string, error) {
	service, err := instruction.serviceNetwork.GetService(ctx, instruction.serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Updating service '%s' failed as it could not be retrieved from the enclave", instruction.serviceId)
	}

	updateServiceConfigMap := map[kurtosis_backend_service.ServiceID]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig{
		instruction.serviceId: instruction.updateServiceConfig,
	}

	serviceSuccessful, serviceFailed, err := instruction.serviceNetwork.UpdateService(ctx, updateServiceConfigMap)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed updating service '%s' with an unexpected error", instruction.serviceId)
	}
	if failure, found := serviceFailed[instruction.serviceId]; found {
		return nil, stacktrace.Propagate(failure, "Failed updating service '%s'", instruction.serviceNetwork)
	}
	_, found := serviceSuccessful[instruction.serviceId]
	if !found {
		return nil, stacktrace.NewError("Service '%s' wasn't accounted as failed nor successfully updated. This is a product bug", instruction.serviceId)
	}
	instructionResult := fmt.Sprintf("Service '%s' with GUID '%s' updated", instruction.serviceId, service.GetRegistration().GetGUID())
	return &instructionResult, nil
}

func (instruction *UpdateServiceInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(UpdateServiceBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *UpdateServiceInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if partition_topology.ParsePartitionId(instruction.updateServiceConfig.Subnetwork) != partition_topology.DefaultPartitionId {
		if !environment.IsNetworkPartitioningEnabled() {
			return startosis_errors.NewValidationError("Service was about to be moved to subnetwork '%s' but the Kurtosis enclave was started with subnetwork capabilities disabled. Make sure to run the Starlark script with subnetwork enabled.", *instruction.updateServiceConfig.Subnetwork)
		}
	}
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return startosis_errors.NewValidationError("There was an error validating '%v' as service ID '%v' does not exist", UpdateServiceBuiltinName, instruction.serviceId)
	}
	return nil
}

func (instruction *UpdateServiceInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var serviceIdArg starlark.String
	var updateServiceConfigArg *kurtosis_types.UpdateServiceConfig
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, updateServiceConfigArgName, &updateServiceConfigArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", UpdateServiceBuiltinName, args, kwargs)
	}
	instruction.starlarkKwargs[serviceIdArgName] = serviceIdArg
	instruction.starlarkKwargs[updateServiceConfigArgName] = updateServiceConfigArg
	instruction.starlarkKwargs.Freeze()

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.serviceId = serviceId
	instruction.updateServiceConfig = updateServiceConfigArg.ToKurtosisType()
	return nil
}
