package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"time"
)

const (
	ipAddressRuntimeValue = "ip_address"
	hostnameRuntimeValue  = "hostname"
)

func fillAddServiceReturnValueWithRuntimeValues(service *service.Service, resultUuid string, runtimeValueStore *runtime_value_store.RuntimeValueStore) error {
	if err := runtimeValueStore.SetValue(resultUuid, map[string]starlark.Comparable{
		ipAddressRuntimeValue: starlark.String(service.GetRegistration().GetPrivateIP().String()),
		hostnameRuntimeValue:  starlark.String(service.GetRegistration().GetHostname()),
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting value with key '%s' in the runtime value store", resultUuid)
	}
	return nil
}

func makeAddServiceInterpretationReturnValue(serviceName starlark.String, serviceConfig *service.ServiceConfig, resultUuid string) (*kurtosis_types.Service, *startosis_errors.InterpretationError) {
	ports := serviceConfig.GetPrivatePorts()
	portSpecsDict := starlark.NewDict(len(ports))
	for portId, port := range ports {
		number := port.GetNumber()
		transportProtocol := port.GetTransportProtocol()
		maybeApplicationProtocol := port.GetMaybeApplicationProtocol()
		var maybeWaitTimeout string
		if port.GetWait() != nil {
			maybeWaitTimeout = port.GetWait().GetTimeout().String()
		}

		portSpec, interpretationErr := port_spec.CreatePortSpecUsingGoValues(number, transportProtocol, maybeApplicationProtocol, maybeWaitTimeout)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		if err := portSpecsDict.SetKey(starlark.String(portId), portSpec); err != nil {
			return nil, startosis_errors.NewInterpretationError("An error occurred while creating a port spec for values "+
				"(number: '%v', transport_protocol: '%v', application_protocol: '%v') the add instruction return value",
				number, transportProtocol, maybeApplicationProtocol)
		}
	}
	ipAddress := starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, ipAddressRuntimeValue))
	hostname := starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, hostnameRuntimeValue))
	returnValue, interpretationErr := kurtosis_types.CreateService(serviceName, hostname, ipAddress, portSpecsDict)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return returnValue, nil
}

func validateSingleService(validatorEnvironment *startosis_validator.ValidatorEnvironment, serviceName service.ServiceName, serviceConfig *service.ServiceConfig) *startosis_errors.ValidationError {
	if isValidServiceName := service.IsServiceNameValid(serviceName); !isValidServiceName {
		return startosis_errors.NewValidationError(invalidServiceNameErrorText(serviceName))
	}

	if persistentDirectories := serviceConfig.GetPersistentDirectories(); persistentDirectories != nil {
		for _, directory := range persistentDirectories.ServiceDirpathToPersistentDirectory {
			if !service_directory.IsPersistentKeyValid(directory.PersistentKey) {
				return startosis_errors.NewValidationError(invalidPersistentKeyErrorText(directory.PersistentKey))
			}
			validatorEnvironment.AddPersistentKey(directory.PersistentKey)
		}
	}

	if validatorEnvironment.DoesServiceNameExist(serviceName) == startosis_validator.ComponentCreatedOrUpdatedDuringPackageRun {
		return startosis_errors.NewValidationError("There was an error validating '%s' as service with the name '%s' already exists inside the package. Adding two different services with the same name isn't allowed; we recommend prefixing/suffixing the two service names or using two different names entirely.", AddServiceBuiltinName, serviceName)
	}
	if serviceConfig.GetFilesArtifactsExpansion() != nil {
		for _, artifactNames := range serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers {
			for _, artifactName := range artifactNames {
				if validatorEnvironment.DoesArtifactNameExist(artifactName) == startosis_validator.ComponentNotFound {
					return startosis_errors.NewValidationError("There was an error validating '%s' as artifact name '%s' does not exist", AddServiceBuiltinName, artifactName)
				}
			}
		}
	}

	if validationErr := validatorEnvironment.HasEnoughCPU(serviceConfig.GetMinCPUAllocationMillicpus(), serviceName); validationErr != nil {
		return validationErr
	}

	if validationErr := validatorEnvironment.HasEnoughMemory(serviceConfig.GetMinMemoryAllocationMegabytes(), serviceName); validationErr != nil {
		return validationErr
	}

	validatorEnvironment.AddServiceName(serviceName)

	if serviceConfig.GetImageBuildSpec() != nil {
		validatorEnvironment.AppendRequiredImageBuild(serviceConfig.GetContainerImageName(), serviceConfig.GetImageBuildSpec())
	} else if serviceConfig.GetImageRegistrySpec() != nil {
		validatorEnvironment.AppendImageToPullWithAuth(serviceConfig.GetContainerImageName(), serviceConfig.GetImageRegistrySpec())
	} else {
		validatorEnvironment.AppendRequiredImagePull(serviceConfig.GetContainerImageName())
	}

	var portIds []string
	for portId := range serviceConfig.GetPrivatePorts() {
		portIds = append(portIds, portId)
	}
	validatorEnvironment.AddPrivatePortIDForService(portIds, serviceName)
	validatorEnvironment.ConsumeMemory(serviceConfig.GetMinMemoryAllocationMegabytes(), serviceName)
	validatorEnvironment.ConsumeCPU(serviceConfig.GetMinCPUAllocationMillicpus(), serviceName)
	return nil
}

