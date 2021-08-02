/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package docker_network_allocator

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestErrorOnNoFreeIps(t *testing.T) {
	cidrs := []string{
		"0.0.0.0/0",
	}
	networks := parseNetworks(t, cidrs)
	_, err := findRandomFreeNetwork(networks)
	assert.Error(t, err)

}

func TestExactHole(t *testing.T) {
	// This has exactly one hole - at 0.0.0.0/20
	cidrs := []string{
		"0.0.16.0/20",
		"0.0.32.0/19",
		"0.0.64.0/18",
		"0.0.128.0/17",
		"0.1.0.0/16",
		"0.2.0.0/15",
		"0.4.0.0/14",
		"0.8.0.0/13",
		"0.16.0.0/12",
		"0.32.0.0/11",
		"0.64.0.0/10",
		"0.128.0.0/9",
		"1.0.0.0/8",
		"2.0.0.0/7",
		"4.0.0.0/6",
		"8.0.0.0/5",
		"16.0.0.0/4",
		"32.0.0.0/3",
		"64.0.0.0/2",
		"128.0.0.0/1",
	}
	assertExpectedResultGivenCidrs(t, cidrs, []byte{0, 0, 0, 0})
}

func TestNetworkFoundOnVariousCases(t *testing.T) {
	// Because the free network-finding is random, we just test that we don't get an error (i.e. we actually find a network)
	successfulCases := [][]string{
		{
			"0.0.16.0/20",
		},
		{
			"0.0.32.0/20",
		},
		{
			"0.0.4.0/24",
		},
		{
			"0.0.0.0/20",
			"0.0.32.0/20",
		},
		{
			"0.0.4.0/24",
			"0.0.64.0/24",
		},
		{
			"0.0.4.0/24",
			"0.0.18.0/24",
		},
		{
			"0.0.4.0/24",
			"0.0.18.0/24",
			"0.0.32.0/24",
		},
		{
			"0.0.0.0/18",
			"0.80.0.0/24",
		},
	}
	for _, cidrs := range successfulCases {
		parsedNetworks := parseNetworks(t, cidrs)
		_, err := findRandomFreeNetwork(parsedNetworks)
		assert.NoError(t, err, "Got an unexpected error when finding a free network with already-occupied CIDRS %+v", cidrs)
	}
}

func assertExpectedResultGivenCidrs(t *testing.T, cidrs []string, expectedIp net.IP) {
	networks := parseNetworks(t, cidrs)
	result, err := findRandomFreeNetwork(networks)
	assert.NoError(t, err)

	assert.Equal(t, expectedIp, result.IP)

	maskNumOnes, maskTotalBits := result.Mask.Size()
	assert.Equal(t, networkWidthBits, uint32(maskTotalBits - maskNumOnes))
}

func parseNetworks(t *testing.T, cidrs []string) []*net.IPNet {
	result := []*net.IPNet{}
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		assert.NoError(t, err)
		result = append(result, network)
	}
	return result
}
