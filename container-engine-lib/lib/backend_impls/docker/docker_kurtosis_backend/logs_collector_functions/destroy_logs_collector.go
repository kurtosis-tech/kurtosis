package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

func DestroyLogsCollector(
	ctx context.Context,
	filters *logs_collector.LogsCollectorFilters,
	dockerManager *docker_manager.DockerManager,
) error {

	_, logsCollectorContainerId, err := getLogsCollectorObjectAndContainerIdMatching(ctx, filters, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs collector using filters '%+v'", filters)
	}

	if logsCollectorContainerId != "" {
		if err := dockerManager.StopContainer(ctx, logsCollectorContainerId, stopLogsCollectorContainersTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the logs collector container with ID '%v'", logsCollectorContainerId)
		}

		if err := dockerManager.RemoveContainer(ctx, logsCollectorContainerId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing the logs collector container with ID '%v'", logsCollectorContainerId)
		}
	}
	return nil
}
