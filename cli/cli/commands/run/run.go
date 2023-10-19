package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"gopkg.in/yaml.v2"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	command_args_run "github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave/inspect"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	scriptOrPackagePathKey                = "script-or-package-path"
	isScriptOrPackagePathArgumentOptional = false
	defaultScriptOrPackagePathArgument    = ""

	starlarkExtension = ".star"

	inputArgsArgKey                  = "args"
	inputArgsArgIsOptional           = true
	inputArgsAreNonGreedy            = false
	inputArgsAreEmptyBracesByDefault = "{}"

	dryRunFlagKey = "dry-run"
	defaultDryRun = "false"

	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	enclaveProductionModeFlagKey = "production"

	showEnclaveInspectFlagKey = "show-enclave-inspect"
	showEnclaveInspectDefault = "true"

	enclaveIdentifierFlagKey = "enclave"
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveIdentifierKeyword = ""

	verbosityFlagKey = "verbosity"
	defaultVerbosity = "brief"

	parallelismFlagKey = "parallelism"
	defaultParallelism = "4"

	experimentalFeaturesFlagKey   = "experimental"
	defaultExperimentalFeatures   = ""
	experimentalFeaturesSplitChar = ","

	githubDomainPrefix          = "github.com/"
	isNewEnclaveFlagWhenCreated = true
	interruptChanBufferSize     = 5

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	kurtosisYMLFilePath = "kurtosis.yml"

	portMappingSeparatorForLogs = ", "

	mainFileFlagKey      = "main-file"
	mainFileDefaultValue = ""

	mainFunctionNameFlagKey      = "main-function-name"
	mainFunctionNameDefaultValue = ""

	noConnectFlagKey = "no-connect"
	noConnectDefault = "false"

	packageArgsFileFlagKey      = "args-file"
	packageArgsFileDefaultValue = ""

	runFailed    = false
	runSucceeded = true

	imageDownloadFlagKey = "image-download"
	defaultImageDownload = "missing"
)

var StarlarkRunCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.StarlarkRunCmdStr,
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	ShortDescription:          "Run a Starlark script or package",
	LongDescription: "Run a Starlark script or runnable package (" + user_support_constants.StarlarkPackagesReferenceURL + ") in " +
		"an enclave. For a script we expect a path to a " + starlarkExtension + " file. For a runnable package we expect " +
		"either a path to a local runnable package directory, or a path to the " + kurtosisYMLFilePath + " file in the package, or the locator URL (" + user_support_constants.StarlarkLocatorsReferenceURL +
		") to a remote runnable package on Github. If the '" + enclaveIdentifierFlagKey + "' flag argument " +
		"is provided, Kurtosis will run the script inside the specified enclave or create it if it doesn't exist. If no '" +
		enclaveIdentifierFlagKey + "' flag param is provided, Kurtosis will create a new enclave with a random name.",
	Flags: []*flags.FlagConfig{
		{
			Key: dryRunFlagKey,
			// TODO(gb): link to a doc page mentioning what a "Kurtosis instruction" is
			Usage:   "If true, the Kurtosis instructions will not be executed, they will just be printed to the output of the CLI",
			Type:    flags.FlagType_Bool,
			Default: defaultDryRun,
		},
		{
			Key: enclaveIdentifierFlagKey,
			Usage: "The enclave identifier of the enclave in which the script or package will be ran. " +
				"An enclave with this name will be created if it doesn't exist.",
			Type:    flags.FlagType_String,
			Default: autogenerateEnclaveIdentifierKeyword,
		},
		{
			Key:     parallelismFlagKey,
			Usage:   "The parallelism level to be used in Starlark commands that supports it",
			Type:    flags.FlagType_Uint32,
			Default: defaultParallelism,
		},
		{
			Key:       verbosityFlagKey,
			Usage:     fmt.Sprintf("The verbosity of the command output: %s. If unset, it defaults to `brief` for a concise and explicit output. Use `detailed` to display the exhaustive list of arguments for each command. `executable` will generate executable Starlark instructions.", strings.Join(command_args_run.VerbosityStrings(), ", ")),
			Type:      flags.FlagType_String,
			Shorthand: "v",
			Default:   defaultVerbosity,
		},
		{
			Key:     showEnclaveInspectFlagKey,
			Usage:   "If true then Kurtosis runs enclave inspect immediately after running the Starlark Package. Default true",
			Type:    flags.FlagType_Bool,
			Default: showEnclaveInspectDefault,
		},
		{
			Key:     fullUuidsFlagKey,
			Usage:   "If true then Kurtosis prints full UUIDs instead of shortened UUIDs. Default false.",
			Type:    flags.FlagType_Bool,
			Default: fullUuidFlagKeyDefault,
		},
		{
			Key: mainFileFlagKey,
			Usage: "This is the relative (to the package root) main file filepath, the main file is the script file that will be executed first" +
				" and this should contains the main function. The default value is 'main.star'. This flag is only used for running packages",
			Type:    flags.FlagType_String,
			Default: mainFileDefaultValue,
		},
		{
			Key: mainFunctionNameFlagKey,
			Usage: "This is the name of the main function which will be executed first as the entrypoint of the package " +
				"or the module. The default value is 'run'.",
			Type:    flags.FlagType_String,
			Default: mainFunctionNameDefaultValue,
		},
		{
			Key: experimentalFeaturesFlagKey,
			Usage: fmt.Sprintf("Set of experimental features to enable for this Kurtosis package run. Enabling "+
				"multiple experimental features can be done by separating them with a comma, i.e. '--%s FEATURE_1,FEATURE_2'. "+
				"Please reach out to Kurtosis team if you wish to try any of those.", experimentalFeaturesFlagKey),
			Type:    flags.FlagType_String,
			Default: defaultExperimentalFeatures,
		},
		{
			Key:       enclaveProductionModeFlagKey,
			Usage:     "If enabled, services will restart if they fail",
			Shorthand: "p",
			Type:      flags.FlagType_Bool,
			Default:   "false",
		},
		{
			Key:     noConnectFlagKey,
			Usage:   "If true then user service ports are not forwarded locally. Default false",
			Type:    flags.FlagType_Bool,
			Default: noConnectDefault,
		},
		{
			Key:     packageArgsFileFlagKey,
			Usage:   "The file (JSON/YAML) will be used as arguments passed to the Kurtosis Package",
			Type:    flags.FlagType_String,
			Default: packageArgsFileDefaultValue,
		},
		{
			Key:     imageDownloadFlagKey,
			Usage:   "If unset, it defaults to `missing` for fetching the latest image only if not available in local cache. Use `never` to only use local cached image (never fetch new images) and `always` to always fetch the latest image.",
			Type:    flags.FlagType_String,
			Default: defaultImageDownload,
		},
	},
	Args: []*args.ArgConfig{
		// TODO add a `Usage` description here when ArgConfig supports it
		file_system_path_arg.NewFilepathOrDirpathArg(
			scriptOrPackagePathKey,
			isScriptOrPackagePathArgumentOptional,
			defaultScriptOrPackagePathArgument,
			scriptPathValidation,
		),
		{
			Key:            inputArgsArgKey,
			DefaultValue:   inputArgsAreEmptyBracesByDefault,
			IsOptional:     inputArgsArgIsOptional,
			IsGreedy:       inputArgsAreNonGreedy,
			ValidationFunc: validatePackageArgs,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	metricsClient metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	// Args parsing and validation
	packageArgs, err := args.GetNonGreedyArg(inputArgsArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the script/package arguments using flag key '%v'", inputArgsArgKey)
	}

	userRequestedEnclaveIdentifier, err := flags.GetString(enclaveIdentifierFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier using flag key '%s'", enclaveIdentifierFlagKey)
	}

	starlarkScriptOrPackagePath, err := args.GetNonGreedyArg(scriptOrPackagePathKey)
	if err != nil {
		return stacktrace.Propagate(err, "Error reading the Starlark script or package directory at '%s'. Does it exist?", starlarkScriptOrPackagePath)
	}

	dryRun, err := flags.GetBool(dryRunFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", dryRunFlagKey)
	}

	parallelism, err := flags.GetUint32(parallelismFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a integer flag with key '%v' but none was found; this is an error in Kurtosis!", parallelismFlagKey)
	}
	castedParallelism := int32(parallelism)

	verbosity, err := parseVerbosityFlag(flags)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the verbosity using flag key '%s'", verbosityFlagKey)
	}

	imageDownload, err := parseImageDownloadFlag(flags)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the image-download using flag key '%s'", imageDownloadFlagKey)
	}

	showFullUuids, err := flags.GetBool(fullUuidsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidsFlagKey)
	}

	showEnclaveInspect, err := flags.GetBool(showEnclaveInspectFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", showEnclaveInspectFlagKey)
	}

	noConnect, err := flags.GetBool(noConnectFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", noConnectDefault)
	}

	relativePathToTheMainFile, err := flags.GetString(mainFileFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", mainFileFlagKey)
	}

	mainFunctionName, err := flags.GetString(mainFunctionNameFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", mainFunctionNameFlagKey)
	}

	isProduction, err := flags.GetBool(enclaveProductionModeFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", enclaveProductionModeFlagKey)
	}

	experimentalFlags, err := parseExperimentalFlag(flags)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", mainFunctionNameFlagKey)
	}

	packageArgsFile, err := flags.GetString(packageArgsFileFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", packageArgsFileFlagKey)
	}

	if packageArgs == inputArgsAreEmptyBracesByDefault && packageArgsFile != packageArgsFileDefaultValue {
		logrus.Debugf("'%v' is empty but '%v' is provided so we will go with the '%v' value", inputArgsArgKey, packageArgsFileFlagKey, packageArgsFileFlagKey)
		packageArgsFileBytes, err := os.ReadFile(packageArgsFile)
		if err != nil {
			return stacktrace.Propagate(err, "attempted to read file provided by flag '%v' with path '%v' but failed", packageArgsFileFlagKey, packageArgsFile)
		}
		packageArgsFileStr := string(packageArgsFileBytes)
		if packageArgParsingErr := validateSerializedArgs(packageArgsFileStr); packageArgParsingErr != nil {
			return stacktrace.Propagate(err, "attempted to validate '%v' but failed", packageArgsFileFlagKey)
		}
		packageArgs = packageArgsFileStr
	} else if packageArgs != inputArgsAreEmptyBracesByDefault && packageArgsFile != packageArgsFileDefaultValue {
		logrus.Debugf("'%v' arg is not empty; ignoring value of '%v' flag as '%v' arg takes precedence", inputArgsArgKey, packageArgsFileFlagKey, inputArgsArgKey)
	}

	cloudUserId := ""
	cloudInstanceId := ""
	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err != nil {
		logrus.Warnf("Could not retrieve the current context. Kurtosis will assume context is local (no cloud user & instance id) and not" +
			"map the enclave service ports. If you're running on a remote context and are seeing this error, then" +
			"the enclave services will be unreachable locally. Turn on debug logging to see the actual error.")
		logrus.Debugf("Error was: %v", err.Error())
	} else {
		if store.IsRemote(currentContext) {
			cloudUserId = currentContext.GetRemoteContextV0().GetCloudUserId()
			cloudInstanceId = currentContext.GetRemoteContextV0().GetCloudInstanceId()
		}
	}

	starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig(
		starlark_run_config.WithDryRun(dryRun),
		starlark_run_config.WithParallelism(castedParallelism),
		starlark_run_config.WithExperimentalFeatureFlags(experimentalFlags),
		starlark_run_config.WithMainFunctionName(mainFunctionName),
		starlark_run_config.WithRelativePathToMainFile(relativePathToTheMainFile),
		starlark_run_config.WithSerializedParams(packageArgs),
		starlark_run_config.WithCloudUserId(cloudUserId),
		starlark_run_config.WithCloudInstanceId(cloudInstanceId),
		starlark_run_config.WithImageDownloadMode(imageDownload),
	)

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, isNewEnclave, err := getOrCreateEnclaveContext(ctx, userRequestedEnclaveIdentifier, kurtosisCtx, metricsClient, isProduction)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", userRequestedEnclaveIdentifier)
	}

	if showEnclaveInspect {
		defer func() {
			if err = inspect.PrintEnclaveInspect(ctx, kurtosisBackend, kurtosisCtx, enclaveCtx.GetEnclaveName(), showFullUuids); err != nil {
				logrus.Errorf("An error occurred while printing enclave status and contents:\n%s", err)
			}
		}()
	}

	if isNewEnclave {
		defer output_printers.PrintEnclaveName(enclaveCtx.GetEnclaveName())
	}

	var responseLineChan <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine
	var cancelFunc context.CancelFunc
	var errRunningKurtosis error

	connect := kurtosis_core_rpc_api_bindings.Connect_CONNECT
	if noConnect {
		connect = kurtosis_core_rpc_api_bindings.Connect_NO_CONNECT
	}

	isRemotePackage := strings.HasPrefix(starlarkScriptOrPackagePath, githubDomainPrefix)
	if isRemotePackage {
		responseLineChan, cancelFunc, errRunningKurtosis = executeRemotePackage(ctx, enclaveCtx, starlarkScriptOrPackagePath, starlarkRunConfig)
	} else {
		fileOrDir, err := os.Stat(starlarkScriptOrPackagePath)
		if err != nil {
			return stacktrace.Propagate(err, "There was an error reading file or package from disk at '%v'", starlarkScriptOrPackagePath)
		}

		if isStandaloneScript(fileOrDir, kurtosisYMLFilePath) {
			if !strings.HasSuffix(starlarkScriptOrPackagePath, starlarkExtension) {
				return stacktrace.NewError("Expected a script with a '%s' extension but got file '%v' with a different extension", starlarkExtension, starlarkScriptOrPackagePath)
			}
			responseLineChan, cancelFunc, errRunningKurtosis = executeScript(ctx, enclaveCtx, starlarkScriptOrPackagePath, starlarkRunConfig)
		} else {
			// if the path is a file with `kurtosis.yml` at the end it's a module dir
			// we remove the `kurtosis.yml` to get just the Dir containing the module
			if isKurtosisYMLFileInPackageDir(fileOrDir, kurtosisYMLFilePath) {
				starlarkScriptOrPackagePath = path.Dir(starlarkScriptOrPackagePath)
			}
			// we pass the sanitized path and look for a Kurtosis YML within it to get the package name
			if err != nil {
				return stacktrace.Propagate(err, "Tried parsing Kurtosis YML at '%v' to get package name but failed", starlarkScriptOrPackagePath)
			}
			responseLineChan, cancelFunc, errRunningKurtosis = executePackage(ctx, enclaveCtx, starlarkScriptOrPackagePath, starlarkRunConfig)
		}
	}
	if errRunningKurtosis != nil {
		return stacktrace.Propagate(errRunningKurtosis, "An error starting the Kurtosis code execution '%v'", starlarkScriptOrPackagePath)
	}

	errRunningKurtosis = ReadAndPrintResponseLinesUntilClosed(responseLineChan, cancelFunc, verbosity, dryRun)
	var runStatusForMetrics bool
	if errRunningKurtosis != nil {
		runStatusForMetrics = runFailed
	} else {
		runStatusForMetrics = runSucceeded
	}

	if err = enclaveCtx.ConnectServices(ctx, connect); err != nil {
		logrus.Warnf("An error occurred configuring the user services port forwarding\nError was: %v", err)
	}

	servicesInEnclavePostRun, servicesInEnclaveError := enclaveCtx.GetServices()
	if servicesInEnclaveError != nil {
		logrus.Warn("Tried getting number of services in the enclave to log metrics but failed")
	} else {
		// TODO(gyani-cloud-metrics) move this to APIC
		if err = metricsClient.TrackKurtosisRunFinishedEvent(starlarkScriptOrPackagePath, len(servicesInEnclavePostRun), runStatusForMetrics, cloudInstanceId, cloudUserId); err != nil {
			logrus.Warn("An error occurred tracking kurtosis run finished event")
		}
	}

	if errRunningKurtosis != nil {
		return errRunningKurtosis
	}

	if servicesInEnclaveError != nil {
		logrus.Warnf("Unable to retrieve the services running inside the enclave so their ports will not be" +
			" mapped to local ports.")
		return nil
	}

	if currentContext != nil {
		return nil
	}

	if !store.IsRemote(currentContext) {
		logrus.Debugf("Current context is local, not mapping enclave service ports")
		return nil
	}

	if connect == kurtosis_core_rpc_api_bindings.Connect_NO_CONNECT {
		logrus.Info("Not forwarding user service ports locally as requested")
		return nil
	}

	// Context is remote. All enclave service ports will be mapped locally
	portalManager := portal_manager.NewPortalManager()
	portsMapping := map[uint16]*services.PortSpec{}
	for serviceInEnclaveName, servicesInEnclaveUuid := range servicesInEnclavePostRun {
		serviceCtx, err := enclaveCtx.GetServiceContext(string(servicesInEnclaveUuid))
		if err != nil {
			return stacktrace.Propagate(err, "Error getting service object for service '%s' with UUID '%s'", serviceInEnclaveName, servicesInEnclaveUuid)
		}
		for _, portSpec := range serviceCtx.GetPublicPorts() {
			portsMapping[portSpec.GetNumber()] = portSpec
		}
	}
	successfullyMappedPorts, failedPorts, err := portalManager.MapPorts(ctx, portsMapping)
	if err != nil {
		var stringifiedPortMapping []string
		for localPort, remotePort := range failedPorts {
			stringifiedPortMapping = append(stringifiedPortMapping, fmt.Sprintf("%d:%d", localPort, remotePort.GetNumber()))
		}
		// TODO: once we have a manual `kurtosis port map` command, suggest using it here to manually map the failed port
		logrus.Warnf("The enclave was successfully run but the following port(s) could not be mapped locally: %s. "+
			"The associated service(s) will not be reachable on the local host",
			strings.Join(stringifiedPortMapping, portMappingSeparatorForLogs))
		logrus.Debugf("Error was: %v", err.Error())
		return nil
	}
	logrus.Infof("Successfully mapped %d ports. All services running inside the enclave are reachable locally on"+
		" their ephemeral port numbers", len(successfullyMappedPorts))
	return nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func executeScript(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, scriptPath string, runConfig *starlark_run_config.StarlarkRunConfig) (<-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	fileContentBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unable to read content of Starlark script file '%s'", scriptPath)
	}
	return enclaveCtx.RunStarlarkScript(ctx, string(fileContentBytes), runConfig)
}

