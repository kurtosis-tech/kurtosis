package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

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

	network.containerIds[serviceId] = containerId
	network.services[serviceId] = service

	availabilityChecker := services.NewServiceAvailabilityChecker(config.availabilityCheckerCore, service, dependencyServices)
	return availabilityChecker, nil
}

/*
Stops the container with the given service ID, and stops tracking it in the network
 */
func (network *ServiceNetwork) RemoveService(serviceId int) error {
	// TODO
	return nil
}

/*
Makes a best-effort attempt to stop all the containers in the network, waiting for the given timeout.
Args:
	containerStopTimeout: How long to wait for each container to stop before force-killing it
*/
func (network *ServiceNetwork) Stop(containerStopTimeout time.Duration) error {
	for serviceId, containerId := range network.containerIds {
		logrus.Debugf("Stopping service ID %v with container ID '%v'", serviceId, containerId)
		err := network.dockerManager.StopContainer(containerId, &containerStopTimeout)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred stopping service ID %v with container ID %v; proceeding to stop other containers:",
				serviceId,
				containerId)
		}
		logrus.Debugf("Container with ID '%v' successfully stopped", containerId)
	}
	return nil
}
