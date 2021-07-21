/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_network

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/palantir/stacktrace"
	"net"
	"time"
)

type MockServiceNetwork struct {
	serviceIps map[service_network_types.ServiceID]net.IP
	serviceSuiteExecutionVolMntDirpaths map[service_network_types.ServiceID]string
}

func NewMockServiceNetwork(serviceIps map[service_network_types.ServiceID]net.IP, serviceSuiteExecutionVolMntDirpaths map[service_network_types.ServiceID]string) *MockServiceNetwork {
	return &MockServiceNetwork{serviceIps: serviceIps, serviceSuiteExecutionVolMntDirpaths: serviceSuiteExecutionVolMntDirpaths}
}

func (m MockServiceNetwork) Repartition(ctx context.Context, newPartitionServices map[service_network_types.PartitionID]*service_network_types.ServiceIDSet, newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection, newDefaultConnection partition_topology.PartitionConnection) error {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) RegisterService(serviceId service_network_types.ServiceID, partitionId service_network_types.PartitionID) (net.IP, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) GenerateFiles(serviceId service_network_types.ServiceID, filesToGenerate map[string]*kurtosis_core_rpc_api_bindings.FileGenerationOptions) (map[string]string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) LoadStaticFiles(serviceId service_network_types.ServiceID, staticFileIdKeys map[string]bool) (map[string]string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) StartService(ctx context.Context, serviceId service_network_types.ServiceID, imageName string, usedPorts map[nat.Port]bool, entrypointArgs []string, cmdArgs []string, dockerEnvVars map[string]string, suiteExecutionVolMntDirpath string, filesArtifactMountDirpaths map[string]string) (map[nat.Port]*nat.PortBinding, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) RemoveService(ctx context.Context, serviceId service_network_types.ServiceID, containerStopTimeout time.Duration) error {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) ExecCommand(ctx context.Context, serviceId service_network_types.ServiceID, command []string) (int32, *bytes.Buffer, error) {
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
	volMntDirPath, found := m.serviceSuiteExecutionVolMntDirpaths[serviceId]
	if !found {
		return "", stacktrace.NewError("No volume directory path defined for service with ID '%v'", serviceId)
	}
	return volMntDirPath, nil
}

func (m MockServiceNetwork) Destroy(ctx context.Context, containerStopTimeout time.Duration) error {
	panic("This is unimplemented for the mock network")
}