func executePackage(
	ctx context.Context,
	enclaveCtx *enclaves.EnclaveContext,
	packagePath string,
	runConfig *starlark_run_config.StarlarkRunConfig,
) (<-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	// we get the absolute path so that the logs make more sense
	absolutePackagePath, err := filepath.Abs(packagePath)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while getting the absolute path for '%v'", packagePath)
	}
	logrus.Infof("Executing Starlark package at '%v' as the passed argument '%v' looks like a directory", absolutePackagePath, packagePath)

	return enclaveCtx.RunStarlarkPackage(ctx, packagePath, runConfig)
}

func executeRemotePackage(
	ctx context.Context,
	enclaveCtx *enclaves.EnclaveContext,
	packageId string,
	runConfig *starlark_run_config.StarlarkRunConfig,
) (<-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	return enclaveCtx.RunStarlarkRemotePackage(ctx, packageId, runConfig)
}

// ReadAndPrintResponseLinesUntilClosed TODO(victor.colombo): Extract this to somewhere reasonable
func ReadAndPrintResponseLinesUntilClosed(responseLineChan <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, cancelFunc context.CancelFunc, verbosity command_args_run.Verbosity, dryRun bool) error {
	defer cancelFunc()

	// This channel will receive a signal when the user presses an interrupt
	interruptChan := make(chan os.Signal, interruptChanBufferSize)
	signal.Notify(interruptChan, os.Interrupt)
	defer close(interruptChan)

	printer := output_printers.NewExecutionPrinter()
	if err := printer.Start(); err != nil {
		return stacktrace.Propagate(err, "Unable to start the printer for this execution. The execution will continue in the background but nothing will be printed.")
	}
	defer printer.Stop()

	isRunSuccessful := false // defaults to false such that we fail loudly if something unexpected happens
	for {
		select {
		case responseLine, isChanOpen := <-responseLineChan:
			if !isChanOpen {
				if !isRunSuccessful && !dryRun {
					// This error thrown by the APIC is not informative right now as it just tells the user to look at errors
					// in the above log. For this reason we're ignoring it and returning nil. This is exceptional to not clutter
					// the CLI output. We should still use stacktrace.Propagate for other errors.
					return stacktrace.Propagate(command_str_consts.ErrorMessageDueToStarlarkFailure, "Error occurred while running kurtosis package")
				}
				return nil
			}
			err := printer.PrintKurtosisExecutionResponseLineToStdOut(responseLine, verbosity, dryRun)
			if err != nil {
				logrus.Errorf("An error occurred trying to write the output of Starlark execution to stdout. The script execution will continue, but the output printed here is incomplete. Error was: \n%s", err.Error())
			}
			// If the run finished, persist its status to the isRunSuccessful bool to throw an error and return a non-zero status code
			if responseLine.GetRunFinishedEvent() != nil {
				isRunSuccessful = responseLine.GetRunFinishedEvent().GetIsRunSuccessful()
			}
		case <-interruptChan:
			return stacktrace.NewError("User manually interrupted the execution, returning. Note that the execution will continue in the Kurtosis enclave")
		}
	}
}

