package run

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	enclave_consts "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/enclave"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	scriptOrModulePathKey = "script-or-module-path"
	starlarkExtension     = ".star"

	moduleArgsFlagKey = "args"
	defaultModuleArgs = "{}"

	dryRunFlagKey = "dry-run"
	defaultDryRun = "false"

	enclaveIdFlagKey = "enclave-id"
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveIdKeyword = ""

	isPartitioningEnabledFlagKey = "with-partitioning"
	defaultIsPartitioningEnabled = false

	githubDomainPrefix          = "github.com/"
	isNewEnclaveFlagWhenCreated = true

	interruptChanBufferSize = 5

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var StarlarkExecCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.StarlarkRunCmdStr,
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	ShortDescription:          "Run a Starlark script or module",
	LongDescription: "Run a Starlark module or script in an enclave. For a script we expect a path to a " + starlarkExtension +
		" file. For a module we expect path to a directory containing kurtosis.mod or a fully qualified Github repository path containing a module. If the enclave-id param is provided, Kurtosis " +
		"will exec the script inside this enclave, or create it if it doesn't exist. If no enclave-id param is " +
		"provided, Kurtosis will create a new enclave with a default name derived from the script or module name.",
	Flags: []*flags.FlagConfig{
		{
			Key: dryRunFlagKey,
			// TODO(gb): link to a doc page mentioning what a "Kurtosis instruction" is
			Usage:   "If true, the Kurtosis instructions will not be executed, they will just be printed to the output of the CLI",
			Type:    flags.FlagType_Bool,
			Default: defaultDryRun,
		},
		{
			Key: moduleArgsFlagKey,
			// TODO(gb): Link to a proper doc page explaining what a proto file is, etc. when we have it
			Usage:   "The parameters that should be passed to the Kurtosis module when running it. It is expected to be a serialized JSON string. Note that if a standalone Kurtosis script is being run, no parameter should be passed.",
			Type:    flags.FlagType_String,
			Default: defaultModuleArgs,
		},
		{
			Key: enclaveIdFlagKey,
			Usage: fmt.Sprintf(
				"The enclave ID in which the script or module will be ran, which must match regex '%v' "+
					"(emptystring will autogenerate an enclave ID). An enclave with this ID will be created if it doesn't exist.",
				enclave_consts.AllowedEnclaveIdCharsRegexStr,
			),
			Type:    flags.FlagType_String,
			Default: autogenerateEnclaveIdKeyword,
		},
		{
			Key: isPartitioningEnabledFlagKey,
			Usage: "If set to true, the enclave that the module runs in will have partitioning enabled so " +
				"network partitioning simulations can be run",
			Type:    flags.FlagType_Bool,
			Default: strconv.FormatBool(defaultIsPartitioningEnabled),
		},
	},
	Args: []*args.ArgConfig{
		&args.ArgConfig{
			// for a module we expect a path to a directory
			// for a script we expect a script with a `.star` extension
			// TODO add a `Usage` description here when ArgConfig supports it
			Key:             scriptOrModulePathKey,
			IsOptional:      false,
			DefaultValue:    "",
			IsGreedy:        false,
			CompletionsFunc: nil,
			ValidationFunc:  validateScriptOrModulePath,
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
	serializedJsonArgs, err := flags.GetString(moduleArgsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the module parameters using flag key '%v'", moduleArgsFlagKey)
	}
	if err = validateModuleArgs(serializedJsonArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the module parameters '%v'", serializedJsonArgs)
	}

	userRequestedEnclaveId, err := flags.GetString(enclaveIdFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using flag key '%s'", enclaveIdFlagKey)
	}
	isPartitioningEnabled, err := flags.GetBool(isPartitioningEnabledFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the is-partitioning-enabled setting using flag key '%v'", isPartitioningEnabledFlagKey)
	}

	starlarkScriptOrModulePath, err := args.GetNonGreedyArg(scriptOrModulePathKey)
	if err != nil {
		return stacktrace.Propagate(err, "Error reading the Starlark script or module dir at '%s'. Does it exist?", starlarkScriptOrModulePath)
	}

	dryRun, err := flags.GetBool(dryRunFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a boolean flag with key '%v' but none was found; this is an error in Kurtosis!", dryRunFlagKey)
	}

	// Get or create enclave in Kurtosis
	enclaveIdStr := userRequestedEnclaveId
	enclaveId := enclaves.EnclaveID(enclaveIdStr)

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, isNewEnclave, err := getOrCreateEnclaveContext(ctx, enclaveId, kurtosisCtx, isPartitioningEnabled)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", enclaveId)
	}
	if isNewEnclave {
		defer output_printers.PrintEnclaveId(enclaveCtx.GetEnclaveID())
	}

	var responseLineChan <-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine
	var cancelFunc context.CancelFunc
	var errRunningKurtosis error

	isRemoteModule := strings.HasPrefix(starlarkScriptOrModulePath, githubDomainPrefix)
	if isRemoteModule {
		responseLineChan, cancelFunc, errRunningKurtosis = executeRemoteModule(ctx, enclaveCtx, starlarkScriptOrModulePath, serializedJsonArgs, dryRun)
	} else {
		fileOrDir, err := os.Stat(starlarkScriptOrModulePath)
		if err != nil {
			return stacktrace.Propagate(err, "There was an error reading file or module from disk at '%v'", starlarkScriptOrModulePath)
		}

		isStandaloneScript := fileOrDir.Mode().IsRegular()
		if isStandaloneScript {
			if !strings.HasSuffix(starlarkScriptOrModulePath, starlarkExtension) {
				return stacktrace.NewError("Expected a script with a '%s' extension but got file '%v' with a different extension", starlarkExtension, starlarkScriptOrModulePath)
			}
			responseLineChan, cancelFunc, errRunningKurtosis = executeScript(ctx, enclaveCtx, starlarkScriptOrModulePath, dryRun)
		} else {
			responseLineChan, cancelFunc, errRunningKurtosis = executeModule(ctx, enclaveCtx, starlarkScriptOrModulePath, serializedJsonArgs, dryRun)
		}
	}
	if errRunningKurtosis != nil {
		return stacktrace.Propagate(errRunningKurtosis, "An error starting the Kurtosis code execution '%v'", starlarkScriptOrModulePath)
	}

	scriptOutput, errRunningKurtosis := readResponseLinesUntilClosed(responseLineChan, cancelFunc)
	if errRunningKurtosis != nil {
		return stacktrace.Propagate(errRunningKurtosis, "Error executing Kurtosis code")
	}
	logrus.Infof("Kurtosis script executed successfully. Output of the module was: \n%v", scriptOutput)
	return nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func validateScriptOrModulePath(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	scriptOrModulePath, err := args.GetNonGreedyArg(scriptOrModulePathKey)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to get argument '%s'", scriptOrModulePathKey)
	}

	scriptOrModulePath = strings.TrimSpace(scriptOrModulePath)
	if scriptOrModulePath == "" {
		return stacktrace.NewError("Received an empty '%v'. It should be a non empty string.", scriptOrModulePathKey)
	}

	if strings.HasPrefix(scriptOrModulePath, githubDomainPrefix) {
		// if it's a Github path we don't validate further, the APIC will do it for us
		return nil
	}

	fileInfo, err := os.Stat(scriptOrModulePath)
	if err != nil {
		return stacktrace.Propagate(err, "Error reading script file or module dir '%s'", scriptOrModulePath)
	}
	if !fileInfo.Mode().IsRegular() && !fileInfo.Mode().IsDir() {
		return stacktrace.Propagate(err, "Script or module path should point to a file on disk or to a directory '%s'", scriptOrModulePath)
	}
	return nil
}

