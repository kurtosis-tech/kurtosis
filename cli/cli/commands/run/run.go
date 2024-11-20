package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/backend_creator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/configs"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/user_support_constants"
	"gopkg.in/yaml.v2"
	"io"
	"k8s.io/utils/strings/slices"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"

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
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
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

	dependenciesFlagKey         = "dependencies"
	dependenciesFlagDefault     = "false"
	pullDependenciesFlagKey     = "pull"
	pullDependenciesFlagDefault = "false"

	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	enclaveProductionModeFlagKey = "production"

	showEnclaveInspectFlagKey = "show-enclave-inspect"
	showEnclaveInspectDefault = "true"

	enclaveIdentifierFlagKey = "enclave"
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveIdentifierKeyword = ""

	verbosityFlagKey = "verbosity"
	defaultVerbosity = "description"

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

	kurtosisYMLFilePath  = "kurtosis.yml"
	kurtosisYMLFilePerms = 0644

	portMappingSeparatorForLogs = ", "

	mainFileFlagKey      = "main-file"
	mainFileDefaultValue = ""

	mainFunctionNameFlagKey      = "main-function-name"
	mainFunctionNameDefaultValue = ""

	noConnectFlagKey = "no-connect"
	noConnectDefault = "false"

	packageArgsFileFlagKey      = "args-file"
	packageArgsFileDefaultValue = ""

	imageDownloadFlagKey = "image-download"
	defaultImageDownload = "missing"

	nonBlockingModeFlagKey = "non-blocking-tasks"
	defaultBlockingMode    = "false"

	httpProtocolRegexStr           = "^(http|https)://"
	shouldCloneNormalRepo          = false
	packageReplaceKeyInKurtosisYml = "replace:"
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
			Key:     dependenciesFlagKey,
			Usage:   "If true, a yaml will be output (to stdout) with a list of images and packages that this run depends on.",
			Type:    flags.FlagType_Bool,
			Default: dependenciesFlagDefault,
		},
		{
			Key:     pullDependenciesFlagKey,
			Usage:   fmt.Sprintf("If true, and the %s flag is passed, attempts to pull all images and packages that the run depends on locally. %s is updated with replace directives pointing to locally pulled packages. If a replace directive already exists an error is thrown. Note: this currently only works on the Docker backend.", dependenciesFlagKey, kurtosisYMLFilePath),
			Type:    flags.FlagType_Bool,
			Default: pullDependenciesFlagDefault,
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
			Usage:     fmt.Sprintf("The verbosity of the command output: %s. If unset, it defaults to `description` for a crisp output that explains whats about to happen. Use `brief` for a concise yet explicit ouptut, to see the entire instruction thats about to execute.  Use `detailed` to display the exhaustive list of arguments for each instruction. `executable` will generate executable Starlark instructions.", strings.Join(command_args_run.VerbosityStrings(), ", ")),
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
			Usage:   "The file (JSON/YAML) that will be used to pass in arguments to the Kurtosis package. Can be a URL or file path.",
			Type:    flags.FlagType_String,
			Default: packageArgsFileDefaultValue,
		},
		{
			Key:     imageDownloadFlagKey,
			Usage:   "If unset, it defaults to `missing` which will only download the latest image tag if the image does not already exist locally (irrespective of the tag of the locally cached image). Use `always` to have Kurtosis always check and download the latest image tag, even if the image exists locally.",
			Type:    flags.FlagType_String,
			Default: defaultImageDownload,
		},
		{
			Key:     nonBlockingModeFlagKey,
			Usage:   "If set, Kurtosis will not block on removing services from tasks from run_sh and run_python instructions. These services will remain and must be manually cleaned up.",
			Type:    flags.FlagType_Bool,
			Default: defaultBlockingMode,
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
	_ metrics_client.MetricsClient,
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

	isDependenciesOnly, err := flags.GetBool(dependenciesFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", dependenciesFlagKey)
	}

	pullDependencies, err := flags.GetBool(pullDependenciesFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", pullDependenciesFlagKey)
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

	nonBlockingMode, err := flags.GetBool(nonBlockingModeFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", nonBlockingModeFlagKey)
	}

	if packageArgs == inputArgsAreEmptyBracesByDefault && packageArgsFile != packageArgsFileDefaultValue {
		logrus.Debugf("'%v' is empty but '%v' is provided so we will go with the '%v' value", inputArgsArgKey, packageArgsFileFlagKey, packageArgsFileFlagKey)
		packageArgs, err = getArgsFromFilepathOrURL(packageArgsFile)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting the package args from filepath or URL '%s'", packageArgsFile)
		}
	} else if packageArgs != inputArgsAreEmptyBracesByDefault && packageArgsFile != packageArgsFileDefaultValue {
		logrus.Debugf("'%v' arg is not empty; ignoring value of '%v' flag as '%v' arg takes precedence", inputArgsArgKey, packageArgsFileFlagKey, inputArgsArgKey)
	}

	starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig(
		starlark_run_config.WithDryRun(dryRun),
		starlark_run_config.WithParallelism(castedParallelism),
		starlark_run_config.WithExperimentalFeatureFlags(experimentalFlags),
		starlark_run_config.WithMainFunctionName(mainFunctionName),
		starlark_run_config.WithRelativePathToMainFile(relativePathToTheMainFile),
		starlark_run_config.WithSerializedParams(packageArgs),
		starlark_run_config.WithImageDownloadMode(*imageDownload),
		starlark_run_config.WithNonBlockingMode(nonBlockingMode),
	)

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, isNewEnclave, err := getOrCreateEnclaveContext(ctx, userRequestedEnclaveIdentifier, kurtosisCtx, isProduction)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", userRequestedEnclaveIdentifier)
	}

	if showEnclaveInspect {
		defer func() {
			if err = inspect.PrintEnclaveInspect(ctx, kurtosisCtx, enclaveCtx.GetEnclaveName(), showFullUuids); err != nil {
				logrus.Errorf("An error occurred while printing enclave status and contents:\n%s", err)
			}
		}()
	}

	if isNewEnclave {
		defer output_printers.PrintEnclaveName(enclaveCtx.GetEnclaveName())
	}

	isRemotePackage := strings.HasPrefix(starlarkScriptOrPackagePath, githubDomainPrefix)

	if isDependenciesOnly {
		dependencyYaml, err := getPackageDependencyYaml(ctx, enclaveCtx, starlarkScriptOrPackagePath, isRemotePackage, packageArgs)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting package dependencies.")
		}

		type PackageDependencies struct {
			Images   []string `yaml:"images"`
			Packages []string `yaml:"packageDependencies"`
		}
		var pkgDeps PackageDependencies
		err = yaml.Unmarshal([]byte(dependencyYaml.PlanYaml), &pkgDeps)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred unmarshalling dependency yaml string")
		}
		out.PrintOutLn("Images:")
		for _, imageStr := range pkgDeps.Images {
			out.PrintOutLn(fmt.Sprintf(" %s", imageStr))
		}
		out.PrintOutLn("Packages:")
		for _, packageStr := range pkgDeps.Packages {
			out.PrintOutLn(fmt.Sprintf(" %s", packageStr))
		}

		if pullDependencies {
			// errors below already wrapped w propagate
			err = pullImagesLocally(ctx, pkgDeps.Images)
			if err != nil {
				return err
			}

			packageNamesToLocalFilepaths, err := pullPackagesLocally(pkgDeps.Packages)
			if err != nil {
				return err
			}

			err = updateKurtosisYamlWithReplaceDirectives(packageNamesToLocalFilepaths)
			if err != nil {
				return err
			}
		}
		return nil
	}

	var responseLineChan <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine
	var cancelFunc context.CancelFunc
	var errRunningKurtosis error

	connect := kurtosis_core_rpc_api_bindings.Connect_CONNECT
	if noConnect {
		connect = kurtosis_core_rpc_api_bindings.Connect_NO_CONNECT
	}

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

	if err = enclaveCtx.ConnectServices(ctx, connect); err != nil {
		logrus.Warnf("An error occurred configuring the user services port forwarding\nError was: %v", err)
	}

	servicesInEnclavePostRun, servicesInEnclaveError := enclaveCtx.GetServices()
	if servicesInEnclaveError != nil {
		logrus.Warn("Tried getting number of services in the enclave to log metrics but failed")
	}

	if errRunningKurtosis != nil {
		return errRunningKurtosis
	}

	if servicesInEnclaveError != nil {
		logrus.Warnf("Unable to retrieve the services running inside the enclave so their ports will not be" +
			" mapped to local ports.")
		return nil
	}

	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err != nil {
		logrus.Warnf("Could not retrieve the current context. Kurtosis will assume context is local (no cloud user & instance id) and not" +
			"map the enclave service ports. If you're running on a remote context and are seeing this error, then" +
			"the enclave services will be unreachable locally. Turn on debug logging to see the actual error.")
		logrus.Debugf("Error was: %v", err.Error())
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
	logrus.Infof("Enclave '%v' created successfully", enclaveContext.GetEnclaveName())
	return enclaveContext, isNewEnclaveFlagWhenCreated, nil
}

