package remove_connection

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"reflect"
)

const (
	RemoveConnectionBuiltinName = "remove_connection"

	SubnetworksArgName = "subnetworks"
)

func NewRemoveConnection(serviceNetwork service_network.ServiceNetwork) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RemoveConnectionBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              SubnetworksArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Tuple],
					Validator:         validateSubnetworks,
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RemoveConnectionCapabilities{
				serviceNetwork: serviceNetwork,

				subnetwork1: "", // populated at interpretation time
				subnetwork2: "", // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			SubnetworksArgName: true,
		},
	}
}

type RemoveConnectionCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	subnetwork1 service_network_types.PartitionID
	subnetwork2 service_network_types.PartitionID
}

func (builtin *RemoveConnectionCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	subnetworks, err := builtin_argument.ExtractArgumentValue[starlark.Tuple](arguments, SubnetworksArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SubnetworksArgName)
	}
	subnetwork1, subnetwork2, interpretationErr := kurtosis_instruction.ParseSubnetworks(subnetworks)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.subnetwork1 = subnetwork1
	builtin.subnetwork2 = subnetwork2
	return starlark.None, nil
}

func (builtin *RemoveConnectionCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if !validatorEnvironment.IsNetworkPartitioningEnabled() {
		return startosis_errors.NewValidationError("Removing connection between two subnetworks cannot be performed because the Kurtosis enclave was started with subnetwork capabilities disabled. Make sure to run the Starlark script with subnetwork enabled.")
	}
	return nil
}

func (builtin *RemoveConnectionCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	if err := builtin.serviceNetwork.UnsetConnection(ctx, builtin.subnetwork1, builtin.subnetwork2); err != nil {
		return "", stacktrace.Propagate(err, "Failed setting connection between subnetwork '%s' and subnetwork '%s'", builtin.subnetwork1, builtin.subnetwork2)
	}
	instructionResult := fmt.Sprintf("Removed subnetwork connection override between '%s' and '%s'", builtin.subnetwork1, builtin.subnetwork2)
	return instructionResult, nil
}

func validateSubnetworks(value starlark.Value) *startosis_errors.InterpretationError {
	subnetworks, ok := value.(starlark.Tuple)
	if !ok {
		return startosis_errors.NewInterpretationError("'%s' argument should be a 'starlark.Tuple', got '%s'", SubnetworksArgName, reflect.TypeOf(value))
	}
	_, _, interpretationErr := kurtosis_instruction.ParseSubnetworks(subnetworks)
	if interpretationErr != nil {
		return interpretationErr
	}
	return nil
}
