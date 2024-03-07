package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
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
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
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
				description:                  "", // populated at interpretation time
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

	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore

	resultUuid  string
	description string
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
	apiServiceConfig, readyCondition, interpretationErr := validateAndConvertConfigAndReadyCondition(
		builtin.serviceNetwork,
		serviceConfig,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		builtin.packageId,
		builtin.packageContentProvider,
		builtin.packageReplaceOptions,
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

	returnValue, interpretationErr := makeAddServiceInterpretationReturnValue(serviceName, builtin.serviceConfig, builtin.resultUuid)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	err = builtin.interpretationTimeValueStore.PutService(builtin.serviceName, returnValue)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while persisting return value for service '%v'", serviceName)
	}
	return returnValue, nil
}

func (builtin *AddServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validationErr := validateSingleService(validatorEnvironment, builtin.serviceName, builtin.serviceConfig); validationErr != nil {
		return validationErr
	}
	return nil
}

func (builtin *AddServiceCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
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
	//kurtosisInstruction := addServiceInstruction.GetInstruction()
	//arguments := kurtosisInstruction.GetArguments()
	//
	//// start building Service Yaml object
	//service := &Service{} //nolint:exhaustruct
	//uuid := pyg.generateUuid()
	//service.Uuid = strconv.Itoa(uuid)
	//
	//// store future references of this service
	//returnValue := addServiceInstruction.GetReturnedValue()
	//returnedService, ok := returnValue.(*kurtosis_types.Service)
	//if !ok {
	//	return stacktrace.NewError("Cast to service didn't work")
	//}
	//futureRefIPAddress, err := returnedService.GetIpAddress()
	//if err != nil {
	//	return err
	//}
	//pyg.futureReferenceIndex[futureRefIPAddress] = fmt.Sprintf("{{ kurtosis.%v.ip_address }}", uuid)
	//futureRefHostName, err := returnedService.GetHostname()
	//if err != nil {
	//	return err
	//}
	//pyg.futureReferenceIndex[futureRefHostName] = fmt.Sprintf("{{ kurtosis.%v.hostname }}", uuid)
	//
	//var regErr error
	//serviceName, regErr := builtin_argument.ExtractArgumentValue[starlark.String](arguments, add_service.ServiceNameArgName)
	//if regErr != nil {
	//	return startosis_errors.WrapWithInterpretationError(regErr, "Unable to extract value for '%s' argument", add_service.ServiceNameArgName)
	//}
	//service.Name = pyg.swapFutureReference(serviceName.GoString()) // swap future references in the strings
	//
	//starlarkServiceConfig, regErr := builtin_argument.ExtractArgumentValue[*service_config.ServiceConfig](arguments, add_service.ServiceConfigArgName)
	//if regErr != nil {
	//	return startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", add_service.ServiceConfigArgName)
	//}
	//serviceConfig, serviceConfigErr := starlarkServiceConfig.ToKurtosisType( // is this an expensive call? // TODO: add this error back in
	//	pyg.serviceNetwork,
	//	kurtosisInstruction.GetPositionInOriginalScript().GetFilename(),
	//	pyg.planYaml.PackageId,
	//	pyg.packageContentProvider,
	//	pyg.packageReplaceOptions)
	//if serviceConfigErr != nil {
	//	return serviceConfigErr
	//}
	//
	//// get image info
	//rawImageAttrValue, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Value](starlarkServiceConfig.KurtosisValueTypeDefault, service_config.ImageAttr)
	//if interpretationErr != nil {
	//	return interpretationErr
	//}
	//image := &ImageSpec{ //nolint:exhaustruct
	//	ImageName: serviceConfig.GetContainerImageName(),
	//}
	//imageBuildSpec := serviceConfig.GetImageBuildSpec()
	//if imageBuildSpec != nil {
	//	switch img := rawImageAttrValue.(type) {
	//	case *service_config.ImageBuildSpec:
	//		contextLocator, err := img.GetBuildContextLocator()
	//		if err != nil {
	//			return err
	//		}
	//		image.BuildContextLocator = contextLocator
	//	}
	//	image.TargetStage = imageBuildSpec.GetTargetStage()
	//}
	//imageSpec := serviceConfig.GetImageRegistrySpec()
	//if imageSpec != nil {
	//	image.Registry = imageSpec.GetRegistryAddr()
	//}
	//service.Image = image
	//
	//// detect future references
	//cmdArgs := []string{}
	//for _, cmdArg := range serviceConfig.GetCmdArgs() {
	//	realCmdArg := pyg.swapFutureReference(cmdArg)
	//	cmdArgs = append(cmdArgs, realCmdArg)
	//}
	//service.Cmd = cmdArgs
	//
	//entryArgs := []string{}
	//for _, entryArg := range serviceConfig.GetEntrypointArgs() {
	//	realEntryArg := pyg.swapFutureReference(entryArg)
	//	entryArgs = append(entryArgs, realEntryArg)
	//}
	//service.Entrypoint = entryArgs
	//
	//// ports
	//service.Ports = []*Port{}
	//for portName, configPort := range serviceConfig.GetPrivatePorts() { // TODO: support public ports
	//
	//	port := &Port{ //nolint:exhaustruct
	//		TransportProtocol: TransportProtocol(configPort.GetTransportProtocol().String()),
	//		Name:              portName,
	//		Number:            configPort.GetNumber(),
	//	}
	//	if configPort.GetMaybeApplicationProtocol() != nil {
	//		port.ApplicationProtocol = ApplicationProtocol(*configPort.GetMaybeApplicationProtocol())
	//	}
	//
	//	service.Ports = append(service.Ports, port)
	//}
	//
	//// env vars
	//service.EnvVars = []*EnvironmentVariable{}
	//for key, val := range serviceConfig.GetEnvVars() {
	//	// detect and future references
	//	value := pyg.swapFutureReference(val)
	//	envVar := &EnvironmentVariable{
	//		Key:   key,
	//		Value: value,
	//	}
	//	service.EnvVars = append(service.EnvVars, envVar)
	//}
	//
	//// file mounts have two cases:
	//// 1. the referenced files artifact already exists in the plan, in which case add the referenced files artifact
	//// 2. the referenced files artifact does not already exist in the plan, in which case the file MUST have been passed in via a top level arg OR is invalid
	//// 	  in this case,
	//// 	  - create new files artifact
	////	  - add it to the service's file mount accordingly
	////	  - add the files artifact to the plan
	//service.Files = []*FileMount{}
	//serviceFilesArtifactExpansions := serviceConfig.GetFilesArtifactsExpansion()
	//if serviceFilesArtifactExpansions != nil {
	//	for mountPath, artifactIdentifiers := range serviceFilesArtifactExpansions.ServiceDirpathsToArtifactIdentifiers {
	//		fileMount := &FileMount{ //nolint:exhaustruct
	//			MountPath: mountPath,
	//		}
	//
	//		var serviceFilesArtifacts []*FilesArtifact
	//		for _, identifier := range artifactIdentifiers {
	//			var filesArtifact *FilesArtifact
	//			// if there's already a files artifact that exists with this name from a previous instruction, reference that
	//			if potentialFilesArtifact, ok := pyg.filesArtifactIndex[identifier]; ok {
	//				filesArtifact = &FilesArtifact{ //nolint:exhaustruct
	//					Name: potentialFilesArtifact.Name,
	//					Uuid: potentialFilesArtifact.Uuid,
	//				}
	//			} else {
	//				// otherwise create a new one
	//				// the only information we have about a files artifact that didn't already exist is the name
	//				// if it didn't already exist AND interpretation was successful, it MUST HAVE been passed in via args
	//				filesArtifact = &FilesArtifact{ //nolint:exhaustruct
	//					Name: identifier,
	//					Uuid: strconv.Itoa(pyg.generateUuid()),
	//				}
	//				pyg.planYaml.FilesArtifacts = append(pyg.planYaml.FilesArtifacts, filesArtifact)
	//				pyg.filesArtifactIndex[identifier] = filesArtifact
	//			}
	//			serviceFilesArtifacts = append(serviceFilesArtifacts, filesArtifact)
	//		}
	//
	//		fileMount.FilesArtifacts = serviceFilesArtifacts
	//		service.Files = append(service.Files, fileMount)
	//	}
	//
	//}
	//
	//pyg.planYaml.Services = append(pyg.planYaml.Services, service)
	//pyg.serviceIndex[service.Name] = service

	return stacktrace.NewError("IMPLEMENT ME")
}

func (builtin *AddServiceCapabilities) Description() string {
	return builtin.description
}

func validateAndConvertConfigAndReadyCondition(
	serviceNetwork service_network.ServiceNetwork,
	rawConfig starlark.Value,
	locatorOfModuleInWhichThisBuiltInIsBeingCalled string,
	packageId string,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
) (*service.ServiceConfig, *service_config.ReadyCondition, *startosis_errors.InterpretationError) {
	config, ok := rawConfig.(*service_config.ServiceConfig)
	if !ok {
		return nil, nil, startosis_errors.NewInterpretationError("The '%s' argument is not a ServiceConfig (was '%s').", ConfigsArgName, reflect.TypeOf(rawConfig))
	}
	apiServiceConfig, interpretationErr := config.ToKurtosisType(
		serviceNetwork,
		locatorOfModuleInWhichThisBuiltInIsBeingCalled,
		packageId,
		packageContentProvider,
		packageReplaceOptions)
	if interpretationErr != nil {
		return nil, nil, interpretationErr
	}

	readyCondition, interpretationErr := config.GetReadyCondition()
	if interpretationErr != nil {
		return nil, nil, interpretationErr
	}

	return apiServiceConfig, readyCondition, nil
}
