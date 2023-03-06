package add_service

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	ipAddressRuntimeValue = "ip_address"
	hostnameRuntimeValue  = "hostname"
)

func fillAddServiceReturnValueWithRuntimeValues(service *service.Service, resultUuid string, runtimeValueStore *runtime_value_store.RuntimeValueStore) {
	runtimeValueStore.SetValue(resultUuid, map[string]starlark.Comparable{
		ipAddressRuntimeValue: starlark.String(service.GetRegistration().GetPrivateIP().String()),
		hostnameRuntimeValue:  starlark.String(service.GetRegistration().GetHostname()),
	})
}

func makeAddServiceInterpretationReturnValue(serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig, resultUuid string) (*kurtosis_types.Service, *startosis_errors.InterpretationError) {
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
	ipAddress := starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, ipAddressRuntimeValue))
	hostname := starlark.String(fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, resultUuid, hostnameRuntimeValue))
	returnValue := kurtosis_types.NewService(hostname, ipAddress, portSpecsDict)
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

	if serviceConfig.EntrypointArgs != nil {
		newEntryPointArgs := make([]string, len(serviceConfig.EntrypointArgs))
		for index, entryPointArg := range serviceConfig.EntrypointArgs {
			entryPointArgWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(entryPointArg, runtimeValueStore)
			if err != nil {
				return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in entry point args for '%v'", entryPointArg)
			}
			newEntryPointArgs[index] = entryPointArgWithRuntimeValueReplaced
		}
		serviceConfigBuilder.WithEntryPointArgs(newEntryPointArgs)
	}

	if serviceConfig.CmdArgs != nil {
		newCmdArgs := make([]string, len(serviceConfig.CmdArgs))
		for index, cmdArg := range serviceConfig.CmdArgs {
			cmdArgWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(cmdArg, runtimeValueStore)
			if err != nil {
				return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%v'", cmdArg)
			}
			newCmdArgs[index] = cmdArgWithRuntimeValueReplaced
		}
		serviceConfigBuilder.WithCmdArgs(newCmdArgs)
	}

	if serviceConfig.EnvVars != nil {
		newEnvVars := make(map[string]string, len(serviceConfig.EnvVars))
		for envVarName, envVarValue := range serviceConfig.EnvVars {
			if err != nil {
				return "", nil, stacktrace.Propagate(err, "Error occurred while replacing IP address in env vars for '%v'", envVarValue)
			}
			envVarValueWithRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(envVarValue, runtimeValueStore)
			if err != nil {
				return "", nil, stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%s': '%s'", envVarName, envVarValue)
			}
			newEnvVars[envVarName] = envVarValueWithRuntimeValueReplaced
		}
		serviceConfigBuilder.WithEnvVars(newEnvVars)
	}

	return service.ServiceName(serviceNameStr), serviceConfigBuilder.Build(), nil
}
