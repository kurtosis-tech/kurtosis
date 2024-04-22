package set_service

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
	SetServiceBuiltinName   = "set_service"
	ServiceNameArgName      = "name"
	SetServiceConfigArgName = "config"

	descriptionFormatStr = "Updating config of service '%v'"
)

func NewSetService(
	serviceNetwork service_network.ServiceNetwork,
	interpretationTimeStore *interpretation_time_value_store.InterpretationTimeValueStore,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
	imageDownloadMode image_download_mode.ImageDownloadMode,
) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: SetServiceBuiltinName,
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
					Name:              SetServiceConfigArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*service_config.ServiceConfig],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// we just try to convert the configs here to validate their shape, to avoid code duplication with Interpret
						_, ok := value.(*service_config.ServiceConfig)
						if !ok {
							return startosis_errors.NewInterpretationError("The '%s' argument is not a ServiceConfig (was '%s').", SetServiceConfigArgName, reflect.TypeOf(value))
						}
						return nil
					},
				},
			},
			Deprecation: nil,
		},
		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &SetServiceCapabilities{
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

type SetServiceCapabilities struct {
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

func (builtin *SetServiceCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	serviceNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}
	serviceName := service.ServiceName(serviceNameArgumentValue.GoString())

	builtin.serviceName = serviceName

	updatedServiceConfig, err := builtin_argument.ExtractArgumentValue[*service_config.ServiceConfig](arguments, SetServiceConfigArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SetServiceConfigArgName)
	}
	rawImageVal, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Value](updatedServiceConfig.KurtosisValueTypeDefault, service_config.ImageAttr)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract raw image attribute.")
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Unable to extract image attribute off of service config.")
	}
	builtin.imageVal = rawImageVal
	apiServiceConfigOverride, _, interpretationErr := add_service.ValidateAndConvertConfigAndReadyCondition(
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

	// get existing service config for service and merge with apiServiceConfigOverride
	currApiServiceConfig, err := builtin.interpretationTimeStore.GetServiceConfig(builtin.serviceName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred retrieving service config for service: %v'", builtin.serviceName)
	}

	builtin.interpretationTimeStore.PutServiceConfig(serviceName, upsertServiceConfigs(currApiServiceConfig, apiServiceConfigOverride))

	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.serviceName))
	return starlark.None, nil
}

func (builtin *SetServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if exists := validatorEnvironment.DoesServiceNameExist(builtin.serviceName); exists == startosis_validator.ComponentNotFound {
		return startosis_errors.NewValidationError("Service '%v' required by '%v' instruction doesn't exist", builtin.serviceName, SetServiceBuiltinName)
	}
	return nil
}

func (builtin *SetServiceCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// Note this is a no-op.
	return fmt.Sprintf("Set service config on service '%v'.", builtin.serviceName), nil
}

func (builtin *SetServiceCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *SetServiceCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(SetServiceBuiltinName).AddServiceName(builtin.serviceName)
}

func (builtin *SetServiceCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYaml) error {
	// update service does not affect the plan
	return nil
}

func (builtin *SetServiceCapabilities) Description() string {
	return builtin.description
}

// Takes values set in [serviceConfigOverride] and sets them on [currServiceConfig], leaving other values of [currServiceConfig] untouched
func upsertServiceConfigs(currServiceConfig, serviceConfigOverride *service.ServiceConfig) *service.ServiceConfig {
	// only one of these image values will be set, the others will be nil or empty string
	// as Starlark service config gurantees that the both service config objects has one set
	currServiceConfig.SetContainerImageName(serviceConfigOverride.GetContainerImageName())
	currServiceConfig.SetImageBuildSpec(serviceConfigOverride.GetImageBuildSpec())
	currServiceConfig.SetImageRegistrySpec(serviceConfigOverride.GetImageRegistrySpec())
	currServiceConfig.SetNixBuildSpec(serviceConfigOverride.GetNixBuildSpec())

	// TODO: impl logic for updating other fields
	// TODO: note: entrypoint, cmd, env vars, and ports will require special behavior to handle future references that could be overriden
	return currServiceConfig
}
