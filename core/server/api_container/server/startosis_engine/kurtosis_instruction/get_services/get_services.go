package get_services

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
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
	GetServicesBuiltinName = "get_services"
	descriptionStr         = "Fetching services"
)

func NewGetServices(interpretationTimeStore *interpretation_time_value_store.InterpretationTimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name:        GetServicesBuiltinName,
			Arguments:   []*builtin_argument.BuiltinArgument{},
			Deprecation: nil,
		},
		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &GetServicesCapabilities{
				interpretationTimeStore: interpretationTimeStore,
				serviceNames:            []service.ServiceName{}, // populated at interpretation time
				description:             "",                      // populated at interpretation time
			}
		},
		DefaultDisplayArguments: map[string]bool{},
	}
}

type GetServicesCapabilities struct {
	interpretationTimeStore *interpretation_time_value_store.InterpretationTimeValueStore
	serviceNames            []service.ServiceName
	description             string
}

func (builtin *GetServicesCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, descriptionStr)

	services, err := builtin.interpretationTimeStore.GetServices()
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while fetching service.")
	}
	servicesList := &starlark.List{}
	for _, serviceVal := range services {
		name, err := serviceVal.GetName()
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred getting name of service: %v", serviceVal)
		}
		builtin.serviceNames = append(builtin.serviceNames, name)

		_ = servicesList.Append(serviceVal)
	}

	return servicesList, nil
}

func (builtin *GetServicesCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// validate if all services exist in the validation environment
	for _, serviceName := range builtin.serviceNames {
		if exists := validatorEnvironment.DoesServiceNameExist(serviceName); exists == startosis_validator.ComponentNotFound {
			return startosis_errors.NewValidationError("Service '%v' required by '%v' instruction doesn't exist", serviceName, GetServicesBuiltinName)
		}
	}
	return nil
}

func (builtin *GetServicesCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// note: this is a no op
	return descriptionStr, nil
}

func (builtin *GetServicesCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *GetServicesCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(GetServicesBuiltinName)
}

func (builtin *GetServicesCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYaml) error {
	// get services does not affect the plan
	return nil
}

func (builtin *GetServicesCapabilities) Description() string {
	return builtin.description
}
