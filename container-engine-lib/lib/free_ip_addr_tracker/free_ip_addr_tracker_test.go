/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package free_ip_addr_tracker

import (
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"net"
	"os"
	"testing"
)

func TestGetIp(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	defer db.Close()
	require.Nil(t, err)
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	require.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(parsedSubnetMask, map[string]bool{
		"1.2.0.2": true,
		"1.2.0.3": true,
	}, db)
	require.Nil(t, err)

	ip, err := addrTracker.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.1", ip.String())

	ip2, err := addrTracker.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.4", ip2.String())

	ip3, err := addrTracker.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.5", ip3.String())
}

func TestReleaseIp(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	defer db.Close()
	require.Nil(t, err)
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	require.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(parsedSubnetMask, map[string]bool{}, db)
	require.Nil(t, err)

	ip, err := addrTracker.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.1", ip.String())

	ip2, err := addrTracker.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.2", ip2.String())

	err = addrTracker.ReleaseIpAddr(ip)
	require.Nil(t, err)

	ip3, err := addrTracker.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.1", ip3.String())
}

func TestIpTrackerDiskPersistence(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	defer db.Close()
	require.Nil(t, err)
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	require.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(parsedSubnetMask, map[string]bool{
		"1.2.0.2": true,
		"1.2.0.3": true,
	}, db)
	require.Nil(t, err)

	ip, err := addrTracker.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.1", ip.String())

	addrTracker2, err := GetOrCreateNewFreeIpAddrTracker(parsedSubnetMask, map[string]bool{}, db)
	require.Nil(t, err)

	ip2, err := addrTracker2.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.4", ip2.String())
}
