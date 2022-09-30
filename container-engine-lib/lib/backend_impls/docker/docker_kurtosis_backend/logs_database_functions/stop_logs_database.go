package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
)

func StopLogsDatabase(
	ctx context.Context,
	filters *logs_database.LogsDatabaseFilters,
	dockerManager *docker_manager.DockerManager,
) error {

	_, logsDatabaseContainerId, err := getLogsDatabaseObjectAndContainerIdMatching(ctx, filters, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs database using filters '%+v'", filters)
	}

	if logsDatabaseContainerId != "" {
		if err := dockerManager.StopContainer(ctx, logsDatabaseContainerId, stopLogsDatabaseContainersTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the logs database container with ID '%v'", logsDatabaseContainerId)
		}
	}

	return nil
}
