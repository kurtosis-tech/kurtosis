package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
)

type serviceNode struct {

}

type ServiceNetwork struct {
	freeIpTracker *FreeIpAddrTracker

	dockerManager *docker.DockerManager

	// If Go had generics, we'd make this object genericized and use that as the return type here
	services map[int]services.Service

	// NOTE: we'll likely need to create an object encapsulating all information about a service - its containerId, its IP,
	containerIds map[int]string

	configurations map[int]serviceConfig
}

// Adds a service to the graph, with the specified dependencies (with the map used only as a set - the values are ignored)
// Returns an AvailabilityChecker for checking when the service is actually availabile
// If no dependencies should be specified, the dependencies map should be empty (not nil)
func (network *ServiceNetwork) AddService(configurationId int, serviceId int, dependencies map[int]bool) (*services.ServiceAvailabilityChecker, error) {
	config, found := network.configurations[configurationId]
	if !found {
		return nil, stacktrace.NewError("No service configuration with ID '%v' has been registered", configurationId)
	}

	if _, exists := network.services[serviceId]; exists {
		return nil, stacktrace.NewError("Service ID %d already exists in the network", serviceId)
	}

	if dependencies == nil {
		return nil, stacktrace.NewError("Dependencies map was nil; use an empty map to specify no dependencies")
	}

	// Golang maps are passed by-ref, so we do a defensive copy here so user can't change their input and mess
	// with our internal data structure
	dependencyServices := make([]services.Service, 0, len(dependencies))
	for dependencyId, _ := range dependencies  {
		dependency, found := network.services[dependencyId]
		if !found {
			return nil, stacktrace.NewError("Declared a dependency on %v but no service with this ID has been registered", dependencyId)
		}
		dependencyServices = append(dependencyServices, dependency)
	}

	staticIp, err := network.freeIpTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to allocate static IP for service %d", serviceId)
	}

	service, containerId, err := config.initializer.CreateService(config.dockerImage, staticIp, network.dockerManager, dependencyServices)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service %v from configuration %v", serviceId, configurationId)
	}

	network.services[serviceId] = service

	runningServices[serviceId] = service
	serviceIps[serviceId] = staticIp
	serviceContainerIds[serviceId] = containerId
	allServiceDependencies[serviceId] = serviceDependencies

	// Because we require the dependencies in the set to already be in the network config, we can simply use the
	// order in which AddService is called to generate a traversal through the dependency DAG (no need to use any
	// DAG traversal algorithms)
	builder.servicesStartOrder = append(builder.servicesStartOrder, serviceId)
	builder.serviceDependencies[serviceId] = dependenciesCopy
	return serviceId, nil
}
