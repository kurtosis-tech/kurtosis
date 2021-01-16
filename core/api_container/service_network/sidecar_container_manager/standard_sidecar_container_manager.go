/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package sidecar_container_manager

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

const (
	iproute2ContainerImage = "kurtosistech/iproute2"

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

// ==========================================================================================
//                               Private types
// ==========================================================================================
type ipTablesChain string

// We sleep forever because all the commands this container will run will be executed
//  via Docker exec
var ipRouteContainerCommand = []string{
	"sleep","infinity",
}

type sidecarContainerState struct {
	containerId string
	ipAddr net.IP

	// Tracks which Kurtosis chain is the primary chain, so we know
	//  which chain is in the background that we can flush and rebuild
	//  when we're changing iptables
	chainInUse ipTablesChain
}

// ==========================================================================================
//                               Private types
// ==========================================================================================
type StandardSidecarContainerManager struct {
	dockerManager *docker_manager.DockerManager

	freeIpAddrTracker *commons.FreeIpAddrTracker

	mutexes map[topology_types.ServiceID]*sync.Mutex

	sidecarContainerStates map[topology_types.ServiceID]*sidecarContainerState
}

func (manager *StandardSidecarContainerManager) UpdateIpTablesForService(ctx context.Context, serviceId topology_types.ServiceID, blockedIps []net.IP) error {
	return manager.wrapWithMutexLocking(
		serviceId,
		func() error { return manager.internalUpdateIpTablesForService(ctx, serviceId, blockedIps) },
	)
}

// ==========================================================================================
//                      Functions that will get wrapped with mutex locking
// ==========================================================================================
// TODO Write tests for this, by extracting the logic to run exec commands on the sidecar into a separate, mockable
//  interface
func (manager *StandardSidecarContainerManager) internalUpdateIpTablesForService(ctx context.Context, serviceId topology_types.ServiceID, blockedIps []net.IP) error {
	sidecarState, found := manager.sidecarContainerStates[serviceId]
	if !found {
		return stacktrace.NewError("No sidecar container state was found for service '%v'", serviceId)
	}

	primaryChain := sidecarState.chainInUse
	var backgroundChain ipTablesChain
	if primaryChain == kurtosisIpTablesChain1 {
		backgroundChain = kurtosisIpTablesChain2
	} else if primaryChain == kurtosisIpTablesChain2 {
		backgroundChain = kurtosisIpTablesChain1
	} else {
		return stacktrace.NewError("Unrecognized iptables chain '%v' in use; this is a code bug", primaryChain)
	}

	sidecarContainerCmd, err := getSidecarContainerCommand(backgroundChain, blockedIps)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the iptables command to run in the sidecar container")
	}

	sidecarContainerId := sidecarState.containerId

	logrus.Debugf(
		"Running iptables command '%v' in sidecar container '%v' to update blocklist for service '%v'...",
		sidecarContainerCmd,
		sidecarContainerId,
		serviceId)
	execOutputBuf := &bytes.Buffer{}
	if err := manager.dockerManager.RunExecCommand(ctx, sidecarContainerId, sidecarContainerCmd, execOutputBuf); err != nil {
		logrus.Error("-------------------- iptables blocklist-updating exec command output --------------------")
		if _, err := io.Copy(logrus.StandardLogger().Out, execOutputBuf); err != nil {
			logrus.Errorf("An error occurred printing the exec logs: %v", err)
		}
		logrus.Error("------------------ End iptables blocklist-updating exec command output --------------------")
		return stacktrace.Propagate(
			err,
			"An error occurred running iptables command '%v' in sidecar container '%v' to update the blocklist of service '%v'",
			sidecarContainerCmd,
			sidecarContainerId,
			serviceId)
	}
	sidecarState.chainInUse = backgroundChain
	logrus.Infof("Successfully updated blocklist for service '%v'", serviceId)
	return nil
}

// ==========================================================================================
//                                    Private helper functions
// ==========================================================================================
func (manager *StandardSidecarContainerManager) wrapWithMutexLocking(serviceId topology_types.ServiceID, delegate func() error) error {
	mutex, found := manager.mutexes[serviceId]
	if !found {
		return stacktrace.NewError("Could not find mutex for service ID '%v'", serviceId)
	}
	mutex.Lock()
	defer mutex.Unlock()

	if err := delegate(); err != nil {
		return stacktrace.Propagate(err, "An error occurred in the delegate function")
	}
	return nil
}

/*
Given the new set of services that should be in the service's iptables, calculate the command that needs to be
	run in the sidecar container to make the service's iptables match the desired state.
*/
func getSidecarContainerCommand(
		backgroundChain ipTablesChain,
		blockedIps []net.IP) ([]string, error) {
	// Pass through map to deduplicate
	blockedIpStrs := map[string]bool{}
	for _, ipAddr := range blockedIps {
		blockedIpStrs[ipAddr.String()] = true
	}

	// NOTE: we could sort this (at a perf cost) if we need to for easier debugging
	ipsToBlockStrSlice := []string{}
	for ipAddr := range blockedIpStrs {
		ipsToBlockStrSlice = append(ipsToBlockStrSlice, ipAddr)
	}

	sidecarContainerCommand := []string{
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
			sidecarContainerCommand = append(sidecarContainerCommand, "&&")
			sidecarContainerCommand = append(sidecarContainerCommand, addBlockedSourceIpsCommand...)
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
		sidecarContainerCommand = append(sidecarContainerCommand, "&&")
		sidecarContainerCommand = append(sidecarContainerCommand, setBackgroundChainInFirstPositionCommand...)
	}

	// Because the command contains '&&', we need to wrap this in 'sh -c' else iptables
	//  will think the '&&' is an argument intended for itself
	shWrappedCommand := []string{
		"sh",
		"-c",
		strings.Join(sidecarContainerCommand, " "),
	}
	return shWrappedCommand, nil
}
