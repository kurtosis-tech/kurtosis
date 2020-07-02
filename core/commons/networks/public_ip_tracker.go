package networks

import (
	"encoding/binary"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

type FreeIpAddrTracker struct {
	log *logrus.Logger
	subnet *net.IPNet
	takenIps map[string]bool
}

/*
Creates a new tracker that will dole out free IP addresses from the given subnet, making sure not to dole out any IPs
from the list of already-taken IPs
 */
func NewFreeIpAddrTracker(log *logrus.Logger, subnetMask string, alreadyTakenIps []string) (ipAddrTracker *FreeIpAddrTracker, err error) {
	_, ipv4Net, err := net.ParseCIDR(subnetMask)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse subnet %s as CIDR.", subnetMask)
	}
	takenIps := map[string]bool{}

	for _, ipAddr := range alreadyTakenIps {
		takenIps[ipAddr] = true
	}

	ipAddrTracker = &FreeIpAddrTracker{
		log: log,
		subnet: ipv4Net,
		takenIps: takenIps,
	}
	return ipAddrTracker, nil
}

// TODO rework this entire function to handle IPv6 as well (currently breaks on IPv6)
func (networkManager FreeIpAddrTracker) GetFreeIpAddr() (ipAddr string, err error){
	// convert IPNet struct mask and address to uint32
	// network is BigEndian
	mask := binary.BigEndian.Uint32(networkManager.subnet.Mask)

	subnetIp := networkManager.subnet.IP

	// The IP can be either 4 bytes or 16 bytes long; we need to handle both!
	// See https://gist.github.com/ammario/649d4c0da650162efd404af23e25b86b
	var intIp uint32
	if len(subnetIp) == 16 {
		intIp = binary.BigEndian.Uint32(subnetIp[12:16])
	} else {
		intIp = binary.BigEndian.Uint32(subnetIp)
	}

	// We remove the zeroth IP because it's only used for specifying the network itself
	start := intIp + 1

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