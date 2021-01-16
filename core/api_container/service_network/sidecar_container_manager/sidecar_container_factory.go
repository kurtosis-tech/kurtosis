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

type SidecarContainerFactory struct {
	dockerManager *docker_manager.DockerManager

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerNetworkId string

	sidecarImageName string

	// Command that will be executed on each new sidecar container to run it forever
	runForeverCmd []string

	shWrappingCmd func([]string) []string
}

func NewSidecarContainerFactory(dockerManager *docker_manager.DockerManager, freeIpAddrTracker *commons.FreeIpAddrTracker, dockerNetworkId string, sidecarImageName string, runForeverCmd []string, shWrappingCmd func([]string) []string) *SidecarContainerFactory {
	return &SidecarContainerFactory{dockerManager: dockerManager, freeIpAddrTracker: freeIpAddrTracker, dockerNetworkId: dockerNetworkId, sidecarImageName: sidecarImageName, runForeverCmd: runForeverCmd, shWrappingCmd: shWrappingCmd}
}

func (factory SidecarContainerFactory) Create(
		ctx context.Context,
		serviceId topology_types.ServiceID,
		serviceContainerId string) (SidecarContainer, error) {
	sidecarIp, err := factory.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting a free IP address for the sidecar container attached to service with ID '%v'")
	}
	sidecarContainerId, err := factory.dockerManager.CreateAndStartContainer(
		ctx,
		factory.sidecarImageName,
		factory.dockerNetworkId,
		sidecarIp,
		map[docker_manager.ContainerCapability]bool{
			docker_manager.NetAdmin: true,
		},
		docker_manager.NewContainerNetworkMode(serviceContainerId),
		map[nat.Port]*nat.PortBinding{},
		factory.runForeverCmd,
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
		factory.dockerManager,
		sidecarContainerId,
		factory.shWrappingCmd)

	sidecarContainer := NewStandardSidecarContainer(
		serviceId,
		sidecarContainerId,
		sidecarIp,
		*execCmdExecutor,
	)

	return sidecarContainer, nil
}