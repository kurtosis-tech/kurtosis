package status

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
)

var StatusCmd = &cobra.Command{
	Use:   command_str_consts.EngineStatusCmdStr,
	Short: "Reports the status of the Kurtosis engine",
	RunE:  run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineManager := engine_manager.NewEngineManager(dockerManager)
	status, _, maybeApiVersion, err := engineManager.GetEngineStatus(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis engine status")
	}
	prettyPrintingStatusVisitor := newPrettyPrintingEngineStatusVisitor(maybeApiVersion)
	if err := status.Accept(prettyPrintingStatusVisitor); err != nil {
		return stacktrace.Propagate(err, "An error occurred printing the engine status")
	}

	return nil
}
