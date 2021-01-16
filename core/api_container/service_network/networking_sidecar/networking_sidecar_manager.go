/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
)

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

	sidecarImageName string

	// Command that will be executed on each new sidecar container to run it forever
	runForeverCmd []string

	shWrappingCmd func([]string) []string
}

func NewStandardNetworkingSidecarManager(dockerManager *docker_manager.DockerManager, freeIpAddrTracker *commons.FreeIpAddrTracker, dockerNetworkId string, sidecarImageName string, runForeverCmd []string, shWrappingCmd func([]string) []string) *StandardNetworkingSidecarManager {
	return &StandardNetworkingSidecarManager{dockerManager: dockerManager, freeIpAddrTracker: freeIpAddrTracker, dockerNetworkId: dockerNetworkId, sidecarImageName: sidecarImageName, runForeverCmd: runForeverCmd, shWrappingCmd: shWrappingCmd}
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
			"An error occurred getting a free IP address for the sidecar container attached to service with ID '%v'")
	}
	sidecarContainerId, err := manager.dockerManager.CreateAndStartContainer(
		ctx,
		manager.sidecarImageName,
		manager.dockerNetworkId,
		sidecarIp,
		map[docker_manager.ContainerCapability]bool{
			docker_manager.NetAdmin: true,
		},
		docker_manager.NewContainerNetworkMode(serviceContainerId),
		map[nat.Port]*nat.PortBinding{},
		manager.runForeverCmd,
		map[string]string{}, // No environment variables
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		map[string]string{}, // No volume mounts either
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred starting the sidecar container attached to service with ID '%v'",
		)
	}

	execCmdExecutor := newSidecarExecCmdExecutor(
		manager.dockerManager,
		sidecarContainerId,
		manager.shWrappingCmd)

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
