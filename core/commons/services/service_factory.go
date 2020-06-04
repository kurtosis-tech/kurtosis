package services

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
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

	filepathsToMount := factory.config.GetFilepathsToMount()
	logrus.Debugf("Filepaths to mount: %+v", filepathsToMount)
	// TODO create a temp file on the parent host, just like we do for the controller's network info file
	// TODO call factory.config.InitializeMountedFiles to fill in the file contents (closing the temporary file after)

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
			// TODO pass in the mappings for each tempfile -> user-desired mount location
			make(map[string]string))
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Could not start docker service for image %v", dockerImage)
	}
	return factory.config.GetServiceFromIp(ipAddr), containerId, nil
}
