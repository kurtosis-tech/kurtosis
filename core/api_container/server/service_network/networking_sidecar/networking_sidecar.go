/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
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

	undefinedIpTablesChain       ipTablesChain = ""
	initialKurtosisIpTablesChain               = kurtosisIpTablesChain1 // The Kurtosois chain that will be in use on service launch

	ipTablesCommand = "iptables"
	ipTablesNewChainFlag = "-N"
	ipTablesInsertRuleFlag = "-I"
	ipTablesFlushChainFlag = "-F"
	ipTablesAppendRuleFlag  = "-A"
	ipTablesReplaceRuleFlag = "-R"
	ipTablesDropAction = "DROP"
	ipTablesFirstRuleIndex = 1	// iptables chains are 1-indexed

	ipTablesInputChain = "INPUT"
	ipTablesOutputChain = "OUTPUT"
)

var intrinsicChainsToUpdate = map[string]bool{
	ipTablesInputChain:  true,
	ipTablesOutputChain: true,
}

// ==========================================================================================
//                                        Interface
// ==========================================================================================
// Extracted as interface for testing
type NetworkingSidecar interface {
	GetIPAddr() net.IP
	GetContainerID() string
	InitializeIpTables(ctx context.Context) error
	UpdateIpTables(ctx context.Context, blockedIps []net.IP) error
}


// ==========================================================================================
//                                      Implementation
// ==========================================================================================
type ipTablesChain string

// Provides a handle into manipulating the network state of a service container indirectly, via the sidecar
type StandardNetworkingSidecar struct {
	mutex *sync.Mutex

	// GUID of the service this sidecar container is attached to
	serviceGUID service_network_types.ServiceGUID

	// Tracks which Kurtosis chain is the primary chain, so we know
	//  which chain is in the background that we can flush and rebuild
	//  when we're changing iptables
	chainInUse ipTablesChain

	containerId string

	ipAddr net.IP

	execCmdExecutor sidecarExecCmdExecutor
}

func NewStandardNetworkingSidecar(serviceGUID service_network_types.ServiceGUID, containerId string, ipAddr net.IP, execCmdExecutor sidecarExecCmdExecutor) *StandardNetworkingSidecar {
	return &StandardNetworkingSidecar{
		mutex:           &sync.Mutex{},
		serviceGUID:     serviceGUID,
		chainInUse:      undefinedIpTablesChain,
		containerId:     containerId,
		ipAddr:          ipAddr,
		execCmdExecutor: execCmdExecutor,
	}
}

func (sidecar *StandardNetworkingSidecar) GetIPAddr() net.IP {
	return sidecar.ipAddr
}

func (sidecar *StandardNetworkingSidecar) GetContainerID() string {
	return sidecar.containerId
}

// Initializes the iptables of the attached service to a state where interactions with this NetworkingSidecar instance
//  can modify things
func (sidecar *StandardNetworkingSidecar) InitializeIpTables(ctx context.Context) error {
	// Yes, initializers are bad, but I'm deeming having initialization logic in the constructor as even worse ~ ktoday, 2021-01-16
	sidecar.mutex.Lock()
	defer sidecar.mutex.Unlock()
	if sidecar.chainInUse != undefinedIpTablesChain {
		return nil
	}

	initCmd := generateIpTablesInitCmd()

	logrus.Infof(
		"Running iptables init command '%v' in sidecar container '%v' attached to service with GUID '%v'...",
		initCmd,
		sidecar.containerId,
		sidecar.serviceGUID)
	if err := sidecar.execCmdExecutor.exec(ctx, initCmd); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred running sidecar iptables init command '%v'",
			initCmd)
	}
	sidecar.chainInUse = initialKurtosisIpTablesChain
	logrus.Infof("Successfully executed iptables update command against service with GUID '%v'", sidecar.serviceGUID)
	return nil
}

func (sidecar *StandardNetworkingSidecar) UpdateIpTables(ctx context.Context, blockedIps []net.IP) error {
	// TODO extract this boilerplate into a separate function
	sidecar.mutex.Lock()
	defer sidecar.mutex.Unlock()
	if sidecar.chainInUse == undefinedIpTablesChain {
		return stacktrace.NewError("Cannot update iptables because they haven't yet been initialized")
	}

	primaryChain := sidecar.chainInUse
	var backgroundChain ipTablesChain
	if primaryChain == kurtosisIpTablesChain1 {
		backgroundChain = kurtosisIpTablesChain2
	} else if primaryChain == kurtosisIpTablesChain2 {
		backgroundChain = kurtosisIpTablesChain1
	} else {
		return stacktrace.NewError("Unrecognized iptables chain '%v' in use; this is a code bug", primaryChain)
	}

	updateCmd := generateIpTablesUpdateCmd(backgroundChain, blockedIps)

	logrus.Infof(
		"Running iptables update command '%v' in sidecar container '%v' attached to service with GUID '%v'...",
		updateCmd,
		sidecar.containerId,
		sidecar.serviceGUID)
	if err := sidecar.execCmdExecutor.exec(ctx, updateCmd); err != nil {
		return stacktrace.Propagate(err, "An error occurred running sidecar update command '%v'", updateCmd)
	}
	sidecar.chainInUse = backgroundChain
	logrus.Infof("Successfully executed iptables update command against service with GUID '%v'", sidecar.serviceGUID)
	return nil
}

// ==========================================================================================
//                                   Private helper functions
// ==========================================================================================
func generateIpTablesInitCmd() []string {
	resultCmd := []string{
		ipTablesCommand,
		ipTablesNewChainFlag,
		string(kurtosisIpTablesChain1),
		"&&",
		ipTablesCommand,
		ipTablesNewChainFlag,
		string(kurtosisIpTablesChain2),
	}

	// Very important that we set the Kurtosis chain for both INPUT *and* OUTPUT chain, to truly simulate
	//  a network partition
	for chain := range intrinsicChainsToUpdate {
		addKurtosisChainInFirstPositionCommand := []string{
			ipTablesCommand,
			ipTablesInsertRuleFlag,
			chain,
			strconv.Itoa(ipTablesFirstRuleIndex),
			"-j",
			string(initialKurtosisIpTablesChain),
		}
		resultCmd = append(resultCmd, "&&")
		resultCmd = append(
			resultCmd,
			addKurtosisChainInFirstPositionCommand...)
	}
	return resultCmd
}

/*
Given the new IPs that should be blocked, generate the exec command that needs to be
	run in the sidecar container to make the service's iptables match the desired state.
*/
func generateIpTablesUpdateCmd(
		backgroundChain ipTablesChain,
		blockedIps []net.IP) []string {
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
	for intrinsicChain := range intrinsicChainsToUpdate {
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

	return resultCmd
}