func getOrCreateEnclaveContext(
	ctx context.Context,
	enclaveIdentifierOrName string,
	kurtosisContext *kurtosis_context.KurtosisContext,
	metricsClient metrics_client.MetricsClient,
	isProduction bool,
) (*enclaves.EnclaveContext, bool, error) {

	if enclaveIdentifierOrName != autogenerateEnclaveIdentifierKeyword {
		_, err := kurtosisContext.GetEnclave(ctx, enclaveIdentifierOrName)
		if err == nil {
			enclaveContext, err := kurtosisContext.GetEnclaveContext(ctx, enclaveIdentifierOrName)
			if err != nil {
				return nil, false, stacktrace.Propagate(err, "An error occurred while getting context for existing enclave with identifier '%v'", enclaveIdentifierOrName)
			}
			return enclaveContext, false, nil
		}
	}
	logrus.Infof("Creating a new enclave for Starlark to run inside...")
	var enclaveContext *enclaves.EnclaveContext
	var err error
	if isProduction {
		enclaveContext, err = kurtosisContext.CreateProductionEnclave(ctx, enclaveIdentifierOrName)
	} else {
		enclaveContext, err = kurtosisContext.CreateEnclave(ctx, enclaveIdentifierOrName)
	}
	if err != nil {
		return nil, false, stacktrace.Propagate(err, fmt.Sprintf("Unable to create new enclave with name '%s'", enclaveIdentifierOrName))
	}
	subnetworkDisableBecauseItIsDeprecated := false
	if err = metricsClient.TrackCreateEnclave(enclaveIdentifierOrName, subnetworkDisableBecauseItIsDeprecated); err != nil {
		logrus.Error("An error occurred while logging the create enclave event")
	}
	logrus.Infof("Enclave '%v' created successfully", enclaveContext.GetEnclaveName())
	return enclaveContext, isNewEnclaveFlagWhenCreated, nil
}

