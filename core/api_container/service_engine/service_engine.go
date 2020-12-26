/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_engine

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/service_engine/partition_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/service_engine/topology_types"
	"github.com/kurtosis-tech/kurtosis/api_container/service_engine/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultPartitionId                   topology_types.PartitionID = "default"
	startingDefaultConnectionBlockStatus                            = false

	iproute2ContainerImage = "kurtosistech/iproute2"

	// We'll create a new chain at the 0th spot for every service
	kurtosisIpTablesChain = "KURTOSIS"

	ipTablesCommand = "iptables"
	ipTablesInputChain = "INPUT"
	ipTablesNewChainFlag = "-N"
	ipTablesInsertRuleFlag = "-I"
	ipTablesFlushChainFlag = "-F"
	ipTablesAppendRuleFlag  = "-A"
	ipTablesDropAction = "DROP"

	// How long we'll wait when making a best-effort attempt to stop a container
	containerStopTimeout = 15 * time.Second
)

type containerInfo struct {
	containerId string
	ipAddr net.IP
}

/**
This is the engine
 */
type ServiceEngine struct {
	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	dockerNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerManager *commons.DockerManager

	userServiceLauncher *user_service_launcher.UserServiceLauncher

	mutex *sync.Mutex

	topology *partition_topology.PartitionTopology

	// == Per-service info ==================================================================
	serviceContainerInfo map[topology_types.ServiceID]containerInfo

	sidecarContainerInfo map[topology_types.ServiceID]containerInfo

	// Mapping of serviceID -> set of serviceIDs tracking what's currently being dropped in the INPUT chain of the service
	ipTablesBlocks map[topology_types.ServiceID]*topology_types.ServiceIDSet
}

func NewServiceEngine(
		isPartitioningEnabled bool,
		dockerNetworkId string,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerManager *commons.DockerManager,
		userServiceLauncher *user_service_launcher.UserServiceLauncher) *ServiceEngine {
	defaultPartitionConnection := partition_topology.PartitionConnection{IsBlocked: startingDefaultConnectionBlockStatus}
	return &ServiceEngine{
		isPartitioningEnabled: isPartitioningEnabled,
		dockerNetworkId: dockerNetworkId,
		freeIpAddrTracker: freeIpAddrTracker,
		dockerManager: dockerManager,
		userServiceLauncher: userServiceLauncher,
		mutex:               &sync.Mutex{},
		topology:            partition_topology.NewPartitionTopology(
			defaultPartitionId,
			defaultPartitionConnection,
		),
		serviceContainerInfo: map[topology_types.ServiceID]containerInfo{},
		sidecarContainerInfo: map[topology_types.ServiceID]containerInfo{},
	}
}

/*
Completely repartitions the network, throwing away the old topology
 */
func (engine *ServiceEngine) Repartition(
		context context.Context,
		newPartitionServices map[topology_types.PartitionID]*topology_types.ServiceIDSet,
		newPartitionConnections map[topology_types.PartitionConnectionID]partition_topology.PartitionConnection,
		newDefaultConnection partition_topology.PartitionConnection) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	if !engine.isPartitioningEnabled {
		return stacktrace.NewError("Cannot repartition; partitioning is not enabled")
	}

	if err := engine.topology.Repartition(newPartitionServices, newPartitionConnections, newDefaultConnection); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the network topology")
	}
	if err := engine.updateIpTables(context); err != nil {
		return stacktrace.Propagate(err, "An error occurred updating the IP tables to match the target blocklist after repartitioning")
	}
	return nil
}


/*
Creates the service with the given ID in the given partition

If partitionId is empty string, the default partition ID is used

Returns: The IP address of the new service
 */
