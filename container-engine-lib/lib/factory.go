package lib

import (
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/backends/docker"
	"github.com/kurtosis-tech/container-engine-lib/lib/backends/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

// GetLocalDockerKurtosisBackend is the entrypoint method we expect users of container-engine-lib to call
func GetLocalDockerKurtosisBackend() (KurtosisBackend, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker client connected to the local environment")
	}

	// TODO Remove the logrus.Logger param; it's no longer needed!
	dockerManager := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)

	dockerKurtosisBackend := docker.NewDockerKurtosisBackend(dockerManager)

	return dockerKurtosisBackend, nil
}

// TODO Kubernetes