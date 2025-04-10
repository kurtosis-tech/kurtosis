package add

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/inspect"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/service_helpers"
	"io"
	"os"
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

	defaultLimits = 0

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	privateIPAddressPlaceholderKey     = "ip-address-placeholder"
	privateIPAddressPlaceholderDefault = ""

	emptyApplicationProtocol = ""
	linkDelimiter            = "://"

	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	JsonConfigFlagKey            = "json-service-config"
	JsonConfigFlagKeyDefault     = ""
	readJsonConfigFromStdinInput = "-"

	portMappingSeparatorForLogs = ", "
)

var (
	defaultTransportProtocolStr = strings.ToLower(kurtosis_core_rpc_api_bindings.Port_TCP.String())
	serviceAddSpec              = fmt.Sprintf(
		`%v%v%v%v%v`,
		service_helpers.MaybeApplicationProtocolSpecForHelp,
		service_helpers.PortApplicationProtocolDelimiter,
		service_helpers.PortNumberSpecForHelp,
		service_helpers.PortNumberProtocolDelimiter,
		service_helpers.TransportProtocolSpecForHelp,
	)
	serviceAddSpecWithPortId = fmt.Sprintf(
		`%v%v%v`,
		service_helpers.PortIdSpecForHelp,
		service_helpers.PortIdSpecDelimiter,
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
			Key:          service_helpers.ImageKey,
			IsOptional:   true,
			DefaultValue: "",
		},
	},
	Flags: []*flags.FlagConfig{
		{
			Key:     service_helpers.CmdKey,
			Usage:   "CMD binary that will be used when running the container",
			Type:    flags.FlagType_String,
			Default: "",
		},
		{
			Key:   service_helpers.EntrypointFlagKey,
			Usage: "ENTRYPOINT binary that will be used when running the container, overriding the image's default ENTRYPOINT",
			// TODO Make this a string list
			Type:    flags.FlagType_String,
			Default: "",
		},
		{
			// TODO We currently can't handle commas, so allow users to set the flag multiple times to set multiple envvars
			Key: service_helpers.EnvvarsFlagKey,
			Usage: fmt.Sprintf(
				"String containing environment variables that will be set when running the container, in "+
					"the form \"KEY1%vVALUE1%vKEY2%vVALUE2\"",
				service_helpers.EnvvarKeyValueDelimiter,
				service_helpers.EnvvarDeclarationsDelimiter,
				service_helpers.EnvvarKeyValueDelimiter,
			),
			Type:    flags.FlagType_String,
			Default: "",
		},
		{
			Key: service_helpers.PortsFlagKey,
			Usage: fmt.Sprintf(`String containing declarations of ports that the container will listen on, in the form, %q`+
				` where %q is a user friendly string for identifying the port, %q is required field, %q is an optional field which must be either`+
				` '%v' or '%v' and defaults to '%v' if omitted and %q is user defined optional value. %v`,
				serviceAddSpecWithPortId,
				service_helpers.PortIdSpecForHelp,
				service_helpers.PortNumberSpecForHelp,
				service_helpers.TransportProtocolSpecForHelp,
				strings.ToLower(kurtosis_core_rpc_api_bindings.Port_TCP.String()),
				strings.ToLower(kurtosis_core_rpc_api_bindings.Port_UDP.String()),
				defaultTransportProtocolStr,
				service_helpers.MaybeApplicationProtocolSpecForHelp,
				generateExampleForPortFlag(),
			),
			Type:    flags.FlagType_String,
			Default: "",
		},
		{
			Key: service_helpers.FilesFlagKey,
			Usage: fmt.Sprintf(
				"String containing declarations of files paths on the container -> artifact name  where the contents of those "+
					"files artifacts should be mounted, in the form \"MOUNTPATH1%vARTIFACTNAME1%vMOUNTPATH2%vARTIFACTNAME2\" where "+
					"ARTIFACTNAME is the name returned by Kurtosis when uploading files to the enclave (e.g. via the '%v %v' command)."+
					"directories can be mounted by mounting multiple artifacts to the same mountpath, in the form, \"MOUNTPATH1%vARTIFACTNAME1%vARTIFACTNAME2\"",
				service_helpers.FilesArtifactMountpointDelimiter,
				service_helpers.FilesArtifactMountsDelimiter,
				service_helpers.FilesArtifactMountpointDelimiter,
				command_str_consts.FilesCmdStr,
				command_str_consts.FilesUploadCmdStr,
				service_helpers.FilesArtifactMountpointDelimiter,
				service_helpers.FilesMultipleArtifactsDelimiter,
			),
			Type:    flags.FlagType_String,
			Default: "",
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
		{
			Key: JsonConfigFlagKey,
			Usage: fmt.Sprintf("If a json formatted service config string is provided via stdin using '%v' value for this flag,"+
				"stdin will get parsed to json and used as values for the service config."+
				"The json format should be identical to the json output format from %v %v %v %v %v."+
				"If stdin value not provided, will assume input is a filepath to a service config json and attempts to read from it.",
				readJsonConfigFromStdinInput,
				command_str_consts.KurtosisCmdStr,
				command_str_consts.ServiceCmdStr,
				command_str_consts.ServiceInspectCmdStr,
				inspect.OutputFormatKeyShorthand,
				inspect.JsonOutputFormat,
			),
			Type:    flags.FlagType_String,
			Default: JsonConfigFlagKeyDefault,
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

	image, err := args.GetNonGreedyArg(service_helpers.ImageKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service image value using key '%v'", service_helpers.ImageKey)
	}

	entrypointStr, err := flags.GetString(service_helpers.EntrypointFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the ENTRYPOINT binary using key '%v'", service_helpers.EntrypointFlagKey)
	}

	cmdStr, err := flags.GetString(service_helpers.CmdKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the CMD using key '%v'", service_helpers.CmdKey)
	}

	envVarsStr, err := flags.GetString(service_helpers.EnvvarsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the env vars string using key '%v'", service_helpers.EnvvarsFlagKey)
	}

	portsStr, err := flags.GetString(service_helpers.PortsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the ports string using key '%v'", service_helpers.PortsFlagKey)
	}

	filesArtifactMountsStr, err := flags.GetString(service_helpers.FilesFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the files artifact mounts string using key '%v'", service_helpers.FilesFlagKey)
	}

	privateIPAddressPlaceholder, err := flags.GetString(privateIPAddressPlaceholderKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the private IP address place holder using key '%v'", privateIPAddressPlaceholderKey)
	}

	showFullUuids, err := flags.GetBool(fullUuidsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidsFlagKey)
	}

	jsonServiceConfigFlagValue, err := flags.GetString(JsonConfigFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the json service config string using key '%v'.", JsonConfigFlagKey)
	}
	jsonServiceConfigStr, err := processJsonServiceConfigFlagInput(jsonServiceConfigFlagValue)
	if err != nil {
		return err // already wrapped
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}

	var serviceConfigStarlarkStr string
	if jsonServiceConfigStr != "" {
		var serviceConfigJson services.ServiceConfig
		if err = json.Unmarshal([]byte(jsonServiceConfigStr), &serviceConfigJson); err != nil {
			return stacktrace.Propagate(err, "An error occurred unmarshalling json service config string:\n '%v'.", jsonServiceConfigStr)
		}
		serviceConfigStarlarkStr = services.GetFullServiceConfigStarlark(
			serviceConfigJson.Image,
			services.ConvertJsonPortToApiPort(serviceConfigJson.PrivatePorts),
			serviceConfigJson.Files,
			serviceConfigJson.Entrypoint,
			serviceConfigJson.Cmd,
			serviceConfigJson.EnvVars,
			serviceConfigJson.MaxMillicpus,
			serviceConfigJson.MaxMemory,
			serviceConfigJson.MaxMemory,
			serviceConfigJson.MinMemory,
			serviceConfigJson.User,
			serviceConfigJson.Tolerations,
			serviceConfigJson.NodeSelectors,
			serviceConfigJson.Labels,
			serviceConfigJson.TiniEnabled,
			serviceConfigJson.PrivateIPAddressPlaceholder,
		)
	} else {
		entrypoint := []string{}
		if entrypointStr != "" {
			entrypoint = strings.Split(cmdStr, service_helpers.EntrypointAndCmdDelimiter)
		}
		cmd := []string{}
		if cmdStr != "" {
			cmd = strings.Split(cmdStr, service_helpers.EntrypointAndCmdDelimiter)
		}
		serviceConfigStarlarkStr, err = GetServiceConfigStarlark(image, portsStr, cmd, entrypoint, envVarsStr, filesArtifactMountsStr, defaultLimits, defaultLimits, defaultLimits, defaultLimits, privateIPAddressPlaceholder)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred getting the container config to start image '%v' with CMD '%v', ENTRYPOINT '%v',  envvars '%v' and private IP address placeholder '%v'",
				image,
				cmdStr,
				entrypointStr,
				envVarsStr,
				privateIPAddressPlaceholder,
			)
		}
	}

	addServiceStarlark := service_helpers.GetAddServiceStarlarkScript(serviceName, serviceConfigStarlarkStr)
	_, err = service_helpers.RunAddServiceStarlarkScript(ctx, serviceName, enclaveIdentifier, addServiceStarlark, enclaveCtx)
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
	envvarsMap, err := service_helpers.ParseEnvVarsStr(envvarsStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing environment variables string '%v'", envvarsStr)
	}

	ports, err := service_helpers.ParsePortsStr(portsStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing ports string '%v'", portsStr)
	}

	filesArtifactMounts, err := service_helpers.ParseFilesArtifactMountsStr(filesArtifactMountsStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing files artifact mounts string '%v'", filesArtifactMountsStr)
	}
	tiniEnabled := false

	emptyNodeSelecors := map[string]string{}
	emptyLabels := map[string]string{}
	return services.GetFullServiceConfigStarlark(
		image,
		ports,
		filesArtifactMounts,
		entrypoint,
		cmdArgs,
		envvarsMap,
		uint32(cpuAllocationMillicpus),
		uint32(memoryAllocationMegabytes),
		uint32(minCpuMilliCores),
		uint32(minMemoryMegaBytes),
		nil, //empty user
		nil, //empty tolerations
		emptyNodeSelecors,
		emptyLabels,
		&tiniEnabled,
		privateIPAddressPlaceholder), nil
}

