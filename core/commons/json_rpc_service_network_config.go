package commons

import (
	"github.com/docker/docker/api/types"
	"github.com/palantir/stacktrace"
)


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
		containerConfigPtr, err := manager.GetContainerCfgFromServiceCfg(serviceCfg)

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
