/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"time"
)

type MockServiceNetwork struct {
	servicePrivateIps                map[service.ServiceID]net.IP
	serviceEnclaveDataDirMntDirpaths map[service.ServiceID]string
}

func NewMockServiceNetwork(serviceIps map[service.ServiceID]net.IP, serviceEnclaveDataDirMntDirpaths map[service.ServiceID]string) *MockServiceNetwork {
	return &MockServiceNetwork{servicePrivateIps: serviceIps, serviceEnclaveDataDirMntDirpaths: serviceEnclaveDataDirMntDirpaths}
}

func (m MockServiceNetwork) Repartition(ctx context.Context, newPartitionServices map[service_network_types.PartitionID]map[service.ServiceID]bool, newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection, newDefaultConnection partition_topology.PartitionConnection) error {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) RegisterService(serviceId service.ServiceID, partitionId service_network_types.PartitionID) (net.IP, string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) StartService(ctx context.Context, serviceId service.ServiceID, imageName string, privatePorts map[string]*port_spec.PortSpec, entrypointArgs []string, cmdArgs []string, dockerEnvVars map[string]string, enclaveDataDirMountDirpath string, oldFilesArtifactMountDirpaths map[files_artifact.FilesArtifactID]string, filesArtifactMountDirpaths map[files_artifact.FilesArtifactID]string) (resultPublicIpAddr net.IP, resultPublicPorts map[string]*port_spec.PortSpec, resultErr error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) RemoveService(ctx context.Context, serviceId service.ServiceID, containerStopTimeout time.Duration) error {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) ExecCommand(ctx context.Context, serviceId service.ServiceID, command []string) (int32, string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) GetServiceRegistrationInfo(serviceId service.ServiceID) (privateIpAddr net.IP, relativeServiceDirpath string, resultErr error) {
	ip, found := m.servicePrivateIps[serviceId]
	if !found {
		return nil, "", stacktrace.NewError("No private IP defined for service with ID '%v'", serviceId)
	}
	return ip, "", nil
}

func (m MockServiceNetwork) GetServiceRunInfo(serviceId service.ServiceID) (privatePorts map[string]*port_spec.PortSpec, maybePublicIpAddr net.IP, publicPorts map[string]*port_spec.PortSpec, enclaveDataDirMntDirpath string, resultErr error) {
	dataDirMntDirpath, found := m.serviceEnclaveDataDirMntDirpaths[serviceId]
	if !found {
		return nil, nil, nil, "", stacktrace.NewError("No enclave data directory mount dirpath defined for service with ID '%v'", serviceId)
	}
	return nil, nil, nil, dataDirMntDirpath, nil
}

func (m MockServiceNetwork) GetServiceIDs() map[service.ServiceID]bool {
	panic("This is unimplemented for the mock network")
}
