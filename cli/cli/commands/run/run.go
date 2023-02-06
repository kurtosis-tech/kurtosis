package run

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	command_args_run "github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	scriptOrPackagePathKey                = "script-or-package-path"
	isScriptOrPackagePathArgumentOptional = false

	starlarkExtension = ".star"

	inputArgsArgKey                  = "args"
	inputArgsArgIsOptional           = true
	inputArgsAreNonGreedy            = false
	inputArgsAreEmptyBracesByDefault = "{}"

	dryRunFlagKey = "dry-run"
	defaultDryRun = "false"

	enclaveIdentifierFlagKey = "enclave-identifier"
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveIdentifierKeyword = ""
	// TODO deprecate this flag in the future
	enclaveIdFlagKey = "enclave-id"

	isSubnetworkCapabilitiesEnabledFlagKey = "with-subnetworks"
	defaultIsSubnetworkCapabilitiesEnabled = false

	verbosityFlagKey = "verbosity"
	defaultVerbosity = "brief"

	parallelismFlagKey = "parallelism"
	defaultParallelism = "4"

	githubDomainPrefix          = "github.com/"
	isNewEnclaveFlagWhenCreated = true
	interruptChanBufferSize     = 5

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	kurtosisYMLFilePath = "kurtosis.yml"
)

var (
	githubScriptpathValidationExceptionFunc = func(scriptpath string) bool {
		// if it's a Github path we don't validate further, the APIC will do it for us
		return strings.HasPrefix(scriptpath, githubDomainPrefix)
	}
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
			Usage: "The enclave identifier of the enclave in which the script or package will be ran." +
				"An enclave with this name will be created if it doesn't exist.",
			Type:    flags.FlagType_String,
			Default: autogenerateEnclaveIdentifierKeyword,
		},
		{
			Key: enclaveIdFlagKey,
			Usage: fmt.Sprintf(
				"The enclave identifier of the enclave in which the script or package will be ran."+
					"An enclave with this name will be created if it doesn't exist. This will be deprecated in favor of '%v'",
				enclaveIdentifierFlagKey,
			),
			Type:    flags.FlagType_String,
			Default: autogenerateEnclaveIdentifierKeyword,
		},
		{
			Key: isSubnetworkCapabilitiesEnabledFlagKey,
			Usage: "If set to true, the enclave that the script or package runs in will have subnetwork capabilities" +
				" enabled.",
			Type:    flags.FlagType_Bool,
			Default: strconv.FormatBool(defaultIsSubnetworkCapabilitiesEnabled),
		},
		{
			Key:       parallelismFlagKey,
			Usage:     "The parallelism level to be used in Starlark commands that supports it",
			Type:      flags.FlagType_Uint32,
			Shorthand: "p",
			Default:   defaultParallelism,
		},
		{
			Key:       verbosityFlagKey,
			Usage:     fmt.Sprintf("The verbosity of the command output: %s. If unset, it defaults to `brief` for a concise and explicit output. Use `detailed` to display the exhaustive list of arguments for each command. `executable` will generate executable Starlark instructions.", strings.Join(command_args_run.VerbosityStrings(), ", ")),
			Type:      flags.FlagType_String,
			Shorthand: "v",
			Default:   defaultVerbosity,
		},
	},
	Args: []*args.ArgConfig{
		// TODO add a `Usage` description here when ArgConfig supports it
		file_system_path_arg.NewFilepathOrDirpathArg(
			scriptOrPackagePathKey,
			isScriptOrPackagePathArgumentOptional,
			githubScriptpathValidationExceptionFunc,
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
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	// Args parsing and validation
	serializedJsonArgs, err := args.GetNonGreedyArg(inputArgsArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the script/package arguments using flag key '%v'", inputArgsArgKey)
	}

	userRequestedEnclaveIdentifier, err := flags.GetString(enclaveIdentifierFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier using flag key '%s'", enclaveIdentifierFlagKey)
	}

	if userRequestedEnclaveIdentifier == autogenerateEnclaveIdentifierKeyword {
		userRequestedEnclaveIdentifier, err = flags.GetString(enclaveIdFlagKey)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the enclave identifier using flag key '%s'", enclaveIdFlagKey)
		}
	}

	isPartitioningEnabled, err := flags.GetBool(isSubnetworkCapabilitiesEnabledFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the is-subnetwork-enabled setting using flag key '%v'", isSubnetworkCapabilitiesEnabledFlagKey)
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

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, isNewEnclave, err := getOrCreateEnclaveContext(ctx, userRequestedEnclaveIdentifier, kurtosisCtx, isPartitioningEnabled)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", userRequestedEnclaveIdentifier)
	}
	if isNewEnclave {
		defer output_printers.PrintEnclaveName(enclaveCtx.GetEnclaveName())
	}

	var responseLineChan <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine
	var cancelFunc context.CancelFunc
	var errRunningKurtosis error

	isRemotePackage := strings.HasPrefix(starlarkScriptOrPackagePath, githubDomainPrefix)
	if isRemotePackage {
		responseLineChan, cancelFunc, errRunningKurtosis = executeRemotePackage(ctx, enclaveCtx, starlarkScriptOrPackagePath, serializedJsonArgs, dryRun, castedParallelism)
	} else {
		fileOrDir, err := os.Stat(starlarkScriptOrPackagePath)
		if err != nil {
			return stacktrace.Propagate(err, "There was an error reading file or package from disk at '%v'", starlarkScriptOrPackagePath)
		}

		if isStandaloneScript(fileOrDir, kurtosisYMLFilePath) {
			if !strings.HasSuffix(starlarkScriptOrPackagePath, starlarkExtension) {
				return stacktrace.NewError("Expected a script with a '%s' extension but got file '%v' with a different extension", starlarkExtension, starlarkScriptOrPackagePath)
			}
			responseLineChan, cancelFunc, errRunningKurtosis = executeScript(ctx, enclaveCtx, starlarkScriptOrPackagePath, serializedJsonArgs, dryRun, castedParallelism)
		} else {
			// if the path is a file with `kurtosis.yml` at the end it's a module dir
			// we remove the `kurtosis.yml` to get just the Dir containing the module
			if isKurtosisYMLFileInPackageDir(fileOrDir, kurtosisYMLFilePath) {
				starlarkScriptOrPackagePath = path.Dir(starlarkScriptOrPackagePath)
			}
			responseLineChan, cancelFunc, errRunningKurtosis = executePackage(ctx, enclaveCtx, starlarkScriptOrPackagePath, serializedJsonArgs, dryRun, castedParallelism)
		}
	}
	if errRunningKurtosis != nil {
		return stacktrace.Propagate(errRunningKurtosis, "An error starting the Kurtosis code execution '%v'", starlarkScriptOrPackagePath)
	}

	errRunningKurtosis = readAndPrintResponseLinesUntilClosed(responseLineChan, cancelFunc, verbosity, dryRun)
	if errRunningKurtosis != nil {
		return stacktrace.Propagate(errRunningKurtosis, "Error executing Kurtosis code")
	}
	return nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func executeScript(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, scriptPath string, serializedParams string, dryRun bool, parallelism int32) (<-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	fileContentBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unable to read content of Starlark script file '%s'", scriptPath)
	}
	return enclaveCtx.RunStarlarkScript(ctx, string(fileContentBytes), serializedParams, dryRun, parallelism)
}

