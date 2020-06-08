package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type serviceConfig struct {
	// The Docker image that will be used to run the given service
	// Nil has a special meaning of "use the image being tested"
	dockerImage *string
	availabilityChecker services.ServiceAvailabilityChecker
	initializer services.ServiceInitializer
}

// Builder to ease the declaration of the network state we want
type ServiceNetworkConfigBuilder struct {
	// Maps a node to the configuration that will be used to construct it
	serviceConfigs map[int]int

	// Maps service_id -> set(ids that the service depends on)
	// The 'map' value is only because Go doesn't have a Set type
	serviceDependencies map[int]map[int]bool

	// Ordering in which to start nodes to guarantee we start the graph respecting dependencies
	servicesStartOrder []int

	// All services which aren't depended on by any other service, indicating that these are the last nodes to start up
	// and, when they're all up, the entire network is ready
	onlyDependentServices map[int]bool

	// Factories that will be used to construct the nodes at build time
	configurations map[int]serviceConfig

	// Tracks the next service configuration ID that will be doled out upon a call to AddStaticImageConfiguration
	nextConfigurationId int

}

func NewServiceNetworkConfigBuilder() *ServiceNetworkConfigBuilder {
	// TODO use aliased struct types to make it clear which IDs are which
	serviceConfigs := make(map[int]int)
	serviceDependencies := make(map[int]map[int]bool)
	serviceStartOrder := make([]int, 0)
	onlyDependentServices := make(map[int]bool)
	configurations := make(map[int]serviceConfig)
	return &ServiceNetworkConfigBuilder{
		serviceConfigs:      serviceConfigs,
		serviceDependencies: serviceDependencies,
		servicesStartOrder:  serviceStartOrder,
		onlyDependentServices: onlyDependentServices,
		configurations: 	 configurations,
		nextConfigurationId: 0,
	}
}

// Adds a service configuration to the network that will run a static Docker image
// This configuration can be referenced later with AddService
func (builder *ServiceNetworkConfigBuilder) AddStaticImageConfiguration(
		dockerImage *string,
		initializerCore services.ServiceInitializerCore,
		availabilityCheckerCore services.ServiceAvailabilityCheckerCore) int {
	serviceConfig := serviceConfig{
		dockerImage: dockerImage,
		availabilityChecker: *services.NewServiceAvailabilityChecker(availabilityCheckerCore),
		initializer:         *services.NewServiceInitializer(initializerCore),
	}
	configurationId := builder.nextConfigurationId
	builder.nextConfigurationId = builder.nextConfigurationId + 1
	builder.configurations[configurationId] = serviceConfig
	return configurationId
}

// Adds a service configuration to the network that will run the Docker image being tested
// This configuration can be referenced later with AddService
func (builder *ServiceNetworkConfigBuilder) AddTestImageConfiguration(
		initializerCore services.ServiceInitializerCore,
		availabilityCheckerCore services.ServiceAvailabilityCheckerCore) int {
	return builder.AddStaticImageConfiguration(nil, initializerCore, availabilityCheckerCore)
}

// Adds a serivce to the graph, with the specified dependencies (with the map used only as a set - the values are ignored)
// Returns the ID of the service, to be used with future AddService calls to declare dependencies on the service
// If no dependencies should be specified, the dependencies map should be empty (not nil)
func (builder *ServiceNetworkConfigBuilder) AddService(networkConfigurationId int, serviceId int, dependencies map[int]bool) (int, error) {
	if _, found := builder.configurations[networkConfigurationId]; !found {
		return 0, stacktrace.NewError("No configuration with ID '%v' has been registered", networkConfigurationId)
	}

	if dependencies == nil {
		return 0, stacktrace.NewError("Dependencies map was nil; use an empty map to specify no dependencies")
	}

	// Golang maps are passed by-ref, so we do a defensive copy here so user can't change their input and mess
	// with our internal data structure
	dependenciesCopy := make(map[int]bool)
	for dependencyId, _ := range dependencies  {
		if _, found := builder.serviceConfigs[dependencyId]; !found {
			return 0, stacktrace.NewError("Declared a dependency on %v but no service with this ID has been registered", dependencyId)
		}
		dependenciesCopy[dependencyId] = true
	}

	builder.serviceConfigs[serviceId] = networkConfigurationId

	builder.onlyDependentServices[serviceId] = true
	for dependencyId, _ := range dependencies {
		// This is safe to do even if the key doesn't exist (i.e. another previously-declared service also depends on it)
		delete(builder.onlyDependentServices, dependencyId)
	}

	// Because we require the dependencies in the set to already be in the network config, we can simply use the
	// order in which AddService is called to generate a traversal through the dependency DAG (no need to use any
	// DAG traversal algorithms)
	builder.servicesStartOrder = append(builder.servicesStartOrder, serviceId)
	builder.serviceDependencies[serviceId] = dependenciesCopy
	return serviceId, nil
}

