/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package sidecar_container_manager

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
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

// ==========================================================================================
//                               Private types
// ==========================================================================================

// We sleep forever because all the commands this container will run will be executed
//  via Docker exec
var ipRouteContainerCommand = []string{
	"sleep","infinity",
}

type serviceMetadata struct {
	serviceContainerId string

	sidecarContainerId string

	sidecarIpAddr      net.IP

	// Tracks which Kurtosis chain is the primary chain, so we know
	//  which chain is in the background that we can flush and rebuild
	//  when we're changing iptables
	chainInUse ipTablesChain
}

// ==========================================================================================
//                               Standard sidecar container manager
// ==========================================================================================
type SidecarContainerID string

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// This class's methods are NOT thread-safe - it's up to the caller to ensure that
//  only one change at a time is run on a given sidecar container!!!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
type StandardSidecarContainerManager struct {

	dockerManager *docker_manager.DockerManager

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerNetworkId string

	serviceMetadata map[topology_types.ServiceID]*serviceMetadata
}

// TODO constructor

// Adds a sidecar container attached to the given service ID
func (manager *StandardSidecarContainerManager) AddSidecarContainer(
		ctx context.Context,
		serviceContainerId string) error {
	sidecarIp, err := manager.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred getting a free IP address for the networking sidecar container")
	}
	sidecarContainerIdStr, err := manager.dockerManager.CreateAndStartContainer(
		ctx,
		iproute2ContainerImage,
		manager.dockerNetworkId,
		sidecarIp,
		map[docker_manager.ContainerCapability]bool{
			docker_manager.NetAdmin: true,
		},
		docker_manager.NewContainerNetworkMode(serviceContainerId),
		map[nat.Port]*nat.PortBinding{},
		ipRouteContainerCommand,
		map[string]string{}, // No environment variables
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		map[string]string{}, // No volume mounts either
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred starting the sidecar iproute container for modifying the service container's iptables")
	}
	sidecarContainerId := SidecarContainerID(sidecarContainerIdStr)
	manager.serviceMetadata[sidecarContainerId] = &serviceMetadata{
		sidecarContainerId: sidecarContainerId,
		sidecarIpAddr:      sidecarIp,
	}

	// As soon as we have the sidecar, we need to create the Kurtosis chain and insert it in first position
	//  on both the INPUT *and* the OUTPUT chains
	configureKurtosisChainsCommand := []string{
		ipTablesCommand,
		ipTablesNewChainFlag,
		string(kurtosisIpTablesChain1),
		"&&",
		ipTablesCommand,
		ipTablesNewChainFlag,
		string(kurtosisIpTablesChain2),
	}
	for _, chain := range []string{ipTablesInputChain, ipTablesOutputChain} {
		addKurtosisChainInFirstPositionCommand := []string{
			ipTablesCommand,
			ipTablesInsertRuleFlag,
			chain,
			strconv.Itoa(ipTablesFirstRuleIndex),
			"-j",
			string(initialKurtosisIpTablesChain),
		}
		configureKurtosisChainsCommand = append(configureKurtosisChainsCommand, "&&")
		configureKurtosisChainsCommand = append(
			configureKurtosisChainsCommand,
			addKurtosisChainInFirstPositionCommand...)
	}

	// We need to wrap this command with 'sh -c' because we're using '&&', and if we don't do this then
	//  iptables will think the '&&' is an argument for it and fail
	configureKurtosisChainShWrappedCommand := []string{
		"sh",
		"-c",
		strings.Join(configureKurtosisChainsCommand, " "),
	}

	logrus.Debugf("Running exec command to configure Kurtosis iptables chain: '%v'", configureKurtosisChainShWrappedCommand)
	execOutputBuf := &bytes.Buffer{}
	if err := network.dockerManager.RunExecCommand(
		context,
		sidecarContainerId,
		configureKurtosisChainShWrappedCommand,
		execOutputBuf); err !=  nil {
		logrus.Error("------------------ Kurtosis iptables chain-configuring exec command output --------------------")
		if _, err := io.Copy(logrus.StandardLogger().Out, execOutputBuf); err != nil {
			logrus.Errorf("An error occurred printing the exec logs: %v", err)
		}
		logrus.Error("---------------- End Kurtosis iptables chain-configuring exec command output --------------------")
		return nil, stacktrace.Propagate(err, "An error occurred running the exec command to configure iptables to use the custom Kurtosis chain")
	}
	network.serviceIpTablesChainInUse[serviceId] = initialKurtosisIpTablesChain
}

func (manager *StandardSidecarContainerManager) UpdateIpTablesForService(ctx context.Context, serviceId topology_types.ServiceID, blockedIps []net.IP) error {
	sidecarContainer, found := manager.serviceMetadata[serviceId]
	if !found {
		return stacktrace.NewError("No sidecar container found for service ID '%v'", serviceId)
	}
	return manager.wrapWithMutexLocking(
		serviceId,
		func() error { return manager.internalUpdateIpTablesForService(ctx, *sidecarContainer, blockedIps) },
	)
}

// ==========================================================================================
//                      Functions that will get wrapped with ipTablesMutex locking
// ==========================================================================================
// TODO Write tests for this, by extracting the logic to run exec commands on the sidecar into a separate, mockable
//  interface
func (manager *StandardSidecarContainerManager) internalUpdateIpTablesForService(ctx context.Context, sidecarState serviceMetadata, blockedIps []net.IP) error {
}

// ==========================================================================================
//                                    Private helper functions
// ==========================================================================================
func (manager *StandardSidecarContainerManager) wrapWithMutexLocking(serviceId topology_types.ServiceID, delegate func() error) error {
	mutex, found := manager.mutexes[serviceId]
	if !found {
		return stacktrace.NewError("Could not find ipTablesMutex for service ID '%v'", serviceId)
	}
	mutex.Lock()
	defer mutex.Unlock()

	if err := delegate(); err != nil {
		return stacktrace.Propagate(err, "An error occurred in the delegate function")
	}
	return nil
}
