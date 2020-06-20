package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

type serviceConfig struct {
	dockerImage string
	availabilityCheckerCore services.ServiceAvailabilityCheckerCore
	initializer services.ServiceInitializer
}

type ServiceNetworkBuilder struct {
	testImage string

	dockerManager *docker.DockerManager

	freeIpTracker *FreeIpAddrTracker

	// Factories that will be used to construct the nodes
	configurations map[int]serviceConfig

	// Tracks the next service configuration ID that will be doled out upon a call to AddStaticImageConfiguration
	nextConfigurationId int
}

// The test image is the Docker image of the service being tested
func NewServiceNetworkBuilder(testImage string, dockerManager *docker.DockerManager, freeIpTracker *FreeIpAddrTracker) *ServiceNetworkBuilder {
	configurations := make(map[int]serviceConfig)
	return &ServiceNetworkBuilder{
		testImage: 		testImage,
		dockerManager: dockerManager,
		freeIpTracker: freeIpTracker,
		configurations: 	 configurations,
		nextConfigurationId: 0,
	}
}

// Adds a service configuration to the network that will run a static Docker image
// This configuration can be referenced later with AddService
func (builder *ServiceNetworkBuilder) AddStaticImageConfiguration(
			dockerImage string,
			initializerCore services.ServiceInitializerCore,
			availabilityCheckerCore services.ServiceAvailabilityCheckerCore) int {
	serviceConfig := serviceConfig{
		dockerImage: dockerImage,
		availabilityCheckerCore: availabilityCheckerCore,
		initializer:         *services.NewServiceInitializer(initializerCore),
	}
	configurationId := builder.nextConfigurationId
	builder.nextConfigurationId = builder.nextConfigurationId + 1
	builder.configurations[configurationId] = serviceConfig
	return configurationId
}

// Adds a service configuration to the network that will run the Docker image being tested
// This configuration can be referenced later with AddService
func (builder *ServiceNetworkBuilder) AddTestImageConfiguration(
			initializerCore services.ServiceInitializerCore,
			availabilityCheckerCore services.ServiceAvailabilityCheckerCore) int {
	return builder.AddStaticImageConfiguration(builder.testImage, initializerCore, availabilityCheckerCore)
}

func (builder ServiceNetworkBuilder) Build() *ServiceNetwork {
	// Defensive copy, so user calling functions on the builder after building won't affect the
	// state of the object we already built
	configurationsCopy := make(map[int]serviceConfig)
	for configurationId, config := range builder.configurations {
		configurationsCopy[configurationId] = config
	}

	return &ServiceNetwork{
		freeIpTracker:  builder.freeIpTracker,
		serviceNodes:   make(map[int]ServiceNode),
		configurations: configurationsCopy,
	}
}
