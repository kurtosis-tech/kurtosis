package traefik

import (
	"os"
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
)

func createTraefikContainerConfigProvider(httpPort uint16, dashboardPort uint16, networkId string, dockerManager *docker_manager.DockerManager) *traefikContainerConfigProvider {
	config := reverse_proxy.NewDefaultReverseProxyConfig(httpPort, dashboardPort, networkId)
	
	// Determine socket path, prioritizing DOCKER_HOST environment variable
	socketPath := getSocketPath(dockerManager)
	
	return newTraefikContainerConfigProvider(config, socketPath)
}

func getSocketPath(dockerManager *docker_manager.DockerManager) string {
	// Check if DOCKER_HOST environment variable is set
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost != "" {
		// Extract socket path from DOCKER_HOST (e.g., "unix:///path/to/socket" -> "/path/to/socket")
		if strings.HasPrefix(dockerHost, "unix://") {
			return strings.TrimPrefix(dockerHost, "unix://")
		}
		// If DOCKER_HOST is set but not a unix socket, fall back to defaults
	}

	// Fall back to default paths based on Docker/Podman
	if dockerManager.IsPodman() {
		return "/run/podman/podman.sock"
	} else {
		return "/var/run/docker.sock"
	}
}
