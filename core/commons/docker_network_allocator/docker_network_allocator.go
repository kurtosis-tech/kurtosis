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
	"sort"
	"strings"
	"sync"
)

const (
	supportedIpAddrBitLength = uint32(32)

	firstAllocatableIpUint32 = uint32(0)

	// Docker returns an error with this text when we try to create a network with a CIDR mask
	//  that overlaps with a preexisting network
	overlappingAddressSpaceErrStr = "Pool overlaps with other one on this address space"

	maxNumNetworkAllocationRetries = 5
)

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
		name string,
		widthBits uint32) (newNetworkId string, newNetwork *net.IPNet, newNetworkGatewayIp net.IP, newNetworkIpAddrTracker *commons.FreeIpAddrTracker, resultErr error) {
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

		freeNetwork, err := findFreeNetwork(widthBits, usedSubnets)
		if err != nil {
			return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred finding a free network to fit the requested width of %v bits", widthBits)
		}

		freeIpAddrTracker := commons.NewFreeIpAddrTracker(log, freeNetwork, map[string]bool{})
		gatewayIp, err := freeIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred getting a free IP for the network gateway")
		}

		networkId, err := provider.dockerManager.CreateNetwork(ctx, name, freeNetwork.String(), gatewayIp)
		if err == nil {
			return networkId, freeNetwork, gatewayIp, freeIpAddrTracker, nil
		}

		if !strings.Contains(err.Error(), overlappingAddressSpaceErrStr) {
			return "", nil, nil, nil, stacktrace.Propagate(
				err,
				"A non-recoverable error occurred creating network '%v' with CIDR '%v'",
				name,
				freeNetwork.String(),
			)
		}

		log.Debugf(
			"Tried to create network '%v' with CIDR '%v', but Docker returned the '%v' error indicating that a new " +
				"network was created in between the time when we polled Docker for networks and created a new one",
			name,
			freeNetwork.String(),
			overlappingAddressSpaceErrStr,
		)
		numRetries += 1
	}

	return "", nil, nil, nil, stacktrace.NewError("We couldn't allocate a new network even after retrying %v times", maxNumNetworkAllocationRetries)
}

func findFreeNetwork(desiredWidthBits uint32, networks []*net.IPNet) (*net.IPNet, error) {
	if desiredWidthBits == 0 {
		return nil, stacktrace.NewError("Cannot request a network of 0 bits")
	}
	if desiredWidthBits >= supportedIpAddrBitLength {
		return nil, stacktrace.NewError(
			"Requested a network width of %v bits, but the maximum supported IP address length is %v",
			desiredWidthBits,
			supportedIpAddrBitLength,
		)
	}

	desiredWidth := uint32(math.Pow(
		float64(2),
		float64(desiredWidthBits),
	))

	isLessFunc := func(i, j int) bool {
		iNetwork := networks[i]
		jNetwork := networks[j]
		return binary.BigEndian.Uint32(iNetwork.IP) < binary.BigEndian.Uint32(jNetwork.IP)
	}
	sort.Slice(
		networks,
		isLessFunc,
	)

	// Generate a list of "holes" - blocks of IPs that aren't taken
	if len(networks) == 0 {
		return nil, stacktrace.NewError("Expected at least one preexisting network, but got 0")
	}
	firstNetwork := networks[0]
	firstNetworkIpUint32 := binary.BigEndian.Uint32(firstNetwork.IP)
	firstPotentialHoleWidth := firstNetworkIpUint32 - firstAllocatableIpUint32
	if firstPotentialHoleWidth >= desiredWidth {
		result := createNetworkFromIpAndWidth(firstAllocatableIpUint32, desiredWidthBits)
		return result, nil
	}
	for i, network := range networks {
		networkMaskNumOnes, networkMaskTotalBits := network.Mask.Size()
		if uint32(networkMaskTotalBits) != supportedIpAddrBitLength {
			return nil, stacktrace.NewError(
				"Encountered a network with a net mask whose total bits '%v' didn't " +
					"match the maximum supported IP address bits is '%v' - this requires a Kurtosis code " +
					"change to support larger IP addresses",
				networkMaskTotalBits,
				supportedIpAddrBitLength,
			)
		}
		networkWidth := uint32(math.Pow(
			float64(2),
			float64(networkMaskTotalBits - networkMaskNumOnes),
		))
		networkStartIpUint32 := binary.BigEndian.Uint32(network.IP)
		networkEndIpUint32 := networkStartIpUint32 + networkWidth

		iPlus1 := i + 1
		var holeWidth uint32
		if iPlus1 < len(networks) {
			nextNetwork := networks[iPlus1]
			nextNetworkStartIp := binary.BigEndian.Uint32(nextNetwork.IP)
			holeWidth = nextNetworkStartIp - networkEndIpUint32
		} else {
			holeWidthUint64 := (uint64(math.MaxUint32) + 1) - uint64(networkEndIpUint32)
			holeWidth = uint32(holeWidthUint64)
		}

		if holeWidth > desiredWidth {
			return createNetworkFromIpAndWidth(networkEndIpUint32, desiredWidthBits), nil
		}
	}

	return nil, stacktrace.NewError(
		"Couldn't find a sufficiently large block of free IP addresses to accommodate a new " +
			"network %v bits wide",
		desiredWidthBits,
	)
}

func createNetworkFromIpAndWidth(firstIpUint32 uint32, desiredWidthBits uint32) *net.IPNet {
	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, firstIpUint32)
	mask := net.CIDRMask(int(supportedIpAddrBitLength - desiredWidthBits), int(supportedIpAddrBitLength))
	return &net.IPNet{
		IP:   ip,
		Mask: mask,
	}
}