func invalidServiceNameErrorText(
	serviceName service.ServiceName,
) string {
	return fmt.Sprintf(
		"Service name '%v' is invalid as it contains disallowed characters. Service names must adhere to the RFC 1035 standard, specifically implementing this regex and be 1-63 characters long: %s. This means the service name must only contain lowercase alphanumeric characters or '-', and must start with a lowercase alphabet and end with a lowercase alphanumeric character.",
		serviceName,
		service.WordWrappedServiceNameRegex,
	)
}

func invalidPersistentKeyErrorText(
	persistentKey service_directory.DirectoryPersistentKey,
) string {
	return fmt.Sprintf(
		"Persistent Key '%v' is invalid as it contains disallowed characters. Persistent Key must adhere to the RFC 1035 standard, specifically implementing this regex and be 1-63 characters long: %s. This means the service name must only contain lowercase alphanumeric characters or '-', and must start with a lowercase alphabet and end with a lowercase alphanumeric character.",
		persistentKey,
		service_directory.WordWrappedPersistentKeyRegex,
	)
}

func replaceMagicStrings(
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	serviceName service.ServiceName,
	serviceConfig *service.ServiceConfig,
) (
	service.ServiceName,
	*service.ServiceConfig,
	error,
) {
	// replacing magic string in service name
	serviceNameStr, err := magic_string_helper.ReplaceRuntimeValueInString(string(serviceName), runtimeValueStore)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime values in service name for '%s'", serviceName)
	}

	var entrypoints []string
	if serviceConfig.GetEntrypointArgs() != nil {
		entrypoints = make([]string, len(serviceConfig.GetEntrypointArgs()))
		for index, entryPointArg := range serviceConfig.GetEntrypointArgs() {
			entryPointArgWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(entryPointArg, runtimeValueStore)
			if err != nil {
				return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in entry point args for '%v'", entryPointArg)
			}
			entrypoints[index] = entryPointArgWithRuntimeValueReplaced
		}
	}

	var cmdArgs []string
	if serviceConfig.GetCmdArgs() != nil {
		cmdArgs = make([]string, len(serviceConfig.GetCmdArgs()))
		for index, cmdArg := range serviceConfig.GetCmdArgs() {
			cmdArgWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(cmdArg, runtimeValueStore)
			if err != nil {
				return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%v'", cmdArg)
			}
			cmdArgs[index] = cmdArgWithRuntimeValueReplaced
		}
	}

	var envVars map[string]string
	if serviceConfig.GetEnvVars() != nil {
		envVars = make(map[string]string, len(serviceConfig.GetEnvVars()))
		for envVarName, envVarValue := range serviceConfig.GetEnvVars() {
			if err != nil {
				return "", nil, stacktrace.Propagate(err, "Error occurred while replacing IP address in env vars for '%v'", envVarValue)
			}
			envVarValueWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(envVarValue, runtimeValueStore)
			if err != nil {
				return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%s': '%s'", envVarName, envVarValue)
			}
			envVars[envVarName] = envVarValueWithRuntimeValueReplaced
		}
	}

	renderedServiceConfig, err := service.CreateServiceConfig(
		serviceConfig.GetContainerImageName(),
		serviceConfig.GetImageBuildSpec(),
		serviceConfig.GetImageRegistrySpec(),
		serviceConfig.GetPrivatePorts(),
		serviceConfig.GetPublicPorts(),
		entrypoints,
		cmdArgs,
		envVars,
		serviceConfig.GetFilesArtifactsExpansion(),
		serviceConfig.GetPersistentDirectories(),
		serviceConfig.GetCPUAllocationMillicpus(),
		serviceConfig.GetMemoryAllocationMegabytes(),
		serviceConfig.GetPrivateIPAddrPlaceholder(),
		serviceConfig.GetMinCPUAllocationMillicpus(),
		serviceConfig.GetMinMemoryAllocationMegabytes(),
		serviceConfig.GetLabels(),
		serviceConfig.GetUser(),
		serviceConfig.GetTolerations(),
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating a service config")
	}

	return service.ServiceName(serviceNameStr), renderedServiceConfig, nil
}

