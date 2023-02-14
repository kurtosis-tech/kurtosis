/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"net"
)

type MockNetworkingSidecarWrapper struct {
	updateFunctionCallsPartitionConnectionConfig []map[string]*partition_topology.PartitionConnection
}

func NewMockNetworkingSidecarWrapper() *MockNetworkingSidecarWrapper {
	return &MockNetworkingSidecarWrapper{updateFunctionCallsPartitionConnectionConfig: []map[string]*partition_topology.PartitionConnection{}}
}

func (sidecar *MockNetworkingSidecarWrapper) GetServiceUUID() service.ServiceUUID {
	return ""
}

func (sidecar *MockNetworkingSidecarWrapper) GetIPAddr() net.IP {
	return net.IP("")
}

func (sidecar *MockNetworkingSidecarWrapper) InitializeTrafficControl(ctx context.Context) error {
	return nil
}

func (sidecar *MockNetworkingSidecarWrapper) UpdateTrafficControl(ctx context.Context, partitionConnectionConfigForIpAddresses map[string]*partition_topology.PartitionConnection) error {
	sidecar.updateFunctionCallsPartitionConnectionConfig = append(sidecar.updateFunctionCallsPartitionConnectionConfig, partitionConnectionConfigForIpAddresses)
	return nil
}

func (sidecar *MockNetworkingSidecarWrapper) GetRecordedUpdatedPacketConnectionConfig() []map[string]*partition_topology.PartitionConnection {
	return sidecar.updateFunctionCallsPartitionConnectionConfig
}
