package add

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"strconv"
	"strings"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceNameArgKey   = "service-name"
	serviceNameTitleKey = "Name"
	serviceUuidTitleKey = "UUID"

	serviceImageArgKey = "image"

	cmdArgsArgKey = "cmd-arg"

	entrypointBinaryFlagKey = "entrypoint"

	envvarsFlagKey              = "env"
	envvarKeyValueDelimiter     = "="
	envvarDeclarationsDelimiter = ","

	portsFlagKey                     = "ports"
	portIdSpecDelimiter              = "="
	portNumberProtocolDelimiter      = "/"
	portDeclarationsDelimiter        = ","
	portApplicationProtocolDelimiter = ":"
	portNumberUintParsingBase        = 10
	portNumberUintParsingBits        = 16

	filesFlagKey                     = "files"
	filesArtifactMountsDelimiter     = ","
	filesArtifactMountpointDelimiter = ":"
	defaultLimits                    = 0

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	privateIPAddressPlaceholderKey     = "ip-address-placeholder"
	privateIPAddressPlaceholderDefault = "KURTOSIS_IP_ADDR_PLACEHOLDER"

	// Each envvar should be KEY1=VALUE1, which means we should have two components to each envvar declaration
	expectedNumberKeyValueComponentsInEnvvarDeclaration = 2
	portNumberIndex                                     = 0
	transportProtocolIndex                              = 1
	expectedPortIdSpecComponentsCount                   = 2
	expectedMountFragmentsCount                         = 2

	minRemainingPortSpecComponents = 1
	maxRemainingPortSpecComponents = 2

	emptyApplicationProtocol = ""
	linkDelimiter            = "://"

	maybeApplicationProtocolSpecForHelp = "MAYBE_APPLICATION_PROTOCOL"
	transportProtocolSpecForHelp        = "TRANSPORT_PROTOCOL"
	portNumberSpecForHelp               = "PORT_NUMBER"
	portIdSpecForHelp                   = "PORT_ID"

	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	portMappingSeparatorForLogs = ", "

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

var ServiceAddCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceAddCmdStr,
	ShortDescription:          "Adds a service to an enclave",
	LongDescription:           "Adds a new service with the given parameters to the given enclave",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		{
			Key: serviceNameArgKey,
		},
		{
			Key: serviceImageArgKey,
		},
		{
			Key:          cmdArgsArgKey,
			IsOptional:   true,
			IsGreedy:     true,
			DefaultValue: []string{},
		},
	},
	Flags: []*flags.FlagConfig{
		{
			Key:   entrypointBinaryFlagKey,
			Usage: "ENTRYPOINT binary that will be used when running the container, overriding the image's default ENTRYPOINT",
			// TODO Make this a string list
			Type: flags.FlagType_String,
		},
		{
			// TODO We currently can't handle commas, so allow users to set the flag multiple times to set multiple envvars
			Key: envvarsFlagKey,
			Usage: fmt.Sprintf(
				"String containing environment variables that will be set when running the container, in "+
					"the form \"KEY1%vVALUE1%vKEY2%vVALUE2\"",
				envvarKeyValueDelimiter,
				envvarDeclarationsDelimiter,
				envvarKeyValueDelimiter,
			),
			Type: flags.FlagType_String,
		},
		{
			Key: portsFlagKey,
			Usage: fmt.Sprintf(`String containing declarations of ports that the container will listen on, in the form, %q`+
				` where %q is a user friendly string for identifying the port, %q is required field, %q is an optional field which must be either`+
				` '%v' or '%v' and defaults to '%v' if omitted and %q is user defined optional value. %v`,
				serviceAddSpecWithPortId,
				portIdSpecForHelp,
				portNumberSpecForHelp,
				transportProtocolSpecForHelp,
				strings.ToLower(kurtosis_core_rpc_api_bindings.Port_TCP.String()),
				strings.ToLower(kurtosis_core_rpc_api_bindings.Port_UDP.String()),
				defaultTransportProtocolStr,
				maybeApplicationProtocolSpecForHelp,
				generateExampleForPortFlag(),
			),
			Type: flags.FlagType_String,
		},
		{
			Key: filesFlagKey,
			Usage: fmt.Sprintf(
				"String containing declarations of files paths on the container -> artifact name  where the contents of those "+
					"files artifacts should be mounted, in the form \"MOUNTPATH1%vARTIFACTNAME1%vMOUNTPATH2%vARTIFACTNAME2\" where "+
					"ARTIFACTNAME is the name returned by Kurtosis when uploading files to the enclave (e.g. via the '%v %v' command)",
				filesArtifactMountpointDelimiter,
				filesArtifactMountsDelimiter,
				filesArtifactMountpointDelimiter,
				command_str_consts.FilesCmdStr,
				command_str_consts.FilesUploadCmdStr,
			),
			Type: flags.FlagType_String,
		},
		{
			Key:     privateIPAddressPlaceholderKey,
			Usage:   "Kurtosis will replace occurrences of this string in the ENTRYPOINT args, ENV vars and CMD args with the IP address of the container inside the enclave",
			Type:    flags.FlagType_String,
			Default: privateIPAddressPlaceholderDefault,
		},
		{
			Key:     fullUuidsFlagKey,
			Usage:   "If true then Kurtosis prints full UUIDs instead of shortened UUIDs. Default false.",
			Type:    flags.FlagType_Bool,
			Default: fullUuidFlagKeyDefault,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier value using key '%v'", enclaveIdentifierArgKey)
	}

	serviceName, err := args.GetNonGreedyArg(serviceNameArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service name value using key '%v'", serviceNameArgKey)
	}

	image, err := args.GetNonGreedyArg(serviceImageArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service image value using key '%v'", serviceImageArgKey)
	}

	cmdArgs, err := args.GetGreedyArg(cmdArgsArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the CMD args using key '%v'", cmdArgsArgKey)
	}

	entrypointStr, err := flags.GetString(entrypointBinaryFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the ENTRYPOINT binary using key '%v'", entrypointBinaryFlagKey)
	}

	envvarsStr, err := flags.GetString(envvarsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the env vars string using key '%v'", envvarsFlagKey)
	}

	portsStr, err := flags.GetString(portsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the ports string using key '%v'", portsFlagKey)
	}

	filesArtifactMountsStr, err := flags.GetString(filesFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the files artifact mounts string using key '%v'", filesFlagKey)
	}

	privateIPAddressPlaceholder, err := flags.GetString(privateIPAddressPlaceholderKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the private IP address place holder using key '%v'", privateIPAddressPlaceholderKey)
	}

	showFullUuids, err := flags.GetBool(fullUuidsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidsFlagKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}

	entrypoint := []string{}
	if entrypointStr != "" {
		entrypoint = append(entrypoint, entrypointStr)
	}
	serviceConfigStarlark, err := GetServiceConfigStarlark(image, portsStr, cmdArgs, entrypoint, envvarsStr, filesArtifactMountsStr, defaultLimits, defaultLimits, defaultLimits, defaultLimits, privateIPAddressPlaceholder)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred getting the container config to start image '%v' with CMD '%+v', ENTRYPOINT '%v',  envvars '%v' and private IP address placeholder '%v'",
			image,
			cmdArgs,
			entrypointStr,
			envvarsStr,
			privateIPAddressPlaceholder,
		)
	}

	_, err = RunAddServiceStarlarkScript(ctx, serviceName, enclaveIdentifier, GetAddServiceStarlarkScript(serviceName, serviceConfigStarlark), enclaveCtx)
	if err != nil {
		return err // already wrapped
	}
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "An error has occurred when getting service added using add command")
	}

	privatePorts := serviceCtx.GetPrivatePorts()
	publicPorts := serviceCtx.GetPublicPorts()
	publicIpAddr := serviceCtx.GetMaybePublicIPAddress()

	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err != nil {
		logrus.Warnf("Could not retrieve the current context. Kurtosis will assume context is local and not" +
			"map the service ports. If you're running on a remote context and are seeing this error, then" +
			"the service will be unreachable locally. Turn on debug logging to see the actual error.")
		logrus.Debugf("Error was: %v", err.Error())
		return nil
	}
	if !store.IsRemote(currentContext) {
		logrus.Debugf("Current context is local, not mapping service ports")
		return nil
	}

	// Map the service public ports to their local port
	portalManager := portal_manager.NewPortalManager()
	portsMapping := map[uint16]*services.PortSpec{}
	for _, portSpec := range serviceCtx.GetPublicPorts() {
		portsMapping[portSpec.GetNumber()] = portSpec
	}
	successfullyMappedPorts, failedPorts, err := portalManager.MapPorts(ctx, portsMapping)
	if err != nil {
		var stringifiedPortMapping []string
		for localPort, remotePort := range failedPorts {
			stringifiedPortMapping = append(stringifiedPortMapping, fmt.Sprintf("%d:%d", localPort, remotePort.GetNumber()))
		}
		// TODO: once we have a manual `kurtosis port map` command, suggest using it here to manually map the failed port
		logrus.Warnf("The service is running but the following port(s) could not be mapped locally: %s.",
			strings.Join(stringifiedPortMapping, portMappingSeparatorForLogs))
	}
	logrus.Infof("Successfully mapped %d ports. The service is reachable locally on its ephemeral port numbers",
		len(successfullyMappedPorts))

	fmt.Printf("Service ID: %v\n", serviceName)
	if len(privatePorts) > 0 {
		fmt.Println("Ports Bindings:")
	} else {
		fmt.Println("Port Bindings: <none defined>")
	}
	keyValuePrinter := output_printers.NewKeyValuePrinter()
	keyValuePrinter.AddPair(serviceNameTitleKey, string(serviceCtx.GetServiceName()))
	serviceUuidStr := string(serviceCtx.GetServiceUUID())
	if showFullUuids {
		keyValuePrinter.AddPair(serviceUuidTitleKey, serviceUuidStr)
	} else {
		shortenedUuidStr := uuid_generator.ShortenedUUIDString(serviceUuidStr)
		keyValuePrinter.AddPair(serviceUuidTitleKey, shortenedUuidStr)
	}

	for portId, privatePortSpec := range privatePorts {
		publicPortSpec, found := publicPorts[portId]
		// With Kubernetes, it's possible for a private port not to have a corresponding public port
		if !found {
			continue
		}

		apiProtocolEnum := kurtosis_core_rpc_api_bindings.Port_TransportProtocol(publicPortSpec.GetTransportProtocol())
		protocolStr := strings.ToLower(apiProtocolEnum.String())

		portApplicationProtocolStr := emptyApplicationProtocol
		if privatePortSpec.GetMaybeApplicationProtocol() != emptyApplicationProtocol {
			portApplicationProtocolStr = fmt.Sprintf("%v%v", privatePortSpec.GetMaybeApplicationProtocol(), linkDelimiter)
		}
		portBindingInfo := fmt.Sprintf(
			"%v/%v -> %v%v:%v",
			privatePortSpec.GetNumber(),
			protocolStr,
			portApplicationProtocolStr,
			publicIpAddr,
			publicPortSpec.GetNumber(),
		)
		keyValuePrinter.AddPair(
			fmt.Sprintf("   %v", portId),
			portBindingInfo,
		)
	}
	keyValuePrinter.Print()

	return nil
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