// validatePackageArgs just validates the args is a valid JSON or YAML string
func validatePackageArgs(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	serializedArgs, err := args.GetNonGreedyArg(inputArgsArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the script/package arguments using flag key '%v'", inputArgsArgKey)
	}
	return validateSerializedArgs(serializedArgs)
}

// parseVerbosityFlag Get the verbosity flag is present, and parse it to a valid Verbosity value
func parseVerbosityFlag(flags *flags.ParsedFlags) (command_args_run.Verbosity, error) {
	verbosityStr, err := flags.GetString(verbosityFlagKey)
	if err != nil {
		return 0, stacktrace.Propagate(err, "An error occurred getting the verbosity using flag key '%s'", verbosityFlagKey)
	}
	verbosity, err := command_args_run.VerbosityString(verbosityStr)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Invalid verbosity value: '%s'. Possible values are %s", verbosityStr, strings.Join(command_args_run.VerbosityStrings(), ", "))
	}
	return verbosity, nil
}

// Get the image download flag is present, and parse it to a valid ImageDownload value
func parseImageDownloadFlag(flags *flags.ParsedFlags) (kurtosis_core_rpc_api_bindings.ImageDownloadMode, error) {

	defaultImageMode := kurtosis_core_rpc_api_bindings.ImageDownloadMode_missing

	imageDownloadStr, err := flags.GetString(imageDownloadFlagKey)
	if err != nil {
		return defaultImageMode, stacktrace.Propagate(err, "An error occurred getting the image-download using flag key '%s'", imageDownloadFlagKey)
	}
	imageDownload, err := command_args_run.ImageDownloadString(imageDownloadStr)
	if err != nil {
		return defaultImageMode, stacktrace.Propagate(err, "Invalid image-download value: '%s'. Possible values are %s", imageDownloadStr, strings.Join(command_args_run.ImageDownloadStrings(), ", "))
	}
	imageDownloadRPC := kurtosis_core_rpc_api_bindings.ImageDownloadMode(kurtosis_core_rpc_api_bindings.ImageDownloadMode_value[strings.ToLower(imageDownload.String())])
	return imageDownloadRPC, nil
}

