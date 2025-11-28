package reverse_proxy_functions

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
	"github.com/kurtosis-tech/stacktrace"
)

func GetReverseProxy(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (*reverse_proxy.ReverseProxy, error) {
	maybeReverseProxyObject, _, err := getReverseProxyObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the reverse proxy")
	}

	return maybeReverseProxyObject, nil
}
