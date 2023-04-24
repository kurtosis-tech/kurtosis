package docker_network_allocator

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
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
	allPossibleNetworks := []*net.IPNet{}
	for secondOctet := secondOctetLowestPossibleValue; secondOctet <= secondOctetMaximumPossibleValue; secondOctet++ {
		ipAddressString := fmt.Sprintf("%v.%v.0.0", allowedNetworkFirstOctet, secondOctet)
		resultNetworkIp := net.ParseIP(ipAddressString)
		resultNetwork := &net.IPNet{
			IP:   resultNetworkIp,
			Mask: networkCidrMask,
		}
		allPossibleNetworks = append(allPossibleNetworks, resultNetwork)
	}
	i := 0
	for {
		if i == secondOctetMaximumPossibleValue {
			break
		}
		freeIPAddress, err := findRandomFreeNetwork(takenNetworks)
		require.NoError(t, err, "Got an unexpected error when finding a free network with already-occupied networks %+v", takenNetworks)
		takenNetworks = append(takenNetworks, freeIPAddress)
		require.Equal(t, allPossibleNetworks[i], freeIPAddress)
		i += 1
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
