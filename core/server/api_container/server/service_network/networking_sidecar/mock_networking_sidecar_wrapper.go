/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"net"
)

type MockNetworkingSidecarWrapper struct {
	updateFunctionCallsPacketLossConfig []map[string]float32
}

func NewMockNetworkingSidecarWrapper() *MockNetworkingSidecarWrapper {
	return &MockNetworkingSidecarWrapper{updateFunctionCallsPacketLossConfig: []map[string]float32{}}
}

func (sidecar MockNetworkingSidecarWrapper) GetServiceGUID() service.ServiceGUID {
	return ""
}

func (sidecar MockNetworkingSidecarWrapper) GetIPAddr() net.IP {
	return net.IP("")
}

func (sidecar MockNetworkingSidecarWrapper) InitializeTrafficControl(ctx context.Context) error {
	return nil
}

func (sidecar *MockNetworkingSidecarWrapper) UpdateTrafficControl(ctx context.Context, allPacketLossPercentageForIpAddresses map[string]float32) error {
	sidecar.updateFunctionCallsPacketLossConfig = append(sidecar.updateFunctionCallsPacketLossConfig, allPacketLossPercentageForIpAddresses)
	return nil
}

func (sidecar MockNetworkingSidecarWrapper) GetRecordedUpdatePacketLossConfig() []map[string]float32 {
	return sidecar.updateFunctionCallsPacketLossConfig
}
