/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/topology_types"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"strings"
)

const (
	networkingSidecarImageName = "kurtosistech/iproute2"
)

// We sleep forever because all the commands this container will run will be executed
//  via Docker exec
var sidecarContainerCommand = []string{
	"sleep","infinity",
}

// Embeds the given command in a call to whichever shell is native to the image, so that a command with things
//  like '&&' will get executed as expected
var sidecarContainerShWrapper = func(unwrappedCmd []string) []string {
	return []string{
		"sh",
		"-c",
		strings.Join(unwrappedCmd, " "),
	}
}


// ==========================================================================================
//                                        Interface
// ==========================================================================================
type NetworkingSidecarManager interface {
	Create(ctx context.Context, serviceId topology_types.ServiceID, serviceContainerId string) (NetworkingSidecar, error)
	Destroy(ctx context.Context, sidecar NetworkingSidecar) error
}

// ==========================================================================================
//                                      Implementation
// ==========================================================================================

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// This class's methods are NOT thread-safe - it's up to the caller to ensure that
//  only one change at a time is run on a given sidecar container!!!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
type StandardNetworkingSidecarManager struct {
	dockerManager *docker_manager.DockerManager

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerNetworkId string
}

func NewStandardNetworkingSidecarManager(dockerManager *docker_manager.DockerManager, freeIpAddrTracker *commons.FreeIpAddrTracker, dockerNetworkId string) *StandardNetworkingSidecarManager {
	return &StandardNetworkingSidecarManager{dockerManager: dockerManager, freeIpAddrTracker: freeIpAddrTracker, dockerNetworkId: dockerNetworkId}
}


// Adds a sidecar container attached to the given service ID
func (manager *StandardNetworkingSidecarManager) Create(
		ctx context.Context,
		serviceId topology_types.ServiceID,
		serviceContainerId string) (NetworkingSidecar, error) {
	sidecarIp, err := manager.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting a free IP address for the sidecar container attached to service with ID '%v'",
			serviceId)
	}
	sidecarContainerId, err := manager.dockerManager.CreateAndStartContainer(
		ctx,
		networkingSidecarImageName,
		manager.dockerNetworkId,
		sidecarIp,
		map[docker_manager.ContainerCapability]bool{
			docker_manager.NetAdmin: true,
		},
		docker_manager.NewContainerNetworkMode(serviceContainerId),
		map[nat.Port]*nat.PortBinding{},
		sidecarContainerCommand,
		map[string]string{}, // No environment variables
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		map[string]string{}, // No volume mounts either
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred starting the sidecar container attached to service with ID '%v'",
			serviceId,
		)
	}

	execCmdExecutor := newStandardSidecarExecCmdExecutor(
		manager.dockerManager,
		sidecarContainerId,
		sidecarContainerShWrapper)

	sidecarContainer := NewStandardNetworkingSidecar(
		serviceId,
		sidecarContainerId,
		sidecarIp,
		*execCmdExecutor,
	)

	return sidecarContainer, nil
}

func (manager *StandardNetworkingSidecarManager) Destroy(
		ctx context.Context,
		sidecar NetworkingSidecar) error {
	sidecarContainerId := sidecar.GetContainerID()
	sidecarIp := sidecar.GetIPAddr()
	if err := manager.dockerManager.KillContainer(ctx, sidecarContainerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred stopping the sidecar with container ID '%v'",
			sidecarContainerId)
	}
	manager.freeIpAddrTracker.ReleaseIpAddr(sidecarIp)
	return nil
}
