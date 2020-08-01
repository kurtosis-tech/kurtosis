package networks

import (
	"encoding/binary"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

/*
Object which is intialized from a subnet and doles out IPs from the subnet, tracking which IPs are currently in use and
	making sure not to return any IPs that are in use.
 */
type FreeIpAddrTracker struct {
	log *logrus.Logger
	subnet *net.IPNet
	takenIps map[string]bool
}

/*
Creates a new IP tracker from the given parameters.

Args:
	log: The logger that log messages will be written to.
	subnetMask: The mask of the subnet that the IP tracker should dole IPs from.
	alreadyTakenIps: A set of IPs that should be marked as taken from initialization.
from the list of already-taken IPs
 */
func NewFreeIpAddrTracker(log *logrus.Logger, subnetMask string, alreadyTakenIps map[string]bool) (ipAddrTracker *FreeIpAddrTracker, err error) {
	_, ipv4Net, err := net.ParseCIDR(subnetMask)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse subnet %s as CIDR.", subnetMask)
	}

	// Defensive copy
	takenIps := map[string]bool{}
	for ipAddr, _ := range alreadyTakenIps {
		takenIps[ipAddr] = true
	}

	ipAddrTracker = &FreeIpAddrTracker{
		log: log,
		subnet: ipv4Net,
		takenIps: takenIps,
	}
	return ipAddrTracker, nil
}

// TODO Return IP objects (which are easily convertable to strings) rather than strings themselves
// TODO rework this entire function to handle IPv6 as well (currently breaks on IPv6)
/*
Gets a free IP address from the subnet that the IP tracker was initializd with.

Returns:
	An IP from the subnet the tracker was initialized with that won't collide with any previously-given IP. The
		actual IP returned is undefined.
 */
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