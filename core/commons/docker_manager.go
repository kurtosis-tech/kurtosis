package commons

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"strconv"
)

// TODO TODO TODO - do we ever need to handle different local host IPs?
const LOCAL_HOST_IP = "0.0.0.0"

// TODO TODO TODO get these from serviceconfig
const DEFAULT_GECKO_HTTP_PORT = nat.Port("9650/tcp")
const DEFAULT_GECKO_STAKING_PORT = nat.Port("9651/tcp")

type DockerManager struct {
	dockerCtx           context.Context
	dockerClient        *client.Client
	freeHostPortTracker *FreeHostPortTracker
}

func NewDockerManager(dockerCtx context.Context, dockerClient *client.Client, hostPortRangeStart int, hostPortRangeEnd int) *DockerManager {
	return &DockerManager{
		dockerCtx:           dockerCtx,
		dockerClient:        dockerClient,
		freeHostPortTracker: NewFreeHostPortTracker(hostPortRangeStart, hostPortRangeEnd),
	}
}

func (manager DockerManager) getFreePort() (freePort *nat.Port, err error) {
	freePortInt, err := manager.freeHostPortTracker.GetFreePort()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	port, err := nat.NewPort("tcp", strconv.Itoa(freePortInt))
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	return &port, nil
}

func (manager DockerManager) getLocalHostIp() string {
	return LOCAL_HOST_IP
}

// Creates a Docker-Container-To-Host Port mapping, defining how a Container's JSON RPC and service-specific ports are
// mapped to the host ports
func (manager *DockerManager) GetContainerHostConfig(serviceConfig JsonRpcServiceConfig) (hostConfig *container.HostConfig, err error) {
	freeRpcPort, err := manager.getFreePort()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	// TODO right nwo this is hardcoded - replace these with FreeHostPortProvider in the future, so we can have
	//  arbitrary service-specific ports!
	jsonRpcPortBinding := []nat.PortBinding{
		{
			HostIP: manager.getLocalHostIp(),
			HostPort: freeRpcPort.Port(),
		},
	}

	// TODO cycle through serviceConfig.GetOtherPorts to bind every one, not just default gecko staking port
	freeStakingPort, err := manager.getFreePort()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	stakingPortBinding := []nat.PortBinding{
		{
			HostIP: manager.getLocalHostIp(),
			HostPort: freeStakingPort.Port(),
		},
	}

	httpPort, err := nat.NewPort("tcp", strconv.Itoa(serviceConfig.GetJsonRpcPort()))
	stakingPort, err := nat.NewPort("tcp", strconv.Itoa(serviceConfig.GetOtherPorts()[0]))
	containerHostConfigPtr := &container.HostConfig{
		PortBindings: nat.PortMap{
			httpPort: jsonRpcPortBinding,
			stakingPort: stakingPortBinding,
		},
	}
	return containerHostConfigPtr, nil
}