func (engine *ServiceEngine) CreateServiceInPartition(
		context context.Context,
		serviceId topology_types.ServiceID,
		imageName string,
		usedPorts map[nat.Port]bool,
		partitionId topology_types.PartitionID,
		ipPlaceholder string,
		startCmd []string,
		dockerEnvVars map[string]string,
		testVolumeMountDirpath string) (net.IP, error) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	if partitionId == "" {
		partitionId = defaultPartitionId
	}

	if _, found := engine.topology.GetPartitionServices()[partitionId]; !found {
		return nil, stacktrace.NewError(
			"No partition with ID '%v' exists in the current partition topology",
			partitionId,
		)
	}

	// TODO Modify this to take in an IP, to kill the race condition with the service starting & partition application
	serviceContainerId, serviceIp, err := engine.userServiceLauncher.Launch(
		context,
		imageName,
		usedPorts,
		ipPlaceholder,
		startCmd,
		dockerEnvVars,
		testVolumeMountDirpath)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the user service")
	}
	engine.serviceContainerInfo[serviceId] = containerInfo{
		containerId: serviceContainerId,
		ipAddr:      serviceIp,
	}
	if err := engine.topology.AddService(serviceId, partitionId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v' to partition '%v'", serviceId, partitionId)
	}

	if engine.isPartitioningEnabled {
		sidecarIp, err := engine.freeIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting a free IP address for the networking sidecar container")
		}
		sidecarContainerId, err := engine.dockerManager.CreateAndStartContainer(
			context,
			iproute2ContainerImage,
			engine.dockerNetworkId,
			sidecarIp,
			map[commons.ContainerCapability]bool{
				commons.NetAdmin: true,
			},
			commons.NewContainerNetworkMode(serviceContainerId),
			map[nat.Port]*nat.PortBinding{},
			[]string{"sleep","infinity"},  // We sleep forever since iptables stuff gets executed via 'exec'
			map[string]string{}, // No environment variables
			map[string]string{}, // no bind mounts for services created via the Kurtosis API
			map[string]string{}, // No volume mounts either
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting the sidecar iproute container for modifying the service container's iptables")
		}
		engine.sidecarContainerInfo[serviceId] = containerInfo{
			containerId: sidecarContainerId,
			ipAddr:      sidecarIp,
		}

		// As soon as we have the sidecar, we need to create the Kurtosis chain and insert it in first position on the INPUT chain
		configureKurtosisChainCommand := []string{
			ipTablesCommand,
			ipTablesNewChainFlag,
			kurtosisIpTablesChain,
			"&&",
			ipTablesCommand,
			ipTablesInsertRuleFlag,
			ipTablesInputChain,
			"1",  // We want to insert the Kurtosis chain first, so it always runs
			"-j",
			kurtosisIpTablesChain,
		}
		if err := engine.dockerManager.RunExecCommand(
				context,
				sidecarContainerId,
				configureKurtosisChainCommand,
				logrus.StandardLogger().Out); err !=  nil {
			return nil, stacktrace.Propagate(err, "An error occurred configuring iptables to use the custom Kurtosis chain")
		}

		// TODO Right now, there's a period of time between user service container launch, and the recalculation of
		//  the blocklist and the application of iptables to the user's container
		//  This means there's a race condition period of time where the service container will be able to talk to everyone!
		//  The fix is to, before starting the service, apply the blocklists to every other node
		//  That way, even with the race condition, the other nodes won't accept traffic from the new node
		if err := engine.updateIpTables(context); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred updating IP tables after adding service '%v'", serviceId)
		}
	}

	return serviceIp, nil
}

func (engine *ServiceEngine) RemoveService(context context.Context, serviceId topology_types.ServiceID) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	serviceInfo, found := engine.serviceContainerInfo[serviceId]
	if !found {
		return stacktrace.NewError("Unknown service '%v'", serviceId)
	}
	serviceContainerId := serviceInfo.containerId

	// Make a best-effort attempt to stop the service container
	logrus.Debugf("Removing service ID '%v' with container ID '%v'...", serviceId, serviceContainerId)
	if err := engine.dockerManager.StopContainer(context, serviceContainerId, containerStopTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the container with ID %v", serviceContainerId)
	}
	engine.topology.RemoveService(serviceId)
	// TODO release the IP that the service got
	delete(engine.serviceContainerInfo, serviceId)
	logrus.Debugf("Successfully removed service with container ID %v", serviceContainerId)

	if engine.isPartitioningEnabled {
		sidecarContainerInfo, found := engine.sidecarContainerInfo[serviceId]
		if !found {
			return stacktrace.NewError(
				"Couldn't find sidecar container ID for service '%v'; this is a code bug where the sidecar container ID didn't get stored at creation time",
				serviceId)

		}
		sidecarContainerId := sidecarContainerInfo.containerId

		// Try to stop the sidecar container too
		logrus.Debugf("Removing sidecar container with container ID '%v'...", sidecarContainerId)
		if err := engine.dockerManager.StopContainer(context, sidecarContainerId, containerStopTimeout); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping the sidecar container with ID %v", sidecarContainerId)
		}
		// TODO release the IP that the service received
		delete(engine.sidecarContainerInfo, serviceId)
		logrus.Debugf("Successfully removed sidecar container with container ID %v", sidecarContainerId)

		if err := engine.updateIpTables(context); err != nil {
			return stacktrace.Propagate(err, "An error occurred updating the iptables after removing service '%v'", serviceId)
		}
	}
	return nil
}

// TODO write tests for me!!
/*
Gets the latest target blocklists from the topology and makes sure iptables matches
 */