func getPackageDependencyYaml(
	ctx context.Context,
	enclaveCtx *enclaves.EnclaveContext,
	starlarkScriptOrPackageId string,
	isRemote bool,
	packageArgs string,
) (*kurtosis_core_rpc_api_bindings.PlanYaml, error) {
	var packageYaml *kurtosis_core_rpc_api_bindings.PlanYaml
	var err error
	if isRemote {
		packageYaml, err = enclaveCtx.GetStarlarkPackagePlanYaml(ctx, starlarkScriptOrPackageId, packageArgs)
	} else {
		fileOrDir, err := os.Stat(starlarkScriptOrPackageId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "There was an error reading file or package from disk at '%v'", starlarkScriptOrPackageId)
		}

		if isStandaloneScript(fileOrDir, kurtosisYMLFilePath) {
			scriptContentBytes, err := os.ReadFile(starlarkScriptOrPackageId)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Unable to read content of Starlark script file '%s'", starlarkScriptOrPackageId)
			}
			packageYaml, err = enclaveCtx.GetStarlarkScriptPlanYaml(ctx, string(scriptContentBytes), packageArgs)
		}
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving plan yaml for provided package.")
	}
	return packageYaml, nil
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
func parseImageDownloadFlag(flags *flags.ParsedFlags) (*kurtosis_core_rpc_api_bindings.ImageDownloadMode, error) {
	imageDownloadStr, err := flags.GetString(imageDownloadFlagKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the image-download using flag key '%s'", imageDownloadFlagKey)
	}

	imageDownloadStrNorm := strings.ToLower(imageDownloadStr)
	keys := make([]string, 0, len(kurtosis_core_rpc_api_bindings.ImageDownloadMode_value))
	for k := range kurtosis_core_rpc_api_bindings.ImageDownloadMode_value {
		keys = append(keys, k)
	}

	if !slices.Contains(keys, imageDownloadStrNorm) {
		return nil, stacktrace.NewError("Invalid image-download value: '%s'. Possible values are: '%s'", imageDownloadStr, strings.Join(keys, "', '"))
	}

	imageDownloadCode := kurtosis_core_rpc_api_bindings.ImageDownloadMode_value[imageDownloadStrNorm]
	imageDownloadRPC := kurtosis_core_rpc_api_bindings.ImageDownloadMode(imageDownloadCode)
	return &imageDownloadRPC, nil
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

func getArgsFromFilepathOrURL(packageArgsFile string) (string, error) {
	var packageArgsFileBytes []byte

	if isHttpUrl(packageArgsFile) {
		argsFileURL, parseErr := url.Parse(packageArgsFile)
		if parseErr != nil {
			return "", stacktrace.Propagate(parseErr, "An error occurred while parsing file args URL '%s'", argsFileURL)
		}
		response, getErr := http.Get(argsFileURL.String())
		if getErr != nil {
			return "", stacktrace.Propagate(getErr, "An error occurred getting the args file content from URL '%s'", argsFileURL.String())
		}
		defer response.Body.Close()
		responseBodyBytes, readAllErr := io.ReadAll(response.Body)
		if readAllErr != nil {
			return "", stacktrace.Propagate(readAllErr, "An error occurred reading the args file content")
		}
		packageArgsFileBytes = responseBodyBytes
	} else {
		_, err := os.Stat(packageArgsFile)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred checking for argument's file existence on '%s'", packageArgsFile)
		}

		packageArgsFileBytes, err = os.ReadFile(packageArgsFile)
		if err != nil {
			return "", stacktrace.Propagate(err, "attempted to read file provided by flag '%v' with path '%v' but failed", packageArgsFileFlagKey, packageArgsFile)
		}
	}

	packageArgsFileStr := string(packageArgsFileBytes)
	if packageArgParsingErr := validateSerializedArgs(packageArgsFileStr); packageArgParsingErr != nil {
		return "", stacktrace.Propagate(packageArgParsingErr, "attempted to validate '%v' but failed", packageArgsFileFlagKey)
	}

	return packageArgsFileStr, nil

}

