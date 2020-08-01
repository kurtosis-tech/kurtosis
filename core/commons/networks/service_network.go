package networks

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

type ServiceID string

type ServiceNode struct {
	IpAddr string

	// If Go had generics, we'd make this object genericized and use that as the return type here
	Service services.Service

	ContainerId string
}

type ServiceNetwork struct {
	freeIpTracker *FreeIpAddrTracker

	dockerManager *docker.DockerManager

	dockerNetworkName string

	serviceNodes map[ServiceID]ServiceNode

	configurations map[ConfigurationID]serviceConfig

	testVolume string

	testVolumeControllerDirpath string
}

func NewServiceNetwork(
			freeIpTracker *FreeIpAddrTracker,
			dockerManager *docker.DockerManager,
			dockerNetworkName string,
			serviceNodes map[ServiceID]ServiceNode,
			configurations map[ConfigurationID]serviceConfig,
			testVolume string,
			testVolumeControllerDirpath string) *ServiceNetwork {
	return &ServiceNetwork{
		freeIpTracker:               freeIpTracker,
		dockerManager:               dockerManager,
		dockerNetworkName:           dockerNetworkName,
		serviceNodes:                serviceNodes,
		configurations:              configurations,
		testVolume:                  testVolume,
		testVolumeControllerDirpath: testVolumeControllerDirpath,
	}
}


func (network *ServiceNetwork) GetSize() int {
	return len(network.serviceNodes)
}

/*
Adds a service to the graph that depends on the services with the given IDs
Args:
	configurationId: The ID of the service configuration to use for creating the service
	serviceId: The ID to give the node in the network
	dependencies: A "set" of service IDs that the node-to-create depends on - i.e., whose information the node-to-create
		needs to start up. If the node-to-create doesn't depend on any other services, the dependencies map should be
		empty (not nil).
Return:
	An AvailabilityChecker for checking when the service is actually available
 */
func (network *ServiceNetwork) AddService(configurationId ConfigurationID, serviceId ServiceID, dependencies map[ServiceID]bool) (*services.ServiceAvailabilityChecker, error) {
	// Maybe one day we'll make this flow from somewhere up above (e.g. make the entire network live inside a single context)
	parentCtx := context.Background()

	config, found := network.configurations[configurationId]
	if !found {
		return nil, stacktrace.NewError("No service configuration with ID '%v' has been registered", configurationId)
	}

	if _, exists := network.serviceNodes[serviceId]; exists {
		return nil, stacktrace.NewError("Service ID %s already exists in the network", serviceId)
	}

	if dependencies == nil {
		return nil, stacktrace.NewError("Dependencies map was nil; use an empty map to specify no dependencies")
	}

	// Golang maps are passed by-ref, so we do a defensive copy here so user can't change their input and mess
	// with our internal data structure
	dependencyServices := make([]services.Service, 0, len(dependencies))
	for dependencyId, _ := range dependencies  {
		dependencyNode, found := network.serviceNodes[dependencyId]
		if !found {
			return nil, stacktrace.NewError("Declared a dependency on %v but no service with this ID has been registered", dependencyId)
		}
		dependencyServices = append(dependencyServices, dependencyNode.Service)
	}

	staticIp, err := network.freeIpTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to allocate static IP for service %s", serviceId)
	}

	initializer := services.NewServiceInitializer(config.initializerCore, network.dockerNetworkName)
	service, containerId, err := initializer.CreateService(
			parentCtx,
			network.testVolume,
			network.testVolumeControllerDirpath,
			config.dockerImage,
			staticIp,
			network.dockerManager,
			dependencyServices)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating service %v from configuration %v", serviceId, configurationId)
	}

	network.serviceNodes[serviceId] = ServiceNode{
		IpAddr:      staticIp,
		Service:     service,
		ContainerId: containerId,
	}

	availabilityChecker := services.NewServiceAvailabilityChecker(parentCtx, config.availabilityCheckerCore, service, dependencyServices)
	return availabilityChecker, nil
}

func (network *ServiceNetwork) GetService(serviceId ServiceID) (ServiceNode, error) {
	node, found := network.serviceNodes[serviceId]
	if !found {
		return ServiceNode{}, stacktrace.NewError("No service with ID %v exists in the network", serviceId)
	}

	return node, nil
}

/*
Stops the container with the given service ID, and stops tracking it in the network
 */
func (network *ServiceNetwork) RemoveService(serviceId ServiceID, containerStopTimeout time.Duration) error {
	// Maybe one day we'll store this on the ServiceNetwork itself, to represent the test context that the ServiceNetwork
	//  was created in
	parentCtx := context.Background()

	nodeInfo, found := network.serviceNodes[serviceId]
	if !found {
		return stacktrace.NewError("No service with ID %v found", serviceId)
	}

	logrus.Debugf("Removing service ID %v...", serviceId)
	delete(network.serviceNodes, serviceId)

	// Make a best-effort attempt to stop the container
	err := network.dockerManager.StopContainer(parentCtx, nodeInfo.ContainerId, &containerStopTimeout)
	if err != nil {
		logrus.Errorf(
			"The following error occurred stopping service ID %v with container ID %v; proceeding to stop other containers:",
			serviceId,
			nodeInfo.ContainerId)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
	}
	logrus.Debugf("Successfully removed service ID %v", serviceId)
	return nil
}

/*
Makes a best-effort attempt to remove all the containers in the network, waiting for the given timeout.
Args:
	containerStopTimeout: How long to wait for each container to stop before force-killing it
*/
func (network *ServiceNetwork) RemoveAll(containerStopTimeout time.Duration) error {
	for serviceId, _ := range network.serviceNodes {
		network.RemoveService(serviceId, containerStopTimeout)
	}
	return nil
}
