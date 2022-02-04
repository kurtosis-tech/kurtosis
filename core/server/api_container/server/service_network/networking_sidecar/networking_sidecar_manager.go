/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"strings"

	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	networkingSidecarImageName = "kurtosistech/iproute2"
)

// We sleep forever because all the commands this container will run will be executed
//  via Docker exec
var sidecarContainerCommand = []string{
	"sleep", "infinity",
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
	Add(ctx context.Context, serviceId service_network_types.ServiceGUID, serviceContainerId string) (NetworkingSidecar, error)
	Remove(ctx context.Context, sidecar NetworkingSidecar) error
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

	enclaveObjAttrProvider schema.EnclaveObjectAttributesProvider

	freeIpAddrTracker *lib.FreeIpAddrTracker

	dockerNetworkId string
}

func NewStandardNetworkingSidecarManager(dockerManager *docker_manager.DockerManager, enclaveObjAttrProvider schema.EnclaveObjectAttributesProvider, freeIpAddrTracker *lib.FreeIpAddrTracker, dockerNetworkId string) *StandardNetworkingSidecarManager {
	return &StandardNetworkingSidecarManager{dockerManager: dockerManager, enclaveObjAttrProvider: enclaveObjAttrProvider, freeIpAddrTracker: freeIpAddrTracker, dockerNetworkId: dockerNetworkId}
}

// Adds a sidecar container attached to the given service ID
func (manager *StandardNetworkingSidecarManager) Add(
	ctx context.Context,
	serviceGUID service_network_types.ServiceGUID,
	serviceContainerId string) (NetworkingSidecar, error) {
	sidecarIp, err := manager.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting a free IP address for the sidecar container attached to service with GUID '%v'",
			serviceGUID)
	}

	containerAttrs, err := manager.enclaveObjAttrProvider.ForNetworkingSidecarContainer(string(serviceGUID))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the attributes for the networking sidecar attached to service with GUID '%v'", serviceGUID)
	}
	containerName := containerAttrs.GetName()
	containerLabels := containerAttrs.GetLabels()
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		networkingSidecarImageName,
		containerName,
		manager.dockerNetworkId,
	).WithAlias(
		containerName,
	).WithStaticIP(
		sidecarIp,
	).WithAddedCapabilities(map[docker_manager.ContainerCapability]bool{
		docker_manager.NetAdmin: true,
	}).WithNetworkMode(
		docker_manager.NewContainerNetworkMode(serviceContainerId),
	).WithCmdArgs(
		sidecarContainerCommand,
	).WithLabels(
		containerLabels,
	).Build()
	sidecarContainerId, _, err := manager.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred starting the sidecar container attached to service with GUID '%v'",
			serviceGUID,
		)
	}

	execCmdExecutor := newStandardSidecarExecCmdExecutor(
		manager.dockerManager,
		sidecarContainerId,
		sidecarContainerShWrapper)

	sidecarContainer := NewStandardNetworkingSidecar(
		serviceGUID,
		sidecarContainerId,
		sidecarIp,
		*execCmdExecutor,
	)

	return sidecarContainer, nil
}

func (manager *StandardNetworkingSidecarManager) Remove(
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
