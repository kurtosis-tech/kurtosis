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

/*
The identifier used for services with the network.
 */
type ServiceID int

/*
A package object containing the details that the ServiceNetwork is tracking about a node.
 */
type ServiceNode struct {
	// The node's IP address within the test's Docker network
	IpAddr string

	// The user-defined interface for interacting with the node.
	// NOTE: this will need to be casted to the appropriate interface becaus Go doesn't yet have generics!
	Service services.Service

	// The Docker container ID of the container running the node
	ContainerId string
}

/*
A package object containing the details of a particular service configuration, to give Kurtosis the implementation-specific
	details about how to interact with user-defined services.
 */
type serviceConfig struct {
	// The Docker image that will be used to launch nodes
	dockerImage string

	// The implementation that will be used for launching a Docker image of a node using this configuration
	initializerCore services.ServiceInitializerCore

	// The implementation that will be used for determining whether a node launched using this configuration is available
	availabilityCheckerCore services.ServiceAvailabilityCheckerCore
}


/*
A struct representing a network of services that will be used for a single test (commonly called the "test network"). This
	struct is the low-level access point for modifying the test network.
 */
type ServiceNetwork struct {
	// The tracker used for doling out new IPs within the subnet being used for this particular test network
	freeIpTracker *FreeIpAddrTracker

	// The Docker manager used for interacting with the Docker engine during test network manipulation
	dockerManager *docker.DockerManager

	// The ID of the Docker network that this test network is running on
	dockerNetworkId string

	// A mapping of service ID -> information about a node
	serviceNodes map[ServiceID]ServiceNode

	// A mapping of configuration ID -> configuration details
	configurations map[ConfigurationID]serviceConfig

	// The name of the Docker volume that will be mounted on:
	// 	a) every single Docker image launched on this network
	//  b) the test controller running logic against this test network
	testVolume string

	// The dirpath where the test volume is mounted on the controller (which is where this code will be running in)
	testVolumeControllerDirpath string
}

/*
Creates a new ServiceNetwork object with the given parameters.

Args:
	freeIpTracker: The IP tracker that will be used to provide IPs for new nodes added to the network.
	dockerManager: The Docker manager that will be used for manipulating the Docker engine during test network modification.
	dockerNetworkName: The name of the Docker network this test network is running on.
	configurations: The configurations that are available for spinning up new nodes in the network.
	testVolume: The name of the Docker volume that will be mounted on all the nodes in the network.
	testVolumeControllerDirpath: The dirpath that the test Docker volume is mounted on in the controller image (which will
		be running all the code here).
 */
func NewServiceNetwork(
			freeIpTracker *FreeIpAddrTracker,
			dockerManager *docker.DockerManager,
			dockerNetworkId string,
			configurations map[ConfigurationID]serviceConfig,
			testVolume string,
			testVolumeControllerDirpath string) *ServiceNetwork {
	return &ServiceNetwork{
		freeIpTracker:               freeIpTracker,
		dockerManager:               dockerManager,
		dockerNetworkId:             dockerNetworkId,
		serviceNodes:                make(map[ServiceID]ServiceNode),
		configurations:              configurations,
		testVolume:                  testVolume,
		testVolumeControllerDirpath: testVolumeControllerDirpath,
	}
}

// Gets the number of nodes in the network
func (network *ServiceNetwork) GetSize() int {
	return len(network.serviceNodes)
}

/*
Adds a service to the network with the given service ID, created using the given configuration ID.

Args:
	configurationId: The ID of the service configuration to use for creating the service.
	serviceId: The service ID that will be used to identify this node in the network.
	dependencies: A "set" of service IDs that the node being created will depend on - i.e., whose information the node-to-create
		needs to start up. If the node-to-create doesn't depend on any other services, the dependencies map should be
		empty (not nil).

Return:
	An AvailabilityChecker for checking when the new service is available and ready for use.
 */
func (network *ServiceNetwork) AddService(configurationId ConfigurationID, serviceId ServiceID, dependencies map[ServiceID]bool) (*services.ServiceAvailabilityChecker, error) {
	// Maybe one day we'll make this flow from somewhere up above (e.g. make the entire network live inside a single context)
	parentCtx := context.Background()

	config, found := network.configurations[configurationId]
	if !found {
		return nil, stacktrace.NewError("No service configuration with ID '%v' has been registered", configurationId)
	}

	if _, exists := network.serviceNodes[serviceId]; exists {
		return nil, stacktrace.NewError("Service ID %d already exists in the network", serviceId)
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
		return nil, stacktrace.Propagate(err, "Failed to allocate static IP for service %d", serviceId)
	}

	initializer := services.NewServiceInitializer(config.initializerCore, network.dockerNetworkId, network.testVolumeControllerDirpath)
	service, containerId, err := initializer.CreateService(
			parentCtx,
			network.testVolume,
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

/*
Gets the node information for the service with the given service ID.
 */
func (network *ServiceNetwork) GetService(serviceId ServiceID) (ServiceNode, error) {
	node, found := network.serviceNodes[serviceId]
	if !found {
		return ServiceNode{}, stacktrace.NewError("No service with ID %v exists in the network", serviceId)
	}

	return node, nil
}

/*
Stops the container with the given service ID, and removes it from the network.
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
Makes a best-effort attempt to remove all the containers in the network, waiting for the given timeout and returning
	an error if the timeout is reached.

Args:
	containerStopTimeout: How long to wait for each container to stop before force-killing it
*/
func (network *ServiceNetwork) RemoveAll(containerStopTimeout time.Duration) error {
	for serviceId, _ := range network.serviceNodes {
		network.RemoveService(serviceId, containerStopTimeout)
	}
	return nil
}
