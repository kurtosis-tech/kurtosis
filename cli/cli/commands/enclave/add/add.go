package add

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	enclave_consts "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/enclave"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	apiContainerVersionFlagKey   = "api-container-version"
	apiContainerLogLevelFlagKey  = "api-container-log-level"
	enclaveNameFlagKey           = "name"
	enclaveProductionModeFlagKey = "production"

	// Signifies that an enclave name should be auto-generated
	autogenerateEnclaveNameKeyword = ""

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

// EnclaveAddCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var EnclaveAddCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveAddCmdStr,
	ShortDescription:          "Creates an enclave",
	LongDescription:           "Creates a new, empty Kurtosis enclave",
	RunFunc:                   run,
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:       apiContainerLogLevelFlagKey,
			Shorthand: "l",
			Type:      flags.FlagType_String,
			Usage: fmt.Sprintf(
				"The log level that the API container should log at (%v)",
				strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
			),
			Default: defaults.DefaultApiContainerLogLevel.String(),
		}, {
			Key:       apiContainerVersionFlagKey,
			Shorthand: "a",
			Type:      flags.FlagType_String,
			Default:   defaults.DefaultAPIContainerVersion,
			Usage:     "The version of the Kurtosis API container that should be started inside the enclave (blank tells the engine to use the default version)",
		}, {
			Key:       enclaveNameFlagKey,
			Shorthand: "n",
			Default:   autogenerateEnclaveNameKeyword,
			Usage: fmt.Sprintf(
				"The enclave name to give the new enclave, which must match regex '%v' "+
					"(emptystring will autogenerate an enclave name)",
				enclave_consts.AllowedEnclaveNameCharsRegexStr,
			),
			Type: flags.FlagType_String,
		},
		{
			Key:       enclaveProductionModeFlagKey,
			Usage:     "If enabled, services will restart if they fail",
			Shorthand: "p",
			Type:      flags.FlagType_Bool,
			Default:   "false",
		},
	},
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	_ *args.ParsedArgs,
) error {

	apiContainerVersion, err := flags.GetString(apiContainerVersionFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the API Container Version using flag with key '%v'; this is a bug in Kurtosis", apiContainerVersionFlagKey)
	}

	isDebugMode, err := flags.GetBool(defaults.DebugModeFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", defaults.DebugModeFlagKey)
	}

	shouldApicRunInDebugMode := defaults.DefaultEnableDebugMode
	if isDebugMode && apiContainerVersion == defaults.DefaultAPIContainerVersion {
		apiContainerVersion = fmt.Sprintf("%s-%s", kurtosis_version.KurtosisVersion, defaults.DefaultKurtosisContainerDebugImageNameSuffix)
		shouldApicRunInDebugMode = true
	}

	kurtosisLogLevelStr, err := flags.GetString(apiContainerLogLevelFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the API Container log level using flag with key '%v'; this is a bug in Kurtosis", apiContainerLogLevelFlagKey)
	}

	enclaveName, err := flags.GetString(enclaveNameFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the enclave name using flag with key '%v'; this is a bug in Kurtosis ", enclaveNameFlagKey)
	}

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}

	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, defaults.DefaultEngineLogLevel, defaults.DefaultEngineEnclavePoolSize)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer func() {
		if err = closeClientFunc(); err != nil {
			logrus.Errorf("Error closing the engine client")
		}
	}()

	logrus.Info("Creating new enclave...")

	isProduction, err := flags.GetBool(enclaveProductionModeFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the enclave mode using flag with key '%v'; this is a bug in Kurtosis", enclaveProductionModeFlagKey)
	}

	mode := kurtosis_engine_rpc_api_bindings.EnclaveMode_TEST
	if isProduction {
		mode = kurtosis_engine_rpc_api_bindings.EnclaveMode_PRODUCTION
	}

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveName:              &enclaveName,
		ApiContainerVersionTag:   &apiContainerVersion,
		ApiContainerLogLevel:     &kurtosisLogLevelStr,
		Mode:                     &mode,
		ShouldApicRunInDebugMode: &shouldApicRunInDebugMode,
	}
	createdEnclaveResponse, err := engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveName)
	}

	enclaveInfo := createdEnclaveResponse.GetEnclaveInfo()
	enclaveName = enclaveInfo.Name

	defer output_printers.PrintEnclaveName(enclaveName)

	return nil
}
