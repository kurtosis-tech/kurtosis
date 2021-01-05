/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_network

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/palantir/stacktrace"
	"gotest.tools/assert"
	"net"
	"testing"
)

func TestGetSidecarContainerCommandNormalOperation(t *testing.T) {
	backgroundChain := kurtosisIpTablesChain1
	service1 := topology_types.ServiceID("service1")
	service1Ip := net.IP{1, 2, 3, 4}
	service2 := topology_types.ServiceID("service2")
	service2Ip := net.IP{5, 6, 7, 8}
	newBlocklist := topology_types.NewServiceIDSet(service1, service2)
	serviceIps := map[topology_types.ServiceID]net.IP{
		service1: service1Ip,
		service2: service2Ip,
	}

	actual, err := getSidecarContainerCommand(backgroundChain, *newBlocklist, serviceIps)
	if err != nil {
		t.Fatal(stacktrace.Propagate(
			err,
			"An error occurred getting the sidecar container command for background chain '%v'",
			backgroundChain),
		)
	}
	commandStr := fmt.Sprintf(
		"iptables -F %v && iptables -A %v -s 1.2.3.4,5.6.7.8 -j DROP " +
			"&& iptables -A %v -d 1.2.3.4,5.6.7.8 -j DROP " +
			"&& iptables -R INPUT 1 -j %v " +
			"&& iptables -R OUTPUT 1 -j %v",
		backgroundChain,
		backgroundChain,
		backgroundChain,
		backgroundChain,
		backgroundChain)
	expected := []string{
		"sh",
		"-c",
		commandStr,
	}
	assert.DeepEqual(t, expected, actual)
}
