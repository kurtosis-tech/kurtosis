package logs_collector_functions

import (
	"context"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
)

// DestroyLogsCollector Destroys the logs collector and its volume
func DestroyLogsCollector(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	dockerManager *docker_manager.DockerManager,
) error {

	enclaveNetworkId, err := shared_helpers.GetEnclaveNetworkByEnclaveUuid(ctx, enclaveUuid, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while retrieving the network id for the enclave")
	}

	_, maybeLogsCollectorContainerId, err := getLogsCollectorObjectAndContainerId(ctx, enclaveUuid, enclaveNetworkId, dockerManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs collector for enclave '%v'", enclaveUuid)
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

	maybeLogsCollectorVolumeName, err := getEnclaveLogsCollectorVolumeName(ctx, dockerManager, enclaveUuid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting logs collector volume for  enclave '%v'", enclaveUuid)
	}

	if maybeLogsCollectorVolumeName == "" {
		return nil
	}

	err = dockerManager.RemoveVolume(ctx, maybeLogsCollectorVolumeName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing the logs collector volume '%v' for enclave '%v'", maybeLogsCollectorVolumeName, enclaveUuid)
	}

	return nil
}
