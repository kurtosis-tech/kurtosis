package add

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
	"strings"
)

const (
	enclaveIdArgKey        = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy   = false

	serviceIdArgKey = "service-id"

	serviceImageArgKey = "image"

	cmdArgsArgKey = "cmd-arg"

	entrypointBinaryFlagKey = "entrypoint"

	envvarsFlagKey          = "env"
	envvarKeyValueDelimiter     = "="
	envvarDeclarationsDelimiter = ","

	portsFlagKey = "ports"
	portIdSpecDelimiter     = "="
	portNumberProtocolDelimiter = "/"
	portDeclarationsDelimiter = ","
	portNumberUintParsingBase = 10
	portNumberUintParsingBits = 16

	filesFlagKey                       = "files"
	filesArtifactMountsDelimiter       = ","
	filesArtifactIdMountpointDelimiter = ":"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey  = "engine-client"

	// Magic key that, when found in CMD, ENTRYPOINT, or envvars, will be replaced with the service's IP address
	serviceIpAddrReplaceKeyword = "IPADDR"

	// Each envvar should be KEY1=VALUE1, which means we should have two components to each envvar declaration
	expectedNumberKeyValueComponentsInEnvvarDeclaration = 2
)

var defaultPortProtocolStr = strings.ToLower(kurtosis_core_rpc_api_bindings.Port_TCP.String())

var ServiceAddCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:              command_str_consts.ServiceAddCmdStr,
	ShortDescription:        "Adds a service to an enclave",
	LongDescription:         fmt.Sprintf(
		"Adds a new service with the given parameters to the given enclave (NOTE: any instances of " +
			"the string '%v' in the CMD args will be replaced with the container's IP address inside the enclave)",
		serviceIpAddrReplaceKeyword,
	),
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:  engineClientCtxKey,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		{
			Key: serviceIdArgKey,
		},
		{
			Key: serviceImageArgKey,
		},
		{
			Key:             cmdArgsArgKey,
			IsOptional:      true,
			IsGreedy:        true,
			DefaultValue: []string{},
		},
	},
	Flags: []*flags.FlagConfig{
		{
			Key: entrypointBinaryFlagKey,
			Usage:     fmt.Sprintf(
				"ENTRYPOINT binary that will be used when running the container, overriding the " +
					"image's default ENTRYPOINT (NOTE: any instances of the string '%v' will be " +
					"replaced with the container's IP address inside the enclave)",
				serviceIpAddrReplaceKeyword,
			),
			// TODO Make this a string list
			Type:      flags.FlagType_String,
		},
		{
			// TODO We currently can't handle commas, so allow users to set the flag multiple times to set multiple envvars
			Key:       envvarsFlagKey,
			Usage:     fmt.Sprintf(
				"String containing environment variables that will be set when running the container, in " +
					"the form \"KEY1%vVALUE1%vKEY2%vVALUE2\" (NOTE: any instances of the string '%v' in values " +
					"will be replaced with the container's IP address inside the enclave)",
				envvarKeyValueDelimiter,
				envvarDeclarationsDelimiter,
				envvarKeyValueDelimiter,
				serviceIpAddrReplaceKeyword,
			),
			Type:      flags.FlagType_String,
		},
		{
			Key: portsFlagKey,
			Usage: fmt.Sprintf(
				"String containing declarations of ports that the container will listen on, in the form "+
					"\"PORTID1%v1234%vPROTOCOL1%vPORTID2%v5678\" where the PORTID is a user-friendly string for "+
					"identifying the port, the port number is required, and the PROTOCOL must be either '%v' or '%v' "+
					"and defaults to '%v' if omitted",
				portIdSpecDelimiter,
				portNumberProtocolDelimiter,
				portDeclarationsDelimiter,
				portIdSpecDelimiter,
				strings.ToLower(kurtosis_core_rpc_api_bindings.Port_TCP.String()),
				strings.ToLower(kurtosis_core_rpc_api_bindings.Port_UDP.String()),
				defaultPortProtocolStr,
			),
			Type:      flags.FlagType_String,
		},
		{
			Key:       filesFlagKey,
			Usage:     fmt.Sprintf(
				"String containing declarations of files artifact IDs -> paths on the container where the contents of those " +
					"files artifacts should be mounted, in the form \"ARTIFACTID1%vMOUNTPATH1%vARTIFACTID2%vMOUNTPATH2\" where " +
					"ARTIFACTID is the ID returned by Kurtosis when uploading files to the enclave (e.g. via the '%v %v' command)",
				filesArtifactIdMountpointDelimiter,
				filesArtifactMountsDelimiter,
				filesArtifactIdMountpointDelimiter,
				command_str_consts.FilesCmdStr,
				command_str_consts.FilesUploadCmdStr,
			),
			Type:      flags.FlagType_String,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveId, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID value using key '%v'", enclaveIdArgKey)
	}

	serviceId, err := args.GetNonGreedyArg(serviceIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service ID value using key '%v'", serviceIdArgKey)
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

	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting existing enclaves")
	}

	infoForEnclave, found := getEnclavesResp.EnclaveInfo[enclaveId]
	if !found {
		return stacktrace.Propagate(err, "No enclave with ID '%v' exists", enclaveId)
	}

	enclaveCtx, err := getEnclaveContextFromEnclaveInfo(infoForEnclave)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveId)
	}

	containerConfigSupplier, err := getContainerConfigSupplier(image, portsStr, cmdArgs, entrypointStr, envvarsStr, filesArtifactMountsStr)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred getting the container config supplier to start image '%v' with CMD '%+v', ENTRYPOINT '%v', and envvars '%v'",
			image,
			cmdArgs,
			entrypointStr,
			envvarsStr,
		)
	}

	// TODO Allow adding services to an already-repartitioned enclave
	serviceCtx, err := enclaveCtx.AddService(
		services.ServiceID(serviceId),
		containerConfigSupplier,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding service '%v' to enclave '%v'", serviceId, enclaveId)
	}

	privatePorts := serviceCtx.GetPrivatePorts()
	publicPorts := serviceCtx.GetPublicPorts()
	publicIpAddr := serviceCtx.GetMaybePublicIPAddress()

	fmt.Println(fmt.Sprintf("Service ID: %v", serviceId))
	if len(privatePorts) > 0 {
		fmt.Println("Ports Bindings:")
	} else {
		fmt.Println("Port Bindings: <none defined>")
	}
	keyValuePrinter := output_printers.NewKeyValuePrinter()
	for portId, privatePortSpec := range privatePorts {
		publicPortSpec, found := publicPorts[portId]
		if !found {
			return stacktrace.NewError("Found private port with ID '%v' that doesn't correspond to any public port; this is a bug in Kurtosis!", portId)
		}

		apiProtocolEnum := kurtosis_core_rpc_api_bindings.Port_Protocol(publicPortSpec.GetProtocol())
		protocolStr := strings.ToLower(apiProtocolEnum.String())
		portBindingInfo := fmt.Sprintf(
			"%v/%v -> %v:%v",
			privatePortSpec.GetNumber(),
			protocolStr,
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

// TODO TODO REMOVE ALL THIS WHEN NewEnclaveContext CAN JUST TAKE IN IP ADDR & PORT NUM!!!
func getEnclaveContextFromEnclaveInfo(infoForEnclave *kurtosis_engine_rpc_api_bindings.EnclaveInfo) (*enclaves.EnclaveContext, error) {
	enclaveId := infoForEnclave.EnclaveId

	apiContainerHostMachineIpAddr, apiContainerHostMachineGrpcPortNum, err := enclave_liveness_validator.ValidateEnclaveLiveness(infoForEnclave)
	if err != nil {
		return nil, stacktrace.NewError("Cannot add service because the API container in enclave '%v' is not running", enclaveId)
	}

	apiContainerHostMachineUrl := fmt.Sprintf(
		"%v:%v",
		apiContainerHostMachineIpAddr,
		apiContainerHostMachineGrpcPortNum,
	)
	conn, err := grpc.Dial(apiContainerHostMachineUrl, grpc.WithInsecure())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container grpc port at '%v' in enclave '%v'",
			apiContainerHostMachineUrl,
			enclaveId,
		)
	}
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)
	enclaveCtx := enclaves.NewEnclaveContext(
		apiContainerClient,
		enclaves.EnclaveID(enclaveId),
	)

	return enclaveCtx, nil
}

func getContainerConfigSupplier(
	image string,
	portsStr string,
	cmdArgs []string,
	entrypoint string,
	envvarsStr string,
	filesArtifactMountsStr string,
) (func(ipAddr string) (*services.ContainerConfig, error), error) {
	envvarsMap, err := parseEnvVarsStr(envvarsStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing environment variables string '%v'", envvarsStr)
	}

	ports, err := parsePortsStr(portsStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing ports string '%v'", portsStr)
	}

	filesArtifactMounts, err := parseFilesArtifactMountsStr(filesArtifactMountsStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing files artifact mounts string '%v'", filesArtifactMountsStr)
	}

	return func(ipAddr string) (*services.ContainerConfig, error) {
		ipReplacedCmdArgs := []string{}
		for _, arg := range cmdArgs {
			newArg := strings.ReplaceAll(arg, serviceIpAddrReplaceKeyword, ipAddr)
			ipReplacedCmdArgs = append(ipReplacedCmdArgs, newArg)
		}

		ipReplacedEntrypoint := strings.ReplaceAll(entrypoint, serviceIpAddrReplaceKeyword, ipAddr)

		ipReplacedEnvvars := map[string]string{}
		for key, value := range envvarsMap {
			newValue := strings.ReplaceAll(value, serviceIpAddrReplaceKeyword, ipAddr)
			ipReplacedEnvvars[key] = newValue
		}

		resultBuilder := services.NewContainerConfigBuilder(image)
		if len(ipReplacedCmdArgs) > 0 {
			resultBuilder.WithCmdOverride(ipReplacedCmdArgs)
		}
		if ipReplacedEntrypoint != "" {
			resultBuilder.WithEntrypointOverride([]string{ipReplacedEntrypoint})
		}
		if len(ipReplacedEnvvars) > 0 {
			resultBuilder.WithEnvironmentVariableOverrides(ipReplacedEnvvars)
		}
		if len(ports) > 0 {
			resultBuilder.WithUsedPorts(ports)
		}
		if len(filesArtifactMounts) > 0 {
			resultBuilder.WithFiles(filesArtifactMounts)
		}

		return resultBuilder.Build(), nil
	}, nil
}

