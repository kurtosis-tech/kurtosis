/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package commons

import (
	"encoding/binary"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

/*
Object which is intialized from a subnet and doles out IPs from the subnet, tracking which IPs are currently in use and
	making sure not to return any IPs that are in use.
 */
type FreeIpAddrTracker struct {
	log *logrus.Logger
	subnet *net.IPNet
	takenIps map[string]bool
	mutex *sync.Mutex
}

/*
Creates a new IP tracker from the given parameters.

Args:
	log: The logger that log messages will be written to.
	subnetMask: The mask of the subnet that the IP tracker should dole IPs from.
	alreadyTakenIps: A set of IPs that should be marked as taken from initialization.
from the list of already-taken IPs
 */
func NewFreeIpAddrTracker(log *logrus.Logger, subnet *net.IPNet, alreadyTakenIps map[string]bool) *FreeIpAddrTracker {
	// Defensive copy
	takenIps := map[string]bool{}
	for ipAddr, _ := range alreadyTakenIps {
		takenIps[ipAddr] = true
	}

	return &FreeIpAddrTracker{
		log: log,
		subnet: subnet,
		takenIps: takenIps,
		mutex: &sync.Mutex{},
	}
}

/*
Gets a free IP address from the subnet that the IP tracker was initializd with.

Returns:
	An IP from the subnet the tracker was initialized with that won't collide with any previously-given IP. The
		actual IP returned is undefined.
 */
func (tracker *FreeIpAddrTracker) GetFreeIpAddr() (ipAddr net.IP, err error){
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	// NOTE: This whole function will need to be rewritten if we support IPv6

	// convert IPNet struct mask and address to uint32
	// network is BigEndian
	mask := binary.BigEndian.Uint32(tracker.subnet.Mask)

	subnetIp := tracker.subnet.IP

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
		if !tracker.takenIps[ipStr] {
			tracker.takenIps[ipStr] = true
			return ip, nil
		}
	}
	return nil, stacktrace.NewError("Failed to allocate IpAddr on subnet %v - all taken.", tracker.subnet)
}

// Returns a previously-taken IP address back to the pool
func (tracker *FreeIpAddrTracker) ReleaseIpAddr(ip net.IP) {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	ipStr := ip.String()
	delete(tracker.takenIps, ipStr)
}
