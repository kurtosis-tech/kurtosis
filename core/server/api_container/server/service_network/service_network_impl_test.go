/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/stretchr/testify/require"
	"net"
	"strconv"
	"testing"
)

const (
	packetLossConfigForBlockedPartition = float32(100)
)

func TestUpdateTrafficControl(t *testing.T) {
	numServices := 10
	ctx := context.Background()

	sidecars := map[service.ServiceGUID]networking_sidecar.NetworkingSidecarWrapper{}
	mockSidecars := map[service.ServiceGUID]*networking_sidecar.MockNetworkingSidecarWrapper{}
	serviceGUIDsToIDs := map[service.ServiceGUID]service_network_types.ServiceID{}
	serviceIDsToGUIDs := map[service_network_types.ServiceID]service.ServiceGUID{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		serviceGUID := newServiceGUID(serviceId)
		serviceGUIDsToIDs[serviceGUID] = serviceId
		serviceIDsToGUIDs[serviceId] = serviceGUID
		sidecar := networking_sidecar.NewMockNetworkingSidecarWrapper()
		sidecars[serviceGUID] = sidecar
		mockSidecars[serviceGUID] = sidecar
	}

	registrationInfo := map[service.ServiceGUID]serviceRegistrationInfo{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		serviceGuid := serviceIDsToGUIDs[serviceId]
		ip := testIpFromInt(i)
		registrationInfo[serviceGuid] = serviceRegistrationInfo{privateIpAddr: ip}
	}

	// Creates the pathological "line" of connections, where each service can only see the services adjacent
	targetServicePacketLossConfigs := map[service.ServiceGUID]map[service.ServiceGUID]float32{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		serviceGUID := serviceIDsToGUIDs[serviceId]
		otherServicesPacketLossConfig := map[service.ServiceGUID]float32{}
		for j := 0; j < numServices; j++ {
			if j < i - 1 || j > i + 1 {
				blockedServiceId := testServiceIdFromInt(j)
				blockedServiceGUID := serviceIDsToGUIDs[blockedServiceId]
				otherServicesPacketLossConfig[blockedServiceGUID] = packetLossConfigForBlockedPartition
			}
		}
		targetServicePacketLossConfigs[serviceGUID] = otherServicesPacketLossConfig
	}


	require.Nil(t, updateTrafficControlConfiguration(ctx, targetServicePacketLossConfigs, registrationInfo, sidecars))

	// Verify that each service got told to block exactly the right things
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)

		expected := map[string]float32{}
		for j := 0; j < numServices; j++ {
			if j < i - 1 || j > i + 1 {
				ip := testIpFromInt(j)
				expected[ip.String()] = packetLossConfigForBlockedPartition
			}
		}

		serviceGUID := serviceIDsToGUIDs[serviceId]
		mockSidecar := mockSidecars[serviceGUID]
		recordedPacketLossConfig := mockSidecar.GetRecordedUpdatePacketLossConfig()
		require.Equal(t, 1, len(recordedPacketLossConfig), "Expected sidecar for service ID '%v' to have recorded exactly one call to update")

		actualPacketLossConfigForService := recordedPacketLossConfig[0]

		require.Equal(t, expected, actualPacketLossConfigForService)
	}
}

func testIpFromInt(i int) net.IP {
	return []byte{1, 1, 1, byte(i)}
}

func testServiceIdFromInt(i int) service_network_types.ServiceID {
	return service_network_types.ServiceID("service-" + strconv.Itoa(i))
}
