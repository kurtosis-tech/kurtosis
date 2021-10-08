/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/service_network_types"
	"github.com/stretchr/testify/assert"
	"net"
	"strconv"
	"testing"
)

func TestUpdateIpTables(t *testing.T) {
	numServices := 10
	ctx := context.Background()

	sidecars := map[service_network_types.ServiceID]networking_sidecar.NetworkingSidecar{}
	mockSidecars := map[service_network_types.ServiceID]*networking_sidecar.MockNetworkingSidecar{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		sidecar := networking_sidecar.NewMockNetworkingSidecar()
		sidecars[serviceId] = sidecar
		mockSidecars[serviceId] = sidecar
	}

	registrationInfo := map[service_network_types.ServiceID]serviceRegistrationInfo{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		serviceGUID := newServiceGUID(serviceId)
		ip := testIpFromInt(i)
		registrationInfo[serviceId] = serviceRegistrationInfo{serviceGUID: serviceGUID, ipAddr: ip}
	}

	// Creates the pathological "line" of connections, where each service can only see the services adjacent
	targetBlocklists := map[service_network_types.ServiceID]*service_network_types.ServiceIDSet{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		blockedSet := service_network_types.NewServiceIDSet()
		for j := 0; j < numServices; j++ {
			if j < i - 1 || j > i + 1 {
				blockedServiceId := testServiceIdFromInt(j)
				blockedSet.AddElem(blockedServiceId)
			}
		}
		targetBlocklists[serviceId] = blockedSet
	}

	assert.Nil(t, updateIpTables(ctx, targetBlocklists, registrationInfo, sidecars))

	// Verify that each service got told to block exactly the right things
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)

		expected := map[string]bool{}
		for j := 0; j < numServices; j++ {
			if j < i - 1 || j > i + 1 {
				ip := testIpFromInt(j)
				expected[ip.String()] = true
			}
		}

		mockSidecar := mockSidecars[serviceId]
		recordedUpdateIps := mockSidecar.GetRecordedUpdateIps()
		assert.Equal(t, 1, len(recordedUpdateIps), "Expected sidecar for service ID '%v' to have recorded exactly one call to update")

		firstRecordedIps := recordedUpdateIps[0]
		actual := map[string]bool{}
		for _, ip := range firstRecordedIps {
			actual[ip.String()] = true
		}

		assert.Equal(t, expected, actual)
	}
}

func testIpFromInt(i int) net.IP {
	return []byte{1, 1, 1, byte(i)}
}

func testServiceIdFromInt(i int) service_network_types.ServiceID {
	return service_network_types.ServiceID("service-" + strconv.Itoa(i))
}
