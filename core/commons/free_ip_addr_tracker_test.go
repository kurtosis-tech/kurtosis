/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package commons

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetIp(t *testing.T) {
	subnetMask := "1.2.3.4/16"
	addrTracker, err := NewFreeIpAddrTracker(logrus.StandardLogger(), subnetMask, map[string]bool{
		"1.2.0.2": true,
		"1.2.0.3": true,
	})
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
}

func TestReleaseIp(t *testing.T) {
	subnetMask := "1.2.3.4/16"
	addrTracker, err := NewFreeIpAddrTracker(logrus.StandardLogger(), subnetMask, map[string]bool{})
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
}
