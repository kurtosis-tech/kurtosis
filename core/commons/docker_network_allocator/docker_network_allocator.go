/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
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
	"net"
	"strings"
	"sync"
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
var networkCidrMask = net.CIDRMask(int(supportedIpAddrBitLength - networkWidthBits), int(supportedIpAddrBitLength))
var networkWidthUint64 = uint64(math.Pow(float64(2), float64(networkWidthBits)))

type DockerNetworkAllocator struct {
	// Even though we don't have any internal state, we still want to make sure we're only trying to allocate one new network at a time
	mutex *sync.Mutex
}

func NewDockerNetworkAllocator() *DockerNetworkAllocator {
	return &DockerNetworkAllocator{
		mutex: &sync.Mutex{},
	}
}


func (provider *DockerNetworkAllocator) CreateNewNetwork(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		log *logrus.Logger,
		networkName string) (newNetworkId string, newNetwork *net.IPNet, newNetworkGatewayIp net.IP, newNetworkIpAddrTracker *commons.FreeIpAddrTracker, resultErr error) {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

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

		freeNetworkIpAndMask, err := findFreeNetwork(usedSubnets)
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

		if !strings.Contains(err.Error(), overlappingAddressSpaceErrStr) {
			return "", nil, nil, nil, stacktrace.Propagate(
				err,
				"A non-recoverable error occurred creating network '%v' with CIDR '%v'",
				networkName,
				freeNetworkIpAndMask.String(),
			)
		}

		// Docker does this weird thing where a newly-deleted network's IPs won't be freed right away
		// The network won't show up in the network list (so we can't detect its used IPs), so the best we can
		//  do is just retry
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

	return "", nil, nil, nil, stacktrace.NewError("We couldn't allocate a new network even after retrying %v times", maxNumNetworkAllocationRetries)
}

func findFreeNetwork(networks []*net.IPNet) (*net.IPNet, error) {
	// TODO PERF: This algorithm is very dumb in that it iterates over EVERY possible network, starting from 0
	//  This means that even if there's a preexisting network that takes up the first half of the IP space, we'll
	//  still try *every* possible network inside that already-allocated space (which will burn a ton of CPU cycles)
	for resultNetworkIpUint64 := uint64(0); resultNetworkIpUint64 < math.MaxUint32; resultNetworkIpUint64 += networkWidthUint64 {
		resultNetworkIpUint32 := uint32(resultNetworkIpUint64)
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