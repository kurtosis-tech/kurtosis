package status

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/engine_status_retriever"
	"github.com/kurtosis-tech/kurtosis-cli/cli/output_printers"
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

	status, apiVersion, err := engine_status_retriever.RetrieveEngineStatus(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't get the Kurtosis engine status")
	}

	switch status {
	case engine_status_retriever.EngineStatus_Stopped:
		fmt.Fprintln(logrus.StandardLogger().Out, "No Kurtosis engine is not running")
	case engine_status_retriever.EngineStatus_ContainerRunningButServerNotResponding:
		fmt.Fprintln(logrus.StandardLogger().Out, "A Kurtosis engine container is running, but the server inside couldn't be reached")
	case engine_status_retriever.EngineStatus_Running:
		engineInfoPrinter := output_printers.NewKeyValuePrinter()
		engineInfoPrinter.AddPair(engineApiVersionInfoLabel, apiVersion)

		fmt.Fprintln(logrus.StandardLogger().Out, "A Kurtosis engine is running with the following info:")
		engineInfoPrinter.Print()
	default:
		return stacktrace.NewError("Unhandled engine status '%v'; this is a bug in Kurtosis", status)
	}

	return nil
}