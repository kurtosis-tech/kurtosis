package set_connection

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	SetConnectionBuiltinName = "set_connection"

	subnetworksArgName      = "subnetworks"
	connectionConfigArgName = "config"
)

func GenerateSetConnectionBuiltin(
	instructionsQueue *[]kurtosis_instruction.KurtosisInstruction,
	serviceNetwork service_network.ServiceNetwork,
) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		newInstruction := newEmptySetConnectionInstruction(serviceNetwork, instructionPosition)
		if interpretationError := newInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, newInstruction)
		return starlark.None, nil
	}
}

type SetConnectionInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	// both are optional but one cannot go without the other. If one is set, the other should be set.
	// There's a XOR check in parseStartosisArgs to ensure this
	optionalSubnetwork1 *service_network_types.PartitionID
	optionalSubnetwork2 *service_network_types.PartitionID

	connectionConfig partition_topology.PartitionConnection
}

func NewSetConnectionInstruction(
	serviceNetwork service_network.ServiceNetwork,
	position *kurtosis_instruction.InstructionPosition,
	optionalSubnetwork1 *service_network_types.PartitionID,
	optionalSubnetwork2 *service_network_types.PartitionID,
	connectionConfig partition_topology.PartitionConnection,
	starlarkKwargs starlark.StringDict,
) *SetConnectionInstruction {
	return &SetConnectionInstruction{
		serviceNetwork:      serviceNetwork,
		position:            position,
		starlarkKwargs:      starlarkKwargs,
		optionalSubnetwork1: optionalSubnetwork1,
		optionalSubnetwork2: optionalSubnetwork2,
		connectionConfig:    connectionConfig,
	}
}

func newEmptySetConnectionInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition) *SetConnectionInstruction {
	return &SetConnectionInstruction{
		serviceNetwork:      serviceNetwork,
		position:            position,
		starlarkKwargs:      starlark.StringDict{},
		optionalSubnetwork1: nil,
		optionalSubnetwork2: nil,
		connectionConfig:    partition_topology.ConnectionAllowed,
	}
}

func (instruction *SetConnectionInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *SetConnectionInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	var args []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg
	if instruction.optionalSubnetwork1 == nil {
		args = []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[connectionConfigArgName]), connectionConfigArgName, kurtosis_instruction.Representative),
		}
	} else {
		args = []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[subnetworksArgName]), subnetworksArgName, kurtosis_instruction.Representative),
			binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[connectionConfigArgName]), connectionConfigArgName, kurtosis_instruction.Representative),
		}
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), SetConnectionBuiltinName, instruction.String(), args)
}

func (instruction *SetConnectionInstruction) Execute(ctx context.Context) (*string, error) {
	var instructionResult string
	if instruction.optionalSubnetwork1 == nil {
		// if optionalSubnetwork1 is nil, optionalSubnetwork2 is nil as well and the default connection is being set
		if err := instruction.serviceNetwork.SetDefaultConnection(ctx, instruction.connectionConfig); err != nil {
			return nil, stacktrace.Propagate(err, "Failed setting default connection to %+v", instruction.connectionConfig)
		}
		instructionResult = "Configured default subnetwork connection"
	} else {
		subnetwork1 := *instruction.optionalSubnetwork1
		subnetwork2 := *instruction.optionalSubnetwork2
		if err := instruction.serviceNetwork.SetConnection(ctx, subnetwork1, subnetwork2, instruction.connectionConfig); err != nil {
			return nil, stacktrace.Propagate(err, "Failed setting connection between subnetwork '%s' and subnetwork '%s' with connection config %+v", subnetwork1, subnetwork2, instruction.connectionConfig)
		}
		instructionResult = fmt.Sprintf("Configured subnetwork connection between '%s' and '%s'", subnetwork1, subnetwork2)
	}
	return &instructionResult, nil
}

func (instruction *SetConnectionInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// TODO(gb):  validate that network partitioning is enabled at the network service level
	return nil
}

func (instruction *SetConnectionInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(SetConnectionBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *SetConnectionInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var subnetworks starlark.Tuple
	var starlarkConnectionConfig *kurtosis_types.ConnectionConfig
	if len(args)+len(kwargs) == 1 {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, connectionConfigArgName, &starlarkConnectionConfig); err != nil {
			return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", SetConnectionBuiltinName, args, kwargs)
		}
		instruction.starlarkKwargs[connectionConfigArgName] = starlarkConnectionConfig
	} else {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, subnetworksArgName, &subnetworks, connectionConfigArgName, &starlarkConnectionConfig); err != nil {
			return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", SetConnectionBuiltinName, args, kwargs)
		}
		instruction.starlarkKwargs[subnetworksArgName] = subnetworks
		instruction.starlarkKwargs[connectionConfigArgName] = starlarkConnectionConfig
	}
	instruction.starlarkKwargs.Freeze()

	if subnetworks != nil {
		subnetwork1, subnetwork2, interpretationErr := kurtosis_instruction.ParseSubnetworks(subnetworks)
		if interpretationErr != nil {
			return interpretationErr
		}
		instruction.optionalSubnetwork1 = &subnetwork1
		instruction.optionalSubnetwork2 = &subnetwork2

		// this is a XOR. We need either the 2 to be nil, or none.
		if (instruction.optionalSubnetwork1 == nil) != (instruction.optionalSubnetwork2 == nil) {
			return startosis_errors.NewInterpretationError("One of subnetwork1 subnetwork2 was nil but not the other. This is a Kurtosis bug ('%v', '%v')", instruction.optionalSubnetwork1, instruction.optionalSubnetwork2)
		}
	}

	instruction.connectionConfig = starlarkConnectionConfig.ToKurtosisType()
	return nil
}
