package reverse_proxy_functions

import (
	"context"
	"net"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	emptyAliasForReverseProxy = ""
)

var (
	autoAssignIpAddressToReverseProxy net.IP = nil
)

func ConnectReverseProxyToNetwork(ctx context.Context, dockerManager *docker_manager.DockerManager, networkId string) error {
	maybeReverseProxyObject, maybeReverseProxyContainerId, err := getReverseProxyObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		logrus.Warnf("Attempted to connect reverse proxy to a network but no reverse proxy container was found.")
		return nil
	}

	if maybeReverseProxyObject == nil {
		return nil
	}

	_, found := maybeReverseProxyObject.GetEnclaveNetworksIpAddress()[networkId]
	if found {
		return nil
	}
	
	if err = dockerManager.ConnectContainerToNetwork(ctx, networkId, maybeReverseProxyContainerId, autoAssignIpAddressToReverseProxy, emptyAliasForReverseProxy); err != nil {
		return stacktrace.Propagate(err, "An error occurred while connecting container '%v' to the enclave network '%v'", maybeReverseProxyContainerId, networkId)
	}

	return nil
}

func DisconnectReverseProxyFromNetwork(ctx context.Context, dockerManager *docker_manager.DockerManager, networkId string) error {
	maybeReverseProxyObject, maybeReverseProxyContainerId, err := getReverseProxyObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		logrus.Warnf("Attempted to disconnect reverse proxy from a network but no reverse proxy container was found.")
		return nil
	}

	if maybeReverseProxyContainerId == "" {
		return nil
	}

	_, found := maybeReverseProxyObject.GetEnclaveNetworksIpAddress()[networkId]
	if !found {
		return nil
	}
	
	if err = dockerManager.DisconnectContainerFromNetwork(ctx, maybeReverseProxyContainerId, networkId); err != nil {
		return stacktrace.Propagate(err, "An error occurred while disconnecting container '%v' from the enclave network '%v'", maybeReverseProxyContainerId, networkId)
	}

	return nil
}

func ConnectReverseProxyToEnclaveNetworks(ctx context.Context, dockerManager *docker_manager.DockerManager) error {
	kurtosisNetworkLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString(): label_value_consts.AppIDDockerLabelValue.GetString(),
	}
	enclaveNetworks, err := dockerManager.GetNetworksByLabels(ctx, kurtosisNetworkLabels)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave networks")
	}

	for _, enclaveNetwork := range enclaveNetworks {
		if err = ConnectReverseProxyToNetwork(ctx, dockerManager, enclaveNetwork.GetId()); err != nil {
			return stacktrace.Propagate(err, "An error occurred connecting the reverse proxy to the enclave network with id '%v'", enclaveNetwork.GetId())
		}
	}

	return nil
}
