package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

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

	// Tracks the next service ID that will be doled out upon a call to AddService
	nextServiceId int

	// Factories that will be used to construct the nodes at build time
	configurations map[int]services.ServiceFactory

	// Tracks the next service configuration ID that will be doled out upon a call to AddServiceConfiguration
	nextConfigurationId int

}

func NewServiceNetworkConfigBuilder() *ServiceNetworkConfigBuilder {
	serviceConfigs := make(map[int]int)
	serviceDependencies := make(map[int]map[int]bool)
	serviceStartOrder := make([]int, 0)
	onlyDependentServices := make(map[int]bool)
	configurations := make(map[int]services.ServiceFactory)
	return &ServiceNetworkConfigBuilder{
		serviceConfigs:      serviceConfigs,
		serviceDependencies: serviceDependencies,
		servicesStartOrder:  serviceStartOrder,
		onlyDependentServices: onlyDependentServices,
		nextServiceId:       0,
		configurations: 	 configurations,
		nextConfigurationId: 0,
	}
}

// Adds a service configuration to the network, that can be referenced later with AddService
func (builder *ServiceNetworkConfigBuilder) AddServiceConfiguration(factory services.ServiceFactory) int {
	configurationId := builder.nextConfigurationId
	builder.nextConfigurationId = builder.nextConfigurationId + 1
	builder.configurations[configurationId] = factory
	return configurationId
}

// Adds a serivce to the graph, with the specified dependencies (with the map used only as a set - the values are ignored)
// Returns the ID of the service, to be used with future AddService calls to declare dependencies on the service
// If no dependencies should be specified, the dependencies map should be empty (not nil)
func (builder *ServiceNetworkConfigBuilder) AddService(configurationId int, dependencies map[int]bool) (int, error) {
	if _, found := builder.configurations[configurationId]; !found {
		return 0, stacktrace.NewError("No configuration with ID '%v' has been registered", configurationId)
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

	serviceId := builder.nextServiceId
	builder.nextServiceId = builder.nextServiceId + 1
	builder.serviceConfigs[serviceId] = configurationId

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

	configurationsCopy := make(map[int]services.ServiceFactory)
	for configurationId, factory := range builder.configurations {
		configurationsCopy[configurationId] = factory
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
	configurations map[int]services.ServiceFactory
}

// TODO use the network name to create a new network!!
func (networkCfg ServiceNetworkConfig) CreateAndRun(publicIpProvider *FreeIpAddrTracker, manager *docker.DockerManager) (*RawServiceNetwork, error) {
	runningServices := make(map[int]services.Service)
	serviceIps := make(map[int]string)
	serviceContainerIds := make(map[int]string)
	for _, serviceId := range networkCfg.servicesStartOrder {
		serviceDependenciesIds := networkCfg.serviceDependencies[serviceId]
		serviceDependencies := make([]services.Service, 0, len(serviceDependenciesIds))
		for dependencyId, _ := range serviceDependenciesIds {
			// We're guaranteed that this dependency will already be running due to the ordering we enforce in the builder
			serviceDependencies = append(serviceDependencies, runningServices[dependencyId])
		}

		configId := networkCfg.serviceConfigs[serviceId]
		factory := networkCfg.configurations[configId]

		staticIp, err := publicIpProvider.GetFreeIpAddr()
		if err != nil {
			// TODO an error here means we have a half-created network that we need to return to the user so they can shut down!
			return nil, stacktrace.Propagate(err, "Failed to allocate static IP for service %d", serviceId)
		}
		service, containerId, err := factory.Construct(staticIp, manager, serviceDependencies)
		if err != nil {
			// TODO an error here means we have a half-created network that we need to return to the user so they can shut down!
			return nil, stacktrace.Propagate(err, "Failed to construct service from factory")
		}
		runningServices[serviceId] = service
		serviceIps[serviceId] = staticIp
		serviceContainerIds[serviceId] = containerId
	}

	return &RawServiceNetwork{
		ContainerIds:   serviceContainerIds,
		ServiceIPs: serviceIps,
	}, nil
}