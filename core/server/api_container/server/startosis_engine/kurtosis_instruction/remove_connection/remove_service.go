package remove_connection

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	RemoveConnectionBuiltinName = "remove_connection"

	subnetworksArgName = "subnetworks"
)

func GenerateRemoveConnectionBuiltin(
	instructionsQueue *[]kurtosis_instruction.KurtosisInstruction,
	serviceNetwork service_network.ServiceNetwork,
) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		newInstruction := newEmptyRemoveConnectionInstruction(serviceNetwork, instructionPosition)
		if interpretationError := newInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, newInstruction)
		return starlark.None, nil
	}
}

type RemoveConnectionInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	subnetwork1 service_network_types.PartitionID
	subnetwork2 service_network_types.PartitionID
}

func newEmptyRemoveConnectionInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition) *RemoveConnectionInstruction {
	return &RemoveConnectionInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		starlarkKwargs: starlark.StringDict{},
		subnetwork1:    "",
		subnetwork2:    "",
	}
}

func NewRemoveConnectionInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, subnetwork1 service_network_types.PartitionID, subnetwork2 service_network_types.PartitionID, starlarkKwargs starlark.StringDict) *RemoveConnectionInstruction {
	return &RemoveConnectionInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		subnetwork1:    subnetwork1,
		subnetwork2:    subnetwork2,
		starlarkKwargs: starlarkKwargs,
	}
}

func (instruction *RemoveConnectionInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *RemoveConnectionInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[subnetworksArgName]), subnetworksArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), RemoveConnectionBuiltinName, instruction.String(), args)
}

func (instruction *RemoveConnectionInstruction) Execute(ctx context.Context) (*string, error) {
	if err := instruction.serviceNetwork.UnsetConnection(ctx, instruction.subnetwork1, instruction.subnetwork2); err != nil {
		return nil, stacktrace.Propagate(err, "Failed setting connection between subnetwork '%s' and subnetwork '%s'", instruction.subnetwork1, instruction.subnetwork2)
	}
	instructionResult := fmt.Sprintf("Removed subnetwork connection override between '%s' and '%s'", instruction.subnetwork1, instruction.subnetwork2)
	return &instructionResult, nil
}

func (instruction *RemoveConnectionInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// TODO(gb):  validate that network partitioning is enabled at the network service level
	return nil
}

func (instruction *RemoveConnectionInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(RemoveConnectionBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *RemoveConnectionInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var subnetworks starlark.Tuple
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, subnetworksArgName, &subnetworks); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", RemoveConnectionBuiltinName, args, kwargs)
	}
	instruction.starlarkKwargs[subnetworksArgName] = subnetworks
	instruction.starlarkKwargs.Freeze()

	subnetwork1, subnetwork2, interpretationErr := kurtosis_instruction.ParseSubnetworks(subnetworks)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.subnetwork1 = subnetwork1
	instruction.subnetwork2 = subnetwork2
	return nil
}
