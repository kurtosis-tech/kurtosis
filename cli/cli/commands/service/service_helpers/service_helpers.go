package service_helpers

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_services"
	"github.com/kurtosis-tech/stacktrace"
	"strconv"
	"strings"
)

const (
	ImageKey          = "image"
	CmdKey            = "cmd"
	EntrypointFlagKey = "entrypoint"

	EnvvarsFlagKey              = "env"
	envvarKeyValueDelimiter     = "="
	envvarDeclarationsDelimiter = ","

	PortsFlagKey                     = "ports"
	portIdSpecDelimiter              = "="
	portNumberProtocolDelimiter      = "/"
	portDeclarationsDelimiter        = ","
	portApplicationProtocolDelimiter = ":"
	portNumberUintParsingBase        = 10
	portNumberUintParsingBits        = 16

	FilesFlagKey                     = "files"
	filesArtifactMountsDelimiter     = ","
	filesArtifactMountpointDelimiter = ":"
	multipleFilesArtifactsDelimiter  = "|"

	emptyApplicationProtocol = ""

	maybeApplicationProtocolSpecForHelp = "MAYBE_APPLICATION_PROTOCOL"
	transportProtocolSpecForHelp        = "TRANSPORT_PROTOCOL"
	portNumberSpecForHelp               = "PORT_NUMBER"
	portIdSpecForHelp                   = "PORT_ID"

	// Each envvar should be KEY1=VALUE1, which means we should have two components to each envvar declaration
	expectedNumberKeyValueComponentsInEnvvarDeclaration = 2
	portNumberIndex                                     = 0
	transportProtocolIndex                              = 1
	expectedPortIdSpecComponentsCount                   = 2
	expectedMountFragmentsCount                         = 2

	minRemainingPortSpecComponents = 1
	maxRemainingPortSpecComponents = 2

	defaultPortWaitTimeoutStr = "30s"
)

var (
	defaultTransportProtocolStr = strings.ToLower(kurtosis_core_rpc_api_bindings.Port_TCP.String())
	serviceAddSpec              = fmt.Sprintf(
		`%v%v%v%v%v`,
		maybeApplicationProtocolSpecForHelp,
		portApplicationProtocolDelimiter,
		portNumberSpecForHelp,
		portNumberProtocolDelimiter,
		transportProtocolSpecForHelp,
	)
	serviceAddSpecWithPortId = fmt.Sprintf(
		`%v%v%v`,
		portIdSpecForHelp,
		portIdSpecDelimiter,
		serviceAddSpec,
	)
)

// Helpers
func GetServiceInfo(ctx context.Context, kurtosisCtx *kurtosis_context.KurtosisContext, enclaveIdentifier, serviceIdentifier string) (*kurtosis_core_rpc_api_bindings.ServiceInfo, *services.ServiceConfig, error) {
	enclaveInfo, err := kurtosisCtx.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the enclave for identifier '%v'", enclaveIdentifier)
	}

	enclaveApiContainerStatus := enclaveInfo.ApiContainerStatus
	isApiContainerRunning := enclaveApiContainerStatus == kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING

	userServices := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	if isApiContainerRunning {
		var err error
		serviceMap := map[string]bool{
			serviceIdentifier: true,
		}
		userServices, err = user_services.GetUserServiceInfoMapFromAPIContainer(ctx, enclaveInfo, serviceMap)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Failed to get service info from API container in enclave '%v'", enclaveInfo.GetEnclaveUuid())
		}
	}

	var service *kurtosis_core_rpc_api_bindings.ServiceInfo
	for _, serviceInfo := range userServices {
		service = serviceInfo
		break
	}

	isTiniEnabled := service.GetTiniEnabled()
	serviceConfig := &services.ServiceConfig{
		Image:                       service.GetContainer().GetImageName(),
		PrivatePorts:                services.ConvertApiPortToJsonPort(service.GetPrivatePorts()),
		PublicPorts:                 services.ConvertApiPortToJsonPort(service.GetMaybePublicPorts()),
		Files:                       services.ConvertApiFilesArtifactsToJsonFiles(service.GetServiceDirPathsToFilesArtifactsList()),
		Entrypoint:                  service.GetContainer().GetEntrypointArgs(),
		Cmd:                         service.GetContainer().GetCmdArgs(),
		EnvVars:                     service.GetContainer().GetEnvVars(),
		PrivateIPAddressPlaceholder: service.GetPrivateIpAddr(),
		MaxMillicpus:                service.GetMaxMillicpus(),
		MinMillicpus:                service.GetMinMillicpus(),
		MaxMemory:                   service.GetMaxMemoryMegabytes(),
		MinMemory:                   service.GetMinMemoryMegabytes(),
		User:                        services.ConvertApiUserToJsonUser(service.GetUser()),
		Tolerations:                 services.ConvertApiTolerationsToJsonTolerations(service.GetTolerations()),
		NodeSelectors:               service.GetNodeSelectors(),
		TiniEnabled:                 &isTiniEnabled,
	}

	return service, serviceConfig, nil
}

