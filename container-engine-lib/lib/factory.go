package lib

import (
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/metrics_reporting"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
)

// GetLocalDockerKurtosisBackend is the entrypoint method we expect users of container-engine-lib to call
func GetLocalDockerKurtosisBackend() (backend_interface.KurtosisBackend, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker client connected to the local environment")
	}

	dockerManager := docker_manager.NewDockerManager(dockerClient)

	dockerKurtosisBackend := docker.NewDockerKurtosisBackend(dockerManager)

	wrappedBackend := metrics_reporting.NewMetricsReportingKurtosisBackend(dockerKurtosisBackend)

	return wrappedBackend, nil
}

// TODO Kubernetes