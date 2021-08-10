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
	updateFunctionCallsIps [][]net.IP
}

func NewMockNetworkingSidecar() *MockNetworkingSidecar {
	return &MockNetworkingSidecar{updateFunctionCallsIps: [][]net.IP{}}
}

func (sidecar MockNetworkingSidecar) GetIPAddr() net.IP {
	return net.IP("")
}

func (sidecar MockNetworkingSidecar) GetContainerID() string {
	return ""
}

func (sidecar MockNetworkingSidecar) InitializeIpTables(ctx context.Context) error {
	return nil
}

func (sidecar *MockNetworkingSidecar) UpdateIpTables(ctx context.Context, blockedIps []net.IP) error {
	sidecar.updateFunctionCallsIps = append(sidecar.updateFunctionCallsIps, blockedIps)
	return nil
}

func (sidecar MockNetworkingSidecar) GetRecordedUpdateIps() [][]net.IP {
	return sidecar.updateFunctionCallsIps
}