func executePackage(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, packagePath string, serializedParams string, dryRun bool, parallelism int32) (<-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	// we get the absolute path so that the logs make more sense
	absolutePackagePath, err := filepath.Abs(packagePath)
	logrus.Infof("Executing Starlark package at '%v' as the passed argument '%v' looks like a directory", absolutePackagePath, packagePath)

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while getting the absolute path for '%v'", packagePath)
	}
	return enclaveCtx.RunStarlarkPackage(ctx, packagePath, serializedParams, dryRun, parallelism)
}

func executeRemotePackage(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, packageId string, serializedParams string, dryRun bool, parallelism int32) (<-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	return enclaveCtx.RunStarlarkRemotePackage(ctx, packageId, serializedParams, dryRun, parallelism)
}

func readAndPrintResponseLinesUntilClosed(responseLineChan <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, cancelFunc context.CancelFunc, verbosity command_args_run.Verbosity, dryRun bool) error {
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
				if !isRunSuccessful {
					return stacktrace.NewError("Kurtosis execution threw an error. See output above for more details")
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
	isPartitioningEnabled bool,
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
	enclaveContext, err := kurtosisContext.CreateEnclave(ctx, enclaveIdentifierOrName, isPartitioningEnabled)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, fmt.Sprintf("Unable to create new enclave with name '%s'", enclaveIdentifierOrName))
	}
	logrus.Infof("Enclave '%v' created successfully", enclaveContext.GetEnclaveName())
	return enclaveContext, isNewEnclaveFlagWhenCreated, nil
}

// validatePackageArgs just validates the args is a valid JSON string
func validatePackageArgs(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	serializedJsonArgs, err := args.GetNonGreedyArg(inputArgsArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the script/package arguments using flag key '%v'", inputArgsArgKey)
	}
	var result interface{}
	if err := json.Unmarshal([]byte(serializedJsonArgs), &result); err != nil {
		return stacktrace.Propagate(err, "Error validating args, likely because it is not a valid JSON.")
	}
	return nil
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

// isStandaloneScript returns true if the fileInfo points to a non `kurtosis.yml` regular file
func isStandaloneScript(fileInfo os.FileInfo, kurtosisYMLFilePath string) bool {
	return fileInfo.Mode().IsRegular() && fileInfo.Name() != kurtosisYMLFilePath
}

func isKurtosisYMLFileInPackageDir(fileInfo os.FileInfo, kurtosisYMLFilePath string) bool {
	return fileInfo.Mode().IsRegular() && fileInfo.Name() == kurtosisYMLFilePath
}
