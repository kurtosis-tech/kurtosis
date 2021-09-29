package v0

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/api_container_launcher_lib/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/api_container_launcher_lib/api_container_starter"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

type V0APIContainerLauncher struct {
	dockerManager *docker_manager.DockerManager
	log *logrus.Logger
	containerImage string
	containerLabels map[string]string
	listenPort uint
	listenProtocol string
	logLevel logrus.Level
}

func NewV0APIContainerLauncher(
		dockerManager *docker_manager.DockerManager,
		log *logrus.Logger,
		containerImage string,
		listenPort uint,
		listenProtocol string,
		logLevel logrus.Level) api_container_launcher.APIContainerLauncher {
	return &V0APIContainerLauncher{
		dockerManager: dockerManager,
		log: log,
		containerImage: containerImage,
		listenPort: listenPort,
		listenProtocol: listenProtocol,
		logLevel: logLevel,
	}
}

func (launcher V0APIContainerLauncher) Launch(
		ctx context.Context,
		containerName string,
		containerLabels map[string]string,
		enclaveId string,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		apiContainerIpAddr net.IP,
		otherTakenIpAddrsInEnclave []net.IP,
		isPartitioningEnabled bool,
		externalMountedContainerIds map[string]bool,
		shouldPublishAllPorts bool) (string, error){
	takenIpAddrStrSet := map[string]bool{
		gatewayIpAddr.String(): true,
		apiContainerIpAddr.String(): true,
	}
	for _, takenIp := range otherTakenIpAddrsInEnclave {
		takenIpAddrStrSet[takenIp.String()] = true
	}
	args := newV0LaunchAPIArgs(
		containerName,
		containerLabels,
		launcher.logLevel.String(),
		enclaveId,
		networkId,
		subnetMask,
		apiContainerIpAddr.String(),
		takenIpAddrStrSet,
		isPartitioningEnabled,
		shouldPublishAllPorts,
		externalMountedContainerIds,
	)
	
	containerId, err := api_container_starter.StartAPIContainer(
		ctx,
		launcher.dockerManager,
		launcher.containerImage,
		containerName,
		containerLabels,
		launcher.listenPort,
		launcher.listenProtocol,
		networkId,
		apiContainerIpAddr,
		enclaveId,
		args,
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred starting the API container using the V0 launch API")
	}
	return containerId, nil
}
