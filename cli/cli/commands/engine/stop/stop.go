package stop

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/version_checker"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)


var StopCmd = &cobra.Command{
	Use:   command_str_consts.EngineStopCmdStr,
	Short: "Stops the Kurtosis engine",
	Long: "Stops the Kurtosis engine, doing nothing if no engine is running",
	RunE:  run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	version_checker.CheckLatestVersion()

	logrus.Infof("Stopping Kurtosis engine...")

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineManager := engine_manager.NewEngineManager(dockerManager)
	if err := engineManager.StopEngineIdempotently(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the Kurtosis engine")
	}

	logrus.Info("Kurtosis engine successfully stopped")
	return nil
}