func (engine *ServiceEngine) updateIpTables(context context.Context) error {
	targetBlocklists, err := engine.topology.GetBlocklists()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the current blocklists")
	}

	toUpdate, err := getServicesNeedingIpTablesUpdates(engine.ipTablesBlocks, targetBlocklists)

	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the services that need iptables updated")
	}

	sidecarContainerCmds, err := getSidecarContainerCommands(toUpdate, engine.serviceContainerInfo)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the sidecar container commands for the " +
			"services that need iptables updates")
	}

	// TODO Run the container updates in parallel, with the container being modified being the most important
	for serviceId, command := range sidecarContainerCmds {
		sidecarContainerInfo, found := engine.sidecarContainerInfo[serviceId]
		if !found {
			// TODO maybe start one if we can't find it?
			return stacktrace.NewError(
				"Couldn't find a sidecar container info for Service ID '%v', which means there's no way " +
					"to update its iptables",
				serviceId,
			)
		}
		sidecarContainerId := sidecarContainerInfo.containerId

		logrus.Infof(
			"Running iptables command '%v' in sidecar container '%v' to update blocklist for service '%v'...",
			command,
			sidecarContainerId,
			serviceId)
		commandLogsBuf := &bytes.Buffer{}
		if err := engine.dockerManager.RunExecCommand(context, sidecarContainerId, command, commandLogsBuf); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred running iptables command '%v' in sidecar container '%v' to update the blocklist of service '%v'",
				command,
				sidecarContainerId,
				serviceId)
		}
		logrus.Infof("Successfully updated blocklist for service '%v'", serviceId)
		logrus.Info("---------------------------------- Command Logs ---------------------------------------")
		if _, err := io.Copy(logrus.StandardLogger().Out, commandLogsBuf); err != nil {
			logrus.Errorf("An error occurred printing the exec logs: %v", err)
		}
		logrus.Info("---------------------------------End Command Logs -------------------------------------")
	}

	// Defensive copy when we store
	blockListToStore := map[topology_types.ServiceID]*topology_types.ServiceIDSet{}
	for serviceId, newBlockedServicesForId := range targetBlocklists {
		blockListToStore[serviceId] = newBlockedServicesForId.Copy()
	}
	engine.ipTablesBlocks = blockListToStore
	return nil
}

// TODO Write tests for me!!
/*
Compares the target state of the world with the current state of the world, and returns only a list of
	the services that need to be updated.
 */
func getServicesNeedingIpTablesUpdates(
		currentBlockedServices map[topology_types.ServiceID]*topology_types.ServiceIDSet,
		newBlockedServices map[topology_types.ServiceID]*topology_types.ServiceIDSet) (map[topology_types.ServiceID]*topology_types.ServiceIDSet, error) {
	result := map[topology_types.ServiceID]*topology_types.ServiceIDSet{}
	for serviceId, newBlockedServicesForId := range newBlockedServices {
		if newBlockedServicesForId.Contains(serviceId) {
			return nil, stacktrace.NewError("Requested for service ID '%v' to block itself!", serviceId)
		}

		// To avoid unnecessary Docker work, we won't update any iptables if the result would be the same
		//  as the current state
		currentBlockedServicesForId, found := currentBlockedServices[serviceId]
		if !found {
			currentBlockedServicesForId = topology_types.NewServiceIDSet()
		}

		noChangesNeeded := newBlockedServicesForId.Equals(currentBlockedServicesForId)
		if noChangesNeeded {
			continue
		}

		result[serviceId] = newBlockedServicesForId
	}
	return result, nil
}

// TODO write tests for me!!
/*
Given a list of updates that need to happen to a service's iptables, a map of serviceID -> commands that
	will be executed on the sidecar Docker container for the service

Args:
	toUpdate: A mapping of serviceID -> set of serviceIDs to block
 */
func getSidecarContainerCommands(
		toUpdate map[topology_types.ServiceID]*topology_types.ServiceIDSet,
		serviceContainerInfo map[topology_types.ServiceID]containerInfo) (map[topology_types.ServiceID][]string, error) {
	result := map[topology_types.ServiceID][]string{}

	// TODO We build two separate commands - flush the Kurtosis iptables chain, and then populate it with new stuff
	//  This means that there's a (very small) window of time where the iptables aren't blocked
	//  To fix this, we should really have two Kurtosis chains, and while one is running build the other one and
	//  then switch over in one atomic operation.
	for serviceId, newBlockedServicesForId := range toUpdate {
		// TODO If this is performance-inhibitive, we could do the extra work to insert only what's needed
		// When modifying a service's iptables, we always want to flush the old and set the new, rather
		//  than trying to update
		sidecarContainerCommand := []string{
			ipTablesCommand,
			ipTablesFlushChainFlag,
			kurtosisIpTablesChain,
		}

		if newBlockedServicesForId.Size() >= 0 {
			ipsToBlockStrSlice := []string{}
			for _, serviceIdToBlock := range newBlockedServicesForId.Elems() {
				toBlockContainerInfo, found := serviceContainerInfo[serviceIdToBlock]
				if !found {
					return nil, stacktrace.NewError("Service ID '%v' needs to block the IP of target service ID '%v', but " +
						"the target service doesn't have an IP associated to it",
						serviceId,
						serviceIdToBlock)
				}
				ipToBlock := toBlockContainerInfo.ipAddr
				ipsToBlockStrSlice = append(ipsToBlockStrSlice, ipToBlock.String())
			}
			ipsToBlockCommaList := strings.Join(ipsToBlockStrSlice, ",")

			addBlockedIpsCommand := []string{
				ipTablesCommand,
				ipTablesAppendRuleFlag,
				kurtosisIpTablesChain,
				"-s",
				ipsToBlockCommaList,
				"-j",
				ipTablesDropAction,
			}
			sidecarContainerCommand = append(sidecarContainerCommand, "&&")
			sidecarContainerCommand = append(sidecarContainerCommand, addBlockedIpsCommand...)
		}
		result[serviceId] = sidecarContainerCommand
	}
	return result, nil
}

