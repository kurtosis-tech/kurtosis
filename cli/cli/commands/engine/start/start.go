package start

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
	"time"
)

const (
	engineImageArg = "image"

	shouldGetStoppedContainersWhenCheckingForExistingEngines = false

	engineWaitForReadyTimeout = 10 * time.Second

	engineImageInfoLabel = "Image"
	engineApiVersionInfoLabel = "API Version"
)

var engineImage string

var StartCmd = &cobra.Command{
	Use:   command_str_consts.EngineStartCmdStr,
	Short: "Starts the Kurtosis engine",
	Long: "Starts the Kurtosis engine, doing nothing if an engine is already running",
	RunE:  run,
}

func init() {
	StartCmd.Flags().StringVar(
		&engineImage,
		engineImageArg,
		defaults.DefaultEngineImage,
		"The image of the Kurtosis engine that should be started",
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logrus.Infof("Starting Kurtosis engine from image '%v'...", engineImage)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineManager := engine_manager.NewEngineManager(dockerManager)
	_, clientCloseFunc, err := engineManager.StartEngineIdempotently(ctx, defaults.DefaultEngineImage)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine")
	}
	defer clientCloseFunc()

	logrus.Info("Kurtosis engine started")

	return nil
}