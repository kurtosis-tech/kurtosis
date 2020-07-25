package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

type serviceConfig struct {
	dockerImage string
	availabilityCheckerCore services.ServiceAvailabilityCheckerCore
	initializerCore services.ServiceInitializerCore
}

type ConfigurationID int

type ServiceNetworkBuilder struct {
	dockerManager *docker.DockerManager

	dockerNetworkName string

	freeIpTracker *FreeIpAddrTracker

	// Factories that will be used to construct the nodes
	configurations map[ConfigurationID]serviceConfig

	// Name of volume that will be mounted on each new service
	testVolume string

	// Location where the test volume is mounted on the controller
	testVolumeControllerDirpath string
}

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

// Adds a service configuration to the network that will run a static Docker image
// This configuration can be referenced later with AddService
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
		make(map[ServiceID]ServiceNode),
		configurationsCopy,
		builder.testVolume,
		builder.testVolumeControllerDirpath)
}
