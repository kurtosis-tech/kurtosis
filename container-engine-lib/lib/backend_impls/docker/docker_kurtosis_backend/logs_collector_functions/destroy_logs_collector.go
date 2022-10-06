package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

func DestroyLogsCollector(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) error {

	_, maybeLogsCollectorContainerId, err := getLogsCollectorObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	if maybeLogsCollectorContainerId == "" {
		return nil
	}

	if err := dockerManager.StopContainer(ctx, maybeLogsCollectorContainerId, stopLogsCollectorContainersTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs collector container with ID '%v'", maybeLogsCollectorContainerId)
	}

	// TODO Allow removeContainer to gracefully stop
	if err := dockerManager.RemoveContainer(ctx, maybeLogsCollectorContainerId); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the logs collector container with ID '%v'", maybeLogsCollectorContainerId)
	}

	return nil
}
