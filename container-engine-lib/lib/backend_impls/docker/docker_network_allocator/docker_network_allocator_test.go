package docker_network_allocator

import (
	"context"
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

func TestEntireNetworkingSpace(t *testing.T) {
	takenNetworks := []*net.IPNet{}
	for i := 0; i < (1 << enclaveWidthBits); i++ {
		freeIPAddress, err := findRandomFreeNetwork(takenNetworks)
		require.NoError(t, err, "Got an unexpected error when finding a free network with already-occupied networks %+v (len %v)", takenNetworks, len(takenNetworks))
		require.NotContains(t, takenNetworks, freeIPAddress)
		takenNetworks = append(takenNetworks, freeIPAddress)
	}
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
