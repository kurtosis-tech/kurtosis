package testnet

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
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

// If Go had generics, this would be genericized so that the arg type = return type
func (factory ServiceFactory) Construct(
			staticIp string,
			manager *docker.DockerManager,
			dependencies []Service) (Service, string, error) {
	dockerImage := factory.config.GetDockerImage()
	startCmdArgs := factory.config.GetStartCommand(staticIp, dependencies)
	usedPorts := factory.config.GetUsedPorts()

	ipAddr, containerId, err := manager.CreateAndStartContainer(dockerImage, staticIp, usedPorts, startCmdArgs)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Could not start docker service for image %v", dockerImage)
	}
	return factory.config.GetServiceFromIp(ipAddr), containerId, nil
}
