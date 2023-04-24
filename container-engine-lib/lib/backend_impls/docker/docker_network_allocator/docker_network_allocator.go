package docker_network_allocator

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/network_helpers"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"
)

const (
	supportedIpAddrBitLength = uint32(32)

	// We hardcode this because the algorithm for finding slots for variable-sized networks is MUCH more complex
	// This will give 4096 IPs per address; if this isn't enough we can up it in the future
	networkWidthBits = uint32(12)

	// Docker returns an error with this text when we try to create a network with a CIDR mask
	//  that overlaps with a preexisting network
	overlappingAddressSpaceErrStr = "Pool overlaps with other one on this address space"

	maxNumNetworkAllocationRetries = 10

	timeBetweenNetworkCreationRetries = 1 * time.Second

	allowedNetworkFirstOctet        = 10
	secondOctetMaximumPossibleValue = 255
	secondOctetLowestPossibleValue  = 1
)

var networkCidrMask = net.CIDRMask(int(supportedIpAddrBitLength-networkWidthBits), int(supportedIpAddrBitLength))
var networkWidthUint64 = uint64(math.Pow(float64(2), float64(networkWidthBits)))
var maxUint32PlusOne = uint64(math.MaxUint32) + 1
var emptyIpSet = map[string]bool{}

type DockerNetworkAllocator struct {
	// Our constructor sets the rand.Seed, so we want to force users to use the constructor
	// This private variable guarantees it
	isConstructedViaConstructor bool
	dockerManager               *docker_manager.DockerManager
}

func NewDockerNetworkAllocator(dockerManager *docker_manager.DockerManager) *DockerNetworkAllocator {
	// NOTE: If we need a deterministic rand seed anywhere else in the program, this will break it! The reason we do this
	//  here is because it's way more likely that we'll forget to seed the rand when using this class than it is that we need
	//  a deterministic rand seed
	rand.Seed(time.Now().UnixNano())

	return &DockerNetworkAllocator{
		isConstructedViaConstructor: true,
		dockerManager:               dockerManager,
	}
}

func (provider *DockerNetworkAllocator) CreateNewNetwork(
	ctx context.Context,
	networkName string,
	labels map[string]string,
) (resultNetworkId string, resultErr error) {
	if !provider.isConstructedViaConstructor {
		return "", stacktrace.NewError("This instance of Docker network allocator was constructed without the constructor, which means that the rand.Seed won't have been initialized!")
	}

	numRetries := 0
	for numRetries < maxNumNetworkAllocationRetries {
		networks, err := provider.dockerManager.ListNetworks(ctx)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred listing the Docker networks")
		}

		usedSubnets := []*net.IPNet{}
		for _, network := range networks {
			for _, ipamConfig := range network.IPAM.Config {
				subnetCidrStr := ipamConfig.Subnet
				_, parsedSubnet, err := net.ParseCIDR(subnetCidrStr)
				if err != nil {
					return "", stacktrace.Propagate(
						err,
						"An error occurred parsing CIDR string '%v' associated with network '%v'",
						subnetCidrStr,
						network.Name,
					)
				}
				usedSubnets = append(usedSubnets, parsedSubnet)
			}
		}

		freeNetworkIpAndMask, err := findRandomFreeNetwork(usedSubnets)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred finding a free network")
		}

		gatewayIp, err := network_helpers.GetFreeIpAddrFromSubnet(emptyIpSet, freeNetworkIpAndMask)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred getting a free IP for the network gateway")
		}

		networkId, err := provider.dockerManager.CreateNetwork(ctx, networkName, freeNetworkIpAndMask.String(), gatewayIp, labels)
		if err == nil {
			return networkId, nil
		}

		// Docker does this weird thing where a newly-deleted network won't show up in DockerClient.ListNetworks, but its IPs
		//  will still be counted as used for several seconds after deletion. The best we can do here is catch the "overlapping
		//  IP pool" error and retry with a new random network
		if !strings.Contains(err.Error(), overlappingAddressSpaceErrStr) {
			return "", stacktrace.Propagate(
				err,
				"A non-recoverable error occurred creating network '%v' with CIDR '%v'",
				networkName,
				freeNetworkIpAndMask.String(),
			)
		}

		logrus.Debugf(
			"Tried to create network '%v' with CIDR '%v', but Docker returned the '%v' error indicating that either:\n"+
				" 1) there used to be a Docker network that used those IPs that was just deleted (Docker will report a network as deleted earlier than its IPs are freed)\n"+
				" 2) a new network was created after we scanned for used IPs but before we made the call to create the network\n"+
				"Either way, we'll sleep for %v and retry",
			networkName,
			freeNetworkIpAndMask.String(),
			overlappingAddressSpaceErrStr,
			timeBetweenNetworkCreationRetries,
		)
		numRetries += 1

		// Docker does this weird thing where a newly-deleted network won't show up in DockerClient.ListNetworks (so it
		//  it won't show up in our list already-allocated pools), but when we try to creat a new network that uses
		//  the newly-freed IPs Docker will fail with "pool overlaps with current space"
		time.Sleep(timeBetweenNetworkCreationRetries)
	}

	return "", stacktrace.NewError(
		"We couldn't allocate a new network even after retrying %v times with %v between retries",
		maxNumNetworkAllocationRetries,
		timeBetweenNetworkCreationRetries,
	)
}

// This algorithm only picks a network in the 10.0.0.0/20 - 10.255.0.0.0/20 list
// https://github.com/hashicorp/serf/issues/385#issuecomment-208755148 - we try to follow RFC 6890
// https://www.rfc-editor.org/rfc/rfc6890.html calls 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16 Private-Use (docker usually picks from 172.16.0.0/12)
// Just with 10.0.0.0/8 we can get 2^24 ips; if we limit services per APIC to 4096(/20) we can get 4096 per docker engine
// For simplicity we limit it to 256 APICs and allow networks 10.1.0.0/20 - 10.255.0.0.0/20
func findRandomFreeNetwork(networks []*net.IPNet) (*net.IPNet, error) {
	for secondOctet := secondOctetLowestPossibleValue; secondOctet <= secondOctetMaximumPossibleValue; secondOctet++ {
		ipAddressString := fmt.Sprintf("%v.%v.0.0", allowedNetworkFirstOctet, secondOctet)
		resultNetworkIp := net.ParseIP(ipAddressString)
		resultNetwork := &net.IPNet{
			IP:   resultNetworkIp,
			Mask: networkCidrMask,
		}
		hasCollision := false
		for _, network := range networks {
			if resultNetwork.Contains(network.IP) || network.Contains(resultNetworkIp) {
				hasCollision = true
				break
			}
		}
		if !hasCollision {
			return resultNetwork, nil
		}
	}

	return nil, stacktrace.NewError("There is no IP address space available for a new network with %v bits of width", networkWidthBits)
}
