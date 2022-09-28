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

func (tracker *FreeIpAddrTracker) GetFreeIpAddr() (ipAddr net.IP, err error) {
	err = tracker.db.Update(func(tx *bolt.Tx) error {
		ipAddr, err = tracker.freeIpAddrTracker.GetFreeIpAddr()
		return tx.Bucket([]byte("taken-ip-addresses")).Put([]byte(ipAddr.String()), []byte{})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ipAddr, nil
}

func (tracker *FreeIpAddrTracker) ReleaseIpAddr(ip net.IP) (err error) {
	err = tracker.db.Update(func(tx *bolt.Tx) error {
		tracker.freeIpAddrTracker.ReleaseIpAddr(ip)
		return tx.Bucket([]byte("taken-ip-addresses")).Delete([]byte(ip.String()))
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func GetOrCreateNewFreeIpAddrTracker(log *logrus.Logger, subnet *net.IPNet, alreadyTakenIps map[string]bool, db *bolt.DB) (*FreeIpAddrTracker, error) {
	tracker := lib.NewFreeIpAddrTracker(log, subnet, alreadyTakenIps)
	err := db.Update(func(tx *bolt.Tx) error {
		for ipAddr, _ := range alreadyTakenIps {
			bucket, err := tx.CreateBucketIfNotExists([]byte("taken-ip-addresses"))
			if err != nil {
				return err
			}
			err = bucket.Put([]byte(ipAddr), []byte{})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &FreeIpAddrTracker{
		tracker,
		db,
	}, nil
}
