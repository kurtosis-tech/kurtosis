package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/stacktrace"
)

func DestroyLogsDatabase(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) error {

	maybeLogsDatabase, maybeLogsDatabaseContainerId, err := getLogsDatabaseObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs database")
	}

	if maybeLogsDatabase == nil || maybeLogsDatabaseContainerId == ""{
		return nil
	}

	//Do not destroy a running logs database if there is a logs collector running because it can be connected to it
	if maybeLogsDatabase.GetStatus() == container_status.ContainerStatus_Running {
		logsCollector, err := logs_collector_functions.GetLogsCollector(ctx, dockerManager)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the logs collector")
		}

		if logsCollector != nil && logsCollector.GetStatus() == container_status.ContainerStatus_Running {
			return stacktrace.NewError("The logs database can't be destroyed because the logs collector is still running, meaning it wouldn't have anywhere to send the logs to")
		}
	}

	if err := dockerManager.StopContainer(ctx, maybeLogsDatabaseContainerId, stopLogsDatabaseContainersTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs database container with ID '%v'", maybeLogsDatabaseContainerId)
	}

	if err := dockerManager.RemoveContainer(ctx, maybeLogsDatabaseContainerId); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the logs database container with ID '%v'", maybeLogsDatabaseContainerId)
	}

	return nil
}
