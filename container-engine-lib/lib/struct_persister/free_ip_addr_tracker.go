package struct_persister

import (
	"encoding/binary"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"net"
)

type FreeIpAddrTracker struct {
	log    *logrus.Logger
	subnet *net.IPNet
	db     *bolt.DB
}

const (
	dbBucketName = "taken-ip-addresses"
)

func (tracker *FreeIpAddrTracker) GetFreeIpAddr() (ipAddr net.IP, err error) {
	err = tracker.db.Update(func(tx *bolt.Tx) error {
		takenIps, err := getTakenIpAddr(tx)
		if err != nil {
			return err
		}
		ipAddr, err = getFreeIpAddrFromSubnet(takenIps, tracker.subnet)
		if err != nil {
			return err
		}
		return tx.Bucket([]byte(dbBucketName)).Put([]byte(ipAddr.String()), []byte{})
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting free IP address")
	}
	return ipAddr, nil
}

func (tracker *FreeIpAddrTracker) ReleaseIpAddr(ip net.IP) (err error) {
	err = tracker.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(dbBucketName)).Delete([]byte(ip.String()))
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while releasing used IP address")
	}
	return nil
}

func GetOrCreateNewFreeIpAddrTracker(log *logrus.Logger, subnet *net.IPNet, alreadyTakenIps map[string]bool, db *bolt.DB) (*FreeIpAddrTracker, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket([]byte(dbBucketName))
		if err != nil {
			return err
		}
		// Bucket does not exist, populate database
		for ipAddr, _ := range alreadyTakenIps {
			err = bucket.Put([]byte(ipAddr), []byte{})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil && err != bolt.ErrBucketExists {
		return nil, stacktrace.Propagate(err, "An error occurred while building free IP address tracker")
	}
	// Bucket does exist, skipping population step
	if err == bolt.ErrBucketExists {
		log.Debugf("Taken IP addresses loaded from database")
	} else {
		log.Debugf("Taken IP addresses saved to database")
	}
	return &FreeIpAddrTracker{
		log,
		subnet,
		db,
	}, nil
}

func getTakenIpAddr(tx *bolt.Tx) (map[string]bool, error) {
	takenIps := map[string]bool{}
	err := tx.Bucket([]byte(dbBucketName)).ForEach(func(k, v []byte) error {
		takenIps[string(k)] = true
		return nil
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching free IP address")
	}
	return takenIps, nil
}

/*
Gets a free IP address from the subnet that the IP tracker was initialized with.

Returns:

	An IP from the subnet the tracker was initialized with that won't collide with any previously-given IP. The
		actual IP returned is undefined.
*/
func getFreeIpAddrFromSubnet(takenIps map[string]bool, subnet *net.IPNet) (ipAddr net.IP, err error) {
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
