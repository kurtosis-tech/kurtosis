package get_service

import (
	"context"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
)

const (
	GetServiceBuiltinName = "get_service"
	ServiceNameArgName    = "name"

	descriptionFormatStr = "Fetching service '%v'"
)

func NewGetService(interpretationTimeStore *interpretation_time_value_store.InterpretationTimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: GetServiceBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ServiceNameArgName)
					},
				},
			},
			Deprecation: nil,
		},
		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &GetServiceCapabilities{
				interpretationTimeStore: interpretationTimeStore,
				serviceName:             "",
				description:             "", // populated at interpretation time
			}
		},
		DefaultDisplayArguments: map[string]bool{
			ServiceNameArgName: true,
		},
	}
}

type GetServiceCapabilities struct {
	interpretationTimeStore *interpretation_time_value_store.InterpretationTimeValueStore
	serviceName             service.ServiceName
	description             string
}

func (builtin *GetServiceCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	serviceNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}
	serviceName := service.ServiceName(serviceNameArgumentValue.GoString())

	builtin.serviceName = serviceName
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.serviceName))

	serviceStarlarkValue, err := builtin.interpretationTimeStore.GetService(serviceName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while fetching service '%v' from the store", serviceName)
	}

	return serviceStarlarkValue, nil
}

func (builtin *GetServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if exists := validatorEnvironment.DoesServiceNameExist(builtin.serviceName); exists == startosis_validator.ComponentNotFound {
		return startosis_errors.NewValidationError("Service '%v' required by '%v' instruction doesn't exist", builtin.serviceName, GetServiceBuiltinName)
	}
	return nil
}

func (builtin *GetServiceCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// Note this is a no-op.
	// Perhaps this instruction should be like `read_file` instead and not a part of any plan
	// But that shouldn't be done outside a function; so it's here for now
	return fmt.Sprintf("Fetched service '%v'", builtin.serviceName), nil
}

func (builtin *GetServiceCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual && enclaveComponents.HasServiceBeenUpdated(builtin.serviceName) {
		return enclave_structure.InstructionIsUpdate
	} else if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *GetServiceCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(GetServiceBuiltinName).AddServiceName(builtin.serviceName)
}

func (builtin *GetServiceCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYamlGenerator) error {
	// get service does not affect the plan
	return nil
}

func (builtin *GetServiceCapabilities) Description() string {
	return builtin.description
}

// UpdateDependencyGraph updates the dependency graph with the effects of running this instruction.
func (builtin *GetServiceCapabilities) UpdateDependencyGraph(instructionUuid dependency_graph.ScheduledInstructionUuid, dependencyGraph *dependency_graph.InstructionsDependencyGraph) error {
	// TODO: Implement dependency graph updates for get_service instruction
	return nil
}
