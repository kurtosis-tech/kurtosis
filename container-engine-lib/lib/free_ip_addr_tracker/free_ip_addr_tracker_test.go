/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package free_ip_addr_tracker

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
	"net"
	"os"
	"testing"
)

func TestGetIp(t *testing.T) {
	db, err := bolt.Open("test.db", 0666, nil)
	assert.Nil(t, err)
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	assert.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(logrus.StandardLogger(), parsedSubnetMask, map[string]bool{
		"1.2.0.2": true,
		"1.2.0.3": true,
	}, db)
	assert.Nil(t, err)

	ip, err := addrTracker.GetFreeIpAddr()
	assert.Nil(t, err)
	assert.Equal(t, "1.2.0.1", ip.String())

	ip2, err := addrTracker.GetFreeIpAddr()
	assert.Nil(t, err)
	assert.Equal(t, "1.2.0.4", ip2.String())

	ip3, err := addrTracker.GetFreeIpAddr()
	assert.Nil(t, err)
	assert.Equal(t, "1.2.0.5", ip3.String())
	err = db.Close()
	assert.Nil(t, err)
	os.Remove("test.db")
}

func TestReleaseIp(t *testing.T) {
	db, err := bolt.Open("test.db", 0666, nil)
	assert.Nil(t, err)
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	assert.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(logrus.StandardLogger(), parsedSubnetMask, map[string]bool{}, db)
	assert.Nil(t, err)

	ip, err := addrTracker.GetFreeIpAddr()
	assert.Nil(t, err)
	assert.Equal(t, "1.2.0.1", ip.String())

	ip2, err := addrTracker.GetFreeIpAddr()
	assert.Nil(t, err)
	assert.Equal(t, "1.2.0.2", ip2.String())

	addrTracker.ReleaseIpAddr(ip)

	ip3, err := addrTracker.GetFreeIpAddr()
	assert.Nil(t, err)
	assert.Equal(t, "1.2.0.1", ip3.String())
	err = db.Close()
	assert.Nil(t, err)
	os.Remove("test.db")
}

func TestIpTrackerDiskPersistence(t *testing.T) {
	db, err := bolt.Open("test.db", 0666, nil)
	assert.Nil(t, err)
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	assert.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(logrus.StandardLogger(), parsedSubnetMask, map[string]bool{
		"1.2.0.2": true,
		"1.2.0.3": true,
	}, db)
	assert.Nil(t, err)

	ip, err := addrTracker.GetFreeIpAddr()
	assert.Nil(t, err)
	assert.Equal(t, "1.2.0.1", ip.String())

	addrTracker2, err := GetOrCreateNewFreeIpAddrTracker(logrus.StandardLogger(), parsedSubnetMask, map[string]bool{}, db)
	assert.Nil(t, err)

	ip2, err := addrTracker2.GetFreeIpAddr()
	assert.Nil(t, err)
	assert.Equal(t, "1.2.0.4", ip2.String())

	assert.Nil(t, err)

	os.Remove("test.db")
}
