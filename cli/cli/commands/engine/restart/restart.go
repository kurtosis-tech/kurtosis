package restart

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/engine/common"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	engineVersionArg         = "version"
	logLevelArg              = "log-level"
	enclavePoolSizeFlag      = "enclave-pool-size"
	gitAuthTokenOverrideFlag = "git-auth-token"

	defaultEngineVersion                   = ""
	restartEngineOnSameVersionIfAnyRunning = false
)

var engineVersion string
var logLevelStr string
var enclavePoolSize uint8
var gitAuthTokenOverride string

// RestartCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var RestartCmd = &cobra.Command{
	Use:   command_str_consts.EngineRestartCmdStr,
	Short: "Restart the Kurtosis engine",
	Long:  "Stops any existing Kurtosis engine, then starts a new one",
	RunE:  run,
}

func init() {
	RestartCmd.Flags().StringVar(
		&engineVersion,
		engineVersionArg,
		defaultEngineVersion,
		"The version (Docker tag) of the Kurtosis engine that should be started (blank will start the default version)",
	)
	RestartCmd.Flags().StringVar(
		&logLevelStr,
		logLevelArg,
		defaults.DefaultEngineLogLevel.String(),
		fmt.Sprintf(
			"The level that the started engine should log at (%v)",
			strings.Join(
				logrus_log_levels.GetAcceptableLogLevelStrs(),
				"|",
			),
		),
	)
	RestartCmd.Flags().Uint8Var(
		&enclavePoolSize,
		enclavePoolSizeFlag,
		defaults.DefaultEngineEnclavePoolSize,
		fmt.Sprintf(
			"The enclave pool size, the default value is '%v' which means it will be disabled. CAUTION: This is only available for Kubernetes, and this command will fail if you want to use it for Docker.",
			defaults.DefaultEngineEnclavePoolSize,
		),
	)
	RestartCmd.Flags().StringVar(
		&gitAuthTokenOverride,
		gitAuthTokenOverrideFlag,
		"",
		"Git personal access token for providing authorization for github operation. If existing github user is logged into Kurtosis CLI, the provided personal access token will override the user.",
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if err := common.ValidateEnclavePoolSizeFlag(enclavePoolSize); err != nil {
		return stacktrace.Propagate(err, "An error occurred validating the '%v' flag", enclavePoolSizeFlag)
	}

	logrus.Infof("Restarting Kurtosis engine...")

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing log level string '%v'", logLevelStr)
	}

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}

	var engineClientCloseFunc func() error
	var restartEngineErr error
	_, engineClientCloseFunc, restartEngineErr = engineManager.RestartEngineIdempotently(ctx, logLevel, engineVersion, restartEngineOnSameVersionIfAnyRunning, enclavePoolSize, gitAuthTokenOverride)
	if restartEngineErr != nil {
		return stacktrace.Propagate(restartEngineErr, "An error occurred restarting the Kurtosis engine")
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()

	logrus.Infof("Engine restarted successfully")
	return nil
}
