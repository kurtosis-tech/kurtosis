package restart

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	engineImageArg = "image"
	logLevelArg    = "log-level"
)

var engineImage string
var logLevelStr string

var RestartCmd = &cobra.Command{
	Use:   command_str_consts.EngineRestartCmdStr,
	Short: "Restart the Kurtosis engine",
	Long:  "Restart the Kurtosis engine, doing nothing if no engine is running",
	RunE:  run,
}

func init() {
	RestartCmd.Flags().StringVar(
		&engineImage,
		engineImageArg,
		defaults.DefaultEngineImage,
		"The image of the Kurtosis engine that should be started",
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
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	logrus.Infof("Restarting Kurtosis engine...")

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing log level string '%v'", logLevelStr)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, engineImage)

	engineManager := engine_manager.NewEngineManager(dockerManager)

	if err := engineManager.StopEngineIdempotently(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the Kurtosis engine")
	}

	_, clientCloseFunc, err := engineManager.StartEngineIdempotently(ctx, engineImage, logLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine")
	}
	defer func() {
		if err := clientCloseFunc(); err != nil {
			logrus.Infof("We tried to close the engine client, but doing so threw an error:\n%v", err)
		}
	}()

	logrus.Infof("Restarted successfully")

	return nil
}
