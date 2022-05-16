package restart

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

var RestartCmd = &cobra.Command{
	Use:   command_str_consts.EngineRestartCmdStr,
	Short: "Restart the Kurtosis engine",
	Long:  "Stops any existing Kurtosis engine, then starts a new one",
	RunE:  run,
}

var WithKubernetes bool

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
	// TODO Remove this in favor of actual Kubernetes info in the config file
	RestartCmd.Flags().BoolVarP(&WithKubernetes, "with-kubernetes", "k", false, "Operate on the engine in kubernetes")
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	logrus.Infof("Restarting Kurtosis engine...")

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing log level string '%v'", logLevelStr)
	}

	// TODO Hack; remove when we read cluster state from disk
	var clusterName = "docker"
	if WithKubernetes {
		clusterName = "minikube"
	}
	engineManager, err := engine_manager.NewEngineManager()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager connected to cluster '%v'", clusterName)
	}

	if err := engineManager.StopEngineIdempotently(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the Kurtosis engine")
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
	defer engineClientCloseFunc()

	logrus.Infof("Engine restarted successfully")

	return nil
}
