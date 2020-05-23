package commons

import (
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"github.com/palantir/stacktrace"
)


// TODO replace these with FreeHostPortProvider in the future (this class shouldn't know anything about Ava)
const DEFAULT_GECKO_HTTP_PORT = nat.Port("9650/tcp")
const DEFAULT_GECKO_STAKING_PORT = nat.Port("9651/tcp")


type JsonRpcServiceNetworkConfig struct {
	services map[int]JsonRpcServiceConfig
}

// TODO replace this with a fluent builder, to make it a bunch easier to add nodes with dependencies
func NewJsonRpcServiceNetworkConfig(serviceCfgs map[int]JsonRpcServiceConfig) *JsonRpcServiceNetworkConfig {
	return &JsonRpcServiceNetworkConfig{
		services:  serviceCfgs,
	}
}

func (networkCfg JsonRpcServiceNetworkConfig) CreateAndRun(manager *DockerManager) (network *JsonRpcServiceNetwork, err error) {
	serviceContainerIds := make(map[int]string)
	for serviceId, serviceCfg := range networkCfg.services {
		containerConfigPtr, err := getContainerCfgFromServiceCfg(serviceCfg)

		// TODO need to use FreeHostPortProvider here
		containerHostConfigPtr, err := getContainerHostConfig(serviceCfg, manager)
		if err != nil {
			return nil, stacktrace.Propagate(err, "")
		}
		// TODO probably use a UUID for the network name (and maybe include test name too)
		resp, err := manager.DockerClient.ContainerCreate(manager.DockerCtx, containerConfigPtr, containerHostConfigPtr, nil, "")
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not create Docker container from image %v.", serviceCfg.GetDockerImage())
		}
		containerId := resp.ID
		if err := manager.DockerClient.ContainerStart(manager.DockerCtx, containerId, types.ContainerStartOptions{}); err != nil {
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
func getContainerHostConfig(serviceConfig JsonRpcServiceConfig,  manager *DockerManager) (hostConfig *container.HostConfig, err error) {
	freeRpcPort, err := manager.GetFreePort()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	// TODO right nwo this is hardcoded - replace these with FreeHostPortProvider in the future, so we can have
	//  arbitrary service-specific ports!
	jsonRpcPortBinding := []nat.PortBinding{
		{
			HostIP: manager.GetLocalHostIp(),
			HostPort: freeRpcPort.Port(),
		},
	}

	// TODO cycle through serviceConfig.GetOtherPorts to bind every one, not just default gecko staking port
	freeStakingPort, err := manager.GetFreePort()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	stakingPortBinding := []nat.PortBinding{
		{
			HostIP: manager.GetLocalHostIp(),
			HostPort: freeStakingPort.Port(),
		},
	}

	containerHostConfigPtr := &container.HostConfig{
		PortBindings: nat.PortMap{
			DEFAULT_GECKO_HTTP_PORT: jsonRpcPortBinding,
			DEFAULT_GECKO_STAKING_PORT: stakingPortBinding,
		},
	}
	return containerHostConfigPtr, nil
}