func isHttpUrl(maybeHttpUrl string) bool {
	httpProtocolRegex := regexp.MustCompile(httpProtocolRegexStr)
	return httpProtocolRegex.MatchString(maybeHttpUrl)
}

func pullImagesLocally(ctx context.Context, images []string) error {
	kurtosisBackend, err := backend_creator.GetDockerKurtosisBackend(backend_creator.NoAPIContainerModeArgs, configs.NoRemoteBackendConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving Docker Kurtosis Backend")
	}
	for _, img := range images {
		_, _, err := kurtosisBackend.FetchImage(ctx, img, nil, image_download_mode.ImageDownloadMode_Always)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred pulling '%v' locally.", img)
		}
	}
	return nil
}

func pullPackagesLocally(packageDependencies []string) (map[string]string, error) {
	localPackagesToRelativeFilepaths := map[string]string{}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting current working directory.")
	}
	// ensure a kurtosis yml exists here so that packages get cloned to the right place (one dir above/nested in the same dir as the package)
	file, err := os.OpenFile(kurtosisYMLFilePath, os.O_WRONLY|os.O_APPEND, kurtosisYMLFilePerms)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred opening '%s' file. Make sure this command is being run from within the directory of a kurtosis package.", kurtosisYMLFilePath)
	}
	defer file.Close()

	parentCwd := filepath.Dir(workingDirectory)
	relParentCwd, err := filepath.Rel(workingDirectory, parentCwd)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting rel path between '%v' and '%v'.", workingDirectory, relParentCwd)
	}

	if _, err := os.Stat(fmt.Sprintf("%s/%s", parentCwd, kurtosisYMLFilePath)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, stacktrace.Propagate(err, "%s does not exist in current working directory. Make sure you are running this at the root of the package with the %s.", kurtosisYMLFilePath, kurtosisYMLFilePath)
		} else {
			return nil, stacktrace.Propagate(err, "An error occurred checking if %s exists.", kurtosisYMLFilePath)
		}
	}
	for _, dependency := range packageDependencies {
		packageIdParts := strings.Split(dependency, "/")
		packageName := packageIdParts[len(packageIdParts)-1]
		logrus.Infof("Pulling package: %v", dependency)

		var repoUrl string
		if !strings.HasPrefix("http://", dependency) {
			repoUrl = "http://" + dependency
		}
		if !strings.HasSuffix(".git", dependency) {
			repoUrl += ".git"
		}
		localPackagePath := fmt.Sprintf("%s/%s", parentCwd, packageName)
		// find the kurtosis.yml
		_, err := git.PlainClone(localPackagePath, shouldCloneNormalRepo, &git.CloneOptions{
			URL:               repoUrl,
			Auth:              nil,
			RemoteName:        "",
			ReferenceName:     "",
			SingleBranch:      false,
			Mirror:            false,
			NoCheckout:        false,
			Depth:             0,
			RecurseSubmodules: 0,
			ShallowSubmodules: false,
			Progress:          nil,
			Tags:              0,
			InsecureSkipTLS:   false,
			CABundle:          nil,
			ProxyOptions: transport.ProxyOptions{
				URL:      "",
				Username: "",
				Password: "",
			},
			Shared: false,
		})
		if err != nil && !errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return nil, stacktrace.Propagate(err, "An error occurred cloning package '%s' to '%s'.", dependency, localPackagePath)
		}
		localPackagesToRelativeFilepaths[dependency] = fmt.Sprintf("%s/%s", relParentCwd, packageName)
	}

	return localPackagesToRelativeFilepaths, nil
}

