package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	AddServiceBuiltinName = "add_service"

	serviceNameArgName   = "service_name"
	serviceConfigArgName = "config"
)

func GenerateAddServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		addServiceInstruction := newEmptyAddServiceInstruction(serviceNetwork, instructionPosition, runtimeValueStore)
		if interpretationError := addServiceInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, addServiceInstruction)
		returnValue, interpretationError := makeAddServiceInterpretationReturnValue(addServiceInstruction.serviceName, addServiceInstruction.serviceConfig)
		if interpretationError != nil {
			return nil, interpretationError
		}
		return returnValue, nil
	}
}

type AddServiceInstruction struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceName   kurtosis_backend_service.ServiceName
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig
}

func newEmptyAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, runtimeValueStore *runtime_value_store.RuntimeValueStore) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		starlarkKwargs:    starlark.StringDict{},
		serviceName:       "",
		serviceConfig:     nil,
		runtimeValueStore: runtimeValueStore,
	}
}

func NewAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceName kurtosis_backend_service.ServiceName, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig, starlarkKwargs starlark.StringDict, runtimeValueStore *runtime_value_store.RuntimeValueStore) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		serviceName:       serviceName,
		serviceConfig:     serviceConfig,
		starlarkKwargs:    starlarkKwargs,
		runtimeValueStore: runtimeValueStore,
	}
}

func (instruction *AddServiceInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *AddServiceInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceNameArgName]), serviceNameArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceConfigArgName]), serviceConfigArgName, kurtosis_instruction.NotRepresentative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), AddServiceBuiltinName, instruction.String(), args)
}

func (instruction *AddServiceInstruction) Execute(ctx context.Context) (*string, error) {
	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(instruction.serviceNetwork, instruction.runtimeValueStore, instruction.serviceName, instruction.serviceConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing a magic string in '%s' instruction arguments. Execution cannot proceed", AddServiceBuiltinName)
	}

	startedService, err := instruction.serviceNetwork.StartService(ctx, replacedServiceName, replacedServiceConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed adding service '%s' to enclave with an unexpected error", replacedServiceName)
	}
	instructionResult := fmt.Sprintf("Service '%s' added with service UUID '%s'", replacedServiceName, startedService.GetRegistration().GetUUID())
	return &instructionResult, nil
}

func (instruction *AddServiceInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if err := validateSingleService(environment, instruction.serviceName, instruction.serviceConfig); err != nil {
		return err
	}
	return nil
}

func (instruction *AddServiceInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(AddServiceBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *AddServiceInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var serviceNameArg starlark.String
	var serviceConfigArg *kurtosis_types.ServiceConfig
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceNameArgName, &serviceNameArg, serviceConfigArgName, &serviceConfigArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", AddServiceBuiltinName, args, kwargs)
	}
	instruction.starlarkKwargs[serviceNameArgName] = serviceNameArg
	instruction.starlarkKwargs[serviceConfigArgName] = serviceConfigArg
	instruction.starlarkKwargs.Freeze()

	serviceName, interpretationErr := kurtosis_instruction.ParseServiceName(serviceNameArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	var serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig
	serviceConfig, interpretationErr = serviceConfigArg.ToKurtosisType()
	if interpretationErr != nil {
		return interpretationErr
	}
	instruction.serviceName = serviceName
	instruction.serviceConfig = serviceConfig
	return nil
}
