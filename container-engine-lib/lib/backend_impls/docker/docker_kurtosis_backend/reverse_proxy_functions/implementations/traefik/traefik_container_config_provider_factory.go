package traefik

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
)

func createTraefikContainerConfigProvider(httpPort uint16, dashboardPort uint16, networkId string, dockerManager *docker_manager.DockerManager) *traefikContainerConfigProvider {
	config := reverse_proxy.NewDefaultReverseProxyConfig(httpPort, dashboardPort, networkId)

	// Determine socket path using the shared helper function
	socketPath := shared_helpers.GetDockerSocketPath(dockerManager)

	return newTraefikContainerConfigProvider(config, socketPath)
}
