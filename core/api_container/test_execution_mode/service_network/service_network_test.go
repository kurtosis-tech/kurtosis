/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/topology_types"
	"github.com/stretchr/testify/assert"
	"net"
	"strconv"
	"testing"
)

func TestUpdateIpTables(t *testing.T) {
	numServices := 10
	ctx := context.Background()

	sidecars := map[topology_types.ServiceID]networking_sidecar.NetworkingSidecar{}
	mockSidecars := map[topology_types.ServiceID]*networking_sidecar.MockNetworkingSidecar{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		sidecar := networking_sidecar.NewMockNetworkingSidecar()
		sidecars[serviceId] = sidecar
		mockSidecars[serviceId] = sidecar
	}

	serviceIps := map[topology_types.ServiceID]net.IP{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		ip := testIpFromInt(i)
		serviceIps[serviceId] = ip
	}

	// Creates the pathological "line" of connections, where each service can only see the services adjacent
	targetBlocklists := map[topology_types.ServiceID]*topology_types.ServiceIDSet{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		blockedSet := topology_types.NewServiceIDSet()
		for j := 0; j < numServices; j++ {
			if j < i - 1 || j > i + 1 {
				blockedServiceId := testServiceIdFromInt(j)
				blockedSet.AddElem(blockedServiceId)
			}
		}
		targetBlocklists[serviceId] = blockedSet
	}

	assert.Nil(t, updateIpTables(ctx, targetBlocklists, serviceIps, sidecars))

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

func testServiceIdFromInt(i int) topology_types.ServiceID {
	return topology_types.ServiceID("service-" + strconv.Itoa(i))
}
