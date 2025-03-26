package update

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/add"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/inspect"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey        = "service"
	isServiceIdentifierArgOptional = false
	isServiceIdentifierArgGreedy   = false

	serviceImageFlagKey     = "image"
	cmdArgsFlagKey          = "cmd"
	entrypointBinaryFlagKey = "entrypoint"

	envvarsFlagKey              = "env"
	envvarKeyValueDelimiter     = "="
	envvarDeclarationsDelimiter = ","

	portsFlagKey                     = "ports"
	portIdSpecDelimiter              = "="
	portNumberProtocolDelimiter      = "/"
	portDeclarationsDelimiter        = ","
	portApplicationProtocolDelimiter = ":"

	filesFlagKey                     = "files"
	filesArtifactMountsDelimiter     = ","
	filesArtifactMountpointDelimiter = ":"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	maybeApplicationProtocolSpecForHelp = "MAYBE_APPLICATION_PROTOCOL"
	transportProtocolSpecForHelp        = "TRANSPORT_PROTOCOL"
	portNumberSpecForHelp               = "PORT_NUMBER"
	portIdSpecForHelp                   = "PORT_ID"
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

var ServiceUpdateCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceUpdateCmdStr,
	ShortDescription:          "Update a service",
	LongDescription:           "Update a service",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     serviceImageFlagKey,
			Usage:   "image",
			Type:    flags.FlagType_String,
			Default: "",
		},
		{
			Key:   cmdArgsFlagKey,
			Usage: "cmd",
			// TODO Make this a string list
			Type:    flags.FlagType_String,
			Default: "",
		},
		{
			Key:   entrypointBinaryFlagKey,
			Usage: "ENTRYPOINT binary that will be used when running the container, overriding the image's default ENTRYPOINT",
			// TODO Make this a string list
			Type:    flags.FlagType_String,
			Default: "",
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
			Type:    flags.FlagType_String,
			Default: "",
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
			Type:    flags.FlagType_String,
			Default: "",
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
			Type:    flags.FlagType_String,
			Default: "",
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		service_identifier_arg.NewServiceIdentifierArg(
			serviceIdentifierArgKey,
			enclaveIdentifierArgKey,
			isServiceIdentifierArgGreedy,
			isServiceIdentifierArgOptional,
		),
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for non-greedy enclave identifier arg '%v' but none was found; this is a bug in the Kurtosis CLI!", enclaveIdentifierArgKey)
	}

	// TODO: need to enforce this is always service name, and not service uuid or short uuid
	serviceName, err := args.GetNonGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for non-greedy enclave identifier arg '%v' but none was found; this is a bug in the Kurtosis CLI!", serviceIdentifierArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}

	var overrideImage string
	var overridePorts map[string]*kurtosis_core_rpc_api_bindings.Port
	var overrideFilesArtifactsMountpoint map[string]string
	var overrideEntrypoint []string
	var overrideCmd []string
	var overrideEnvVars map[string]string

	imageStr, err := flags.GetString(serviceImageFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the image using key '%v'", serviceImageFlagKey)
	}
	overrideImage = imageStr

	cmdStr, err := flags.GetString(cmdArgsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the cmd using key '%v'", cmdArgsFlagKey)
	}
	if cmdStr != "" {
		overrideCmd = []string{cmdStr}
	}

	entrypointStr, err := flags.GetString(entrypointBinaryFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the ENTRYPOINT binary using key '%v'", entrypointBinaryFlagKey)
	}
	if entrypointStr != "" {
		overrideEntrypoint = []string{entrypointStr}
	}

	envVarsStr, err := flags.GetString(envvarsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the env vars string using key '%v'", envvarsFlagKey)
	}
	if envVarsStr != "" {
		overrideEnvVars, err = add.ParseEnvVarsStr(envVarsStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing env vars string: %v", envVarsStr)
		}
	}

	portsStr, err := flags.GetString(portsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the ports string using key '%v'", portsFlagKey)
	}
	if portsStr != "" {
		overridePorts, err = add.ParsePortsStr(portsStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing ports string: %v", portsStr)
		}
	}

	filesArtifactMountsStr, err := flags.GetString(filesFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the files artifact mounts string using key '%v'", filesFlagKey)
	}
	if filesArtifactMountsStr != "" {
		overrideFilesArtifactsMountpoint, err = add.ParseFilesArtifactMountsStr(filesArtifactMountsStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing files artifacts mount points string: %v", filesArtifactMountsStr)
		}
	}

	serviceInspectOutputMap, err := inspect.PrintServiceInspect(ctx, kurtosisBackend, kurtosisCtx, enclaveIdentifier, serviceName, false, "json", false)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting")
	}

	var currImage string
	if img, ok := serviceInspectOutputMap[inspect.ServiceImageTitleName]; ok {
		currImage = img.(string)
	}

	var currPorts map[string]*kurtosis_core_rpc_api_bindings.Port
	if ports, ok := serviceInspectOutputMap[inspect.ServicePortsTitleName]; ok {
		currPorts = ports.(map[string]*kurtosis_core_rpc_api_bindings.Port)
	}

	currFilesArtifactsMountpoint := map[string]string{}
	if filesArtifactsMountpoints, ok := serviceInspectOutputMap[inspect.ServiceFilesTitleName]; ok {
		currFilesArtifactsMountpointWithMultipleMounts := filesArtifactsMountpoints.(map[string][]string)
		for svcName, mountpoints := range currFilesArtifactsMountpointWithMultipleMounts {
			currFilesArtifactsMountpoint[svcName] = mountpoints[0]
		}
	}

	var currEntrypoint []string
	if entrypoint, ok := serviceInspectOutputMap[inspect.ServiceEntrypointArgsTitleName]; ok {
		currEntrypoint = entrypoint.([]string)
	}

	var currCmd []string
	if cmd, ok := serviceInspectOutputMap[inspect.ServiceCmdArgsTitleName]; ok {
		currCmd = cmd.([]string)
	}

	var currEnvVarsMap map[string]string
	if cmd, ok := serviceInspectOutputMap[inspect.ServiceEnvVarsTitleName]; ok {
		currEnvVarsMap = cmd.(map[string]string)
	}

	var currCpuAllocationMillicpus int
	if cpuAlloc, ok := serviceInspectOutputMap[inspect.ServiceMaxCpuAllocationTitleName]; ok {
		currCpuAllocationMillicpus = cpuAlloc.(int)
	}
	var currMemoryMegabytes int
	if memoryAlloc, ok := serviceInspectOutputMap[inspect.ServiceMemoryAllocationTitleName]; ok {
		currMemoryMegabytes = memoryAlloc.(int)
	}
	var currMinCpuMillicores int
	if minCpuMillis, ok := serviceInspectOutputMap[inspect.ServiceMinCpuAllocationTitleName]; ok {
		currMinCpuMillicores = minCpuMillis.(int)
	}
	var currMinMemoryBytes int
	if minMem, ok := serviceInspectOutputMap[inspect.ServiceMinMemoryAllocationTitleName]; ok {
		currMinMemoryBytes = minMem.(int)
	}

	// merge
	var mergedImage string
	var mergedEntrypoint []string
	var mergedCmd []string

	// if override image was provided, use that as the image, otherwise keep curr
	if overrideImage != "" {
		mergedImage = overrideImage
	} else {
		mergedImage = currImage
	}

	// if override entrypoint was provided, use that as the entrypoint, otherwise keep curr
	if len(overrideEntrypoint) > 0 {
		mergedEntrypoint = overrideEntrypoint
	} else {
		mergedEntrypoint = currEntrypoint
	}

	// if override cmd was provided, use that as the cmd, otherwise keep curr
	if len(overrideCmd) > 0 {
		mergedCmd = overrideCmd
	} else {
		mergedCmd = currCmd
	}

	// combine current ports with override ports
	mergedPorts := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	for portId, port := range overridePorts {
		mergedPorts[portId] = port
	}
	for portId, port := range currPorts {
		mergedPorts[portId] = port
	}

	// combine current env vars with override env vars
	mergedEnvVarsMap := map[string]string{}
	for key, val := range overrideEnvVars {
		mergedEnvVarsMap[key] = val
	}
	for key, val := range currEnvVarsMap {
		mergedEnvVarsMap[key] = val
	}

	// combine current files artifacts mount points with override mount points
	mergedFilesArtifactsMountpoint := map[string]string{}
	for key, val := range overrideFilesArtifactsMountpoint {
		mergedFilesArtifactsMountpoint[key] = val
	}
	for key, val := range currFilesArtifactsMountpoint {
		mergedFilesArtifactsMountpoint[key] = val
	}

	var mergedCpuAllocationMillicpus int
	mergedCpuAllocationMillicpus = currCpuAllocationMillicpus

	var mergedMemoryMegabytes int
	mergedMemoryMegabytes = currMemoryMegabytes

	var mergedMinCpuMillicores int
	mergedMinCpuMillicores = currMinCpuMillicores

	var mergedMinMemoryBytes int
	mergedMinMemoryBytes = currMinMemoryBytes

	// call getServiceConfig
	serviceConfigStr := services.GetServiceConfigStarlark(
		mergedImage,
		mergedPorts,
		mergedFilesArtifactsMountpoint, // TODO: get in svc inspect
		mergedEntrypoint,
		mergedCmd,
		mergedEnvVarsMap,
		service_config.PrivateIpAddressPlaceholderAttr,
		mergedCpuAllocationMillicpus, // TODO: get in svc inspect
		mergedMemoryMegabytes,        // TODO: get in svc inspect
		mergedMinCpuMillicores,       // TODO: get in svc inspect
		mergedMinMemoryBytes)         // TODO: get in service inspect
	//logrus.Infof("SERVICE CONFIG STRING: %v", serviceConfigStr)

	// call run add service starlark script
	addServiceStarlarkStr := add.GetAddServiceStarlarkScript(serviceName, serviceConfigStr)

	logrus.Infof("Running update service starlark for service '%v' in enclave '%v'...", serviceName, enclaveIdentifier)
	starlarkRunResult, err := add.RunAddServiceStarlarkScript(ctx, serviceName, enclaveIdentifier, addServiceStarlarkStr, enclaveCtx)
	if err != nil {
		return err //already wrapped
	}

	out.PrintOutLn(string(starlarkRunResult.RunOutput))

	return nil
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
