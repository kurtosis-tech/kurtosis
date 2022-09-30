package free_ip_addr_tracker

import (
	"encoding/binary"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"net"
)

type FreeIpAddrTracker struct {
	subnet *net.IPNet
	db     *bolt.DB
}

var (
	emptyValueForKeySet = []byte{}
	dbBucketName        = []byte("taken-ip-addresses")
)

func (tracker *FreeIpAddrTracker) GetFreeIpAddr() (net.IP, error) {
	var ipAddr net.IP
	err := tracker.db.Update(func(tx *bolt.Tx) error {
		takenIps, err := getTakenIpAddrs(tx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting taken IP addresses")
		}
		ipAddr, err = GetFreeIpAddrFromSubnet(takenIps, tracker.subnet)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting a free IP address from subnet")
		}
		return tx.Bucket(dbBucketName).Put([]byte(ipAddr.String()), emptyValueForKeySet)
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting a free IP address")
	}
	return ipAddr, nil
}

func (tracker *FreeIpAddrTracker) ReleaseIpAddr(ip net.IP) error {
	err := tracker.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(dbBucketName).Delete([]byte(ip.String()))
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while releasing used IP address '%v'", ip)
	}
	return nil
}

func GetOrCreateNewFreeIpAddrTracker(subnet *net.IPNet, alreadyTakenIps map[string]bool, db *bolt.DB) (*FreeIpAddrTracker, error) {
	bucketExists := false
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket(dbBucketName)
		if err != nil {
			bucketExists = true
			return stacktrace.Propagate(err, "An error occurred while creating IP tracker database bucket")
		}
		// Bucket does not exist, populate database
		for ipAddr, _ := range alreadyTakenIps {
			if err != bucket.Put([]byte(ipAddr), emptyValueForKeySet) {
				return stacktrace.Propagate(err, "An error occurred writing IP to database '%v'", ipAddr)
			}
		}
		return nil
	})
	if err != nil && !bucketExists {
		return nil, stacktrace.Propagate(err, "An error occurred while building free IP address tracker")
	}
	// Bucket does exist, skipping population step
	if err == bolt.ErrBucketExists {
		logrus.Debugf("Taken IP addresses loaded from database")
	} else {
		logrus.Debugf("Taken IP addresses saved to database")
	}
	return &FreeIpAddrTracker{
		subnet,
		db,
	}, nil
}

func getTakenIpAddrs(tx *bolt.Tx) (map[string]bool, error) {
	takenIps := map[string]bool{}
	bucket := tx.Bucket(dbBucketName)
	err := bucket.ForEach(func(k, v []byte) error {
		takenIps[string(k)] = true
		return nil
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching free IP address")
	}
	return takenIps, nil
}

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
