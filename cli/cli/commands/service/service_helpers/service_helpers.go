package service_helpers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/enclaves"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/services"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/user_services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	ImageKey                  = "image"
	CmdKey                    = "cmd"
	EntrypointFlagKey         = "entrypoint"
	EntrypointAndCmdDelimiter = " "

	EnvvarsFlagKey              = "env"
	EnvvarKeyValueDelimiter     = "="
	EnvvarDeclarationsDelimiter = ","

	PortsFlagKey                     = "ports"
	PortIdSpecDelimiter              = "="
	PortNumberProtocolDelimiter      = "/"
	PortDeclarationsDelimiter        = ","
	PortApplicationProtocolDelimiter = ":"
	PortNumberUintParsingBase        = 10
	PortNumberUintParsingBits        = 16

	FilesFlagKey                     = "files"
	FilesArtifactMountsDelimiter     = ","
	FilesArtifactMountpointDelimiter = ":"
	FilesMultipleArtifactsDelimiter  = "|"

	EmptyApplicationProtocol = ""

	MaybeApplicationProtocolSpecForHelp = "MAYBE_APPLICATION_PROTOCOL"
	TransportProtocolSpecForHelp        = "TRANSPORT_PROTOCOL"
	PortNumberSpecForHelp               = "PORT_NUMBER"
	PortIdSpecForHelp                   = "PORT_ID"

	// Each envvar should be KEY1=VALUE1, which means we should have two components to each envvar declaration
	ExpectedNumberKeyValueComponentsInEnvvarDeclaration = 2
	PortNumberIndex                                     = 0
	TransportProtocolIndex                              = 1
	ExpectedPortIdSpecComponentsCount                   = 2
	ExpectedMountFragmentsCount                         = 2

	MinRemainingPortSpecComponents = 1
	MaxRemainingPortSpecComponents = 2

	defaultPortWaitTimeoutStr = "30s"
)

