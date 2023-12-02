package reverse_proxy_functions

import (
	"context"
	"net"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	emptyAliasForLogsCollector = ""
)

var (
	autoAssignIpAddressToLogsCollector net.IP = nil
)

func ConnectReverseProxyToNetwork(ctx context.Context, dockerManager *docker_manager.DockerManager, networkId string) error {
	_, maybeReverseProxyContainerId, err := getReverseProxyObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		logrus.Warnf("Attempted to connect reverse proxy to a network but no reverse proxy container was found.")
		return nil
	}

	if maybeReverseProxyContainerId == "" {
		return nil
	}

	if err = dockerManager.ConnectContainerToNetwork(ctx, networkId, maybeReverseProxyContainerId, autoAssignIpAddressToLogsCollector, emptyAliasForLogsCollector); err != nil {
		return stacktrace.Propagate(err, "An error occurred while connecting container '%v' to the enclave network '%v'", maybeReverseProxyContainerId, networkId)
	}

	return nil
}

func DisconnectReverseProxyFromNetwork(ctx context.Context, dockerManager *docker_manager.DockerManager, networkId string) error {
	_, maybeReverseProxyContainerId, err := getReverseProxyObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		logrus.Warnf("Attempted to disconnect reverse proxy from a network but no reverse proxy container was found.")
		return nil
	}

	if maybeReverseProxyContainerId == "" {
		return nil
	}

	if err = dockerManager.DisconnectContainerFromNetwork(ctx, maybeReverseProxyContainerId, networkId); err != nil {
		return stacktrace.Propagate(err, "An error occurred while disconnecting container '%v' from the enclave network '%v'", maybeReverseProxyContainerId, networkId)
	}

	return nil
}

func DisconnectReverseProxyFromEnclaveNetworks(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaveNetworkIds map[enclave.EnclaveUUID]string,
) (
	map[enclave.EnclaveUUID]bool,
	map[enclave.EnclaveUUID]error,
	error,
) {
	networkIdsToRemove := map[string]bool{}
	enclaveUuidsForNetworkIds := map[string]enclave.EnclaveUUID{}
	for enclaveUuid, networkId := range enclaveNetworkIds {
		networkIdsToRemove[networkId] = true
		enclaveUuidsForNetworkIds[networkId] = enclaveUuid
	}

	var disconnectNetworkOperation docker_operation_parallelizer.DockerOperation = func(ctx context.Context, dockerManager *docker_manager.DockerManager, dockerObjectId string) error {
		if err := DisconnectReverseProxyFromNetwork(ctx, dockerManager, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred disconnecting the reverse proxy from the enclave network with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulNetworkIds, erroredNetworkIds := docker_operation_parallelizer.RunDockerOperationInParallel(
		ctx,
		networkIdsToRemove,
		dockerManager,
		disconnectNetworkOperation,
	)

	successfulEnclaveUuids := map[enclave.EnclaveUUID]bool{}
	for networkId := range successfulNetworkIds {
		enclaveUuid, found := enclaveUuidsForNetworkIds[networkId]
		if !found {
			return nil, nil, stacktrace.NewError("The reverse proxy was successfully disconnected from the Docker network '%v', but wasn't requested to be disconnected", networkId)
		}
		successfulEnclaveUuids[enclaveUuid] = true
	}

	erroredEnclaveUuids := map[enclave.EnclaveUUID]error{}
	for networkId, networkRemovalErr := range erroredNetworkIds {
		enclaveUuid, found := enclaveUuidsForNetworkIds[networkId]
		if !found {
			return nil, nil, stacktrace.NewError("Docker network '%v' had the following error during disconnect, but wasn't requested to be disconnected:\n%v", networkId, networkRemovalErr)
		}
		erroredEnclaveUuids[enclaveUuid] = networkRemovalErr
	}
	return successfulEnclaveUuids, erroredEnclaveUuids, nil
}
