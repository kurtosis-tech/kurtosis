package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"reflect"
)

const (
	AddServiceBuiltinName = "add_service"

	ServiceNameArgName   = "name"
	ServiceConfigArgName = "config"

	addServiceDescriptionFormatStr = "Adding service with name '%v' and image '%v'"
)

func NewAddService(
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore,
	imageDownloadMode image_download_mode.ImageDownloadMode) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: AddServiceBuiltinName,

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
					Name:              ServiceConfigArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*service_config.ServiceConfig],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// we just try to convert the configs here to validate their shape, to avoid code duplication with Interpret
						_, ok := value.(*service_config.ServiceConfig)
						if !ok {
							return startosis_errors.NewInterpretationError("The '%s' argument is not a ServiceConfig (was '%s').", ConfigsArgName, reflect.TypeOf(value))
						}
						return nil
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &AddServiceCapabilities{
				serviceNetwork:         serviceNetwork,
				runtimeValueStore:      runtimeValueStore,
				packageId:              packageId,
				packageContentProvider: packageContentProvider,
				packageReplaceOptions:  packageReplaceOptions,
				serviceName:            "",  // populated at interpretation time
				serviceConfig:          nil, // populated at interpretation time

				resultUuid:     "",  // populated at interpretation time
				readyCondition: nil, // populated at interpretation time

				interpretationTimeValueStore: interpretationTimeValueStore,
				description:                  "",  // populated at interpretation time
				returnValue:                  nil, // populated at interpretation time
				imageVal:                     nil, // populated at interpretation time
				imageDownloadMode:            imageDownloadMode,
			}
		},

		DefaultDisplayArguments: map[string]bool{
			ServiceNameArgName:   true,
			ServiceConfigArgName: true,
		},
	}
}

type AddServiceCapabilities struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	serviceName    service.ServiceName
	serviceConfig  *service.ServiceConfig
	readyCondition *service_config.ReadyCondition

	// These params are needed to successfully convert service config if an ImageBuildSpec was provided
	packageId              string
	packageContentProvider startosis_packages.PackageContentProvider
	packageReplaceOptions  map[string]string
	imageVal               starlark.Value

	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore

	resultUuid  string
	returnValue *kurtosis_types.Service
	description string

	imageDownloadMode image_download_mode.ImageDownloadMode
}

func (builtin *AddServiceCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	serviceName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}

	serviceConfig, err := builtin_argument.ExtractArgumentValue[*service_config.ServiceConfig](arguments, ServiceConfigArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceConfigArgName)
	}
	rawImageVal, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Value](serviceConfig.KurtosisValueTypeDefault, service_config.ImageAttr)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract raw image attribute.")
	}
	if !found {
		return nil, startosis_errors.NewInterpretationError("Unable to extract image attribute off of service config.")
	}
	builtin.imageVal = rawImageVal
	apiServiceConfig, readyCondition, interpretationErr := shared_helpers.ValidateAndConvertConfigAndReadyCondition(
		builtin.serviceNetwork,
		serviceConfig,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		builtin.packageId,
		builtin.packageContentProvider,
		builtin.packageReplaceOptions,
		builtin.imageDownloadMode,
	)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	builtin.serviceName = service.ServiceName(serviceName.GoString())
	builtin.serviceConfig = apiServiceConfig
	builtin.readyCondition = readyCondition
	builtin.resultUuid, err = builtin.runtimeValueStore.GetOrCreateValueAssociatedWithService(builtin.serviceName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to create runtime value to hold '%v' command return values", AddServiceBuiltinName)
	}

	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(addServiceDescriptionFormatStr, builtin.serviceName, builtin.serviceConfig.GetContainerImageName()))

	builtin.returnValue, interpretationErr = makeAddServiceInterpretationReturnValue(serviceName, builtin.serviceConfig, builtin.resultUuid)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	err = builtin.interpretationTimeValueStore.PutService(builtin.serviceName, builtin.returnValue)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while persisting return value for service '%v'", serviceName)
	}
	builtin.interpretationTimeValueStore.PutServiceConfig(builtin.serviceName, builtin.serviceConfig)

	return builtin.returnValue, nil
}

func (builtin *AddServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validationErr := validateSingleService(validatorEnvironment, builtin.serviceName, builtin.serviceConfig); validationErr != nil {
		return validationErr
	}
	return nil
}

