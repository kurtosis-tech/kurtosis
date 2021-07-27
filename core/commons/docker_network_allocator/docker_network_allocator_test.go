/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package docker_network_allocator

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

/*
8 = 1000
9 = 1001
10 = 1010

 */

func TestCreateNetworkFromIp(t *testing.T) {
	ipBytes := []byte{192, 168, 0, 1}
	desiredWidthBits := uint32(8)
	network := createNetworkFromIpAndWidth(binary.BigEndian.Uint32(ipBytes), desiredWidthBits)
	assert.Equal(t, net.IP(ipBytes), network.IP)

	numMaskOnes, totalMaskBits := network.Mask.Size()
	assert.Equal(t, desiredWidthBits, uint32(totalMaskBits - numMaskOnes))
}

func TestErrorOnNoNetworks(t *testing.T) {
	_, err := findFreeNetwork(uint32(8), []*net.IPNet{})
	assert.Error(t, err)
}

func TestErrorOnTooBigNetwork(t *testing.T) {
	cidrs := []string{
		"0.0.1.0/24",
	}
	networks := parseNetworks(t, cidrs)
	_, err := findFreeNetwork(uint32(400), networks)
	assert.Error(t, err)
}

func TestErrorOnZeroWidthNetwork(t *testing.T) {
	cidrs := []string{
		"0.0.1.0/24",
	}
	networks := parseNetworks(t, cidrs)
	_, err := findFreeNetwork(uint32(0), networks)
	assert.Error(t, err)
}

func TestErrorOnNoFreeIps(t *testing.T) {
	cidrs := []string{
		"0.0.0.0/0",
	}
	networks := parseNetworks(t, cidrs)
	_, err := findFreeNetwork(uint32(1), networks)
	assert.Error(t, err)

}

func TestExactHoleBeforeNetwork(t *testing.T) {
	cidrs := []string{
		"0.0.1.0/24",
	}
	assertExpectedResultGivenCidrs(t, cidrs, net.IP([]byte{0, 0, 0, 0}), uint32(8))
}

func TestLooseHoleBeforeNetwork(t *testing.T) {
	cidrs := []string{
		"0.0.2.0/24",
	}
	assertExpectedResultGivenCidrs(t, cidrs, net.IP([]byte{0, 0, 0, 0}), uint32(8))
}

func TestTooSmallHoleBeforeNetwork(t *testing.T) {
	cidrs := []string{
		"0.0.0.128/25",
	}
	assertExpectedResultGivenCidrs(t, cidrs, net.IP([]byte{0, 0, 1, 0}), uint32(8))
}

func TestExactHoleBetweenNetworks(t *testing.T) {
	cidrs := []string{
		"0.0.0.0/24",
		"0.0.2.0/24",
	}
	assertExpectedResultGivenCidrs(t, cidrs, net.IP([]byte{0, 0, 1, 0}), uint32(8))
}

func TestLooseHoleBetweenNetworks(t *testing.T) {
	cidrs := []string{
		"0.0.0.0/24",
		"0.0.3.0/24",
	}
	assertExpectedResultGivenCidrs(t, cidrs, net.IP([]byte{0, 0, 1, 0}), uint32(8))
}

func TestTooSmallHoleBetweenNetworks(t *testing.T) {
	cidrs := []string{
		"0.0.0.0/24",
		"0.0.1.128/25",
	}
	assertExpectedResultGivenCidrs(t, cidrs, net.IP([]byte{0, 0, 2, 0}), uint32(8))
}

func TestMultipleTooSmallHoleBetweenNetworks(t *testing.T) {
	cidrs := []string{
		"0.0.0.0/24",
		"0.0.1.128/25",
		"0.0.2.128/25",
	}
	assertExpectedResultGivenCidrs(t, cidrs, net.IP([]byte{0, 0, 3, 0}), uint32(8))
}

func assertExpectedResultGivenCidrs(t *testing.T, cidrs []string, expectedIp net.IP, desiredWidthBits uint32) {
	networks := parseNetworks(t, cidrs)
	result, err := findFreeNetwork(desiredWidthBits, networks)
	assert.NoError(t, err)

	assert.Equal(t, expectedIp, result.IP)

	maskNumOnes, maskTotalBits := result.Mask.Size()
	assert.Equal(t, desiredWidthBits, uint32(maskTotalBits - maskNumOnes))
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
