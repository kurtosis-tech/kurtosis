package testnet

import (
	"encoding/binary"
	"github.com/palantir/stacktrace"
	"net"
)

type FreeIpAddrTracker struct {
	networkName string
	subnet *net.IPNet
	takenIps map[string]bool
}

func NewFreeIpAddrTracker(networkName string, subnetMask string) (ipAddrTracker *FreeIpAddrTracker, err error) {
	_, ipv4Net, err := net.ParseCIDR(subnetMask)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse subnet %s as CIDR.", subnetMask)
	}
	takenIps := map[string]bool{}
	ipAddrTracker = &FreeIpAddrTracker{
		networkName: networkName,
		subnet: ipv4Net,
		takenIps: takenIps,
	}
	// TODO TODO TODO: Explicitly pass gatewayIP to Docker Network and count from there.
	// HACK: remove the zeroth IP - Docker doesn't use this for container space..
	_, err = ipAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get gatewayIP for network %v.", ipv4Net)
	}
	// HACK: remove the first IP - by default Docker uses this as the gateway.
	_, err = ipAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get gatewayIP for network %v.", ipv4Net)
	}
	return ipAddrTracker, nil
}

func (networkManager FreeIpAddrTracker) GetFreeIpAddr() (ipAddr string, err error){
	// convert IPNet struct mask and address to uint32
	// network is BigEndian
	mask := binary.BigEndian.Uint32(networkManager.subnet.Mask)
	start := binary.BigEndian.Uint32(networkManager.subnet.IP)
	// find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)
	// loop through addresses as uint32
	for i := start; i <= finish; i++ {
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		ipStr := ip.String()
		if !networkManager.takenIps[ipStr] {
			networkManager.takenIps[ipStr] = true
			return ipStr, nil
		}
	}
	return "", stacktrace.NewError("Failed to allocate IpAddr on subnet %v - all taken.", networkManager.subnet)
}