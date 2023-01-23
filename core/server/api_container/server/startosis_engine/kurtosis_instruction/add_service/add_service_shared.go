package add_service

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

func makeAddServiceInterpretationReturnValue(serviceName service.ServiceName, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig) (*kurtosis_types.Service, *startosis_errors.InterpretationError) {
	ports := serviceConfig.GetPrivatePorts()
	portSpecsDict := starlark.NewDict(len(ports))
	for portId, port := range ports {
		number := port.GetNumber()
		transportProtocol := port.GetTransportProtocol()
		maybeApplicationProtocol := port.GetMaybeApplicationProtocol()

		portSpec := kurtosis_types.NewPortSpec(number, transportProtocol, maybeApplicationProtocol)
		if err := portSpecsDict.SetKey(starlark.String(portId), portSpec); err != nil {
			return nil, startosis_errors.NewInterpretationError("An error occurred while creating a port spec for values "+
				"(number: '%v', transport_protocol: '%v', application_protocol: '%v') the add instruction return value",
				number, transportProtocol, maybeApplicationProtocol)
		}
	}
	ipAddress := starlark.String(fmt.Sprintf(magic_string_helper.IpAddressReplacementPlaceholderFormat, serviceName))
	returnValue := kurtosis_types.NewService(ipAddress, portSpecsDict)
	return returnValue, nil
}

func validateSingleService(validatorEnvironment *startosis_validator.ValidatorEnvironment, serviceName service.ServiceName, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig) *startosis_errors.ValidationError {
	if partition_topology.ParsePartitionId(serviceConfig.Subnetwork) != partition_topology.DefaultPartitionId {
		if !validatorEnvironment.IsNetworkPartitioningEnabled() {
			return startosis_errors.NewValidationError("Service was about to be started inside subnetwork '%s' but the Kurtosis enclave was started with subnetwork capabilities disabled. Make sure to run the Starlark code with subnetwork enabled.", *serviceConfig.Subnetwork)
		}
	}
	if validatorEnvironment.DoesServiceNameExist(serviceName) {
		return startosis_errors.NewValidationError("There was an error validating '%s' as service '%s' already exists", AddServiceBuiltinName, serviceName)
	}
	for _, artifactName := range serviceConfig.FilesArtifactMountpoints {
		if !validatorEnvironment.DoesArtifactNameExist(artifactName) {
			return startosis_errors.NewValidationError("There was an error validating '%s' as artifact name '%s' does not exist", AddServiceBuiltinName, artifactName)
		}
	}
	validatorEnvironment.AddServiceName(serviceName)
	validatorEnvironment.AppendRequiredContainerImage(serviceConfig.ContainerImageName)
	return nil
}

func replaceMagicStrings(
	serviceNetwork service_network.ServiceNetwork,
	runtimeValueStore *runtime_value_store.RuntimeValueStore,
	serviceName service.ServiceName,
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig,
) (
	service.ServiceName,
	*kurtosis_core_rpc_api_bindings.ServiceConfig,
	error,
) {
	// replacing magic string in service name
	serviceNameStr, err := magic_string_helper.ReplaceRuntimeValueInString(string(serviceName), runtimeValueStore)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime values in service name for '%s'", serviceName)
	}

	// replacing all magic strings in service config
	serviceConfigBuilder := services.NewServiceConfigBuilderFromServiceConfig(serviceConfig)

	newEntryPointArgs := make([]string, len(serviceConfig.EntrypointArgs))
	for index, entryPointArg := range serviceConfig.EntrypointArgs {
		entryPointArgWithIPAddressReplaced, err := magic_string_helper.ReplaceIPAddressInString(entryPointArg, serviceNetwork, serviceNameStr)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Error occurred while replacing IP address in entry point args for '%v'", entryPointArg)
		}
		entryPointArgWithIPAddressAndRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(entryPointArgWithIPAddressReplaced, runtimeValueStore)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in entry point args for '%v'", entryPointArg)
		}
		newEntryPointArgs[index] = entryPointArgWithIPAddressAndRuntimeValueReplaced
	}
	serviceConfigBuilder.WithEntryPointArgs(newEntryPointArgs)

	newCmdArgs := make([]string, len(serviceConfig.CmdArgs))
	for index, cmdArg := range serviceConfig.CmdArgs {
		cmdArgWithIPAddressReplaced, err := magic_string_helper.ReplaceIPAddressInString(cmdArg, serviceNetwork, serviceNameStr)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Error occurred while replacing IP address in command args for '%v'", cmdArg)
		}
		cmdArgWithIPAddressAndRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(cmdArgWithIPAddressReplaced, runtimeValueStore)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%v'", cmdArg)
		}
		newCmdArgs[index] = cmdArgWithIPAddressAndRuntimeValueReplaced
	}
	serviceConfigBuilder.WithCmdArgs(newCmdArgs)

	newEnvVars := make(map[string]string, len(serviceConfig.EnvVars))
	for envVarName, envVarValue := range serviceConfig.EnvVars {
		envVarValueWithIPAddressReplaced, err := magic_string_helper.ReplaceIPAddressInString(envVarValue, serviceNetwork, serviceNameStr)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Error occurred while replacing IP address in env vars for '%v'", envVarValue)
		}
		envVarValueWithIPAddressAndRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(envVarValueWithIPAddressReplaced, runtimeValueStore)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%s': '%s'", envVarName, envVarValue)
		}
		newEnvVars[envVarName] = envVarValueWithIPAddressAndRuntimeValueReplaced
	}
	serviceConfigBuilder.WithEnvVars(newEnvVars)

	return service.ServiceName(serviceNameStr), serviceConfigBuilder.Build(), nil
}
