package services

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	TIME_BETWEEN_STARTUP_POLLS = 1 * time.Second
)

// TODO Rename to ServiceInitializer
// This implicitly is a Docker container factory, but we could abstract to other backends if we wanted later
type ServiceFactory struct {
	config ServiceFactoryConfig
}

func NewServiceFactory(config ServiceFactoryConfig) *ServiceFactory {
	return &ServiceFactory{
		config: config,
	}
}

// TODO Rename to NewInstance
// If Go had generics, this would be genericized so that the arg type = return type
func (factory ServiceFactory) Construct(
			staticIp string,
			manager *docker.DockerManager,
			dependencies []Service) (Service, string, error) {
	dockerImage := factory.config.GetDockerImage()
	startCmdArgs := factory.config.GetStartCommand(staticIp, dependencies)
	usedPorts := factory.config.GetUsedPorts()

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
	return factory.config.GetServiceFromIp(ipAddr), containerId, nil
}

// Waits for the given service to start up by making requests (configured by the core) to the service until the service
//  is reported as up or the timeout is reached
func (factory ServiceFactory) WaitForStartup(toCheck Service, dependencies []Service) error {
	startupTimeout := factory.config.GetStartupTimeout()
	pollStartTime := time.Now()
	for time.Since(pollStartTime) < startupTimeout {
		if factory.config.IsServiceUp(toCheck, dependencies) {
			return nil
		}
		logrus.Tracef("Service is not yet available; sleeping for %v before retrying...", TIME_BETWEEN_STARTUP_POLLS)
		time.Sleep(TIME_BETWEEN_STARTUP_POLLS)
	}
	return stacktrace.NewError("Hit timeout (%v) while waiting for service to start", startupTimeout)
}
