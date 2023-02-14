/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package free_ip_addr_tracker

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/test_helpers"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestGetIp(t *testing.T) {
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	require.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(parsedSubnetMask, map[string]bool{
		"1.2.0.2": true,
		"1.2.0.3": true,
	}, enclaveDb)
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
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	require.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(parsedSubnetMask, map[string]bool{}, enclaveDb)
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
	enclaveDb, cleaningFunction, err := test_helpers.CreateEnclaveDbForTesting()
	require.Nil(t, err)
	defer cleaningFunction()
	subnetMask := "1.2.3.4/16"
	_, parsedSubnetMask, err := net.ParseCIDR(subnetMask)
	require.Nil(t, err)
	addrTracker, err := GetOrCreateNewFreeIpAddrTracker(parsedSubnetMask, map[string]bool{
		"1.2.0.2": true,
		"1.2.0.3": true,
	}, enclaveDb)
	require.Nil(t, err)

	ip, err := addrTracker.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.1", ip.String())

	addrTracker2, err := GetOrCreateNewFreeIpAddrTracker(parsedSubnetMask, map[string]bool{}, enclaveDb)
	require.Nil(t, err)

	ip2, err := addrTracker2.GetFreeIpAddr()
	require.Nil(t, err)
	require.Equal(t, "1.2.0.4", ip2.String())
}
