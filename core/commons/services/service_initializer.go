package services

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/palantir/stacktrace"
)

// This implicitly is a Docker container-backed service initializer, but we could abstract to other backends if we wanted later
type ServiceInitializer struct {
	core ServiceInitializerCore
}

func NewServiceInitializer(core ServiceInitializerCore) *ServiceInitializer {
	return &ServiceInitializer{
		core: core,
	}
}

// If Go had generics, this would be genericized so that the arg type = return type
func (initializer ServiceInitializer) CreateService(
			dockerImage string,
			staticIp string,
			manager *docker.DockerManager,
			dependencies []Service) (Service, string, error) {
	startCmdArgs := initializer.core.GetStartCommand(staticIp, dependencies)
	usedPorts := initializer.core.GetUsedPorts()

	// TODO mount volumes when we want services to read/write state to disk
	// TODO we really want GetEnvVariables instead of GetStartCmd because every image should be nicely parameterized to avoid
	//   the testing code knowing about the specifics of the image (like where the binary is located). However, this relies
	//   on the service images being parameterized with environment variables.
	ipAddr, containerId, err := manager.CreateAndStartContainer(
			dockerImage,
			staticIp,
			usedPorts,
			startCmdArgs,
			make(map[string]string),
			make(map[string]string))
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Could not start docker service for image %v", dockerImage)
	}
	return initializer.core.GetServiceFromIp(ipAddr), containerId, nil
}

func (initializer ServiceInitializer) LoadService(ipAddr string) Service {
	return initializer.core.GetServiceFromIp(ipAddr)
}
