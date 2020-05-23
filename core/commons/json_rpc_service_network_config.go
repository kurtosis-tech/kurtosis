package commons

import (
	"container/list"
	"context"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
)

const LOCAL_HOST_IP = "0.0.0.0"

// TODO replace these with FreeHostPortProvider in the future (this class shouldn't know anything about Ava)
const DEFAULT_GECKO_HTTP_PORT = nat.Port("9650/tcp")
const DEFAULT_GECKO_STAKING_PORT = nat.Port("9651/tcp")

type JsonRpcServiceNetworkConfigBuilder struct {
	serviceConfigs map[int]JsonRpcServiceConfig

	// Maps service_id -> set(ids of services that depend on the service)
	// The 'map' value is only because Go doesn't have a set type
	serviceDependents map[int]map[int]bool

	// A "set" of service IDs that don't depend on anything and so can be started first
	servicesWithoutDependencies map[int]bool

	// Tracks the next node ID that will be doled out upon a call to AddNode
	nextNodeId int
}

func NewJsonRpcServiceNetworkConfigBuilder() *JsonRpcServiceNetworkConfigBuilder {
	serviceConfigs := make(map[int]JsonRpcServiceConfig)
	serviceDependents := make(map[int]map[int]bool)
	servicesWithoutDependencies := make(map[int]bool)
	return &JsonRpcServiceNetworkConfigBuilder{
		serviceConfigs:    serviceConfigs,
		serviceDependents: serviceDependents,
		servicesWithoutDependencies: servicesWithoutDependencies,
	}
}

// Adds a node to the graph, with the specified dependencies (with the map used only as a set - the values are ignored)
// If no dependencies should be specified, the dependencies map should be empty (not nil)
func (builder JsonRpcServiceNetworkConfigBuilder) AddNode(config JsonRpcServiceConfig, dependencies map[int]bool) (int, error) {
	for dependencyId, _ := range(dependencies) {
		if _, found := builder.serviceConfigs[dependencyId]; !found {
			return 0, stacktrace.NewError("Declared a dependency on %v but no service with this ID has been registered", dependencyId)
		}
	}
	nodeId := builder.nextNodeId
	builder.nextNodeId++
	builder.serviceConfigs[nodeId] = config
	if len(dependencies) > 0 {
		for dependencyId, _ := range(dependencies) {
			if dependents, found := builder.serviceDependents[dependencyId]; found {
				dependents[nodeId] = true
			} else {
				newDependents := make(map[int]bool)
				newDependents[nodeId] = true
				builder.serviceDependents[dependencyId] = newDependents
			}
		}
	} else {
		builder.servicesWithoutDependencies[nodeId] = true
	}

	return nodeId, nil
}

func (builder JsonRpcServiceNetworkConfigBuilder) Build() *JsonRpcServiceNetworkConfig {
	return &JsonRpcServiceNetworkConfig{
		services:          builder.serviceConfigs,
		serviceDependents: builder.serviceDependents,
		servicesWithoutDependencies: builder.servicesWithoutDependencies,
	}
}


type JsonRpcServiceNetworkConfig struct {
	services map[int]JsonRpcServiceConfig
	serviceDependents map[int]map[int]bool
	servicesWithoutDependencies map[int]bool
}

func (networkCfg JsonRpcServiceNetworkConfig) CreateAndRun(dockerCtx context.Context, dockerClient *client.Client) (network *JsonRpcServiceNetwork, err error) {
	toStartQueue := make([]int, 0)
	alreadyStarted := make(map[int]bool)
	for serviceId, _ := range networkCfg.servicesWithoutDependencies {
		toStartQueue = append(toStartQueue, serviceId)
	}

	for len(toStartQueue) > 0 {
		serviceToStart := toStartQueue[0]
		toStartQueue = toStartQueue[1:]

	}






	serviceContainerIds := make(map[int]string)
	for serviceId, serviceCfg := range networkCfg.services {
		containerConfigPtr, err := getContainerCfgFromServiceCfg(serviceCfg)

		// TODO need to use FreeHostPortProvider here
		containerHostConfigPtr := getContainerHostConfig(serviceCfg)
		// TODO probably use a UUID for the network name (and maybe include test name too)
		resp, err := dockerClient.ContainerCreate(dockerCtx, containerConfigPtr, containerHostConfigPtr, nil, "")
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not create Docker container from image %v.", serviceCfg.GetDockerImage())
		}
		containerId := resp.ID
		if err := dockerClient.ContainerStart(dockerCtx, containerId, types.ContainerStartOptions{}); err != nil {
			return nil, stacktrace.Propagate(err, "Could not start Docker container from image %v.", serviceCfg.GetDockerImage())
		}
		serviceContainerIds[serviceId] = containerId
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

// TODO should I actually be passing sorta-complex objects like JsonRpcServiceConfig by value???
// Creates a more generalized Docker Container configuration for Gecko, with a 5-parameter initialization command.
// Gecko HTTP and Staking ports inside the Container are the standard defaults.
func getContainerCfgFromServiceCfg(serviceConfig JsonRpcServiceConfig) (config *container.Config, err error) {
	jsonRpcPort, err := nat.NewPort("tcp", strconv.Itoa(serviceConfig.GetJsonRpcPort()))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not parse port int.")
	}

	portSet := nat.PortSet{
		jsonRpcPort: struct{}{},
	}
	for _, port := range serviceConfig.GetOtherPorts() {
		otherPort, err := nat.NewPort("tcp", strconv.Itoa(port))
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not parse port int.")
		}
		portSet[otherPort] = struct{}{}
	}

	nodeConfigPtr := &container.Config{
		Image: serviceConfig.GetDockerImage(),
		// TODO allow modifying of protocol at some point
		ExposedPorts: portSet,
		Cmd: serviceConfig.GetContainerStartCommand(),
		Tty: false,
	}
	return nodeConfigPtr, nil
}

// Creates a Docker-Container-To-Host Port mapping, defining how a Container's JSON RPC and service-specific ports are
// mapped to the host ports
func getContainerHostConfig(serviceConfig JsonRpcServiceConfig) *container.HostConfig {
	// TODO right nwo this is hardcoded - replace these with FreeHostPortProvider in the future, so we can have
	//  arbitrary service-specific ports!
	jsonRpcPortBinding := []nat.PortBinding{
		{
			HostIP: LOCAL_HOST_IP,
			HostPort: strconv.Itoa(serviceConfig.GetJsonRpcPort()),
		},
	}
	// TODO this shouldn't be here - this class should have nothing Ava-specific
	stakingPortInt := strconv.Itoa(serviceConfig.GetOtherPorts()[0]) // This is actually STAKING_PORT_ID, but I get an import cycle if I actually use it
	stakingPortBinding := []nat.PortBinding{
		{
			HostIP: LOCAL_HOST_IP,
			HostPort: stakingPortInt,
		},
	}

	containerHostConfigPtr := &container.HostConfig{
		PortBindings: nat.PortMap{
			DEFAULT_GECKO_HTTP_PORT: jsonRpcPortBinding,
			DEFAULT_GECKO_STAKING_PORT: stakingPortBinding,
		},
	}
	return containerHostConfigPtr
}

