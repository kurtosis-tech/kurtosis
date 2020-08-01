package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

// Identifier used for service configurations
type ConfigurationID int

/*
A builder for configuring & constructing a test ServiceNetwork.
 */
type ServiceNetworkBuilder struct {
	// The Docker manager that will be used for manipulating the Docker engine during the test
	dockerManager *docker.DockerManager

	// The name of the Docker network that the test network runs in
	dockerNetworkName string

	// IP address tracker for doling out IPs to new services in the test network
	freeIpTracker *FreeIpAddrTracker

	// Mapping of configuration ID -> factories used to construct new nodes
	configurations map[ConfigurationID]serviceConfig

	// Name of the Docker volume that will be mounted on each new service
	testVolume string

	// Directory path where the test Docker volume is mounted on the controller
	testVolumeControllerDirpath string
}

/*
Creates a new builder for configuring a ServiceNetwork.

Args:
	dockerManager: Docker manager that will be used to manipulate the Docker engine when adding services
	dockerNetworkName: Name of the Docker network that the test network is running in
	freeIpTracker: IP tracker for doling out IPs to new services that will be added to the network
	testVolume: Name of the Docker volume mounted on the controller, that will be mounted on every service
	testVolumeControllerDirpath: The dirpath where the test volume is mounted on the controller (which is where this code
		will be executing)
 */
func NewServiceNetworkBuilder(
			dockerManager *docker.DockerManager,
			dockerNetworkName string,
			freeIpTracker *FreeIpAddrTracker,
			testVolume string,
			testVolumeContrllerDirpath string) *ServiceNetworkBuilder {
	configurations := make(map[ConfigurationID]serviceConfig)
	return &ServiceNetworkBuilder{
		dockerManager:               dockerManager,
		dockerNetworkName:           dockerNetworkName,
		freeIpTracker:               freeIpTracker,
		configurations:              configurations,
		testVolume:                  testVolume,
		testVolumeControllerDirpath: testVolumeContrllerDirpath,
	}
}

/*
Defines a new service configuration to the network that can later be used to launch Docker containers

Args:
	configurationId: The ID by which this configuration will be referenced later
	dockerImage: The Docker image that containers launched with this configuration will run with
	initializerCore: The user-defined logic for how to launch the Docker container
	availabilityCheckerCore: The user-defined logic for how to report services launched with this configuration
		as available
 */
func (builder *ServiceNetworkBuilder) AddConfiguration(
			configurationId ConfigurationID,
			dockerImage string,
			initializerCore services.ServiceInitializerCore,
			availabilityCheckerCore services.ServiceAvailabilityCheckerCore) error {
	if _, found := builder.configurations[configurationId]; found {
		return stacktrace.NewError("Configuration ID %v is already registered", configurationId)
	}

	serviceConfig := serviceConfig{
		dockerImage: dockerImage,
		availabilityCheckerCore: availabilityCheckerCore,
		initializerCore:         initializerCore,
	}
	builder.configurations[configurationId] = serviceConfig
	return nil
}

/*
Constructs a ServiceNetwork with the configurations that were defined for this builder
 */
func (builder ServiceNetworkBuilder) Build() *ServiceNetwork {
	// Defensive copy, so user calling functions on the builder after building won't affect the
	// state of the object we already built
	configurationsCopy := make(map[ConfigurationID]serviceConfig)
	for configurationId, config := range builder.configurations {
		configurationsCopy[configurationId] = config
	}
	return NewServiceNetwork(
		builder.freeIpTracker,
		builder.dockerManager,
		builder.dockerNetworkName,
		configurationsCopy,
		builder.testVolume,
		builder.testVolumeControllerDirpath)
}