func generateExampleForPortFlag() string {
	return fmt.Sprintf(
		`Example: "PORTID1%v1234%vudp%vPORTID2%vhttp%v5678%vPORTID3%vhttp%v6000%vudp"`,
		service_helpers.PortIdSpecDelimiter,
		service_helpers.PortNumberProtocolDelimiter,
		service_helpers.PortDeclarationsDelimiter,
		service_helpers.PortIdSpecDelimiter,
		service_helpers.PortApplicationProtocolDelimiter,
		service_helpers.PortDeclarationsDelimiter,
		service_helpers.PortIdSpecDelimiter,
		service_helpers.PortApplicationProtocolDelimiter,
		service_helpers.PortNumberProtocolDelimiter,
	)
}

func processJsonServiceConfigFlagInput(jsonServiceConfigFlagInput string) (string, error) {
	var configBytes []byte
	var err error
	switch jsonServiceConfigFlagInput {
	case JsonConfigFlagKeyDefault:
		return JsonConfigFlagKeyDefault, nil
	case readJsonConfigFromStdinInput:
		configBytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred reading json service config from stdin.")
		}
	default:
		logrus.Infof("No json provided via stdin so attempting to read json service config input from file '%v' instead.", jsonServiceConfigFlagInput)
		configBytes, err = os.ReadFile(jsonServiceConfigFlagInput)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred err when attempting to read service config from file '%v'.", jsonServiceConfigFlagInput)
		}
	}
	return string(configBytes), nil
}
