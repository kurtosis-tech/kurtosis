package commons

import (
	"github.com/docker/docker/api/types"
	"github.com/palantir/stacktrace"
)

// Builder to ease the declaration of the network state we want
type JsonRpcServiceNetworkConfigBuilder struct {
	serviceConfigs map[int]JsonRpcServiceConfig

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
}

func NewJsonRpcServiceNetworkConfigBuilder() *JsonRpcServiceNetworkConfigBuilder {
	serviceConfigs := make(map[int]JsonRpcServiceConfig)
	serviceDependencies := make(map[int]map[int]bool)
	serviceStartOrder := make([]int, 0)
	onlyDependentServices := make(map[int]bool)
	return &JsonRpcServiceNetworkConfigBuilder{
		serviceConfigs:      serviceConfigs,
		serviceDependencies: serviceDependencies,
		servicesStartOrder:  serviceStartOrder,
		onlyDependentServices: onlyDependentServices,
		nextServiceId:       0,
	}
}

// Adds a serivce to the graph, with the specified dependencies (with the map used only as a set - the values are ignored)
// Returns the ID of the service, to be used with future AddService calls to declare dependencies on the service
// If no dependencies should be specified, the dependencies map should be empty (not nil)
func (builder *JsonRpcServiceNetworkConfigBuilder) AddService(config JsonRpcServiceConfig, dependencies map[int]bool) (int, error) {
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
	builder.serviceConfigs[serviceId] = config

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

func (builder JsonRpcServiceNetworkConfigBuilder) Build() *JsonRpcServiceNetworkConfig {
	// Defensive copy, so user calling functions on the builder after building won't affect the
	// state of the object we already built
	serviceConfigsCopy := make(map[int]JsonRpcServiceConfig)
	for serviceId, serviceCfg := range builder.serviceConfigs {
		serviceConfigsCopy[serviceId] = serviceCfg
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

	return &JsonRpcServiceNetworkConfig{
		serviceConfigs:      serviceConfigsCopy,
		serviceDependencies: serviceDependenciesCopy,
		servicesStartOrder:  serviceStartOrderCopy,
		onlyDependentServices: onlyDependentServicesCopy,
	}
}

// Object declaring the state of the network to be created
type JsonRpcServiceNetworkConfig struct {
	// TODO make this be a single map[int]RunningService objects
	serviceConfigs map[int]JsonRpcServiceConfig
	serviceDependencies map[int]map[int]bool
	servicesStartOrder []int
	onlyDependentServices map[int]bool
}

// TODO use the network name to create a new network!!
func (networkCfg JsonRpcServiceNetworkConfig) CreateAndRun(networkName string, manager *DockerManager) (network *JsonRpcServiceNetwork, err error) {
	serviceLivenessReqs := make(map[int]JsonRpcRequest)
	for serviceId, serviceCfg := range networkCfg.serviceConfigs {
		serviceLivenessReqs[serviceId] = serviceCfg.GetLivenessRequest()
	}

	// TODO this isn't sufficient - we'll need to also store service-specific ports (e.g. staking port)
	runningServices := make(map[int]JsonRpcServiceSocket)
	serviceContainerIds := make(map[int]string)
	for _, serviceId := range networkCfg.servicesStartOrder {
		serviceDependenciesIds := networkCfg.serviceDependencies[serviceId]
		serviceDependenciesLivenessReqs := make(map[JsonRpcServiceSocket]JsonRpcRequest)
		for dependencyId, _ := range serviceDependenciesIds {
			// We're guaranteed that this dependency will already be running due to the ordering we enforce in the builder
			dependencySocket := runningServices[dependencyId]
			serviceDependenciesLivenessReqs[dependencySocket] = serviceLivenessReqs[dependencyId]
		}

		serviceCfg := networkCfg.serviceConfigs[serviceId]
		// TODO this relies on serviceId being incremental, and is a total hack until --public-ips flag is gone from Gecko!
		containerConfigPtr, err := manager.GetContainerCfgFromServiceCfg(serviceId, serviceCfg, serviceDependenciesLivenessReqs)


		
		containerHostConfigPtr, err := manager.GetContainerHostConfig(serviceCfg)
		if err != nil {
			return nil, stacktrace.Propagate(err, "")
		}
		// TODO probably use a UUID for the network name (and maybe include test name too)
		resp, err := manager.dockerClient.ContainerCreate(manager.dockerCtx, containerConfigPtr, containerHostConfigPtr, nil, "")
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not create Docker container from image %v.", serviceCfg.GetDockerImage())
		}
		containerId := resp.ID
		if err := manager.dockerClient.ContainerStart(manager.dockerCtx, containerId, types.ContainerStartOptions{}); err != nil {
			return nil, stacktrace.Propagate(err, "Could not start Docker container from image %v.", serviceCfg.GetDockerImage())
		}

		containerJson, err := manager.dockerClient.ContainerInspect(manager.dockerCtx, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Inspect container failed, which is necessary to get the contaienr's IP")
		}
		containerIpAddr := containerJson.NetworkSettings.IPAddress

		serviceContainerIds[serviceId] = containerId
		runningServices[serviceId] = JsonRpcServiceSocket{
			IPAddress: containerIpAddr,
			Port:      serviceCfg.GetJsonRpcPort(),
		}
	}


	// TODO actually fill in all the other stuff besides container ID
	return &JsonRpcServiceNetwork{
		NetworkId:               "",
		ServiceContainerIds:     serviceContainerIds,
		ServiceIps:              nil,
		ServiceJsonRpcPorts:     nil,
		ServiceCustomPorts:      nil,
		NetworkLivenessRequests: nil,
	}, nil
}