// parseExperimentalFlag parses the sert of experimental features enabled for this run
func parseExperimentalFlag(flags *flags.ParsedFlags) ([]kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag, error) {
	experimentalFeaturesStr, err := flags.GetString(experimentalFeaturesFlagKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the experimental features flag using flag key '%s'", experimentalFeaturesFlagKey)
	}

	var parsedExperimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag
	experimentalFeatures := strings.Split(experimentalFeaturesStr, experimentalFeaturesSplitChar)
	for _, experimentalFeatureStr := range experimentalFeatures {
		trimmedExperimentalFeatureStr := strings.TrimSpace(experimentalFeatureStr)
		if trimmedExperimentalFeatureStr == "" {
			continue
		}
		experimentalFeatureInt, found := kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag_value[trimmedExperimentalFeatureStr]
		if !found {
			return nil, stacktrace.NewError(
				"Unable to parse '%s' as a valid Kurtosis feature flag.",
				trimmedExperimentalFeatureStr)
		}
		parsedExperimentalFeatures = append(parsedExperimentalFeatures, kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag(experimentalFeatureInt))
	}
	return parsedExperimentalFeatures, nil
}

// isStandaloneScript returns true if the fileInfo points to a non `kurtosis.yml` regular file
func isStandaloneScript(fileInfo os.FileInfo, kurtosisYMLFilePath string) bool {
	return fileInfo.Mode().IsRegular() && fileInfo.Name() != kurtosisYMLFilePath
}

func isKurtosisYMLFileInPackageDir(fileInfo os.FileInfo, kurtosisYMLFilePath string) bool {
	return fileInfo.Mode().IsRegular() && fileInfo.Name() == kurtosisYMLFilePath
}

func scriptPathValidation(scriptPath string) (error, bool) {
	// if it's a Github path we don't validate further, the APIC will do it for us
	if strings.HasPrefix(scriptPath, githubDomainPrefix) {
		return nil, file_system_path_arg.DoNotContinueWithDefaultValidation
	}

	return nil, file_system_path_arg.ContinueWithDefaultValidation
}

func validateSerializedArgs(serializedArgs string) error {
	var result interface{}
	var jsonError error
	if jsonError = json.Unmarshal([]byte(serializedArgs), &result); jsonError == nil {
		return nil
	}
	var yamlError error
	if yamlError = yaml.Unmarshal([]byte(serializedArgs), &result); yamlError == nil {
		return nil
	}
	return stacktrace.Propagate(
		fmt.Errorf("JSON parsing error '%v', YAML parsing error '%v'", jsonError, yamlError),
		"Error validating args, because it is not a valid JSON or YAML.")
}
