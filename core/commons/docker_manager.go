package commons

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"strconv"
)

// TODO TODO TODO - do we ever need to handle different local host IPs?
const LOCAL_HOST_IP = "0.0.0.0"

type DockerManager struct {
	DockerCtx context.Context
	DockerClient *client.Client
	freeHostPortTracker *FreeHostPortTracker
}

func NewDockerManager(dockerCtx context.Context, dockerClient *client.Client, hostPortRangeStart int, hostPortRangeEnd int) *DockerManager {
	return &DockerManager{
		DockerCtx: dockerCtx,
		DockerClient: dockerClient,
		freeHostPortTracker: NewFreeHostPortTracker(hostPortRangeStart, hostPortRangeEnd),
	}
}

func (manager DockerManager) GetFreePort() (freePort *nat.Port, err error) {
	freePortInt, err := manager.freeHostPortTracker.GetFreePort()
	if err != nil {
		return nil, stacktrace.Propagate(err, "")
	}
	port := nat.Port(strconv.Itoa(freePortInt) + "/tcp")
	return &port, nil
}

func (manager DockerManager) GetLocalHostIp() string {
	return LOCAL_HOST_IP
}
