package network_helpers

import (
	"encoding/binary"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

/*
GetFreeIpAddrFromSubnet
Gets a free IP address from the subnet that the IP tracker was initialized with.

Returns:

	An IP from the subnet the tracker was initialized with that won't collide with any previously-given IP. The
		actual IP returned is undefined.
*/
func GetFreeIpAddrFromSubnet(takenIps map[string]bool, subnet *net.IPNet) (net.IP, error) {
	// NOTE: This whole function will need to be rewritten if we support IPv6

	// convert IPNet struct mask and address to uint32
	// network is BigEndian
	mask := binary.BigEndian.Uint32(subnet.Mask)

	subnetIp := subnet.IP

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
		if !takenIps[ipStr] {
			takenIps[ipStr] = true
			return ip, nil
		}
	}
	return nil, stacktrace.NewError("Failed to allocate IpAddr on subnet %v - all taken.", subnet)
}