// Parses a string in the form KEY1=VALUE1,KEY2=VALUE2 into a map of strings
// An empty string will result in an empty map
// Empty strings will be skipped (e.g. ',,,' will result in an empty map)
func parseEnvVarsStr(envvarsStr string) (map[string]string, error) {
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
func parsePortsStr(portsStr string) (map[string]*services.PortSpec, error) {
	result := map[string]*services.PortSpec{}
	if strings.TrimSpace(portsStr) == "" {
		return result, nil
	}

	allPortDeclarationStrs := strings.Split(portsStr, portDeclarationsDelimiter)
	for _, portDeclarationStr := range allPortDeclarationStrs {
		if len(strings.TrimSpace(portDeclarationStr)) == 0 {
			continue
		}

		portIdSpecComponents := strings.Split(portDeclarationStr, portIdSpecDelimiter)
		if len(portIdSpecComponents) != 2 {
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

func parsePortSpecStr(specStr string) (*services.PortSpec, error) {
	if len(strings.TrimSpace(specStr)) == 0 {
		return nil, stacktrace.NewError("Cannot parse empty spec string")
	}

	portSpecComponents := strings.Split(specStr, portNumberProtocolDelimiter)
	if len(portSpecComponents) == 0 {
		// Not sure how this would even happen
		return nil, stacktrace.NewError("Port spec string '%v' was split into 0 components, which should never happen", specStr)
	}
	numPortSpecComponents := len(portSpecComponents)
	if numPortSpecComponents > 2 {
		return nil, stacktrace.NewError(
			"Port spec string '%v' was split on delimiter '%v' into '%v' components, which indicates too many delimiters",
			specStr,
			portNumberProtocolDelimiter,
			numPortSpecComponents,
		)
	}
	portNumberStr := portSpecComponents[0]
	portProtocolStr := defaultPortProtocolStr
	if numPortSpecComponents > 1 {
		portProtocolStr = portSpecComponents[1]
	}

	portNumberUint64, err := strconv.ParseUint(portNumberStr, portNumberUintParsingBase, portNumberUintParsingBits)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred parsing port number string '%v' with base '%v' and bits '%v'",
			portNumberStr,
			portNumberUintParsingBase,
			portNumberUintParsingBits,
		)
	}
	portNumberUint16 := uint16(portNumberUint64)

	portProtocolEnumInt, found := kurtosis_core_rpc_api_bindings.Port_Protocol_value[strings.ToUpper(portProtocolStr)]
	if !found {
		return nil, stacktrace.NewError("Unrecognized port protocol '%v'", portProtocolStr)
	}
	portProtocol := services.PortProtocol(portProtocolEnumInt)

	return services.NewPortSpec(portNumberUint16, portProtocol), nil
}

func parseFilesArtifactMountsStr(filesArtifactMountsStr string) (map[services.FilesArtifactID]string, error) {
	result := map[services.FilesArtifactID]string{}
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

		mountFragments := strings.Split(trimmedMountStr, filesArtifactIdMountpointDelimiter)
		if len(mountFragments) != 2 {
			return nil, stacktrace.NewError(
				"Files artifact mountpoint string %v was '%v' but should be in the form 'files_artifact_id%vmountpoint'",
				idx,
				trimmedMountStr,
				filesArtifactIdMountpointDelimiter,
			)
		}
		filesArtifactId := services.FilesArtifactID(mountFragments[0])
		mountpoint := mountFragments[1]

		if existingMountpoint, found := result[filesArtifactId]; found {
			return nil, stacktrace.NewError(
				"Files artifact with ID '%v' is declared twice; once to mountpoint '%v' and again to mountpoint '%v'",
				filesArtifactId,
				existingMountpoint,
				mountpoint,
			)
		}

		result[filesArtifactId] = mountpoint
	}

	return result, nil
}