package kurtosis_instruction

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const AddServiceBuiltinName = "add_service"

type AddServiceInstruction struct {
	serviceNetwork *service_network.ServiceNetwork

	position      InstructionPosition
	serviceId     kurtosis_backend_service.ServiceID
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig
}

func NewAddServiceInstruction(serviceNetwork *service_network.ServiceNetwork, position InstructionPosition, serviceId kurtosis_backend_service.ServiceID, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      serviceId,
		serviceConfig:  serviceConfig,
	}
}

func AddServiceBuiltin(instructionsQueue *[]KurtosisInstruction, serviceNetwork *service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceId, serviceConfig, err := parseStartosisArgs(b, args, kwargs)
		if err != nil {
			return nil, err
		}
		addServiceInstruction := NewAddServiceInstruction(serviceNetwork, getPosition(thread), serviceId, serviceConfig)
		*instructionsQueue = append(*instructionsQueue, addServiceInstruction)
		return starlark.None, nil
	}
}

func (instruction *AddServiceInstruction) GetPositionInOriginalScript() *InstructionPosition {
	return &instruction.position
}

func (instruction *AddServiceInstruction) GetCanonicalInstruction() string {
	// TODO(gb): implement when we need to return the canonicalized version of the script.
	//  Maybe there's a way to retrieve the serialized instruction from starlark-go
	return "add_service(...)"
}

func (instruction *AddServiceInstruction) Execute(ctx context.Context) error {
	serviceConfigMap := map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig{
		instruction.serviceId: instruction.serviceConfig,
	}

	serviceSuccessful, serviceFailed, err := instruction.serviceNetwork.StartServices(ctx, serviceConfigMap, service_network_types.PartitionID(""))
	if err != nil {
		return err
	}
	if serviceFailed[instruction.serviceId] != nil {
		return serviceFailed[instruction.serviceId]
	}
	if serviceSuccessful[instruction.serviceId] == nil {
		return stacktrace.NewError("Service wasn't accounted as failed not successfully added. This is a product bug")
	}
	return nil
}

func (instruction *AddServiceInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, *kurtosis_core_rpc_api_bindings.ServiceConfig, error) {
	// TODO(gb): Right now, we expect the Startosis script to be very "untyped" like:
	//  ```startosis
	//  my_service_port = struct(port = 1234, protocol = "TCP")
	//  my_service_config = struct(private_port = port, other_arg = "blah")
	//  ```
	//  But we can do better than this defining our own structures:
	//  ```
	//  my_service_port = port_spec(port = 1234, protocol = "TCP") # port() is a Startosis defined struct
	//  my_service_config = service_config(port = port, other_arg = "blah")
	//  ```
	//  With custom types, we can parse the args directly to our own Go types and potentially isolate the checks

	var serviceIdArg starlark.String
	var serviceConfigArg *starlarkstruct.Struct
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "service_id", &serviceIdArg, "service_config", &serviceConfigArg); err != nil {
		return "", nil, err
	}

	serviceId, err := ParseServiceId(serviceIdArg)
	if err != nil {
		return "", nil, err
	}

	serviceConfig, err := ParseServiceConfigArg(serviceConfigArg)
	if err != nil {
		return "", nil, err
	}
	return serviceId, serviceConfig, nil
}

func getPosition(thread *starlark.Thread) InstructionPosition {
	// TODO(gb): can do better by returning the entire callstack positions, but it's a good start
	if len(thread.CallStack()) == 0 {
		panic("empty call stack is unexpected, this should not happen")
	}
	// position of current instruction is  store at the bottom of the call stack
	callFrame := thread.CallStack().At(len(thread.CallStack()) - 1)
	return *NewInstructionPosition(callFrame.Pos.Line, callFrame.Pos.Col)
}
