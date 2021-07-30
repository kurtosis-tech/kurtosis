/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package docker_network_allocator

import (
	"context"
	"encoding/binary"
	"github.com/docker/docker/api/types"
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

	firstAllocatableIpUint32 = uint32(0)

	// We hardcode this because the algorithm for finding slots for variable-sized networks is MUCH more complex
	// This will give 4096 IPs per address; if this isn't enough we can up it in the future
	networkWidthBits = uint32(12)

	// Docker returns an error with this text when we try to create a network with a CIDR mask
	//  that overlaps with a preexisting network
	overlappingAddressSpaceErrStr = "Pool overlaps with other one on this address space"

	maxNumNetworkAllocationRetries = 5

	maxNumNetworkAvailabilityRetries = 5

	timeBetweenNetworkAvailabilityPolls = 500 * time.Millisecond
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

		// TODO DEBUGGING
		log.Debug("============= FINDING FREE BLOCKS ==============")
		printDockerNetworks(log, networks)
		log.Debug("")

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
			// TODO DEBUGGING
			// time.Sleep(1 * time.Second)

			/*
			// TODO DEBUGGING
			log.Debug("============= AFTER CREATING NETWORK ==============")
			afterCreationNetworks, _ := dockerManager.ListNetworks(ctx); printDockerNetworks(log, afterCreationNetworks)
			log.Debug("")

			functionExitedSuccessfully := false
			defer func() {
				// If the function hasn't exited successfully, we clean up the
				if !functionExitedSuccessfully {
					// We use the background context, in case the normal context was destroyed
					if err := dockerManager.RemoveNetwork(context.Background(), networkId); err != nil {
						log.Errorf(
							"We successfully allocated Docker network '%v' with CIDR '%v' and ID '%v', but we didn't " +
								"successfully finish the allocation function so we tried to destroy the newly-created network",
							networkName,
							freeNetworkIpAndMask.String(),
							networkId,
						)
						log.Errorf("That failed with the following error:")
						fmt.Fprintln(log.Out, err)
						log.Errorf("!!! ACTION REQUIRED !!!! You'll need to manually remove Docker network with ID '%v'!", networkId)
					}
				}
			}()

			// Wait for the newly-created network to be up & available
			availabilityRetries := 0
			for availabilityRetries < maxNumNetworkAvailabilityRetries {
				matchingNetworks, err := dockerManager.GetNetworkIdsByName(ctx, networkName)
				if err == nil && len(matchingNetworks) == 1 && matchingNetworks[0] == networkId {
					break
				}
				availabilityRetries += 1
				logrus.Debugf(
					"Newly-created network '%v' with ID '%v' doesn't show up yet in the Docker network list; sleeping " +
						"for %v and trying again...",
					networkName,
					networkId,
					timeBetweenNetworkAvailabilityPolls,
				)
				time.Sleep(timeBetweenNetworkAvailabilityPolls)
			}
			if availabilityRetries >= maxNumNetworkAvailabilityRetries {
				return "", nil, nil, nil, stacktrace.NewError("Successfully created network '%v', but it didn't show up in the Docker network list even after ", networkName)
			}

			// TODO DEBUGGING
			log.Debug("============= BEFORE EXITING FUNCTION ==============")
			beforeFunctionExitNetworks, _ := dockerManager.ListNetworks(ctx); printDockerNetworks(log, beforeFunctionExitNetworks)
			log.Debug("")

			functionExitedSuccessfully = true

			 */
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

		log.Debugf("ERROR: %v", err)

		log.Debugf(
			"Tried to create network '%v' with CIDR '%v', but Docker returned the '%v' error indicating that a new " +
				"network was created in between the time when we polled Docker for networks and created a new one",
			networkName,
			freeNetworkIpAndMask.String(),
			overlappingAddressSpaceErrStr,
		)
		numRetries += 1

		// TODO DEBUGGING
		log.Debug("============= JUST BEFORE RETRY SLEEP ==============")
		beforeRetrySleepExitNetworks, _ := dockerManager.ListNetworks(ctx); printDockerNetworks(log, beforeRetrySleepExitNetworks)
		log.Debug("")

		// TODO DEBUGGING
		time.Sleep(1 * time.Second)
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

	/*




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

		if holeWidth >= desiredWidth {
			return createNetworkFromIpAndWidth(networkEndIpUint32, desiredWidthBits), nil
		}
	}

	return nil, stacktrace.NewError(
		"Couldn't find a sufficiently large block of free IP addresses to accommodate a new " +
			"network %v bits wide",
		desiredWidthBits,
	)

	 */
}

func printDockerNetworks(log *logrus.Logger, networks []types.NetworkResource) {
	log.Debugf("Got the following network CIDRs back from Docker network-listing:")
	for _, network := range networks {
		log.Debugf(" - %v", network.Name)
		for _, ipamConfig := range network.IPAM.Config {
			log.Debugf("   - %v (gateway: %v)", ipamConfig.Subnet, ipamConfig.Gateway)
		}
	}
}

/*
func createNetworkFromIpAndWidth(firstIpUint32 uint32, desiredWidthBits uint32) *net.IPNet {
	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, firstIpUint32)
	mask := net.CIDRMask(int(supportedIpAddrBitLength - desiredWidthBits), int(supportedIpAddrBitLength))
	return &net.IPNet{
		IP:   ip,
		Mask: mask,
	}
}


 */