func (builder ServiceNetworkConfigBuilder) Build() *ServiceNetworkConfig {
	// Defensive copy, so user calling functions on the builder after building won't affect the
	// state of the object we already built
	serviceConfigsCopy := make(map[int]int)
	for serviceId, configId := range builder.serviceConfigs {
		serviceConfigsCopy[serviceId] = configId
	}
	serviceDependenciesCopy := make(map[int]map[int]bool)
	for serviceId, dependencies := range builder.serviceDependencies {
		dependenciesCopy := make(map[int]bool)
		for dependencyId, _ := range dependencies {
			dependenciesCopy[dependencyId] = true
		}
		serviceDependenciesCopy[serviceId] = dependenciesCopy
	}
	serviceStartOrderCopy := make([]int, len(builder.servicesStartOrder))
	copy(serviceStartOrderCopy, builder.servicesStartOrder)

	onlyDependentServicesCopy := make(map[int]bool)
	for dependencyId, _ := range builder.onlyDependentServices {
		onlyDependentServicesCopy[dependencyId] = true
	}

	configurationsCopy := make(map[int]serviceConfig)
	for configurationId, config := range builder.configurations {
		configurationsCopy[configurationId] = config
	}

	return &ServiceNetworkConfig{
		serviceConfigs:      serviceConfigsCopy,
		serviceDependencies: serviceDependenciesCopy,
		servicesStartOrder:  serviceStartOrderCopy,
		onlyDependentServices: onlyDependentServicesCopy,
		configurations:      configurationsCopy,
	}
}

// Object declaring the state of the network to be created
type ServiceNetworkConfig struct {
	serviceConfigs map[int]int
	serviceDependencies map[int]map[int]bool
	servicesStartOrder []int
	// Do we actually need to keep this onlyDependents list ?? We've been doing it for liveness-checking, but maybe we just
	// push that to the implementer of the network (make them do the calls based off what they know)
	// Don't want to rip it out yet though because it was a pain to put in
	onlyDependentServices map[int]bool
	configurations map[int]serviceConfig
}

/*
This method will create a running instantion of the configured network

Returns:
	A struct containing information about the network. If an error occurs midway through creation, there will be several
		containers left hanging around and *the network return value will contain only the already-started containers*! A
		user of this method should check if the error result is set and, if so, shut down the running containers!
 */