// GetServiceConfigStarlark TODO(victor.colombo): Extract this to a more reasonable place
func GetServiceConfigStarlark(
	image string,
	portsStr string,
	cmdArgs []string,
	entrypoint []string,
	envvarsStr string,
	filesArtifactMountsStr string,
	cpuAllocationMillicpus int,
	memoryAllocationMegabytes int,
	minCpuMilliCores int,
	minMemoryMegaBytes int,
	privateIPAddressPlaceholder string,
) (string, error) {
	envvarsMap, err := ParseEnvVarsStr(envvarsStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing environment variables string '%v'", envvarsStr)
	}

	ports, err := ParsePortsStr(portsStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing ports string '%v'", portsStr)
	}

	filesArtifactMounts, err := ParseFilesArtifactMountsStr(filesArtifactMountsStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing files artifact mounts string '%v'", filesArtifactMountsStr)
	}
	return services.GetServiceConfigStarlark(image, ports, filesArtifactMounts, entrypoint, cmdArgs, envvarsMap, privateIPAddressPlaceholder, cpuAllocationMillicpus, memoryAllocationMegabytes, minCpuMilliCores, minMemoryMegaBytes), nil
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

func ParseFilesArtifactMountsStr(filesArtifactMountsStr string) (map[string]string, error) {
	result := map[string]string{}
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
		filesArtifactName := mountFragments[1]

		if existingName, found := result[mountpoint]; found {
			return nil, stacktrace.NewError(
				"Mountpoint '%v' is declared twice; once to artifact name '%v' and again to artifact name '%v'",
				mountpoint,
				existingName,
				filesArtifactName,
			)
		}

		result[mountpoint] = filesArtifactName
	}

	return result, nil
}

func generateExampleForPortFlag() string {
	return fmt.Sprintf(
		`Example: "PORTID1%v1234%vudp%vPORTID2%vhttp%v5678%vPORTID3%vhttp%v6000%vudp"`,
		portIdSpecDelimiter,
		portNumberProtocolDelimiter,
		portDeclarationsDelimiter,
		portIdSpecDelimiter,
		portApplicationProtocolDelimiter,
		portDeclarationsDelimiter,
		portIdSpecDelimiter,
		portApplicationProtocolDelimiter,
		portNumberProtocolDelimiter,
	)
}
