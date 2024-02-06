package start

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/engine/common"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const (
	engineVersionFlagKey   = "version"
	logLevelFlagKey        = "log-level"
	enclavePoolSizeFlagKey = "enclave-pool-size"

	defaultEngineVersion          = ""
	kurtosisTechEngineImagePrefix = "kurtosistech/engine"
	imageVersionDelimiter         = ":"
)

var StartCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.EngineStartCmdStr,
	ShortDescription: "Starts the Kurtosis engine",
	LongDescription:  "Starts the Kurtosis engine, doing nothing if an engine is already running",
	Args:             nil,
	Flags: []*flags.FlagConfig{
		{
			Key:       engineVersionFlagKey,
			Usage:     "The version (Docker tag) of the Kurtosis engine that should be started (blank will start the default version)",
			Shorthand: "",
			Type:      flags.FlagType_String,
			Default:   defaultEngineVersion,
		},
		{
			Key: logLevelFlagKey,
			Usage: fmt.Sprintf(
				"The level that the started engine should log at (%v)",
				strings.Join(
					logrus_log_levels.GetAcceptableLogLevelStrs(),
					"|",
				),
			),
			Shorthand: "",
			Type:      flags.FlagType_String,
			Default:   defaults.DefaultEngineLogLevel.String(),
		},
		{
			Key: enclavePoolSizeFlagKey,
			Usage: fmt.Sprintf(
				"The enclave pool size, the default value is '%v' which means it will be disabled. CAUTION: This is only available for Kubernetes, and this command will fail if you want to use it for Docker.",
				defaults.DefaultEngineEnclavePoolSize,
			),
			Shorthand: "",
			Type:      flags.FlagType_Uint8,
			Default:   strconv.Itoa(int(defaults.DefaultEngineEnclavePoolSize)),
		},
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, flags *flags.ParsedFlags, _ *args.ParsedArgs) error {
	ctx := context.Background()

	enclavePoolSize, err := flags.GetUint8(enclavePoolSizeFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a integer flag with key '%v' but none was found; this is an error in Kurtosis!", enclavePoolSizeFlagKey)
	}

	if err := common.ValidateEnclavePoolSizeFlag(enclavePoolSize); err != nil {
		return stacktrace.Propagate(err, "An error occurred validating the '%v' flag", enclavePoolSizeFlagKey)
	}

	logLevelStr, err := flags.GetString(logLevelFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the Kurtosis engine log level using flag with key '%v'; this is a bug in Kurtosis", logLevelFlagKey)
	}

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing log level string '%v'", logLevelStr)
	}

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager")
	}

	var engineClientCloseFunc func() error
	var startEngineErr error

	engineVersion, err := flags.GetString(engineVersionFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the Kurtosis engine Container Version using flag with key '%v'; this is a bug in Kurtosis", engineVersionFlagKey)
	}

	isDebugMode, err := flags.GetBool(defaults.DebugModeFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", defaults.DebugModeFlagKey)
	}

	if engineVersion == defaultEngineVersion && isDebugMode {
		engineDebugVersion := fmt.Sprintf("%s-%s", kurtosis_version.KurtosisVersion, defaults.DefaultKurtosisContainerDebugImageNameSuffix)
		logrus.Infof("Starting Kurtosis engine in debug mode from image '%v%v%v'...", kurtosisTechEngineImagePrefix, imageVersionDelimiter, engineDebugVersion)
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithCustomVersion(ctx, engineDebugVersion, logLevel, enclavePoolSize, true)
	} else if engineVersion == defaultEngineVersion {

		logrus.Infof("Starting Kurtosis engine from image '%v%v%v'...", kurtosisTechEngineImagePrefix, imageVersionDelimiter, kurtosis_version.KurtosisVersion)
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, logLevel, enclavePoolSize)
	} else {
		logrus.Infof("Starting Kurtosis engine from image '%v%v%v'...", kurtosisTechEngineImagePrefix, imageVersionDelimiter, engineVersion)
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithCustomVersion(ctx, engineVersion, logLevel, enclavePoolSize, defaults.DefaultEnableDebugMode)
	}
	if startEngineErr != nil {
		return stacktrace.Propagate(startEngineErr, "An error occurred starting the Kurtosis engine")
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()

	logrus.Info("Kurtosis engine started")

	return nil
}
