package testnet

import (
	"github.com/palantir/stacktrace"
	"strconv"
	"strings"
)

const NETWORK_SIZE_LIMIT=1000

type FreeIpAddrTracker struct {
	networkName string
	subnetMask string
	gatewayIp string
	subnetSubstring string
	takenIpSuffixes map[int]bool
}

func NewFreeIpAddrTracker(networkName string, subnetMask string) *FreeIpAddrTracker {
	subnetInitialPoint := strings.Split(subnetMask, "/")[0]
	subnetSubstring := subnetInitialPoint[:len(subnetInitialPoint) - 1]
	gatewayIp := subnetSubstring + "1"
	takenIpSuffixes := map[int]bool{1:true}
	return &FreeIpAddrTracker{
		networkName: networkName,
		subnetMask: subnetMask,
		subnetSubstring: subnetSubstring,
		gatewayIp: gatewayIp,
		takenIpSuffixes: takenIpSuffixes,
	}
}

func (networkManager FreeIpAddrTracker) GetFreeIpAddr() (ipAddr string, err error){
	for i := 1; i < NETWORK_SIZE_LIMIT; i++ {
		if !networkManager.takenIpSuffixes[i] {
			networkManager.takenIpSuffixes[i] = true
			return networkManager.subnetSubstring + strconv.Itoa(i), nil
		}
	}
	return "", stacktrace.NewError("Failed to allocate IpAddr on subnet %s - all taken.", networkManager.subnetMask)
}
