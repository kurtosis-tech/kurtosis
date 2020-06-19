package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)



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
