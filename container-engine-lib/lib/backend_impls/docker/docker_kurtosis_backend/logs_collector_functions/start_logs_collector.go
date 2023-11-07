package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
)

func StartLogsCollectorForEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, dockerManager *docker_manager.DockerManager, ) error {

	_, logsCollectorContainerId, err := getLogsCollectorObjectAndContainerId(ctx, enclaveUuid, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting logs collector container ID for enclave '%v'", enclaveUuid)
	}

	if err := dockerManager.StartContainer(ctx, logsCollectorContainerId); err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the logs collector container with ID '%v'", logsCollectorContainerId)
	}

	return nil
}
