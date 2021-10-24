package status

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	output_printers2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	engineApiVersionInfoLabel = "API Version"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
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

	status, apiVersion, err := engine_manager.GetEngineStatus(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't get the Kurtosis engine status")
	}

	switch status {
	case engine_manager.EngineStatus_Stopped:
		fmt.Fprintln(logrus.StandardLogger().Out, "No Kurtosis engine is running")
	case engine_manager.EngineStatus_ContainerRunningButServerNotResponding:
		fmt.Fprintln(logrus.StandardLogger().Out, "A Kurtosis engine container is running, but the server inside couldn't be reached")
	case engine_manager.EngineStatus_Running:
		engineInfoPrinter := output_printers2.NewKeyValuePrinter()
		engineInfoPrinter.AddPair(engineApiVersionInfoLabel, apiVersion)

		fmt.Fprintln(logrus.StandardLogger().Out, "A Kurtosis engine is running with the following info:")
		engineInfoPrinter.Print()
	default:
		return stacktrace.NewError("Unhandled engine status '%v'; this is a bug in Kurtosis", status)
	}

	return nil
}