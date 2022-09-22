package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func StopLogsCollector(ctx context.Context, dockerManager *docker_manager.DockerManager) error {

	allLogsCollectorContainers, err := getAllLogsCollectorContainers(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting all logs collector containers")
	}
	if len(allLogsCollectorContainers) == 0 {
		logrus.Debug("There isn't any logs collector container, so there is nothing to stop")
		return nil
	}

	for _, logsCollectorContainer := range allLogsCollectorContainers {
		if err := dockerManager.StopContainer(ctx, logsCollectorContainer.GetId(), stopLogsCollectorContainersTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the logs collector container with ID '%v'", logsCollectorContainer.GetId())
		}
	}

	return nil
}

