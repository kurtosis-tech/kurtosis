package docker_network_allocator

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestErrorOnInstantiationWithoutConstructor(t *testing.T) {
	allocator := DockerNetworkAllocator{}
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

func TestExactHole(t *testing.T) {
	// This has exactly one hole - at 1.0.0.0/20
	// Note that all 0.X.X.X aren't available because they're reserved
	cidrs := []string{
		"1.0.16.0/20",
		"1.0.32.0/19",
		"1.0.64.0/18",
		"1.0.128.0/17",
		"1.1.0.0/16",
		"1.2.0.0/15",
		"1.4.0.0/14",
		"1.8.0.0/13",
		"1.16.0.0/12",
		"1.32.0.0/11",
		"1.64.0.0/10",
		"1.128.0.0/9",
		"2.0.0.0/7",
		"4.0.0.0/6",
		"8.0.0.0/5",
		"16.0.0.0/4",
		"32.0.0.0/3",
		"64.0.0.0/2",
		"128.0.0.0/1",
	}
	assertExpectedResultGivenCidrs(t, cidrs, []byte{1, 0, 0, 0})
}

func TestNetworkFoundOnVariousCases(t *testing.T) {
	// Because the free network-finding is random, we just test that we don't get an error (i.e. we actually find a network)
	successfulCases := [][]string{
		{
			"1.0.16.0/20",
		},
		{
			"1.0.32.0/20",
		},
		{
			"1.0.4.0/24",
		},
		{
			"1.0.0.0/20",
			"1.0.32.0/20",
		},
		{
			"1.0.4.0/24",
			"1.0.64.0/24",
		},
		{
			"1.0.4.0/24",
			"1.0.18.0/24",
		},
		{
			"1.0.4.0/24",
			"1.0.18.0/24",
			"1.0.32.0/24",
		},
		{
			"1.0.0.0/18",
			"1.80.0.0/24",
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
