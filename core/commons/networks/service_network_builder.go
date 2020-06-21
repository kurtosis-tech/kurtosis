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

type ServiceNetworkBuilder struct {
	testImage string

	dockerManager *docker.DockerManager

	freeIpTracker *FreeIpAddrTracker

	// Factories that will be used to construct the nodes
	configurations map[int]serviceConfig

	// Name of volume that will be mounted on each new service
	testVolume string

	// Location where the test volume is mounted on the controller
	testVolumeControllerDirpath string
}

// The test image is the Docker image of the service being tested
func NewServiceNetworkBuilder(
			testImage string,
			dockerManager *docker.DockerManager,
			freeIpTracker *FreeIpAddrTracker,
			testVolume string,
			testVolumeContrllerDirpath string) *ServiceNetworkBuilder {
	configurations := make(map[int]serviceConfig)
	return &ServiceNetworkBuilder{
		testImage:                   testImage,
		dockerManager:               dockerManager,
		freeIpTracker:               freeIpTracker,
		configurations:              configurations,
		testVolume:                  testVolume,
		testVolumeControllerDirpath: testVolumeContrllerDirpath,
	}
}

// Adds a service configuration to the network that will run a static Docker image
// This configuration can be referenced later with AddService
func (builder *ServiceNetworkBuilder) AddStaticImageConfiguration(
			configurationId int,
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

// Adds a service configuration to the network that will run the Docker image being tested
// This configuration can be referenced later with AddService
func (builder *ServiceNetworkBuilder) AddTestImageConfiguration(
			configurationId int,
			initializerCore services.ServiceInitializerCore,
			availabilityCheckerCore services.ServiceAvailabilityCheckerCore) error {
	return builder.AddStaticImageConfiguration(configurationId, builder.testImage, initializerCore, availabilityCheckerCore)
}

func (builder ServiceNetworkBuilder) Build() *ServiceNetwork {
	// Defensive copy, so user calling functions on the builder after building won't affect the
	// state of the object we already built
	configurationsCopy := make(map[int]serviceConfig)
	for configurationId, config := range builder.configurations {
		configurationsCopy[configurationId] = config
	}
	return NewServiceNetwork(
		builder.freeIpTracker,
		builder.dockerManager,
		make(map[int]ServiceNode),
		configurationsCopy,
		builder.testVolume,
		builder.testVolumeControllerDirpath)
}
