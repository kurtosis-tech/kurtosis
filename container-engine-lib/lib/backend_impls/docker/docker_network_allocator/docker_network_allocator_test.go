package docker_network_allocator

import (
	"context"
	"math/big"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorOnInstantiationWithoutConstructor(t *testing.T) {
	allocator := DockerNetworkAllocator{
		isConstructedViaConstructor: false,
		dockerManager:               nil,
	}
	_, err := allocator.CreateNewNetwork(context.Background(), "", map[string]string{})
	assert.Error(t, err)
}

func TestErrorOnNoFreeIps(t *testing.T) {
	cidrs := []string{
		"0.0.0.0/0",
	}
	networks := parseNetworks(t, cidrs)
	_, err := findRandomFreeNetwork(networks)
	assert.Error(t, err)
}

func TestMaskWidthChanged(t *testing.T) {
	require.EqualValues(t, 16, networkWidthBits+enclaveWidthBits, "findRandomFreeNetwork only works for 16 bit mask width, please review it since the constraint was violated")
}

func TestEntireNetworkingSpace(t *testing.T) {
	takenNetworks := []*net.IPNet{}
	for i := 0; i < 1<<enclaveWidthBits; i++ {
		freeIPAddress, err := findRandomFreeNetwork(takenNetworks)
		require.NoError(t, err, "Got an unexpected error when finding a free network with already-occupied networks %+v (len %v)", takenNetworks, len(takenNetworks))
		require.NotContains(t, takenNetworks, freeIPAddress)
		takenNetworks = append(takenNetworks, freeIPAddress)
		require.EqualValues(t, 1<<networkWidthBits, countIPAddresses(freeIPAddress))
	}
}

func countIPAddresses(ipNet *net.IPNet) int64 {
	ones, bits := ipNet.Mask.Size()
	numIPs := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(bits-ones)), nil)
	return numIPs.Int64()
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