func GetAddServiceStarlarkScript(serviceName string, serviceConfigStarlark string) string {
	return fmt.Sprintf(`def run(plan):
	plan.add_service(name = "%s", config = %s)
`, serviceName, serviceConfigStarlark)
}

func RunAddServiceStarlarkScript(ctx context.Context, serviceName, enclaveIdentifier, starlarkScript string, enclaveCtx *enclaves.EnclaveContext) (*enclaves.StarlarkRunResult, error) {
	starlarkRunResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, starlark_run_config.NewRunStarlarkConfig())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error has occurred when running Starlark to add service")
	}
	if starlarkRunResult.InterpretationError != nil {
		return nil, stacktrace.NewError("An error has occurred when adding service: %s\nThis is a bug in Kurtosis, please report.", starlarkRunResult.InterpretationError)
	}
	if len(starlarkRunResult.ValidationErrors) > 0 {
		return nil, stacktrace.NewError("An error occurred when validating add service '%v' to enclave '%v': %s", serviceName, enclaveIdentifier, starlarkRunResult.ValidationErrors)
	}
	if starlarkRunResult.ExecutionError != nil {
		return nil, stacktrace.NewError("An error occurred adding service '%v' to enclave '%v': %s", serviceName, enclaveIdentifier, starlarkRunResult.ExecutionError)
	}
	return starlarkRunResult, nil
}

// Parses a string in the form KEY1=VALUE1,KEY2=VALUE2 into a map of strings
// An empty string will result in an empty map
// Empty strings will be skipped (e.g. ',,,' will result in an empty map)
func ParseEnvVarsStr(envvarsStr string) (map[string]string, error) {
	result := map[string]string{}
	if envvarsStr == "" {
		return result, nil
	}

	allEnvvarDeclarationStrs := strings.Split(envvarsStr, envvarDeclarationsDelimiter)
	for _, envvarDeclarationStr := range allEnvvarDeclarationStrs {
		if len(strings.TrimSpace(envvarDeclarationStr)) == 0 {
			continue
		}

		envvarKeyValueComponents := strings.SplitN(envvarDeclarationStr, envvarKeyValueDelimiter, expectedNumberKeyValueComponentsInEnvvarDeclaration)
		if len(envvarKeyValueComponents) < expectedNumberKeyValueComponentsInEnvvarDeclaration {
			return nil, stacktrace.NewError("Environment declaration string '%v' must be of the form KEY1%vVALUE1", envvarDeclarationStr, envvarKeyValueDelimiter)
		}
		key := envvarKeyValueComponents[0]
		value := envvarKeyValueComponents[1]

		preexistingValue, found := result[key]
		if found {
			return nil, stacktrace.NewError(
				"Cannot declare environment variable '%v' assigned to value '%v' because the key has previously been assigned to value '%v'",
				key,
				value,
				preexistingValue,
			)
		}

		result[key] = value
	}

	return result, nil
}

// Parses a string in the form PORTID1=1234,PORTID2=5678/udp
// An empty string will result in an empty map
// Empty strings will be skipped (e.g. ',,,' will result in an empty map)
func ParsePortsStr(portsStr string) (map[string]*kurtosis_core_rpc_api_bindings.Port, error) {
	result := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	if strings.TrimSpace(portsStr) == "" {
		return result, nil
	}

	allPortDeclarationStrs := strings.Split(portsStr, portDeclarationsDelimiter)
	for _, portDeclarationStr := range allPortDeclarationStrs {
		if len(strings.TrimSpace(portDeclarationStr)) == 0 {
			continue
		}

		portIdSpecComponents := strings.Split(portDeclarationStr, portIdSpecDelimiter)
		if len(portIdSpecComponents) != expectedPortIdSpecComponentsCount {
			return nil, stacktrace.NewError("Port declaration string '%v' must be of the form PORTID%vSPEC", portDeclarationStr, portIdSpecDelimiter)
		}
		portId := portIdSpecComponents[0]
		specStr := portIdSpecComponents[1]
		if len(strings.TrimSpace(portId)) == 0 {
			return nil, stacktrace.NewError("Port declaration with spec string '%v' has an empty port ID", specStr)
		}
		portSpec, err := parsePortSpecStr(specStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred parsing port spec string '%v' for port with ID '%v'", specStr, portId)
		}

		if _, found := result[portId]; found {
			return nil, stacktrace.NewError(
				"Cannot define port '%v' with spec '%v' because it is already defined",
				portId,
				specStr,
			)
		}

		result[portId] = portSpec
	}

	return result, nil
}

