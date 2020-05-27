package testnet

import (
	"github.com/gmarchetti/kurtosis/commons/docker"
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
			publicIpTracker *FreeIpAddrTracker,
			manager *docker.DockerManager,
			dependencies []Service) (Service, string, error) {
	dockerImage := factory.config.GetDockerImage()
	staticIp, err := publicIpTracker.GetFreeIpAddr()
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Could not allocate static IP address.")
	}
	startCmdArgs := factory.config.GetStartCommand(staticIp, dependencies)
	usedPorts := factory.config.GetUsedPorts()

	ipAddr, containerId, err := manager.CreateAndStartContainerForService(dockerImage, staticIp, usedPorts, startCmdArgs)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Could not start docker service for image %v", dockerImage)
	}
	return factory.config.GetServiceFromIp(ipAddr), containerId, nil
}
