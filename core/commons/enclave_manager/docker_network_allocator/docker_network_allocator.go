/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package docker_network_allocator

import (
	"context"
	"encoding/binary"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
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
)

var networkCidrMask = net.CIDRMask(int(supportedIpAddrBitLength-networkWidthBits), int(supportedIpAddrBitLength))
var networkWidthUint64 = uint64(math.Pow(float64(2), float64(networkWidthBits)))
var maxUint32PlusOne = uint64(math.MaxUint32) + 1

// If we create a subnetwork with this IP, Docker will give a failure like so:
//   "failed to set gateway while updating gateway: route for the gateway X.X.X.X could not be found: network is unreachable"
// This is because these addresses are multicast addresses: https://en.wikipedia.org/wiki/Multicast_address
var multicastIpRangeStart = binary.BigEndian.Uint32([]byte{224, 0, 0, 0})
var multicastIpRangeEnd = binary.BigEndian.Uint32([]byte{240, 0, 0, 0})

type DockerNetworkAllocator struct {
	// Our constructor sets the rand.Seed, so we want to force users to use the constructor
	// This private variable guarantees it
	isConstructedViaConstructor bool
}

func NewDockerNetworkAllocator() *DockerNetworkAllocator {
	// NOTE: If we need a deterministic rand seed anywhere else in the program, this will break it! The reason we do this
	//  here is because it's way more likely that we'll forget to seed the rand when using this class than it is that we need
	//  a deterministic rand seed
	rand.Seed(time.Now().UnixNano())

	return &DockerNetworkAllocator{
		isConstructedViaConstructor: true,
	}
}


func (provider *DockerNetworkAllocator) CreateNewNetwork(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		log *logrus.Logger,
		networkName string) (newNetworkId string, newNetwork *net.IPNet, newNetworkGatewayIp net.IP, newNetworkIpAddrTracker *commons.FreeIpAddrTracker, resultErr error) {
	if !provider.isConstructedViaConstructor {
		return "", nil, nil, nil, stacktrace.NewError("This instance of Docker network allocator was constructed without the constructor, which means that the rand.Seed won't have been initialized!")
	}

	numRetries := 0
	for numRetries < maxNumNetworkAllocationRetries {
		networks, err := dockerManager.ListNetworks(ctx)
		if err != nil {
			return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred listing the Docker networks")
		}

		usedSubnets := []*net.IPNet{}
		for _, network := range networks {
			for _, ipamConfig := range network.IPAM.Config {
				subnetCidrStr := ipamConfig.Subnet
				_, parsedSubnet, err := net.ParseCIDR(subnetCidrStr)
				if err != nil {
					return "", nil, nil, nil, stacktrace.Propagate(
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
			return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred finding a free network")
		}

		freeIpAddrTracker := commons.NewFreeIpAddrTracker(log, freeNetworkIpAndMask, map[string]bool{})
		gatewayIp, err := freeIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting a free IP for the network gateway")
		}

		networkId, err := dockerManager.CreateNetwork(ctx, networkName, freeNetworkIpAndMask.String(), gatewayIp)
		if err == nil {
			return networkId, freeNetworkIpAndMask, gatewayIp, freeIpAddrTracker, nil
		}

		// Docker does this weird thing where a newly-deleted network won't show up in DockerClient.ListNetworks, but its IPs
		//  will still be counted as used for several seconds after deletion. The best we can do here is catch the "overlapping
		//  IP pool" error and retry with a new random network
		if !strings.Contains(err.Error(), overlappingAddressSpaceErrStr) {
			return "", nil, nil, nil, stacktrace.Propagate(
				err,
				"A non-recoverable error occurred creating network '%v' with CIDR '%v'",
				networkName,
				freeNetworkIpAndMask.String(),
			)
		}

		log.Debugf(
			"Tried to create network '%v' with CIDR '%v', but Docker returned the '%v' error indicating that either:\n" +
				" 1) there used to be a Docker network that used those IPs that was just deleted (Docker will report a network as deleted earlier than its IPs are freed)\n" +
				" 2) a new network was created after we scanned for used IPs but before we made the call to create the network\n" +
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

	return "", nil, nil, nil, stacktrace.NewError(
		"We couldn't allocate a new network even after retrying %v times with %v between retries",
		maxNumNetworkAllocationRetries,
		timeBetweenNetworkCreationRetries,
	)
}

// NOTE: This is an intentionally non-deterministic algorithm!!!! The rationale: when many instances of Kurtosis
//  are running at once, if we make the algorithm deterministic (e.g. start a 0.0.0.0, and keep checking subsequent
//  subnets until you find a free one, which was the first iteration of this algo) then you get contention as the
//  multiple instances are all trying to allocate the same networks at the same time. Therefore, we change the start
//  to be different on every call
func findRandomFreeNetwork(networks []*net.IPNet) (*net.IPNet, error) {
	searchStartNetworksOffsetUint64 := uint64(rand.Uint32()) / networkWidthUint64
	searchStartNetworkIpUint64 := searchStartNetworksOffsetUint64 * networkWidthUint64

	// TODO PERF: This algorithm is very dumb in that it iterates over EVERY possible network, starting from a random
	//  start IP. This means that even if there's a preexisting network that takes up the first half of the IP space, we'll
	//  still try *every* possible network inside that already-allocated space (which will burn a ton of CPU cycles)
	for offsetIpsUint64 := uint64(0); offsetIpsUint64 < maxUint32PlusOne; offsetIpsUint64 += networkWidthUint64 {
		resultNetworkIpUint64UnModulod := searchStartNetworkIpUint64 + offsetIpsUint64

		// Homerolled modulo, because doing modulo in Golang is a pain in the ass
		var resultNetworkIpUint64 uint64
		if resultNetworkIpUint64UnModulod < maxUint32PlusOne {
			resultNetworkIpUint64 = resultNetworkIpUint64UnModulod
		} else {
			resultNetworkIpUint64 = resultNetworkIpUint64UnModulod - maxUint32PlusOne
		}
		resultNetworkIpUint32 := uint32(resultNetworkIpUint64)

		if resultNetworkIpUint32 >= multicastIpRangeStart && resultNetworkIpUint32 < multicastIpRangeEnd {
			continue
		}

		resultNetworkIp := make([]byte, 4)
		binary.BigEndian.PutUint32(resultNetworkIp, resultNetworkIpUint32)
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