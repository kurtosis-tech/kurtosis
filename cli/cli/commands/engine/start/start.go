package start

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	engineVersionArg = "version"
	logLevelArg      = "log-level"

	defaultEngineVersion = ""
)

var engineVersion string
var logLevelStr string

var StartCmd = &cobra.Command{
	Use:   command_str_consts.EngineStartCmdStr,
	Short: "Starts the Kurtosis engine",
	Long:  "Starts the Kurtosis engine, doing nothing if an engine is already running",
	RunE:  run,
}

func init() {
	StartCmd.Flags().StringVar(
		&engineVersion,
		engineVersionArg,
		defaultEngineVersion,
		"The version (Docker tag) of the Kurtosis engine that should be started (blank will start the default version)",
	)
	StartCmd.Flags().StringVar(
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
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	logrus.Infof("Starting Kurtosis engine from image '%v'...", engineVersion)

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
	if engineVersion == defaultEngineVersion {
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, logLevel)
	} else {
		_, engineClientCloseFunc, startEngineErr = engineManager.StartEngineIdempotentlyWithCustomVersion(ctx, engineVersion, logLevel)
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
