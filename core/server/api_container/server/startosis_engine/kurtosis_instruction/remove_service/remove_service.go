package remove_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	RemoveServiceBuiltinName = "remove_service"

	ServiceNameArgName   = "name"
	descriptionFormatStr = "Removing service '%v'"
)

func NewRemoveService(serviceNetwork service_network.ServiceNetwork) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RemoveServiceBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// TODO: when #903 is merged, validate service name are non emtpy string
						return nil
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RemoveServiceCapabilities{
				serviceNetwork: serviceNetwork,

				serviceName: "", // populated at interpretation time
				description: "", // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			ServiceNameArgName: true,
		},
	}
}

type RemoveServiceCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	serviceName service.ServiceName
	description string
}

func (builtin *RemoveServiceCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	serviceName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}

	builtin.serviceName = service.ServiceName(serviceName.GoString())
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.serviceName))
	return starlark.None, nil
}

func (builtin *RemoveServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validatorEnvironment.DoesServiceNameExist(builtin.serviceName) == startosis_validator.ComponentNotFound {
		return startosis_errors.NewValidationError("There was an error validating '%v' as service name '%v' doesn't exist", RemoveServiceBuiltinName, builtin.serviceName)
	}
	validatorEnvironment.RemoveServiceName(builtin.serviceName)
	validatorEnvironment.RemoveServiceFromPrivatePortIDMapping(builtin.serviceName)
	validatorEnvironment.FreeMemory(builtin.serviceName)
	validatorEnvironment.FreeCPU(builtin.serviceName)
	return nil
}

func (builtin *RemoveServiceCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	serviceUUID, err := builtin.serviceNetwork.RemoveService(ctx, string(builtin.serviceName))
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed removing service with unexpected error")
	}
	instructionResult := fmt.Sprintf("Service '%s' with service UUID '%s' removed", builtin.serviceName, serviceUUID)
	return instructionResult, nil
}

func (builtin *RemoveServiceCapabilities) TryResolveWith(_ bool, _ *enclave_plan_persistence.EnclavePlanInstruction, _ *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	return enclave_structure.InstructionIsNotResolvableAbort
}

func (builtin *RemoveServiceCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(
		RemoveServiceBuiltinName,
	).AddServiceName(
		builtin.serviceName,
	)
}

func (builtin *RemoveServiceCapabilities) UpdatePlan(plan *plan_yaml.PlanYaml) error {
	err := plan.RemoveService(string(builtin.serviceName))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred updating plan with remove service.")
	}
	return nil
}

func (builtin *RemoveServiceCapabilities) Description() string {
	return builtin.description
}
