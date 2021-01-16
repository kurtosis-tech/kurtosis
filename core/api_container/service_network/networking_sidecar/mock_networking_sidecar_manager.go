/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
)

type MockNetworkingSidecarManager struct {
	index int
}

func NewMockNetworkingSidecarManager(index int) *MockNetworkingSidecarManager {
	return &MockNetworkingSidecarManager{index: index}
}

func (m MockNetworkingSidecarManager) Create(ctx context.Context, serviceId topology_types.ServiceID, serviceContainerId string) (NetworkingSidecar, error) {
	panic("implement me")
}

func (m MockNetworkingSidecarManager) Destroy(ctx context.Context, sidecar NetworkingSidecar) error {
	panic("implement me")
}

