package free_ip_addr_tracker

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/consts"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/network_helpers"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"net"
)

type FreeIpAddrTracker struct {
	subnet    *net.IPNet
	enclaveDb *enclave_db.EnclaveDB
}

var (
	takenIpAddressBucketName = []byte("free-ip-addr-tracker")
)

func (tracker *FreeIpAddrTracker) GetFreeIpAddr() (net.IP, error) {
	var ipAddr net.IP
	err := tracker.enclaveDb.Update(func(tx *bolt.Tx) error {
		takenIps, err := getTakenIpAddrs(tx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting taken IP addresses")
		}
		ipAddr, err = network_helpers.GetFreeIpAddrFromSubnet(takenIps, tracker.subnet)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while getting a free IP address from subnet")
		}
		return tx.Bucket(takenIpAddressBucketName).Put([]byte(ipAddr.String()), consts.EmptyValueForKeySet)
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting a free IP address")
	}
	return ipAddr, nil
}

func (tracker *FreeIpAddrTracker) ReleaseIpAddr(ip net.IP) error {
	err := tracker.enclaveDb.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(takenIpAddressBucketName).Delete([]byte(ip.String()))
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while releasing used IP address '%v'", ip)
	}
	return nil
}

func GetOrCreateNewFreeIpAddrTracker(subnet *net.IPNet, alreadyTakenIps map[string]bool, db *enclave_db.EnclaveDB) (*FreeIpAddrTracker, error) {
	bucketExists := false
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket(takenIpAddressBucketName)
		if err != nil {
			bucketExists = true
			return stacktrace.Propagate(err, "An error occurred while creating IP tracker database bucket")
		}
		// Bucket does not exist, populate database
		for ipAddr := range alreadyTakenIps {
			if err != bucket.Put([]byte(ipAddr), consts.EmptyValueForKeySet) {
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
	bucket := tx.Bucket(takenIpAddressBucketName)
	err := bucket.ForEach(func(k, v []byte) error {
		takenIps[string(k)] = true
		return nil
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching free IP address")
	}
	return takenIps, nil
}
