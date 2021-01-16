/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package sidecar_container_manager

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

// ==========================================================================================
//                                        Interface
// ==========================================================================================
type SidecarContainerManager interface {
	CreateSidecarContainer(
	) error

	DestroySidecarContainer(
	) error
}

// ==========================================================================================
//                                      Implementation
// ==========================================================================================

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// This class's methods are NOT thread-safe - it's up to the caller to ensure that
//  only one change at a time is run on a given sidecar container!!!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
type StandardSidecarContainerManager struct {

	serviceMetadata map[topology_types.ServiceID]*serviceMetadata
}

// TODO constructor

// Adds a sidecar container attached to the given service ID
func (manager *StandardSidecarContainerManager) AddSidecarContainer(
		ctx context.Context,
		serviceContainerId string) error {
	// TODO create sidecar container
	sidecarContainerId := SidecarContainerID(sidecarContainerIdStr)
	manager.serviceMetadata[sidecarContainerId] = &serviceMetadata{
		sidecarContainerId: sidecarContainerId,
		sidecarIpAddr:      sidecarIp,
	}

	// TODO initialize sidecar cotainer

}

func (manager *StandardSidecarContainerManager) UpdateIpTablesForService(ctx context.Context, serviceId topology_types.ServiceID, blockedIps []net.IP) error {
	sidecarContainer, found := manager.serviceMetadata[serviceId]
	if !found {
		return stacktrace.NewError("No sidecar container found for service ID '%v'", serviceId)
	}
	return manager.wrapWithMutexLocking(
		serviceId,
		func() error { return manager.internalUpdateIpTablesForService(ctx, *sidecarContainer, blockedIps) },
	)
}

// ==========================================================================================
//                      Functions that will get wrapped with ipTablesMutex locking
// ==========================================================================================
// TODO Write tests for this, by extracting the logic to run exec commands on the sidecar into a separate, mockable
//  interface
func (manager *StandardSidecarContainerManager) internalUpdateIpTablesForService(ctx context.Context, sidecarState serviceMetadata, blockedIps []net.IP) error {
}

// ==========================================================================================
//                                    Private helper functions
// ==========================================================================================
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