// TODO use the network name to create a new network!!
func (networkCfg ServiceNetworkConfig) CreateNetwork(testImage string, publicIpProvider *FreeIpAddrTracker, manager *docker.DockerManager) (*RawServiceNetwork, error) {
	runningServices := make(map[int]services.Service)
	serviceIps := make(map[int]string)
	serviceContainerIds := make(map[int]string)
	allServiceDependencies := make(map[int][]services.Service)

	// First pass: start all services
	logrus.Info("Creating & starting test network containers...")
	for _, serviceId := range networkCfg.servicesStartOrder {
		serviceDependencies := networkCfg.getServiceDependencies(serviceId, runningServices)

		config, err := networkCfg.getServiceConfig(serviceId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not get service config for service ID %v", serviceId)
		}

		staticIp, err := publicIpProvider.GetFreeIpAddr()
		if err != nil {
			return &RawServiceNetwork{
				ServiceIPs:   serviceIps,
				ContainerIds: serviceContainerIds,
			}, stacktrace.Propagate(err, "Failed to allocate static IP for service %d", serviceId)
		}

		dockerImagePtr := config.dockerImage
		if dockerImagePtr == nil {
			dockerImagePtr = &testImage
		}

		service, containerId, err := config.initializer.CreateService(*dockerImagePtr, staticIp, manager, serviceDependencies)
		if err != nil {
			return &RawServiceNetwork{
				ServiceIPs:   serviceIps,
				ContainerIds: serviceContainerIds,
			}, stacktrace.Propagate(err, "Failed to construct service from serviceConfig")
		}

		runningServices[serviceId] = service
		serviceIps[serviceId] = staticIp
		serviceContainerIds[serviceId] = containerId
		allServiceDependencies[serviceId] = serviceDependencies
	}
	logrus.Info("Test network containers created & started")

	return &RawServiceNetwork{
		ServiceIPs:   serviceIps,
		ContainerIds: serviceContainerIds,
	}, nil
}

// Intended for use by the test controller
// Reads basic information about the nodes in the network and uses the topology information stored
//  in this object to:
//		1) translate each IP into the appropriate service using the constructors given in this object and
//		2) wait until all services are available
func (networkCfg ServiceNetworkConfig) LoadNetwork(rawInfo RawServiceNetwork) (map[int]services.Service, error) {
	// First pass: construct the instantions of each service object
	logrus.Info("Loading services from IPs...")
	runningServices := make(map[int]services.Service)
	for _, serviceId := range networkCfg.servicesStartOrder {

		ipAddr, found := rawInfo.ServiceIPs[serviceId]
		if !found {
			return nil, stacktrace.NewError("Missing expected service ID '%v' in network info", serviceId)
		}

		config, err := networkCfg.getServiceConfig(serviceId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not get service config for service ID %v", serviceId)
		}

		service := config.initializer.LoadService(ipAddr)
		runningServices[serviceId] = service
	}
	logrus.Info("All services loaded from IPs")

	// Second pass: wait for all services to come up
	logrus.Info("Waiting for network to become available...")
	for _, serviceId := range networkCfg.servicesStartOrder {
		service := runningServices[serviceId]
		serviceDependencies := networkCfg.getServiceDependencies(serviceId, runningServices)
		config, err := networkCfg.getServiceConfig(serviceId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not get service config for ID %v", serviceId)
		}

		logrus.Debugf("Waiting for service %v to become available...", serviceId)
		if err := config.availabilityChecker.WaitForStartup(service, serviceDependencies); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred waiting for service %v to start up", serviceId)
		}
		logrus.Debugf("Service %v is available", serviceId)
	}
	logrus.Info("Network is available")

	return runningServices, nil
}

// Convenience function to get a service's dependencies as a []Service
func (networkCfg ServiceNetworkConfig) getServiceDependencies(serviceId int, runningServices map[int]services.Service) []services.Service {
	serviceDependenciesIds := networkCfg.serviceDependencies[serviceId]
	serviceDependencies := make([]services.Service, 0, len(serviceDependenciesIds))
	for dependencyId, _ := range serviceDependenciesIds {
		// We're guaranteed that this dependency will already be running due to the ordering we enforce in the builder
		serviceDependencies = append(serviceDependencies, runningServices[dependencyId])
	}
	return serviceDependencies
}

// Convenience function to get a service's serviceConfig
func (networkCfg ServiceNetworkConfig) getServiceConfig(serviceId int) (serviceConfig, error) {
	configId, found := networkCfg.serviceConfigs[serviceId]
	if !found {
		return serviceConfig{}, stacktrace.NewError("Found ID '%v' in the network info but no configuration is defined for this ID in the network config", serviceId)
	}

	config, found := networkCfg.configurations[configId]
	if !found {
		return serviceConfig{}, stacktrace.NewError(
			"Service ID '%v' uses service configuration '%v', but this service config wasn't found in this network configuration; this is likely a code problem",
			serviceId,
			configId)
	}
	return config, nil
}
