package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	enclave_consts "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/enclave"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

const (
	scriptOrModulePathKey = "script-or-module-path"
	startosisExtension    = ".star"

	moduleArgsFlagKey = "args"
	defaultModuleArgs = "{}"

	dryRunFlagKey = "dry-run"
	defaultDryRun = "false"

	enclaveIdFlagKey = "enclave-id"
	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveIdKeyword = ""

	isPartitioningEnabledFlagKey = "with-partitioning"
	defaultIsPartitioningEnabled = false

	scriptArgForLogging = "script"
	moduleArgForLogging = "module"

	githubDomainPrefix = "github.com/"
)

var StartosisExecCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.StartosisExecCmdStr,
	ShortDescription: "Execute a Startosis script or module",
	LongDescription: "Execute a Startosis module or script in an enclave. For a script we expect a path to a " + startosisExtension +
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
			Usage:   "The parameters that should be passed to the Kurtosis module when executing it. It is expected to be a serialized JSON string. Note that if a standalone Kurtosis script is being executed, no parameter should be passed.",
			Type:    flags.FlagType_String,
			Default: defaultModuleArgs,
		},
		{
			Key: enclaveIdFlagKey,
			Usage: fmt.Sprintf(
				"The enclave ID in which the script or module will be executed, which must match regex '%v' "+
					"(emptystring will autogenerate an enclave ID). An enclave with this ID will be created if it doesn't exist.",
				enclave_consts.AllowedEnclaveIdCharsRegexStr,
			),
			Type:    flags.FlagType_String,
			Default: autogenerateEnclaveIdKeyword,
		},
		{
			Key: isPartitioningEnabledFlagKey,
			Usage: "If set to true, the enclave that the module executes in will have partitioning enabled so " +
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
			Key:            scriptOrModulePathKey,
			IsOptional:     false,
			DefaultValue:   "",
			IsGreedy:       false,
			ValidationFunc: validateScriptOrModulePath,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
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

	startosisScriptOrModulePath, err := args.GetNonGreedyArg(scriptOrModulePathKey)
	if err != nil {
		return stacktrace.Propagate(err, "Error reading the Startosis script or module dir at '%s'. Does it exist?", startosisScriptOrModulePath)
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

	enclaveCtx, err := getOrCreateEnclaveContext(ctx, enclaveId, kurtosisCtx, isPartitioningEnabled)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", enclaveId)
	}
	defer output_printers.PrintEnclaveId(enclaveCtx.GetEnclaveID())

	if strings.HasPrefix(startosisScriptOrModulePath, githubDomainPrefix) {
		err = executeRemoteModule(enclaveCtx, startosisScriptOrModulePath, serializedJsonArgs, dryRun)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while running the module '%v'", startosisScriptOrModulePath)
		}
		return nil
	}

	fileOrDir, err := os.Stat(startosisScriptOrModulePath)
	if err != nil {
		return stacktrace.Propagate(err, "There was an error reading file or module from disk at '%v'", startosisScriptOrModulePath)
	}

	if fileOrDir.Mode().IsRegular() {
		if !strings.HasSuffix(startosisScriptOrModulePath, startosisExtension) {
			return stacktrace.NewError("Expected a script with a '%s' extension but got file '%v' with a different extension", startosisExtension, startosisScriptOrModulePath)
		}
		err = executeScript(enclaveCtx, startosisScriptOrModulePath, dryRun)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while executing script '%v'", startosisScriptOrModulePath)
		}
		return nil
	}

	err = executeModule(enclaveCtx, startosisScriptOrModulePath, serializedJsonArgs, dryRun)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while running the module '%v'", startosisScriptOrModulePath)
	}

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

	fileInfo, err := os.Stat(scriptOrModulePath)
	if err != nil {
		return stacktrace.Propagate(err, "Error reading script file or module dir '%s'", scriptOrModulePath)
	}
	if !fileInfo.Mode().IsRegular() && !fileInfo.Mode().IsDir() {
		return stacktrace.Propagate(err, "Script or module path should point to a file on disk or to a directory '%s'", scriptOrModulePath)
	}
	return nil
}

func executeScript(enclaveCtx *enclaves.EnclaveContext, scriptPath string, dryRun bool) error {
	fileContentBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to read content of Startosis script file '%s'", scriptPath)
	}

	executionResponse, err := enclaveCtx.ExecuteStartosisScript(string(fileContentBytes), dryRun)
	if err != nil {
		return stacktrace.Propagate(err, "An unexpected error occurred executing the Startosis script '%s'", scriptPath)
	}

	err = validateExecutionResponse(executionResponse, scriptPath, scriptArgForLogging, dryRun)
	if err != nil {
		return stacktrace.Propagate(err, "Ran into a few errors while interpreting, validating or executing the script '%v'", scriptPath)
	}

	return nil
}