func (builtin *AddServiceCapabilities) Execute(ctx context.Context, arguments *builtin_argument.ArgumentValuesSet) (string, error) {
	// update service config to use new service config set by a set_service instruction, if one exists
	if builtin.interpretationTimeValueStore.ExistsNewServiceConfigForService(builtin.serviceName) {
		newServiceConfig, err := builtin.interpretationTimeValueStore.GetNewServiceConfig(builtin.serviceName)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred retrieving a new service config '%s'.", builtin.serviceName)
		}
		builtin.serviceConfig = newServiceConfig
		builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(addServiceDescriptionFormatStr, builtin.serviceName, newServiceConfig.GetContainerImageName()))
	}
	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(builtin.runtimeValueStore, builtin.serviceName, builtin.serviceConfig)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred replace a magic string in '%s' instruction arguments for service '%s'. Execution cannot proceed", AddServiceBuiltinName, builtin.serviceName)
	}
	var startedService *service.Service
	exist, err := builtin.serviceNetwork.ExistServiceRegistration(builtin.serviceName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting service registration for service '%s'", builtin.serviceName)
	}
	if exist {
		startedService, err = builtin.serviceNetwork.UpdateService(ctx, replacedServiceName, replacedServiceConfig)
	} else {
		startedService, err = builtin.serviceNetwork.AddService(ctx, replacedServiceName, replacedServiceConfig)
	}
	if err != nil {
		return "", stacktrace.Propagate(err, "Unexpected error occurred starting service '%s'", replacedServiceName)
	}

	if err := runServiceReadinessCheck(
		ctx,
		builtin.serviceNetwork,
		builtin.runtimeValueStore,
		replacedServiceName,
		builtin.readyCondition,
	); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while checking if service '%v' is ready", replacedServiceName)
	}

	if err := fillAddServiceReturnValueWithRuntimeValues(startedService, builtin.resultUuid, builtin.runtimeValueStore); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while adding service return values with result key UUID '%s'", builtin.resultUuid)
	}
	instructionResult := fmt.Sprintf("Service '%s' added with service UUID '%s'", replacedServiceName, startedService.GetRegistration().GetUUID())
	return instructionResult, nil
}

func (builtin *AddServiceCapabilities) TryResolveWith(instructionsAreEqual bool, other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	// if other instruction is nil or other instruction is not an add_service instruction, status is unknown
	if other == nil {
		enclaveComponents.AddService(builtin.serviceName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	if other.Type != AddServiceBuiltinName {
		enclaveComponents.AddService(builtin.serviceName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	// if service names don't match, status is unknown, instructions can't be resolved together
	if !other.HasOnlyServiceName(builtin.serviceName) {
		enclaveComponents.AddService(builtin.serviceName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	// if service names are equal but the instructions are not equal, it means the service config has been updated.
	// The instruction should be rerun
	if !instructionsAreEqual {
		enclaveComponents.AddService(builtin.serviceName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}

	// From here instructions are equal
	// We check if there has been some updates to the files it's mounting. If that's the case, it should be rerun
	filesArtifactsExpansion := builtin.serviceConfig.GetFilesArtifactsExpansion()
	if filesArtifactsExpansion != nil {
		for _, filesArtifactNames := range filesArtifactsExpansion.ServiceDirpathsToArtifactIdentifiers {
			for _, filesArtifactName := range filesArtifactNames {
				if enclaveComponents.HasFilesArtifactBeenUpdated(filesArtifactName) {
					enclaveComponents.AddService(builtin.serviceName, enclave_structure.ComponentIsUpdated)
					return enclave_structure.InstructionIsUpdate
				}
			}
		}
	}

	// We check if service config was changed by a set_service instruction. If that's the case, it should be rerun
	if builtin.interpretationTimeValueStore.ExistsNewServiceConfigForService(builtin.serviceName) {
		enclaveComponents.AddService(builtin.serviceName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}

	enclaveComponents.AddService(builtin.serviceName, enclave_structure.ComponentWasLeftIntact)
	return enclave_structure.InstructionIsEqual
}

func (builtin *AddServiceCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(
		AddServiceBuiltinName,
	).AddServiceName(
		builtin.serviceName,
	)
}

func (builtin *AddServiceCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYaml) error {
	var buildContextLocator string
	var targetStage string
	var registryAddress string
	var interpretationErr *startosis_errors.InterpretationError

	// set image values based on type of image
	if builtin.imageVal != nil {
		switch starlarkImgVal := builtin.imageVal.(type) {
		case *service_config.ImageBuildSpec:
			buildContextLocator, interpretationErr = starlarkImgVal.GetBuildContextLocator()
			if interpretationErr != nil {
				return startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred getting build context locator")
			}
			targetStage, interpretationErr = starlarkImgVal.GetTargetStage()
			if interpretationErr != nil {
				return startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred getting target stage.")
			}
		case *service_config.ImageSpec:
			registryAddress, interpretationErr = starlarkImgVal.GetRegistryAddrIfSet()
			if interpretationErr != nil {
				return startosis_errors.WrapWithInterpretationError(interpretationErr, "An error occurred getting registry address.")
			}
		default:
			// assume NixBuildSpec or regular image
		}
	}

	err := planYaml.AddService(builtin.serviceName, builtin.returnValue, builtin.serviceConfig, buildContextLocator, targetStage, registryAddress)
	if err != nil {
		return stacktrace.NewError("An error occurred updating the plan with service: %v", builtin.serviceName)
	}
	return nil
}

func (builtin *AddServiceCapabilities) Description() string {
	return builtin.description
}
