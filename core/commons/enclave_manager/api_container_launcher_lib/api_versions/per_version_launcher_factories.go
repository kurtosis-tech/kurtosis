package api_versions

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_manager/api_container_launcher_lib/api_container_launcher"
	v02 "github.com/kurtosis-tech/kurtosis-core/commons/enclave_manager/api_container_launcher_lib/api_versions/v0"
	"github.com/sirupsen/logrus"
)

type apiContainerLauncherFactory = func(
	dockerManager *docker_manager.DockerManager,
	log *logrus.Logger,
	containerImage string,
	listenPort uint,
	listenProtocol string,
	logLevel logrus.Level,
) api_container_launcher.APIContainerLauncher

// The array index here indicates the version
var PerAPIVersionLauncherFactories = []apiContainerLauncherFactory{
	v02.NewV0APIContainerLauncher,
}
