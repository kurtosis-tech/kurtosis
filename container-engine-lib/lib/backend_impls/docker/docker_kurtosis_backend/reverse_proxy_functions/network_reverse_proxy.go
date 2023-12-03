package reverse_proxy_functions

import (
	"context"
	"net"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
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
	_, maybeReverseProxyContainerId, err := getReverseProxyObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		logrus.Warnf("Attempted to connect reverse proxy to a network but no reverse proxy container was found.")
		return nil
	}

	if maybeReverseProxyContainerId == "" {
		return nil
	}

	if err = dockerManager.ConnectContainerToNetwork(ctx, networkId, maybeReverseProxyContainerId, autoAssignIpAddressToReverseProxy, emptyAliasForReverseProxy); err != nil {
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
