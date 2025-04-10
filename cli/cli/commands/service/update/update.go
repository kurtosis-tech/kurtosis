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
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service/service_helpers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_client_factory"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
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

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
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

var ServiceUpdateCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceUpdateCmdStr,
	ShortDescription:          "Update a service",
	LongDescription:           "Update a service",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     service_helpers.ImageKey,
			Usage:   "Image to use for the service being updated.",
			Type:    flags.FlagType_String,
			Default: "",
		},
		{
			Key:   service_helpers.CmdKey,
			Usage: "CMD to run on the service once it is restarted by the update.",
			// TODO Make this a string list
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
					"ARTIFACTNAME is the name returned by Kurtosis when uploading files to the enclave (e.g. via the '%v %v' command)"+
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

	metricsClient, closeMetricsClientFunc, err := metrics_client_factory.GetMetricsClient()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting metrics client.")
	}
	err = metricsClient.TrackServiceUpdate(enclaveIdentifier, serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred tracking service update metric.")
	}
	defer closeMetricsClientFunc()

	var overrideImage string
	var overridePorts map[string]*kurtosis_core_rpc_api_bindings.Port
	var overrideFilesArtifactsMountpoint map[string][]string
	var overrideEntrypoint []string
	var overrideCmd []string
	var overrideEnvVars map[string]string

	imageStr, err := flags.GetString(service_helpers.ImageKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the image using key '%v'", service_helpers.ImageKey)
	}
	overrideImage = imageStr

	cmdStr, err := flags.GetString(service_helpers.CmdKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the cmd using key '%v'", service_helpers.CmdKey)
	}
	if cmdStr != "" {
		overrideCmd = []string{cmdStr}
	}

	entrypointStr, err := flags.GetString(service_helpers.EntrypointFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the ENTRYPOINT binary using key '%v'", service_helpers.EntrypointFlagKey)
	}
	if entrypointStr != "" {
		overrideEntrypoint = []string{entrypointStr}
	}

	envVarsStr, err := flags.GetString(service_helpers.EnvvarsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the env vars string using key '%v'", service_helpers.EnvvarsFlagKey)
	}
	if envVarsStr != "" {
		overrideEnvVars, err = service_helpers.ParseEnvVarsStr(envVarsStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing env vars string: %v", envVarsStr)
		}
	}

	portsStr, err := flags.GetString(service_helpers.PortsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the ports string using key '%v'", service_helpers.PortsFlagKey)
	}
	if portsStr != "" {
		overridePorts, err = service_helpers.ParsePortsStr(portsStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing ports string: %v", portsStr)
		}
	}

	filesArtifactMountsStr, err := flags.GetString(service_helpers.FilesFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the files artifact mounts string using key '%v'", service_helpers.FilesFlagKey)
	}
	if filesArtifactMountsStr != "" {
		overrideFilesArtifactsMountpoint, err = service_helpers.ParseFilesArtifactMountsStr(filesArtifactMountsStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing files artifacts mount points string: %v", filesArtifactMountsStr)
		}
	}

	_, currServiceConfig, err := service_helpers.GetServiceInfo(ctx, kurtosisCtx, enclaveIdentifier, serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service info of service '%v' in enclave '%v'.", serviceName, enclaveIdentifier)
	}

	// merge overrides with existing service config
	// if override image was provided, use that as the image, otherwise keep curr
	var mergedImage string
	if overrideImage != "" {
		mergedImage = overrideImage
	} else {
		mergedImage = currServiceConfig.Image
	}

	// if override entrypoint was provided, use that as the entrypoint, otherwise keep curr
	var mergedEntrypoint []string
	if len(overrideEntrypoint) > 0 {
		mergedEntrypoint = overrideEntrypoint
	} else {
		mergedEntrypoint = currServiceConfig.Entrypoint
	}

	// if override cmd was provided, use that as the cmd, otherwise keep curr
	var mergedCmd []string
	if len(overrideCmd) > 0 {
		mergedCmd = overrideCmd
	} else {
		mergedCmd = currServiceConfig.Cmd
	}

	// combine current ports with override ports
	mergedPorts := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	currApiPorts := services.ConvertJsonPortToApiPort(currServiceConfig.PrivatePorts)
	for portId, port := range currApiPorts {
		mergedPorts[portId] = port
	}
	for portId, port := range overridePorts {
		mergedPorts[portId] = port
	}

	// combine current env vars with override env vars
	mergedEnvVarsMap := map[string]string{}
	for key, val := range currServiceConfig.EnvVars {
		mergedEnvVarsMap[key] = val
	}
	for key, val := range overrideEnvVars {
		mergedEnvVarsMap[key] = val
	}

	// combine current files artifacts mount points with override mount points
	mergedFilesArtifactsMountpoint := map[string][]string{}
	for key, val := range currServiceConfig.Files {
		mergedFilesArtifactsMountpoint[key] = val
	}
	for key, val := range overrideFilesArtifactsMountpoint {
		mergedFilesArtifactsMountpoint[key] = val
	}

	// call getServiceConfig
	serviceConfigStr := services.GetFullServiceConfigStarlark(
		mergedImage,
		mergedPorts,
		mergedFilesArtifactsMountpoint,
		mergedEntrypoint,
		mergedCmd,
		mergedEnvVarsMap,
		currServiceConfig.MaxMillicpus,
		currServiceConfig.MaxMemory,
		currServiceConfig.MinMillicpus,
		currServiceConfig.MinMemory,
		currServiceConfig.User,
		currServiceConfig.Tolerations,
		currServiceConfig.NodeSelectors,
		currServiceConfig.Labels,
		currServiceConfig.TiniEnabled,
		currServiceConfig.PrivateIPAddressPlaceholder,
	)

	addServiceStarlarkStr := service_helpers.GetAddServiceStarlarkScript(serviceName, serviceConfigStr)
	logrus.Infof("Update Service Starlark:\n%v", addServiceStarlarkStr)

	logrus.Infof("Running update service starlark for service '%v' in enclave '%v'...", serviceName, enclaveIdentifier)
	starlarkRunResult, err := service_helpers.RunAddServiceStarlarkScript(ctx, serviceName, enclaveIdentifier, addServiceStarlarkStr, enclaveCtx)
	if err != nil {
		return err //already wrapped
	}

	out.PrintOutLn(string(starlarkRunResult.RunOutput))

	return nil
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