func executeModule(enclaveCtx *enclaves.EnclaveContext, modulePath string, serializedParams string, dryRun bool) error {
	executionResponse, err := enclaveCtx.ExecuteStartosisModule(modulePath, serializedParams, dryRun)
	if err != nil {
		return stacktrace.Propagate(err, "An unexpected error occurred executing the Startosis module '%s'", modulePath)
	}

	err = validateExecutionResponse(executionResponse, modulePath, moduleArgForLogging, dryRun)
	if err != nil {
		return stacktrace.Propagate(err, "Ran into a few errors while interpreting, validating or executing the module '%v' with dry-run set to '%v'", modulePath, dryRun)
	}

	return nil
}

func executeRemoteModule(enclaveCtx *enclaves.EnclaveContext, moduleId string, serializedParams string, dryRun bool) error {
	executionResponse, err := enclaveCtx.ExecuteStartosisRemoteModule(moduleId, serializedParams, dryRun)
	if err != nil {
		return stacktrace.Propagate(err, "An unexpected error occurred executing the Startosis module '%s'", moduleId)
	}

	err = validateExecutionResponse(executionResponse, moduleId, moduleArgForLogging, dryRun)
	if err != nil {
		return stacktrace.Propagate(err, "Ran into a few errors while interpreting, validating or executing the module '%v' with dry-run set to '%v'", moduleId, dryRun)
	}

	return nil
}

func validateExecutionResponse(executionResponse *kurtosis_core_rpc_api_bindings.ExecuteStartosisResponse, scriptOrModulePath string, scriptOrModuleArg string, dryRun bool) error {
	if executionResponse.InterpretationError != "" {
		return stacktrace.NewError("There was an error interpreting the Startosis %s '%s': \n%v", scriptOrModuleArg, scriptOrModulePath, executionResponse.InterpretationError)
	}
	if len(executionResponse.ValidationErrors) > 0 {
		return stacktrace.NewError("There was an error validating the Startosis %s '%s': \n%v", scriptOrModuleArg, scriptOrModulePath, executionResponse.ValidationErrors)
	}

	concatenatedKurtosisInstructions := make([]string, len(executionResponse.SerializedInstructions))
	for idx := 0; idx < len(executionResponse.SerializedInstructions); idx++ {
		concatenatedKurtosisInstructions[idx] = executionResponse.SerializedInstructions[idx].SerializedInstruction
	}
	logrus.Infof("Kurtosis script successfully interpreted and validated. List of Kurtosis instructions generated:\n%v", strings.Join(concatenatedKurtosisInstructions, "\n"))

	if executionResponse.ExecutionError != "" {
		return stacktrace.NewError("There was an error executing the Startosis %s '%s': \n%v", scriptOrModuleArg, scriptOrModulePath, executionResponse.ExecutionError)
	}

	if dryRun {
		logrus.Infof("Kurtosis script '%s' executed successfully in dry-run mode. The instructions printed above were not submitted to Kurtotis engine.", scriptOrModuleArg)
	} else {
		logrus.Infof("Kurtosis script '%s' executed successfully. All instructions listed above were submitted to Kurtosis engine.", scriptOrModuleArg)
	}
	logrus.Infof("Output of the module was: \n%v", executionResponse.SerializedScriptOutput)
	return nil
}

func getOrCreateEnclaveContext(ctx context.Context, enclaveId enclaves.EnclaveID, kurtosisContext *kurtosis_context.KurtosisContext, isPartitioningEnabled bool) (*enclaves.EnclaveContext, error) {
	enclavesMap, err := kurtosisContext.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get existing enclaves from Kurtosis backend")
	}
	if _, found := enclavesMap[enclaveId]; found {
		enclaveContext, err := kurtosisContext.GetEnclaveContext(ctx, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Unable to get enclave context from the existing enclave '%s'", enclaveId)
		}
		return enclaveContext, nil
	}
	enclaveContext, err := kurtosisContext.CreateEnclave(ctx, enclaveId, isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, fmt.Sprintf("Unable to create new enclave with ID '%s'", enclaveId))
	}
	return enclaveContext, nil
}

// validateModuleArgs just validates the args is a valid JSON string
func validateModuleArgs(serializedJson string) error {
	var result interface{}
	if err := json.Unmarshal([]byte(serializedJson), &result); err != nil {
		return stacktrace.Propagate(err, "Error validating args, likely because it is not a valid JSON.")
	}
	return nil
}
