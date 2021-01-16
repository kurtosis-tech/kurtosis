/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package sidecar_container_manager

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
type SidecarContainerManager interface {
	Create(ctx context.Context, serviceId topology_types.ServiceID, serviceContainerId string) (SidecarContainer, error)
	Destroy(ctx context.Context, sidecar SidecarContainer) error
}

// ==========================================================================================
//                                      Implementation
// ==========================================================================================

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// This class's methods are NOT thread-safe - it's up to the caller to ensure that
//  only one change at a time is run on a given sidecar container!!!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
type StandardSidecarContainerManager struct {
	dockerManager *docker_manager.DockerManager

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerNetworkId string

	sidecarImageName string

	// Command that will be executed on each new sidecar container to run it forever
	runForeverCmd []string

	shWrappingCmd func([]string) []string
}

func NewStandardSidecarContainerManager(dockerManager *docker_manager.DockerManager, freeIpAddrTracker *commons.FreeIpAddrTracker, dockerNetworkId string, sidecarImageName string, runForeverCmd []string, shWrappingCmd func([]string) []string) *StandardSidecarContainerManager {
	return &StandardSidecarContainerManager{dockerManager: dockerManager, freeIpAddrTracker: freeIpAddrTracker, dockerNetworkId: dockerNetworkId, sidecarImageName: sidecarImageName, runForeverCmd: runForeverCmd, shWrappingCmd: shWrappingCmd}
}

// Adds a sidecar container attached to the given service ID
func (manager *StandardSidecarContainerManager) Create(
		ctx context.Context,
		serviceId topology_types.ServiceID,
		serviceContainerId string) (SidecarContainer, error) {
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

	execCmdExecutor := NewSidecarExecCmdExecutor(
		manager.dockerManager,
		sidecarContainerId,
		manager.shWrappingCmd)

	sidecarContainer := NewStandardSidecarContainer(
		serviceId,
		sidecarContainerId,
		sidecarIp,
		*execCmdExecutor,
	)

	return sidecarContainer, nil
}

func (manager *StandardSidecarContainerManager) Destroy(
		ctx context.Context,
		sidecar SidecarContainer) error {
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


	// ==========================================================================================
//                                    Private helper functions
// ==========================================================================================
/*
func (manager *StandardSidecarContainerManager) wrapWithMutexLocking(serviceId topology_types.ServiceID, delegate func() error) error {
	mutex, found := manager.mutexes[serviceId]
	if !found {
		return stacktrace.NewError("Could not find ipTablesMutex for service ID '%v'", serviceId)
	}
	mutex.Lock()
	defer mutex.Unlock()

	if err := delegate(); err != nil {
		return stacktrace.Propagate(err, "An error occurred in the delegate function")
	}
	return nil
}

 */