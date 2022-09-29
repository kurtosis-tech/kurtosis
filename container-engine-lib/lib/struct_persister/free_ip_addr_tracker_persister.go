package struct_persister

import (
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"net"
)

type FreeIpAddrTracker struct {
	freeIpAddrTracker *lib.FreeIpAddrTracker
	db                *bolt.DB
}

const (
	dbBucketName = "taken-ip-addresses"
)

func (tracker *FreeIpAddrTracker) GetFreeIpAddr() (ipAddr net.IP, err error) {
	err = tracker.db.Update(func(tx *bolt.Tx) error {
		ipAddr, err = tracker.freeIpAddrTracker.GetFreeIpAddr()
		return tx.Bucket([]byte(dbBucketName)).Put([]byte(ipAddr.String()), []byte{})
	})
	if err != nil {
		return nil, err
	}
	return ipAddr, nil
}

func (tracker *FreeIpAddrTracker) ReleaseIpAddr(ip net.IP) (err error) {
	err = tracker.db.Update(func(tx *bolt.Tx) error {
		tracker.freeIpAddrTracker.ReleaseIpAddr(ip)
		return tx.Bucket([]byte(dbBucketName)).Delete([]byte(ip.String()))
	})
	if err != nil {
		return err
	}
	return nil
}

func GetOrCreateNewFreeIpAddrTracker(log *logrus.Logger, subnet *net.IPNet, alreadyTakenIps map[string]bool, db *bolt.DB) (*FreeIpAddrTracker, error) {
	// Defensive copy
	takenIps := map[string]bool{}
	for ipAddr, _ := range alreadyTakenIps {
		takenIps[ipAddr] = true
	}
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket([]byte(dbBucketName))
		if err != nil {
			return err
		}
		// Bucket does not exist, populate database
		for ipAddr, _ := range takenIps {
			err = bucket.Put([]byte(ipAddr), []byte{})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err == bolt.ErrBucketExists {
		err = db.View(func(tx *bolt.Tx) error {
			// Bucket does exist, hydrate alreadyTakenIps
			takenIps = map[string]bool{}
			bucket := tx.Bucket([]byte(dbBucketName))
			return bucket.ForEach(func(k, v []byte) error {
				takenIps[string(k)] = true
				return nil
			})
		})
	}
	if err != nil {
		return nil, err
	}
	tracker := lib.NewFreeIpAddrTracker(log, subnet, takenIps)
	return &FreeIpAddrTracker{
		tracker,
		db,
	}, nil
}
