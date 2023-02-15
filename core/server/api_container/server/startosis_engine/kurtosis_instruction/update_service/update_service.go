package update_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	UpdateServiceBuiltinName = "update_service"

	ServiceNameArgName         = "service_name"
	UpdateServiceConfigArgName = "config"
)

func NewUpdateService(serviceNetwork service_network.ServiceNetwork) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: UpdateServiceBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              UpdateServiceConfigArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*kurtosis_types.UpdateServiceConfig],
					Validator:         nil,
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &UpdateServiceCapabilities{
				serviceNetwork: serviceNetwork,

				serviceName:         "",  // populated at interpretation time
				updateServiceConfig: nil, // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			ServiceNameArgName:         true,
			UpdateServiceConfigArgName: false,
		},
	}
}

type UpdateServiceCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	serviceName         service.ServiceName
	updateServiceConfig *kurtosis_core_rpc_api_bindings.UpdateServiceConfig
}

func (builtin *UpdateServiceCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	serviceName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}

	updateServiceConfig, err := builtin_argument.ExtractArgumentValue[*kurtosis_types.UpdateServiceConfig](arguments, UpdateServiceConfigArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", UpdateServiceConfigArgName)
	}

	builtin.serviceName = service.ServiceName(serviceName.GoString())
	builtin.updateServiceConfig = updateServiceConfig.ToKurtosisType()
	return starlark.None, nil
}

func (builtin *UpdateServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if partition_topology.ParsePartitionId(builtin.updateServiceConfig.Subnetwork) != partition_topology.DefaultPartitionId {
		if !validatorEnvironment.IsNetworkPartitioningEnabled() {
			return startosis_errors.NewValidationError("Service was about to be moved to subnetwork '%s' but the Kurtosis enclave was started with subnetwork capabilities disabled. Make sure to run the Starlark script with subnetwork enabled.", *builtin.updateServiceConfig.Subnetwork)
		}
	}
	if !validatorEnvironment.DoesServiceNameExist(builtin.serviceName) {
		return startosis_errors.NewValidationError("There was an error validating '%v' as service name '%v' does not exist", UpdateServiceBuiltinName, builtin.serviceName)
	}
	return nil
}

func (builtin *UpdateServiceCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	runningService, err := builtin.serviceNetwork.GetService(ctx, string(builtin.serviceName))
	if err != nil {
		return "", stacktrace.Propagate(err, "Updating service '%s' failed as it could not be retrieved from the enclave", builtin.serviceName)
	}

	updateServiceConfigMap := map[service.ServiceName]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig{
		builtin.serviceName: builtin.updateServiceConfig,
	}

	serviceSuccessful, serviceFailed, err := builtin.serviceNetwork.UpdateService(ctx, updateServiceConfigMap)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed updating service '%s' with an unexpected error", builtin.serviceName)
	}
	if failure, found := serviceFailed[builtin.serviceName]; found {
		return "", stacktrace.Propagate(failure, "Failed updating service '%s'", builtin.serviceNetwork)
	}
	_, found := serviceSuccessful[builtin.serviceName]
	if !found {
		return "", stacktrace.NewError("Service '%s' wasn't accounted as failed nor successfully updated. This is a product bug", builtin.serviceName)
	}
	instructionResult := fmt.Sprintf("Service '%s' with UUID '%s' updated", builtin.serviceName, runningService.GetRegistration().GetUUID())
	return instructionResult, nil
}