var (
	defaultTransportProtocolStr = strings.ToLower(kurtosis_core_rpc_api_bindings.Port_TCP.String())
	serviceAddSpec              = fmt.Sprintf(
		`%v%v%v%v%v`,
		MaybeApplicationProtocolSpecForHelp,
		PortApplicationProtocolDelimiter,
		PortNumberSpecForHelp,
		PortNumberProtocolDelimiter,
		TransportProtocolSpecForHelp,
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
	isTtyEnabled := service.GetTtyEnabled()
	serviceConfig := &services.ServiceConfig{
		Image:                       service.GetContainer().GetImageName(),
		PrivatePorts:                services.ConvertApiPortToJsonPort(service.GetPrivatePorts()),
		PublicPorts:                 services.ConvertApiPortToJsonPort(service.GetMaybePublicPorts()),
		Files:                       services.ConvertApiFilesArtifactsToJsonFiles(service.GetServiceDirPathsToFilesArtifactsList()),
		Entrypoint:                  service.GetContainer().GetEntrypointArgs(),
		Cmd:                         service.GetContainer().GetCmdArgs(),
		EnvVars:                     service.GetContainer().GetEnvVars(),
		PrivateIPAddressPlaceholder: "", // leave empty for now
		MaxMillicpus:                service.GetMaxMillicpus(),
		MinMillicpus:                service.GetMinMillicpus(),
		MaxMemory:                   service.GetMaxMemoryMegabytes(),
		MinMemory:                   service.GetMinMemoryMegabytes(),
		User:                        services.ConvertApiUserToJsonUser(service.GetUser()),
		Tolerations:                 services.ConvertApiTolerationsToJsonTolerations(service.GetTolerations()),
		NodeSelectors:               service.GetNodeSelectors(),
		Labels:                      service.GetLabels(),
		TiniEnabled:                 &isTiniEnabled,
		TtyEnabled:                  &isTtyEnabled,
	}

	return service, serviceConfig, nil
}

func GetAddServiceStarlarkScript(serviceName string, serviceConfigStarlark string) string {
	return fmt.Sprintf(`def run(plan):
	plan.add_service(name = "%s", config = %s)
`, serviceName, serviceConfigStarlark)
}

func RunAddServiceStarlarkScript(ctx context.Context, serviceName, enclaveIdentifier, starlarkScript string, enclaveCtx *enclaves.EnclaveContext) (*enclaves.StarlarkRunResult, error) {
	logrus.Debugf("Add service starlark:\n%v", starlarkScript)
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

	allEnvvarDeclarationStrs := strings.Split(envvarsStr, EnvvarDeclarationsDelimiter)
	for _, envvarDeclarationStr := range allEnvvarDeclarationStrs {
		if len(strings.TrimSpace(envvarDeclarationStr)) == 0 {
			continue
		}

		envvarKeyValueComponents := strings.SplitN(envvarDeclarationStr, EnvvarKeyValueDelimiter, ExpectedNumberKeyValueComponentsInEnvvarDeclaration)
		if len(envvarKeyValueComponents) < ExpectedNumberKeyValueComponentsInEnvvarDeclaration {
			return nil, stacktrace.NewError("Environment declaration string '%v' must be of the form KEY1%vVALUE1", envvarDeclarationStr, EnvvarKeyValueDelimiter)
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

	allPortDeclarationStrs := strings.Split(portsStr, PortDeclarationsDelimiter)
	for _, portDeclarationStr := range allPortDeclarationStrs {
		if len(strings.TrimSpace(portDeclarationStr)) == 0 {
			continue
		}

		portIdSpecComponents := strings.Split(portDeclarationStr, PortIdSpecDelimiter)
		if len(portIdSpecComponents) != ExpectedPortIdSpecComponentsCount {
			return nil, stacktrace.NewError("Port declaration string '%v' must be of the form PORTID%vSPEC", portDeclarationStr, PortIdSpecDelimiter)
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

	remainingPortSpecComponents := strings.Split(remainingPortSpec, PortNumberProtocolDelimiter)
	numRemainingPortSpecComponents := len(remainingPortSpecComponents)
	if numRemainingPortSpecComponents > MaxRemainingPortSpecComponents {
		return nil, stacktrace.NewError(
			`Invalid port spec string, expected format is %q but got '%v'`,
			serviceAddSpec,
			specStr,
		)
	}

	portNumberUint16, err := getPortNumberFromPortSpecString(remainingPortSpecComponents[PortNumberIndex])
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while parsing port number '%v' in port spec '%v'", remainingPortSpecComponents[PortNumberIndex], specStr)
	}

	transportProtocol := defaultTransportProtocolStr
	if numRemainingPortSpecComponents > MinRemainingPortSpecComponents {
		transportProtocol = remainingPortSpecComponents[TransportProtocolIndex]
	}

	transportProtocolFromEnum, err := getTransportProtocolFromPortSpecString(transportProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while parsing transport protocol '%v' in port spec '%v'", remainingPortSpecComponents[TransportProtocolIndex], specStr)
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

	beforeDelimiter, afterDelimiter, foundDelimiter := strings.Cut(portProtocolStr, PortApplicationProtocolDelimiter)

	if !foundDelimiter {
		return EmptyApplicationProtocol, beforeDelimiter, nil
	}

	if foundDelimiter && beforeDelimiter == EmptyApplicationProtocol {
		return EmptyApplicationProtocol, "", stacktrace.NewError("optional application protocol argument cannot be empty")
	}

	return beforeDelimiter, afterDelimiter, nil
}

func getPortNumberFromPortSpecString(portNumberStr string) (uint32, error) {
	portNumberUint64, err := strconv.ParseUint(portNumberStr, PortNumberUintParsingBase, PortNumberUintParsingBits)
	if err != nil {
		return 0, stacktrace.Propagate(
			err,
			"An error occurred parsing port number string '%v' with base '%v' and bits '%v'",
			portNumberStr,
			PortNumberUintParsingBase,
			PortNumberUintParsingBits,
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
	allMountStrs := strings.Split(filesArtifactMountsStr, FilesArtifactMountsDelimiter)
	for idx, mountStr := range allMountStrs {
		trimmedMountStr := strings.TrimSpace(mountStr)
		if len(trimmedMountStr) == 0 {
			continue
		}

		mountFragments := strings.Split(trimmedMountStr, FilesArtifactMountpointDelimiter)
		if len(mountFragments) != ExpectedMountFragmentsCount {
			return nil, stacktrace.NewError(
				"Files artifact mountpoint string %v was '%v' but should be in the form 'mountpoint%sfiles_artifact_name'",
				idx,
				trimmedMountStr,
				FilesArtifactMountpointDelimiter,
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

		result[mountpoint] = strings.Split(filesArtifactNamesStr, FilesMultipleArtifactsDelimiter)
	}

	return result, nil
}