func executeScript(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, scriptPath string, dryRun bool) (<-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, context.CancelFunc, error) {
	fileContentBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unable to read content of Starlark script file '%s'", scriptPath)
	}
	return enclaveCtx.ExecuteKurtosisScript(ctx, string(fileContentBytes), dryRun)
}

func executeModule(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, modulePath string, serializedParams string, dryRun bool) (<-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, context.CancelFunc, error) {
	// we get the absolute path so that the logs make more sense
	absoluteModulePath, err := filepath.Abs(modulePath)
	logrus.Infof("Executing Starlark package at '%v' as the passed argument '%v' looks like a directory", absoluteModulePath, modulePath)

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while getting the absolute path for '%v'", modulePath)
	}
	return enclaveCtx.ExecuteKurtosisModule(ctx, modulePath, serializedParams, dryRun)
}

func executeRemoteModule(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, moduleId string, serializedParams string, dryRun bool) (<-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, context.CancelFunc, error) {
	return enclaveCtx.ExecuteKurtosisRemoteModule(ctx, moduleId, serializedParams, dryRun)
}

func readResponseLinesUntilClosed(responseLineChan <-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine, cancelFunc context.CancelFunc) (string, error) {
	defer cancelFunc()

	// This channel will receive a signal when the user presses an interrupt
	interruptChan := make(chan os.Signal, interruptChanBufferSize)
	signal.Notify(interruptChan, os.Interrupt)
	defer close(interruptChan)

	scriptOutput := &strings.Builder{}
	isError := false
	for {
		select {
		case responseLine, isChanOpen := <-responseLineChan:
			if !isChanOpen {
				if isError {
					return scriptOutput.String(), stacktrace.NewError("Kurtosis execution threw an error. See output above for more details")
				}
				return scriptOutput.String(), nil
			}
			isError = isError || output_printers.PrintKurtosisExecutionResponseLineToStdOut(responseLine, scriptOutput)
		case <-interruptChan:
			return scriptOutput.String(), stacktrace.NewError("User manually interrupted the execution, returning. Note that the execution will continue in the Kurtosis enclave")
		}
	}
}

func getOrCreateEnclaveContext(
	ctx context.Context,
	enclaveId enclaves.EnclaveID,
	kurtosisContext *kurtosis_context.KurtosisContext,
	isPartitioningEnabled bool,
) (*enclaves.EnclaveContext, bool, error) {

	enclavesMap, err := kurtosisContext.GetEnclaves(ctx)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "Unable to get existing enclaves from Kurtosis backend")
	}
	if _, found := enclavesMap[enclaveId]; found {
		enclaveContext, err := kurtosisContext.GetEnclaveContext(ctx, enclaveId)
		if err != nil {
			return nil, false, stacktrace.Propagate(err, "Unable to get enclave context from the existing enclave '%s'", enclaveId)
		}
		return enclaveContext, false, nil
	}
	logrus.Infof("Creating a new enclave for Starlark to run inside...")
	enclaveContext, err := kurtosisContext.CreateEnclave(ctx, enclaveId, isPartitioningEnabled)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, fmt.Sprintf("Unable to create new enclave with ID '%s'", enclaveId))
	}
	logrus.Infof("Enclave '%v' created successfully", enclaveContext.GetEnclaveID())
	return enclaveContext, isNewEnclaveFlagWhenCreated, nil
}

// validateModuleArgs just validates the args is a valid JSON string
func validateModuleArgs(serializedJson string) error {
	var result interface{}
	if err := json.Unmarshal([]byte(serializedJson), &result); err != nil {
		return stacktrace.Propagate(err, "Error validating args, likely because it is not a valid JSON.")
	}
	return nil
}
