package start

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
)

const (
	engineVersionFlagKey           = "version"
	logLevelFlagKey                = "log-level"
	enclavePoolSizeFlagKey         = "enclave-pool-size"
	githubAuthTokenOverrideFlagKey = "github-auth-token"
	logRetentionPeriodFlagKey      = "log-retention-period"

	defaultEngineVersion          = ""
	kurtosisTechEngineImagePrefix = "kurtosistech/engine"
	imageVersionDelimiter         = ":"

	defaultShouldRestartAPIContainers = "false"
	restartAPIContainersFlagKey       = "restart-api-containers"

	domainFlagKey = "domain"
	defaultDomain = ""
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
		{
			Key:       githubAuthTokenOverrideFlagKey,
			Usage:     "The github auth token that should be used to authorize git operations such as accessing packages in private repositories. Overrides existing github auth config if a user is logged in.",
			Shorthand: "",
			Type:      flags.FlagType_String,
			Default:   defaults.DefaultGitHubAuthTokenOverride,
		},
		{
			Key:       restartAPIContainersFlagKey,
			Usage:     "Restart the current API containers after starting the Kurtosis engine.",
			Shorthand: "",
			Type:      flags.FlagType_Bool,
			Default:   defaultShouldRestartAPIContainers,
		},
		{
			Key:       domainFlagKey,
			Usage:     "The domain name of the enclave manager UI if self-hosting (blank defaults to localhost for the local use case and cloud.kurtosis.com for the Kurtosis cloud use case)",
			Shorthand: "",
			Type:      flags.FlagType_String,
			Default:   defaultDomain,
		},
		{
			Key:       logRetentionPeriodFlagKey,
			Usage:     "The length of time that Kurtosis should keep logs for. Eg. if set to 168h, Kurtosis will remove all logs beyond 1 week. You can specify hours using \"h\" however Kurtosis currently only supports setting retention on a weekly basis.",
			Shorthand: "",
			Type:      flags.FlagType_String,
			Default:   defaults.DefaultLogRetentionPeriod,
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

	githubAuthTokenOverride, err := flags.GetString(githubAuthTokenOverrideFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting GitHub auth token override flag with key '%v'. This is a bug in Kurtosis", githubAuthTokenOverrideFlagKey)
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

	shouldRestartAPIContainers, err := flags.GetBool(restartAPIContainersFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", restartAPIContainersFlagKey)
	}

	domain, err := flags.GetString(domainFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the Kurtosis engine enclave manager UI domain name using the flag with key '%v'; this is a bug in Kurtosis", domainFlagKey)
	}

	logRetentionPeriodStr, err := flags.GetString(logRetentionPeriodFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting the log retention period string from flag: '%v'", logRetentionPeriodStr)
	}
	_, err = time.ParseDuration(logRetentionPeriodStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing provided log retention period '%v' into a duration. Ensure the provided value has the proper format. Valid time units are \"ns\", \"us\" (or \"µs\"), \"ms\", \"s\", \"m\", \"h\".", logRetentionPeriodStr)
	}

	if engineVersion == defaultEngineVersion && isDebugMode {
		engineDebugVersion := fmt.Sprintf("%s-%s", kurtosis_version.KurtosisVersion, defaults.DefaultKurtosisContainerDebugImageNameSuffix)
		logrus.Infof("Starting Kurtosis engine in debug mode from image '%v%v%v'...", kurtosisTechEngineImagePrefix, imageVersionDelimiter, engineDebugVersion)
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithCustomVersion(ctx, engineDebugVersion, logLevel, enclavePoolSize, true, githubAuthTokenOverride, shouldRestartAPIContainers, domain, logRetentionPeriodStr)
	} else if engineVersion == defaultEngineVersion {
		logrus.Infof("Starting Kurtosis engine from image '%v%v%v'...", kurtosisTechEngineImagePrefix, imageVersionDelimiter, kurtosis_version.KurtosisVersion)
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, logLevel, enclavePoolSize, githubAuthTokenOverride, shouldRestartAPIContainers, domain, logRetentionPeriodStr)
	} else {
		logrus.Infof("Starting Kurtosis engine from image '%v%v%v'...", kurtosisTechEngineImagePrefix, imageVersionDelimiter, engineVersion)
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithCustomVersion(ctx, engineVersion, logLevel, enclavePoolSize, defaults.DefaultEnableDebugMode, githubAuthTokenOverride, shouldRestartAPIContainers, domain, logRetentionPeriodStr)
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