func updateKurtosisYamlWithReplaceDirectives(packageNamesToLocalFilepaths map[string]string) error {
	file, err := os.OpenFile(kurtosisYMLFilePath, os.O_WRONLY|os.O_APPEND, kurtosisYMLFilePerms)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening '%s' file.", kurtosisYMLFilePath)
	}
	defer file.Close()

	// assume kurtosis.yml is a small file so okay to read into memory
	kurtosisYmlBytes, err := os.ReadFile(kurtosisYMLFilePath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred reading '%s' file.", kurtosisYMLFilePath)
	}
	replaceDirectiveStr := fmt.Sprintf("%s:\n", packageReplaceKeyInKurtosisYml)
	for packageName, localFilepath := range packageNamesToLocalFilepaths {
		// TODO: this assumes the users kurtosis yml is indented by two spaces which might always not be true and this could break a users kurtosis.yml
		// TODO: find a way to handle other indentation levels
		replaceDirectiveStr += fmt.Sprintf("  %s: %s\n", packageName, localFilepath)
	}
	if strings.Contains(string(kurtosisYmlBytes), packageReplaceKeyInKurtosisYml) {
		logrus.Infof("A replace directive was already detected in '%s' so we will avoid overwriting it. Update the replace directive with the following:\n%s", kurtosisYMLFilePath, replaceDirectiveStr)
		return nil
	}
	_, err = file.Write([]byte(replaceDirectiveStr))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing '%s' to kurtosis.yml", replaceDirectiveStr)
	}
	return nil
}
