/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestGenerateUpdateCmd(t *testing.T) {
	backgroundChain := kurtosisIpTablesChain1
	service1Ip := net.IP{1, 2, 3, 4}
	service2Ip := net.IP{5, 6, 7, 8}
	newIpsToBlock := []net.IP{
		service1Ip,
		service2Ip,
	}

	actual := generateIpTablesUpdateCmd(backgroundChain, newIpsToBlock)

	// The order in which the IPs get iterated and put into a joined string is nondeterministic, so
	//  we have to prepare two versions of the expected string to account for both permutations
	ipStrVersions := []string{
		service1Ip.String() + "," + service2Ip.String(),
		service2Ip.String() + "," + service1Ip.String(),
	}

	expectedCommands := [][]string{}
	for _, ipStrVersion := range ipStrVersions {
		backgroundChainStr := string(backgroundChain)
		firstRuleIdxStr := strconv.Itoa(ipTablesFirstRuleIndex)
		expected := []string{
			"iptables", "-F", backgroundChainStr, "&&",
			"iptables", "-A", backgroundChainStr, "-s", ipStrVersion, "-j", "DROP", "&&",
			"iptables", "-A", backgroundChainStr, "-d", ipStrVersion, "-j", "DROP", "&&",
			"iptables", "-R", ipTablesInputChain, firstRuleIdxStr, "-j", backgroundChainStr, "&&",
			"iptables", "-R", ipTablesOutputChain, firstRuleIdxStr, "-j", backgroundChainStr,
		}
		expectedCommands = append(expectedCommands, expected)
	}
	matches := reflect.DeepEqual(expectedCommands[0], actual) || reflect.DeepEqual(expectedCommands[1], actual)
	assert.True(t, matches, "Expected command doesn't match either IP string combination")
}

func TestInitializationDoesAllNecessaryChains(t *testing.T) {
	neededChains := map[string]bool{}
	for chain := range intrinsicChainsToUpdate {
		neededChains[chain] = true
	}

	cmd := generateIpTablesInitCmd()
	for _, word := range cmd {
		if _, found := neededChains[word]; found {
			delete(neededChains, word)
		}
	}

	for chain := range neededChains {
		t.Fatalf("iptables initialization command doesn't initialize chain '%v'", chain)
	}
}

func TestUpdateDoesAllNecessaryChains(t *testing.T) {
	neededChains := map[string]bool{}
	for chain := range intrinsicChainsToUpdate {
		neededChains[chain] = true
	}

	ips := []net.IP{
		{1, 2, 3, 4},
	}
	cmd := generateIpTablesUpdateCmd("TEST_CHAIN", ips)
	for _, word := range cmd {
		if _, found := neededChains[word]; found {
			delete(neededChains, word)
		}
	}

	for chain := range neededChains {
		t.Fatalf("iptables update command doesn't update chain '%v'", chain)
	}
}

func TestInitialization(t *testing.T) {
	execCmdExecutor := newMockSidecarExecCmdExecutor()

	sidecar := NewStandardNetworkingSidecar(
		"test",
		"abc123",
		[]byte{1, 2, 3, 4},
		execCmdExecutor)
	assert.Equal(t, undefinedIpTablesChain, sidecar.chainInUse)

	err := sidecar.InitializeIpTables(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(execCmdExecutor.commands))
	assert.Equal(t, initialKurtosisIpTablesChain, sidecar.chainInUse)
}

func TestDoubleInitializationIsFine(t *testing.T) {
	execCmdExecutor := newMockSidecarExecCmdExecutor()

	sidecar := NewStandardNetworkingSidecar(
		"test",
		"abc123",
		[]byte{1, 2, 3, 4},
		execCmdExecutor)
	err := sidecar.InitializeIpTables(context.Background())
	assert.Nil(t, err)

	err = sidecar.InitializeIpTables(context.Background())
	assert.Nil(t, err)

	assert.Equal(t, initialKurtosisIpTablesChain, sidecar.chainInUse)
}

func TestChainSwapping(t *testing.T) {
	execCmdExecutor := newMockSidecarExecCmdExecutor()
	ctx := context.Background()

	sidecar := NewStandardNetworkingSidecar(
		"test",
		"abc123",
		[]byte{1, 2, 3, 4},
		execCmdExecutor)
	assert.Nil(t, sidecar.InitializeIpTables(ctx))
	assert.Equal(t, kurtosisIpTablesChain1, sidecar.chainInUse)

	ips := []net.IP{
		{1, 2, 3, 4},
	}
	assert.Nil(t, sidecar.UpdateIpTables(ctx, ips))
	assert.Equal(t, kurtosisIpTablesChain2, sidecar.chainInUse)

	assert.Nil(t, sidecar.UpdateIpTables(ctx, ips))
	assert.Equal(t, kurtosisIpTablesChain1, sidecar.chainInUse)
}

func TestConcurrencySafety(t *testing.T) {
	numProcesses := 20

	execCmdExecutor := newMockSidecarExecCmdExecutor()
	ctx := context.Background()

	sidecar := NewStandardNetworkingSidecar(
		"test",
		"abc123",
		[]byte{1, 2, 3, 4},
		execCmdExecutor)
	sidecar.InitializeIpTables(ctx)

	execCmdExecutor.setBlocked(true)

	for i := 0; i < numProcesses; i++ {
		iByte := byte(i)
		ips := []net.IP{
			{iByte, iByte, iByte, iByte},
		}
		go func() {
			sidecar.UpdateIpTables(ctx, ips)
		}()
		time.Sleep(5 * time.Millisecond)  // Make sure they enter the sidecar in proper order
	}

	// At this point:
	// - If the sidecar isn't controlling concurrency, all the processes will be backed up inside the exec cmd executor
	// - If the sidecar is controlling concurrency, only one thread will be in the ExecCmdExecutor and the rest will be queued
	//     inside the sidecar in FIFO order

	execCmdExecutor.setBlocked(false)

	// Give the now-unblocked threads time to finish
	time.Sleep(500 * time.Millisecond)

	// Verify that concurrency was controlled in the sidecar, so everything is ordered
	// We ignore the first command, because it will be the initialization
	for i := 1; i <= numProcesses; i++ {
		expectedByte := byte(i - 1)
		expectedIpStr := net.IP([]byte{expectedByte, expectedByte, expectedByte, expectedByte}).String()
		assert.Contains(t, execCmdExecutor.commands[i], expectedIpStr)
	}
}
