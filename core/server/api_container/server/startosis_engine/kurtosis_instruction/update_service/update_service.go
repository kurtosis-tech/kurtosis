package update_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
	"reflect"
)

const (
	UpdateServiceBuiltinName   = "update_service"
	ServiceNameArgName         = "name"
	UpdateServiceConfigArgName = "config"

	descriptionFormatStr = "Updating config of service '%v'"
)

func NewUpdateService(
	serviceNetwork service_network.ServiceNetwork,
	interpretationTimeStore *interpretation_time_value_store.InterpretationTimeValueStore,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
	imageDownloadMode image_download_mode.ImageDownloadMode,
) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: UpdateServiceBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ServiceNameArgName)
					},
				},
				{
					Name:              UpdateServiceConfigArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*service_config.ServiceConfig],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// we just try to convert the configs here to validate their shape, to avoid code duplication with Interpret
						_, ok := value.(*service_config.ServiceConfig)
						if !ok {
							return startosis_errors.NewInterpretationError("The '%s' argument is not a ServiceConfig (was '%s').", UpdateServiceConfigArgName, reflect.TypeOf(value))
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &UpdateServiceCapabilities{
				interpretationTimeStore: interpretationTimeStore,
				serviceNetwork:          serviceNetwork,
				serviceName:             "",  // populated at interpretation time
				serviceConfig:           nil, // populated at interpretation time
				imageVal:                nil, // populated at interpretation time
				packageId:               packageId,
				packageContentProvider:  packageContentProvider,
				packageReplaceOptions:   packageReplaceOptions,
				description:             "", // populated at interpretation time
				imageDownloadMode:       imageDownloadMode,
			}
		},
		DefaultDisplayArguments: map[string]bool{
			ServiceNameArgName: true,
		},
	}
}

type UpdateServiceCapabilities struct {
	interpretationTimeStore *interpretation_time_value_store.InterpretationTimeValueStore
	serviceName             service.ServiceName
	serviceConfig           *service.ServiceConfig

	serviceNetwork service_network.ServiceNetwork

	// These params are needed to successfully convert service config if an ImageBuildSpec was provided
	packageId              string
	packageContentProvider startosis_packages.PackageContentProvider
	packageReplaceOptions  map[string]string
	imageVal               starlark.Value

	imageDownloadMode image_download_mode.ImageDownloadMode
	description       string
}

func (builtin *UpdateServiceCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	serviceNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}
	serviceName := service.ServiceName(serviceNameArgumentValue.GoString())

	builtin.serviceName = serviceName

	updatedServiceConfig, err := builtin_argument.ExtractArgumentValue[*service_config.ServiceConfig](arguments, UpdateServiceConfigArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", UpdateServiceConfigArgName)
	}
	rawImageVal, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Value](updatedServiceConfig.KurtosisValueTypeDefault, service_config.ImageAttr)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract raw image attribute.")
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Unable to extract image attribute off of service config.")
	}
	builtin.imageVal = rawImageVal
	updatedApiServiceConfig, _, interpretationErr := add_service.ValidateAndConvertConfigAndReadyCondition(
		builtin.serviceNetwork,
		updatedServiceConfig,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		builtin.packageId,
		builtin.packageContentProvider,
		builtin.packageReplaceOptions,
		builtin.imageDownloadMode,
	)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	// implement a Put Service Config that overwrites whatever service config existed for that service
	err = builtin.interpretationTimeStore.PutServiceConfig(serviceName, updatedApiServiceConfig)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while fetching service '%v' from the store", serviceName)
	}

	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.serviceName))
	return starlark.None, nil
}

func (builtin *UpdateServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if exists := validatorEnvironment.DoesServiceNameExist(builtin.serviceName); exists == startosis_validator.ComponentNotFound {
		return startosis_errors.NewValidationError("Service '%v' required by '%v' instruction doesn't exist", builtin.serviceName, UpdateServiceBuiltinName)
	}
	return nil
}

func (builtin *UpdateServiceCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// Note this is a no-op.
	return fmt.Sprintf("Fetched service '%v'", builtin.serviceName), nil
}

func (builtin *UpdateServiceCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *UpdateServiceCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(UpdateServiceBuiltinName).AddServiceName(builtin.serviceName)
}

func (builtin *UpdateServiceCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYaml) error {
	// update service does not affect the plan
	return nil
}

func (builtin *UpdateServiceCapabilities) Description() string {
	return builtin.description
}
