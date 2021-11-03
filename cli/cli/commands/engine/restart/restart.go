package restart

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RestartCmd = &cobra.Command{
	Use:   command_str_consts.EngineRestartCmdStr,
	Short: "Restart the Kurtosis engine",
	Long:  "Restart the Kurtosis engine, doing nothing if no engine is running",
	RunE:  run,
}

func init() {

}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	logrus.Infof("Restarting Kurtosis engine...")

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineManager := engine_manager.NewEngineManager(dockerManager)

	engineStatus, hostPortBindings, currentEngineVersion, err :=  engineManager.GetEngineStatus(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting engine status")
	}
	logrus.Debugf("Currently running Kurtosis engine version '%v' with host port bindings '%+v'", currentEngineVersion, hostPortBindings)

	if engineStatus != engine_manager.EngineStatus_Stopped {
		if err := engineManager.StopEngineIdempotently(ctx); err != nil {
			return stacktrace.Propagate(err, "An error occurred restarting the Kurtosis engine")
		}
	}

	currentLogrusLevel := logrus.GetLevel()

	_, clientCloseFunc, err := engineManager.StartEngineIdempotently(ctx, defaults.DefaultEngineImage, currentLogrusLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine")
	}
	defer func() {
		if err := clientCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the engine client, but doing so threw an error:\n%v", err)
		}
	}()

	return nil
}