func runServiceReadinessCheck(
	ctx context.Context,
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	serviceName service.ServiceName,
	readyConditions *service_config.ReadyCondition,
) error {
	if readyConditions != nil {

		recipe, intepretationErr := readyConditions.GetRecipe()
		if intepretationErr != nil {
			return stacktrace.Propagate(intepretationErr, "An error occurred getting the recipe value from ready conditions '%v'", readyConditions)
		}

		field, intepretationErr := readyConditions.GetField()
		if intepretationErr != nil {
			return stacktrace.Propagate(intepretationErr, "An error occurred getting the field value from ready conditions '%v'", readyConditions)
		}

		assertion, intepretationErr := readyConditions.GetAssertion()
		if intepretationErr != nil {
			return stacktrace.Propagate(intepretationErr, "An error occurred getting the assertion value from ready conditions '%v'", readyConditions)
		}

		target, intepretationErr := readyConditions.GetTarget()
		if intepretationErr != nil {
			return stacktrace.Propagate(intepretationErr, "An error occurred getting the target value from ready conditions '%v'", readyConditions)
		}

		interval, intepretationErr := readyConditions.GetInterval()
		if intepretationErr != nil {
			return stacktrace.Propagate(intepretationErr, "An error occurred getting the interval value from ready conditions '%v'", readyConditions)
		}

		timeout, intepretationErr := readyConditions.GetTimeout()
		if intepretationErr != nil {
			return stacktrace.Propagate(intepretationErr, "An error occurred getting the timeout value from ready conditions '%v'", readyConditions)
		}

		startTime := time.Now()
		logrus.Infof("Checking service readiness for '%s' at '%v'", serviceName, startTime) //TODO change to debug
		lastResult, tries, err := shared_helpers.ExecuteServiceAssertionWithRecipe(
			ctx,
			serviceNetwork,
			runtimeValueStore,
			serviceName,
			recipe,
			field,
			assertion,
			target,
			interval,
			timeout,
		)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred checking if service '%v' is ready, using "+
					"recipe '%+v', value field '%v', assertion '%v', target '%v', interval '%s' and time-out '%s'.",
				serviceName,
				recipe,
				field,
				assertion,
				target,
				interval,
				timeout,
			)
		}
		//TODO change to debug
		logrus.Infof("Checking if service '%v' is ready took %d tries (%v in total). "+
			"Assertion passed with following:\n%s",
			serviceName,
			tries,
			time.Since(startTime),
			recipe.ResultMapToString(lastResult),
		)
	}
	return nil
}