func parsePortSpecStr(specStr string) (*kurtosis_core_rpc_api_bindings.Port, error) {
	if len(strings.TrimSpace(specStr)) == 0 {
		return nil, stacktrace.NewError("Cannot parse empty spec string")
	}

	maybeApplicationProtocol, remainingPortSpec, err := getMaybeApplicationProtocolFromPortSpecString(specStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while parsing application protocol '%v' in port spec '%v'", maybeApplicationProtocol, specStr)
	}

	remainingPortSpecComponents := strings.Split(remainingPortSpec, portNumberProtocolDelimiter)
	numRemainingPortSpecComponents := len(remainingPortSpecComponents)
	if numRemainingPortSpecComponents > maxRemainingPortSpecComponents {
		return nil, stacktrace.NewError(
			`Invalid port spec string, expected format is %q but got '%v'`,
			serviceAddSpec,
			specStr,
		)
	}

	portNumberUint16, err := getPortNumberFromPortSpecString(remainingPortSpecComponents[portNumberIndex])
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while parsing port number '%v' in port spec '%v'", remainingPortSpecComponents[portNumberIndex], specStr)
	}

	transportProtocol := defaultTransportProtocolStr
	if numRemainingPortSpecComponents > minRemainingPortSpecComponents {
		transportProtocol = remainingPortSpecComponents[transportProtocolIndex]
	}

	transportProtocolFromEnum, err := getTransportProtocolFromPortSpecString(transportProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while parsing transport protocol '%v' in port spec '%v'", remainingPortSpecComponents[transportProtocolIndex], specStr)
	}
	return &kurtosis_core_rpc_api_bindings.Port{
		Number:                   portNumberUint16,
		TransportProtocol:        transportProtocolFromEnum,
		MaybeApplicationProtocol: maybeApplicationProtocol,
		MaybeWaitTimeout:         defaultPortWaitTimeoutStr, //TODO we should add this to the port's arguments instead of using only a default value
		Locked:                   nil,
		Alias:                    nil,
	}, nil
}

/*
*
This method takes in port protocol string and parses it to get application protocol.
It looks for `:` delimiter and splits the string into array of at most size 2. If the length
of array is 2 then application protocol exists, otherwise it does not. This is basically what
strings.Cut() does. // TODO: use that instead once we update go version
*/
func getMaybeApplicationProtocolFromPortSpecString(portProtocolStr string) (string, string, error) {

	beforeDelimiter, afterDelimiter, foundDelimiter := strings.Cut(portProtocolStr, portApplicationProtocolDelimiter)

	if !foundDelimiter {
		return emptyApplicationProtocol, beforeDelimiter, nil
	}

	if foundDelimiter && beforeDelimiter == emptyApplicationProtocol {
		return emptyApplicationProtocol, "", stacktrace.NewError("optional application protocol argument cannot be empty")
	}

	return beforeDelimiter, afterDelimiter, nil
}

func getPortNumberFromPortSpecString(portNumberStr string) (uint32, error) {
	portNumberUint64, err := strconv.ParseUint(portNumberStr, portNumberUintParsingBase, portNumberUintParsingBits)
	if err != nil {
		return 0, stacktrace.Propagate(
			err,
			"An error occurred parsing port number string '%v' with base '%v' and bits '%v'",
			portNumberStr,
			portNumberUintParsingBase,
			portNumberUintParsingBits,
		)
	}
	portNumberUint32 := uint32(portNumberUint64)
	return portNumberUint32, nil
}

func getTransportProtocolFromPortSpecString(portSpec string) (kurtosis_core_rpc_api_bindings.Port_TransportProtocol, error) {
	transportProtocolEnumInt, found := kurtosis_core_rpc_api_bindings.Port_TransportProtocol_value[strings.ToUpper(portSpec)]
	if !found {
		return 0, stacktrace.NewError("Unrecognized port protocol '%v'", portSpec)
	}
	return kurtosis_core_rpc_api_bindings.Port_TransportProtocol(transportProtocolEnumInt), nil
}

func ParseFilesArtifactMountsStr(filesArtifactMountsStr string) (map[string][]string, error) {
	result := map[string][]string{}
	if strings.TrimSpace(filesArtifactMountsStr) == "" {
		return result, nil
	}

	// NOTE: we might actually want to allow the same artifact being mounted in multiple places
	allMountStrs := strings.Split(filesArtifactMountsStr, filesArtifactMountsDelimiter)
	for idx, mountStr := range allMountStrs {
		trimmedMountStr := strings.TrimSpace(mountStr)
		if len(trimmedMountStr) == 0 {
			continue
		}

		mountFragments := strings.Split(trimmedMountStr, filesArtifactMountpointDelimiter)
		if len(mountFragments) != expectedMountFragmentsCount {
			return nil, stacktrace.NewError(
				"Files artifact mountpoint string %v was '%v' but should be in the form 'mountpoint%sfiles_artifact_name'",
				idx,
				trimmedMountStr,
				filesArtifactMountpointDelimiter,
			)
		}
		mountpoint := mountFragments[0]
		filesArtifactNamesStr := mountFragments[1]
		if existingNames, found := result[mountpoint]; found {
			return nil, stacktrace.NewError(
				"Mountpoint '%v' is declared twice; once to artifact name '%v' and again to artifact name '%v'",
				mountpoint,
				existingNames,
				filesArtifactNamesStr,
			)
		}

		result[mountpoint] = strings.Split(filesArtifactNamesStr, multipleFilesArtifactsDelimiter)
	}

	return result, nil
}
