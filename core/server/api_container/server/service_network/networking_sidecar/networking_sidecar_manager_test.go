/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"testing"
)

func TestDestroyReleasesIp(t *testing.T) {
	// TODO need to mock DockerManager to actually test this
	/*
	freeIpAddrTracker, err := commons.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		"1.2.3.4/16",
		map[string]bool{})
	assert.Nil(t, err)

	sidecarManager := NewStandardNetworkingSidecarManager(
		nil,
		freeIpAddrTracker,
		"some-network",
		"test-sidecar-image",
		[]string{},
		func(cmd []string) []string { return cmd },
	)

	ip, err := freeIpAddrTracker.GetFreeIpAddr()
	assert.Equal(t, []byte{1, 2, 0, 1}, ip.String())
	assert.Nil(t, err)

	mockExecCmdExecutor := newMockSidecarExecCmdExecutor()
	sidecar := NewStandardNetworkingSidecarWrapper(
		"test-service-id",
		"test-container-id",
		ip,
		mockExecCmdExecutor)

	sidecarManager.D

	 */
}

