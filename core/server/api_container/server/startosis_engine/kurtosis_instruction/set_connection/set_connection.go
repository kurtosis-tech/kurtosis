package set_connection

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/connection_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"reflect"
)

const (
	SetConnectionBuiltinName = "set_connection"

	SubnetworksArgName      = "subnetworks"
	ConnectionConfigArgName = "config"
)

func NewSetConnection(serviceNetwork service_network.ServiceNetwork) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: SetConnectionBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              SubnetworksArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Tuple],
					Validator:         validateSubnetworks,
				},
				{
					Name:              ConnectionConfigArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*connection_config.ConnectionConfig],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// we just try to convert the configs here to validate their shape, to avoid code duplication
						// with Interpret
						if _, err := validateAndConvertConfig(value); err != nil {
							return err
						}
						return nil
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &SetConnectionCapabilities{
				serviceNetwork: serviceNetwork,

				optionalSubnetwork1: nil, // populated at interpretation time
				optionalSubnetwork2: nil, // populated at interpretation time
				connectionConfig:    nil, // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			SubnetworksArgName:      true,
			ConnectionConfigArgName: true,
		},
	}
}

type SetConnectionCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	// both are optional but one cannot go without the other. If one is set, the other should be set.
	// There's a XOR check in Interpret to ensure this
	optionalSubnetwork1 *service_network_types.PartitionID
	optionalSubnetwork2 *service_network_types.PartitionID

	connectionConfig *partition_topology.PartitionConnection
}

func (builtin *SetConnectionCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	connectionConfigStarlark, err := builtin_argument.ExtractArgumentValue[*connection_config.ConnectionConfig](arguments, ConnectionConfigArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ConnectionConfigArgName)
	}
	connectionConfig, interpretationErr := connectionConfigStarlark.ToKurtosisType()
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.connectionConfig = connectionConfig

	if !arguments.IsSet(SubnetworksArgName) {
		return starlark.None, nil
	}
	subnetworks, err := builtin_argument.ExtractArgumentValue[starlark.Tuple](arguments, SubnetworksArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SubnetworksArgName)
	}
	subnetwork1, subnetwork2, interpretationErr := shared_helpers.ParseSubnetworks(SubnetworksArgName, subnetworks)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.optionalSubnetwork1 = &subnetwork1
	builtin.optionalSubnetwork2 = &subnetwork2
	return starlark.None, nil
}

func (builtin *SetConnectionCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if !validatorEnvironment.IsNetworkPartitioningEnabled() {
		return startosis_errors.NewValidationError("Setting connection between two subnetworks cannot be performed because the Kurtosis enclave was started with subnetwork capabilities disabled. Make sure to run the Starlark script with subnetwork enabled.")
	}
	return nil
}

func (builtin *SetConnectionCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	var instructionResult string
	if builtin.optionalSubnetwork1 == nil {
		// if optionalSubnetwork1 is nil, optionalSubnetwork2 is nil as well and the default connection is being set
		if err := builtin.serviceNetwork.SetDefaultConnection(ctx, *builtin.connectionConfig); err != nil {
			return "", stacktrace.Propagate(err, "Failed setting default connection to %+v", builtin.connectionConfig)
		}
		instructionResult = "Configured default subnetwork connection"
	} else {
		subnetwork1 := *builtin.optionalSubnetwork1
		subnetwork2 := *builtin.optionalSubnetwork2
		if err := builtin.serviceNetwork.SetConnection(ctx, subnetwork1, subnetwork2, *builtin.connectionConfig); err != nil {
			return "", stacktrace.Propagate(err, "Failed setting connection between subnetwork '%s' and subnetwork '%s' with connection config %+v", subnetwork1, subnetwork2, builtin.connectionConfig)
		}
		instructionResult = fmt.Sprintf("Configured subnetwork connection between '%s' and '%s'", subnetwork1, subnetwork2)
	}
	return instructionResult, nil
}

func validateSubnetworks(value starlark.Value) *startosis_errors.InterpretationError {
	subnetworks, ok := value.(starlark.Tuple)
	if !ok {
		return startosis_errors.NewInterpretationError("'%s' argument should be a 'starlark.Tuple', got '%s'", SubnetworksArgName, reflect.TypeOf(value))
	}
	_, _, interpretationErr := shared_helpers.ParseSubnetworks(SubnetworksArgName, subnetworks)
	if interpretationErr != nil {
		return interpretationErr
	}
	return nil
}

func validateAndConvertConfig(rawConfig starlark.Value) (*partition_topology.PartitionConnection, *startosis_errors.InterpretationError) {
	starlarkConnectionConfig, ok := rawConfig.(*connection_config.ConnectionConfig)
	if !ok {
		return nil, startosis_errors.NewInterpretationError("The '%s' argument is not a ConnectionConfig (was '%s').", ConnectionConfigArgName, reflect.TypeOf(rawConfig))
	}
	connectionConfig, interpretationErr := starlarkConnectionConfig.ToKurtosisType()
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return connectionConfig, nil
}
