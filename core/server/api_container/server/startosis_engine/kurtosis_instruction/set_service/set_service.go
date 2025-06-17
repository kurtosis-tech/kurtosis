package set_service

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
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
)

const (
	SetServiceBuiltinName   = "set_service"
	ServiceNameArgName      = "name"
	SetServiceConfigArgName = "config"

	descriptionFormatStr = "Setting config of service '%v'"
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

	serviceConfigOverride, err := builtin_argument.ExtractArgumentValue[*service_config.ServiceConfig](arguments, SetServiceConfigArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SetServiceConfigArgName)
	}
	rawImageVal, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Value](serviceConfigOverride.KurtosisValueTypeDefault, service_config.ImageAttr)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract raw image attribute.")
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Unable to extract image attribute off of service config.")
	}
	builtin.imageVal = rawImageVal
	apiServiceConfigOverride, _, interpretationErr := shared_helpers.ValidateAndConvertConfigAndReadyCondition(
		builtin.serviceNetwork,
		serviceConfigOverride,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		builtin.packageId,
		builtin.packageContentProvider,
		builtin.packageReplaceOptions,
		builtin.imageDownloadMode,
	)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	// get original service config for service and merge with apiServiceConfigOverride
	currApiServiceConfig, err := builtin.interpretationTimeStore.GetServiceConfig(builtin.serviceName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred retrieving service config for service: %v'", builtin.serviceName)
	}

	mergedServiceConfig, err := upsertServiceConfigs(currApiServiceConfig, apiServiceConfigOverride)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while overriding service configs in set service for service: %v", builtin.serviceName)
	}
	builtin.serviceConfig = mergedServiceConfig
	builtin.interpretationTimeStore.SetServiceConfig(serviceName, mergedServiceConfig)

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

func (builtin *SetServiceCapabilities) TryResolveWith(instructionsAreEqual bool, other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual && enclaveComponents.HasServiceBeenUpdated(builtin.serviceName) {
		return enclave_structure.InstructionIsUpdate
	} else if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}

	// if service names are equal but the instructions are not equal, it means the service config has been updated.
	// The instruction should be rerun
	if !instructionsAreEqual {
		return enclave_structure.InstructionIsUpdate
	}

	return enclave_structure.InstructionIsUnknown
}

func (builtin *SetServiceCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(SetServiceBuiltinName).AddServiceName(builtin.serviceName)
}

func (builtin *SetServiceCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYamlGenerator) error {
	// update service does not affect the plan
	return nil
}

func (builtin *SetServiceCapabilities) Description() string {
	return builtin.description
}

// UpdateDependencyGraph updates the dependency graph with the effects of running this instruction.
func (builtin *SetServiceCapabilities) UpdateDependencyGraph(dependencyGraph *instructions_plan.InstructionsDependencyGraph) error {
	// TODO: Implement dependency graph updates for set_service instruction
	return nil
}

// Takes values set in [serviceConfigOverride] and sets them on [currServiceConfig], leaving other values of [currServiceConfig] untouched
func upsertServiceConfigs(currServiceConfig, serviceConfigOverride *service.ServiceConfig) (*service.ServiceConfig, error) {
	// only one of these image values will be set, the others will be nil or empty string
	// as the Starlark service config gurantees that the service config objects only has one set
	currServiceConfig.SetContainerImageName(serviceConfigOverride.GetContainerImageName())
	currServiceConfig.SetImageBuildSpec(serviceConfigOverride.GetImageBuildSpec())
	currServiceConfig.SetImageRegistrySpec(serviceConfigOverride.GetImageRegistrySpec())
	currServiceConfig.SetNixBuildSpec(serviceConfigOverride.GetNixBuildSpec())

	// for other fields, only override if they are explicitly set on serviceConfigOverride
	if cpuAllocationMillicpusOverride := serviceConfigOverride.GetCPUAllocationMillicpus(); cpuAllocationMillicpusOverride != 0 {
		currServiceConfig.SetCPUAllocationMillicpus(cpuAllocationMillicpusOverride)
	}
	if memoryAllocationMegabytesOverride := serviceConfigOverride.GetMemoryAllocationMegabytes(); memoryAllocationMegabytesOverride != 0 {
		currServiceConfig.SetMemoryAllocationMegabytes(memoryAllocationMegabytesOverride)
	}
	if minCPUAllocationMillicpusOverride := serviceConfigOverride.GetMinCPUAllocationMillicpus(); minCPUAllocationMillicpusOverride != 0 {
		currServiceConfig.SetMinCPUAllocationMillicpus(minCPUAllocationMillicpusOverride)
	}
	if minMemoryAllocationMegabytesOverride := serviceConfigOverride.GetMinMemoryAllocationMegabytes(); minMemoryAllocationMegabytesOverride != 0 {
		currServiceConfig.SetMinMemoryAllocationMegabytes(minMemoryAllocationMegabytesOverride)
	}
	if userOverride := serviceConfigOverride.GetUser(); userOverride != nil {
		currServiceConfig.SetUser(userOverride)
	}
	if labelsOverride := serviceConfigOverride.GetLabels(); len(labelsOverride) > 0 {
		currServiceConfig.SetLabels(labelsOverride)
	}
	if tolerationsOverride := serviceConfigOverride.GetTolerations(); len(tolerationsOverride) > 0 {
		currServiceConfig.SetTolerations(tolerationsOverride)
	}

	// TODO: impl logic for overriding entrypoint, cmd, env vars, and ports
	// TODO: note: these will require careful handling of future references that could be potentially be overriden and affect behavior
	if len(serviceConfigOverride.GetEnvVars()) != 0 {
		return nil, startosis_errors.NewInterpretationError("Overriding environment variables is currently not supported.")
	}
	if len(serviceConfigOverride.GetEntrypointArgs()) != 0 {
		return nil, startosis_errors.NewInterpretationError("Overriding entrypoint args is currently not supported.")
	}
	if len(serviceConfigOverride.GetCmdArgs()) != 0 {
		return nil, startosis_errors.NewInterpretationError("Overriding cmd args is currently not supported.")
	}
	if len(serviceConfigOverride.GetPrivatePorts()) != 0 {
		return nil, startosis_errors.NewInterpretationError("Overriding private ports is currently not supported. ")
	}
	if len(serviceConfigOverride.GetPublicPorts()) != 0 {
		return nil, startosis_errors.NewInterpretationError("Overriding public ports is currently not supported.")
	}
	if serviceConfigOverride.GetFilesArtifactsExpansion() != nil {
		return nil, startosis_errors.NewInterpretationError("Overriding files artifacts is currently not supported.")
	}
	if serviceConfigOverride.GetPersistentDirectories() != nil {
		return nil, startosis_errors.NewInterpretationError("Overriding persistent directories is currently not supported.")
	}

	return currServiceConfig, nil
}
