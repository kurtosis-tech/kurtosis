package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func StopLogsDatabase(ctx context.Context, dockerManager *docker_manager.DockerManager) error {

	allLogsDatabaseContainers, err := getAllLogsDatabaseContainers(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting all logs database containers")
	}
	if len(allLogsDatabaseContainers) == 0 {
		logrus.Debug("There isn't any logs database container, so there is nothing to stop")
		return nil
	}

	for _, logsDatabaseContainer := range allLogsDatabaseContainers {
		if err := dockerManager.StopContainer(ctx, logsDatabaseContainer.GetId(), stopLogsDatabaseContainersTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the logs database container with ID '%v'", logsDatabaseContainer.GetId())
		}
	}

	return nil
}
