/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"net"
)

type MockNetworkingSidecar struct {
	updateFunctionCallsPacketLossConfig []map[string]float32
}

func NewMockNetworkingSidecar() *MockNetworkingSidecar {
	return &MockNetworkingSidecar{updateFunctionCallsPacketLossConfig: []map[string]float32{}}
}

func (sidecar MockNetworkingSidecar) GetIPAddr() net.IP {
	return net.IP("")
}

func (sidecar MockNetworkingSidecar) GetContainerID() string {
	return ""
}

func (sidecar MockNetworkingSidecar) InitializeTrafficControl(ctx context.Context) error {
	return nil
}

func (sidecar *MockNetworkingSidecar) UpdateTrafficControl(ctx context.Context, allPacketLossPercentageForIpAddresses map[string]float32) error {
	sidecar.updateFunctionCallsPacketLossConfig = append(sidecar.updateFunctionCallsPacketLossConfig, allPacketLossPercentageForIpAddresses)
	return nil
}

func (sidecar MockNetworkingSidecar) GetRecordedUpdatePacketLossConfig() []map[string]float32 {
	return sidecar.updateFunctionCallsPacketLossConfig
}
