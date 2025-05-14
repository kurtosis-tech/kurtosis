package reverse_proxy_functions

import (
	"context"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	stopReverseProxyContainerTimeout = 2 * time.Second
)

// Destroys reverse proxy idempotently, returns nil if no reverse proxy reverse proxy container was found
func DestroyReverseProxy(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	usePodmanBridgeNetwork bool,
) error {
	_, maybeReverseProxyContainerId, err := getReverseProxyObjectAndContainerId(ctx, dockerManager, usePodmanBridgeNetwork)
	if err != nil {
		logrus.Warnf("Attempted to destroy reverse proxy but no reverse proxy container was found. Error was:\n%s", err.Error())
		return nil
	}

	if maybeReverseProxyContainerId == "" {
		return nil
	}

	if err := destroyReverseProxyWithContainerId(ctx, dockerManager, maybeReverseProxyContainerId); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the reverse proxy container with ID '%v'", maybeReverseProxyContainerId)
	}

	return nil
}

func destroyReverseProxyWithContainerId(ctx context.Context, dockerManager *docker_manager.DockerManager, reverseProxyContainerId string) error {
	if err := dockerManager.StopContainer(ctx, reverseProxyContainerId, stopReverseProxyContainerTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the reverse proxy container with ID '%v'", reverseProxyContainerId)
	}

	if err := dockerManager.RemoveContainer(ctx, reverseProxyContainerId); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the reverse proxy container with ID '%v'", reverseProxyContainerId)
	}
	return nil
}
