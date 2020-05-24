package commons

import (
	"github.com/docker/docker/api/types"
	"github.com/palantir/stacktrace"
)

// This implicitly is a Docker container factory, but we could abstract to other backends if we wanted later
type ServiceFactory struct {
	config ServiceFactoryConfig
}

func NewServiceFactory(config ServiceFactoryConfig) *ServiceFactory {
	return &ServiceFactory{
		config: config,
	}
}

// TODO needing to pass in hte ipAddrOffset is a nasty awful hack here that will go away when the --public-ips flag is gone!
// If Go had generics, this would be genericized so that the arg type = return type
func (factory ServiceFactory) Construct(
			ipAddrOffset int,
			manager *DockerManager,
			dependencies []Service) (Service, string, error) {
	dockerImage := factory.config.GetDockerImage()
	startCmdArgs := factory.config.GetStartCommand(ipAddrOffset, dependencies)
	usedPorts := factory.config.GetUsedPorts()

	containerConfigPtr, err := manager.GetContainerCfgFromServiceCfg(dockerImage, usedPorts, startCmdArgs)

	containerHostConfigPtr, err := manager.GetContainerHostConfig(usedPorts)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "")
	}
	// TODO probably use a UUID for the network name (and maybe include test name too)
	resp, err := manager.dockerClient.ContainerCreate(manager.dockerCtx, containerConfigPtr, containerHostConfigPtr, nil, "")
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Could not create Docker container from image %v.", dockerImage)
	}
	containerId := resp.ID
	if err := manager.dockerClient.ContainerStart(manager.dockerCtx, containerId, types.ContainerStartOptions{}); err != nil {
		return nil, "", stacktrace.Propagate(err, "Could not start Docker container from image %v.", dockerImage)
	}

	containerJson, err := manager.dockerClient.ContainerInspect(manager.dockerCtx, containerId)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Inspect container failed, which is necessary to get the container's IP")
	}
	containerIpAddr := containerJson.NetworkSettings.IPAddress
	return factory.config.GetServiceFromIp(containerIpAddr), containerId, nil
}
