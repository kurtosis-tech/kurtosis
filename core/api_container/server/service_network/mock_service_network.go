/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/service_network_types"
	"github.com/palantir/stacktrace"
	"net"
	"time"
)

type MockServiceNetwork struct {
	serviceIps                       map[service_network_types.ServiceID]net.IP
	serviceEnclaveDataVolMntDirpaths map[service_network_types.ServiceID]string
}

func NewMockServiceNetwork(serviceIps map[service_network_types.ServiceID]net.IP, serviceEnclaveDataVolMntDirpaths map[service_network_types.ServiceID]string) *MockServiceNetwork {
	return &MockServiceNetwork{serviceIps: serviceIps, serviceEnclaveDataVolMntDirpaths: serviceEnclaveDataVolMntDirpaths}
}

func (m MockServiceNetwork) Repartition(ctx context.Context, newPartitionServices map[service_network_types.PartitionID]*service_network_types.ServiceIDSet, newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection, newDefaultConnection partition_topology.PartitionConnection) error {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) RegisterService(serviceId service_network_types.ServiceID, partitionId service_network_types.PartitionID) (net.IP, string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) StartService(ctx context.Context, serviceId service_network_types.ServiceID, imageName string, usedPorts map[nat.Port]bool, entrypointArgs []string, cmdArgs []string, dockerEnvVars map[string]string, enclaveDataVolMntDirpath string, filesArtifactMountDirpaths map[string]string) (map[nat.Port]*nat.PortBinding, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) RemoveService(ctx context.Context, serviceId service_network_types.ServiceID, containerStopTimeout time.Duration) error {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) ExecCommand(ctx context.Context, serviceId service_network_types.ServiceID, command []string) (int32, string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) GetServiceIP(serviceId service_network_types.ServiceID) (net.IP, error) {
	ip, found := m.serviceIps[serviceId]
	if !found {
		return nil, stacktrace.NewError("No IP defined for service with ID '%v'", serviceId)
	}
	return ip, nil
}

func (m MockServiceNetwork) GetServiceEnclaveDataVolMntDirpath(serviceId service_network_types.ServiceID) (string, error) {
	volMntDirPath, found := m.serviceEnclaveDataVolMntDirpaths[serviceId]
	if !found {
		return "", stacktrace.NewError("No volume directory path defined for service with ID '%v'", serviceId)
	}
	return volMntDirPath, nil
}

func (m MockServiceNetwork) GetRelativeServiceDirpath(serviceId service_network_types.ServiceID) (string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) GetServiceIDs() map[service_network_types.ServiceID]bool {
	panic("This is unimplemented for the mock network")
}
