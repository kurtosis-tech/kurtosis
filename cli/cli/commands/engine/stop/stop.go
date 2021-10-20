package stop

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/engine_labels_schema"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
	"time"
)


const (
	engineStopTimeout = 30 * time.Second

	shouldGetStoppedContainersWhenCheckingForExistingEngines = false
)

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the Kurtosis engine",
	Long: "Stops the Kurtosis engine, doing nothing if no engine is running",
	RunE:  run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	logrus.Infof("Stopping Kurtosis engine...")

	ctx := context.Background()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	matchingEngineContainers, err := dockerManager.GetContainersByLabels(
		ctx,
		engine_labels_schema.EngineContainerLabels,
		shouldGetStoppedContainersWhenCheckingForExistingEngines,
	)
	numMatchingEngineContainers := len(matchingEngineContainers)
	if numMatchingEngineContainers == 0 {
		logrus.Info("No Kurtosis engine is running; nothing to do")
	}
	if numMatchingEngineContainers > 1 {
		logrus.Warnf(
			"Found %v Kurtosis engine containers, which is strange because there should never be more than 1 engine container; all will be stopped",
			numMatchingEngineContainers,
		)
	}

	engineStopErrorStrs := []string{}
	for _, engineContainer := range matchingEngineContainers {
		containerName := engineContainer.GetName()
		containerId := engineContainer.GetId()
		if err := dockerManager.StopContainer(ctx, containerId, engineStopTimeout); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred stopping engine container '%v' with ID '%v'",
				containerName,
				containerId,
			)
			engineStopErrorStrs = append(engineStopErrorStrs, wrappedErr.Error())
		}
	}

	if len(engineStopErrorStrs) > 0 {
		return stacktrace.NewError(
			"One or more errors occurred stopping the engine(s):\n%v",
			strings.Join(
				engineStopErrorStrs,
				"\n\n",
			),
		)
	}

	logrus.Info("Kurtosis engine successfully stopped")
	return nil
}
