package logs_collector_functions

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

// If nothing is found returns nil
func GetLogsCollectorForEnclave(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	dockerManager *docker_manager.DockerManager,
	usePodmanBridgeNetwork bool,
) (
	resultMaybeLogsCollector *logs_collector.LogsCollector,
	resultErr error,
) {
	enclaveNetworkId, err := shared_helpers.GetEnclaveNetworkByEnclaveUuid(ctx, enclaveUuid, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while retrieving the network id for the enclave '%v'", enclaveUuid)
	}

	maybeLogsCollectorObject, _, err := getLogsCollectorObjectAndContainerId(ctx, enclaveUuid, enclaveNetworkId, dockerManager, usePodmanBridgeNetwork)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	return maybeLogsCollectorObject, nil
}
