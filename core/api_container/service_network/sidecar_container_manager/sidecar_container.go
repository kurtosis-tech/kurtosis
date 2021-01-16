/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package sidecar_container_manager

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/sidecar_container_manager/sidecar_image_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

const (
	// We create two chains so that during modifications we can flush and rebuild one
	//  while the other one is live
	kurtosisIpTablesChain1 ipTablesChain = "KURTOSIS1"
	kurtosisIpTablesChain2 ipTablesChain = "KURTOSIS2"
	initialKurtosisIpTablesChain = kurtosisIpTablesChain1 // The Kurtosois chain that will be in use on service launch

	ipTablesCommand = "iptables"
	ipTablesInputChain = "INPUT"
	ipTablesOutputChain = "OUTPUT"
	ipTablesNewChainFlag = "-N"
	ipTablesInsertRuleFlag = "-I"
	ipTablesFlushChainFlag = "-F"
	ipTablesAppendRuleFlag  = "-A"
	ipTablesReplaceRuleFlag = "-R"
	ipTablesDropAction = "DROP"
	ipTablesFirstRuleIndex = 1	// iptables chains are 1-indexed
)

type ipTablesChain string

type serviceMetadata struct {
	serviceContainerId string

	sidecarContainerId string

	sidecarIpAddr      net.IP

}

// Extracted as interface for testing
type SidecarContainer interface {
	UpdateIpTables(
		ctx context.Context,
		blockedIps []net.IP,
	) error
}

// Provides a handle into manipulating the network state of a service container indirectly, via the sidecar
type StandardSidecarContainer struct {
	mutex *sync.Mutex

	// ID of the service this sidecar container is attached to
	serviceId topology_types.ServiceID

	// Tracks which Kurtosis chain is the primary chain, so we know
	//  which chain is in the background that we can flush and rebuild
	//  when we're changing iptables
	chainInUse ipTablesChain

	containerId string

	ipAddr string

	execCmdExecutor SidecarExecCmdExecutor
}

func (sidecar StandardSidecarContainer) UpdateIpTables(ctx context.Context, blockedIps []net.IP) error {
	primaryChain := sidecar.chainInUse
	var backgroundChain ipTablesChain
	if primaryChain == kurtosisIpTablesChain1 {
		backgroundChain = kurtosisIpTablesChain2
	} else if primaryChain == kurtosisIpTablesChain2 {
		backgroundChain = kurtosisIpTablesChain1
	} else {
		return stacktrace.NewError("Unrecognized iptables chain '%v' in use; this is a code bug", primaryChain)
	}

	updateCmd, err := generateIpTablesUpdateCmd(backgroundChain, blockedIps)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred generating the command to update the service's iptables")
	}

	logrus.Infof(
		"Running iptables update command '%v' in sidecar container '%v' attached to service with ID '%v'...",
		updateCmd,
		sidecar.containerId,
		sidecar.serviceId)
	if err := sidecar.execCmdExecutor.exec(ctx, updateCmd); err != nil {
		return stacktrace.Propagate(err, "An error occurred running sidecar update command '%v'")
	}
	logrus.Infof("Successfully executed iptables update command '%v'", sidecar.serviceId)
	return nil
}

func generateIpTablesInitCmd(
	) {

}

/*
Given the new IPs that should be blocked, generate the exec command that needs to be
	run in the sidecar container to make the service's iptables match the desired state.
*/
func generateIpTablesUpdateCmd(
		backgroundChain ipTablesChain,
		blockedIps []net.IP) ([]string, error) {
	// Deduplicate IPs for cleanliness
	blockedIpStrs := map[string]bool{}
	for _, ipAddr := range blockedIps {
		blockedIpStrs[ipAddr.String()] = true
	}

	// NOTE: we could sort this (at a perf cost) if we need to for easier debugging
	ipsToBlockStrSlice := []string{}
	for ipAddr := range blockedIpStrs {
		ipsToBlockStrSlice = append(ipsToBlockStrSlice, ipAddr)
	}

	resultCmd := []string{
		ipTablesCommand,
		ipTablesFlushChainFlag,
		string(backgroundChain),
	}

	if len(ipsToBlockStrSlice) > 0 {
		ipsToBlockCommaList := strings.Join(ipsToBlockStrSlice, ",")

		// As of 2020-12-31 the Kurtosis chains get used by both INPUT and OUTPUT intrinsic iptables chains,
		//  so we add rules to the Kurtosis chains to drop traffic both inbound and outbound
		for _, flag := range []string{"-s", "-d"} {
			// PERF NOTE: If it takes iptables a long time to insert all the rules, we could do the
			//  extra work leg work to calculate the diff and insert only what's needed
			addBlockedSourceIpsCommand := []string{
				ipTablesCommand,
				ipTablesAppendRuleFlag,
				string(backgroundChain),
				flag,
				ipsToBlockCommaList,
				"-j",
				ipTablesDropAction,
			}
			resultCmd = append(resultCmd, "&&")
			resultCmd = append(resultCmd, addBlockedSourceIpsCommand...)
		}
	}

	// Lastly, make sure to update which chain is being used for both INPUT and OUTPUT iptables
	for _, intrinsicChain := range []string{ipTablesInputChain, ipTablesOutputChain} {
		setBackgroundChainInFirstPositionCommand := []string{
			ipTablesCommand,
			ipTablesReplaceRuleFlag,
			intrinsicChain,
			strconv.Itoa(ipTablesFirstRuleIndex),
			"-j",
			string(backgroundChain),
		}
		resultCmd = append(resultCmd, "&&")
		resultCmd = append(resultCmd, setBackgroundChainInFirstPositionCommand...)
	}

	return resultCmd, nil
}
