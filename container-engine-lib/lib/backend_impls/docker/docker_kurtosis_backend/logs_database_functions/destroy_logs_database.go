package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

func DestroyLogsDatabase(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) error {

	_, maybeLogsDatabaseContainerId, err := getLogsDatabaseObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs database")
	}

	if maybeLogsDatabaseContainerId == "" {
		return nil
	}

	if err := dockerManager.StopContainer(ctx, maybeLogsDatabaseContainerId, stopLogsDatabaseContainersTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs database container with ID '%v'", maybeLogsDatabaseContainerId)
	}

	if err := dockerManager.RemoveContainer(ctx, maybeLogsDatabaseContainerId); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the logs database container with ID '%v'", maybeLogsDatabaseContainerId)
	}

	return nil
}
