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
	"reflect"
	"testing"
)

func TestUpdateIpTables(t *testing.T) {
	// TODO
}

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

	// The order in which the IPs get iterated and put into a joined string is nondeterministic, so
	//  we have to prepare two versions of the expected string to account for both permutations
	ipStrVersions := []string{
		service1Ip.String() + "," + service2Ip.String(),
		service2Ip.String() + "," + service1Ip.String(),
	}

	expectedCommands := [][]string{}
	for _, ipStrVersion := range ipStrVersions {
		commandStr := fmt.Sprintf(
			"iptables -F %v " +
				"&& iptables -A %v -s %v -j DROP " +
				"&& iptables -A %v -d %v -j DROP " +
				"&& iptables -R INPUT 1 -j %v " +
				"&& iptables -R OUTPUT 1 -j %v",
			backgroundChain,
			backgroundChain, ipStrVersion,
			backgroundChain, ipStrVersion,
			backgroundChain,
			backgroundChain)
		expected := []string{
			"sh",
			"-c",
			commandStr,
		}
		expectedCommands = append(expectedCommands, expected)
	}
	matches := reflect.DeepEqual(expectedCommands[0], actual) || reflect.DeepEqual(expectedCommands[1], actual)
	assert.Assert(t, matches, "Expected command doesn't match either IP string combination")
}
