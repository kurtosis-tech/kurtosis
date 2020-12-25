/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package partitioning

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api"
	"github.com/kurtosis-tech/kurtosis/api_container/partitioning/partition_connection"
	"github.com/kurtosis-tech/kurtosis/api_container/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
	"sync"
)

const (
	defaultPartitionId                   PartitionID = "default"
	startingDefaultConnectionBlockStatus             = false

	iproute2ContainerImage = "kurtosistech/iproute2"

	inputIpTablesChain = "INPUT"

	// We'll create a new chain at the 0th spot for every service
	kurtosisIpTablesChain = "KURTOSIS"

	ipTablesCommand = "iptables"
	ipTablesFlushChainFlag = "-F"
	ipTablesAppendRuleFlag  = "-A"
	ipTablesDropAction = "DROP"
)

type PartitionID string

/**
This is the engine
 */
type PartitioningEngine struct {
	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	dockerNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerManager *commons.DockerManager

	userServiceLauncher *user_service_launcher.UserServiceLauncher

	defaultConnection partition_connection.PartitionConnection

	mutex *sync.Mutex

	// == Per-service info ==================================================================
	serviceIps map[api.ServiceID]net.IP

	serviceContainerIds map[api.ServiceID]string

	servicePartitions map[api.ServiceID]PartitionID

	sidecarContainerIps map[api.ServiceID]net.IP

	// Mapping of Service ID -> the sidecar iproute container that we'll use to manipulate the
	//  container's networking
	sidecarContainerIds map[api.ServiceID]string

	// Mapping of serviceID -> set of serviceIDs that are currently being dropped in the INPUT chain of the service
	// An entry in this list means that the host has a Kurtosis entry as its first entry in the INPUT chain, set to the
	//  IPs of the services
	// A missing entry means that iptables doesn't have a Kurtosis entry present
	blockedServices map[api.ServiceID]ServiceIDSet

	// == Per-partition info ==================================================================
	partitionConnections map[partition_connection.PartitionConnectionID]partition_connection.PartitionConnection

	// A service can be a part of exactly one partition at a time
	partitionServices map[PartitionID]ServiceIDSet // partitionId -> set<serviceId>
}

func NewPartitioningEngine(
		isPartitioningEnabled bool,
		dockerNetworkId string,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerManager *commons.DockerManager,
		userServiceLauncher *user_service_launcher.UserServiceLauncher) *PartitioningEngine {
	return &PartitioningEngine{
		isPartitioningEnabled: isPartitioningEnabled,
		dockerNetworkId: dockerNetworkId,
		freeIpAddrTracker: freeIpAddrTracker,
		dockerManager: dockerManager,
		userServiceLauncher: userServiceLauncher,
		defaultConnection: partition_connection.PartitionConnection{
			IsBlocked: startingDefaultConnectionBlockStatus,
		},
		mutex: &sync.Mutex{},

		serviceIps: map[api.ServiceID]net.IP{},
		sidecarContainerIds: map[api.ServiceID]string{},
		servicePartitions: map[api.ServiceID]PartitionID{},

		partitionConnections: map[partition_connection.PartitionConnectionID]partition_connection.PartitionConnection{},
		partitionServices: map[PartitionID]ServiceIDSet{
			defaultPartitionId: *NewServiceIDSet(),
		},
	}
}

/*
Completely repartitions the network, throwing away the old topology
 */
func (engine *PartitioningEngine) Repartition(
		servicePartitions map[api.ServiceID]PartitionID,
		partitionConnections map[partition_connection.PartitionConnectionID]partition_connection.PartitionConnection) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()



	if !engine.isPartitioningEnabled {
		stacktrace.NewError("Cannot repartition; partitioning is not enabled")
	}

	// TODO FIX THIS!
	return nil
}


/*
Creates the service with the given ID in the given partition

If partitionId is empty string, the default partition ID is used
 */
func (engine *PartitioningEngine) CreateServiceInPartition(
		context context.Context,
		serviceId api.ServiceID,
		imageName string,
		usedPorts map[nat.Port]bool,
		partitionId PartitionID,
		ipPlaceholder string,
		startCmd []string,
		dockerEnvVars map[string]string,
		testVolumeMountDirpath string) error {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()

	if partitionId == "" {
		partitionId = defaultPartitionId
	}

	if _, found := engine.partitionServices[partitionId]; !found {
		return stacktrace.NewError(
			"No partition with ID '%v' exists in the current partition topology",
			partitionId,
		)
	}

	serviceContainerId, serviceIp, err := engine.userServiceLauncher.Launch(
		context,
		imageName,
		usedPorts,
		ipPlaceholder,
		startCmd,
		dockerEnvVars,
		testVolumeMountDirpath)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating the user service")
	}
	engine.serviceIps[serviceId] = serviceIp
	engine.serviceContainerIds[serviceId] = serviceContainerId

	if engine.isPartitioningEnabled {
		sidecarIp, err := engine.freeIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return stacktrace.Propagate(
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
			return stacktrace.Propagate(err, "An error occurred starting the sidecar iproute container for modifying the service container's iptables")
		}
		engine.sidecarContainerIps[serviceId] = sidecarIp
		engine.sidecarContainerIds[serviceId] = sidecarContainerId

		newBlockedServices := engine.recalculateBlocklists()
		if err := engine.updateIpTableBlocklists(context, newBlockedServices); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred updating the iptables blocklists for all services in " +
					"the wake of adding service '%v'",
				serviceId)
		}
	}

	return nil
}

// TODO test me, including speed profiling!!
func (engine PartitioningEngine) recalculateBlocklists() map[api.ServiceID]ServiceIDSet {
	result := map[api.ServiceID]ServiceIDSet{}
	for partitionId, servicesInPartition := range engine.partitionServices {
		for otherPartitionId, servicesInOtherPartition := range engine.partitionServices {
			// Services in a partition will never have services in the same partition on their blocklist
			if partitionId == otherPartitionId {
				continue
			}
			connectionId := *partition_connection.NewPartitionConnectionID(partitionId, otherPartitionId)
			connection := engine.defaultConnection
			if definedConnection, found := engine.partitionConnections[connectionId]; found {
				connection = definedConnection
			}

			if !connection.IsBlocked {
				continue
			}

			for _, serviceId := range servicesInPartition.Elems() {
				blockedServicesForId := *NewServiceIDSet()
				if definedBlockedServicesForId, found := result[serviceId]; found {
					blockedServicesForId = definedBlockedServicesForId
				}
				blockedServicesForId.AddElems(servicesInOtherPartition)
				result[serviceId] = blockedServicesForId
			}
		}
	}
	return result
}

// TODO write tests for me!!
/*
Compares the newBlockedServices with the existing state and, if any services need to have their iptables updated,
	applies the changes before finally updating the engine's state with the new state
An empty blocklist will result in the Kurtosis rule being deleted entirely from IP tables
 */
func (engine *PartitioningEngine) updateIpTableBlocklists(
		context context.Context,
		newBlockedServices map[api.ServiceID]ServiceIDSet) error {
	toUpdate, err := getServicesNeedingIpTablesUpdates(engine.blockedServices, newBlockedServices)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the services that need iptables updated")
	}

	sidecarContainerCmds, err := getSidecarContainerCommands(toUpdate, engine.serviceIps)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the sidecar container commands for the " +
			"services that need iptables updates")
	}

	// TODO Run the container updates in parallel
	for serviceId, command := range sidecarContainerCmds {
		sidecarContainerId, found := engine.sidecarContainerIds[serviceId]
		if !found {
			// TODO maybe start one if we can't find it?
			return stacktrace.NewError(
				"Couldn't find a sidecar container for Service ID '%v', which means there's no way " +
					"to update its iptables",
				serviceId,
			)
		}

		logrus.Info("Running iptables command '%v' in sidecar container '%v' to update blocklist for service '%v'...")
		commandLogsBuf := &bytes.Buffer{}
		if err := engine.dockerManager.RunExecCommand(context, sidecarContainerId, command, commandLogsBuf); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred running iptables command '%v' in sidecar container '%v' to update the blocklist of service '%v'",
				command,
				sidecarContainerId,
				serviceId)
		}
		logrus.Info("Successfully updated blocklist for service '%v'")
		logrus.Info("---------------------------------- Command Logs ---------------------------------------")
		io.Copy(logrus.StandardLogger().Out, commandLogsBuf)
		logrus.Info("---------------------------------End Command Logs -------------------------------------")
	}

	// Defensive copy when we store
	blockListToStore := map[api.ServiceID]ServiceIDSet{}
	for serviceId, newBlockedServicesForId := range newBlockedServices {
		blockListToStore[serviceId] = newBlockedServicesForId.Copy()
	}
	engine.blockedServices = blockListToStore
	return nil
}

// TODO Write tests for me!!
/*
Compares the desired new state of the world with the current state of the world, and returns only a list of
	the services that need to be updated.
 */
func getServicesNeedingIpTablesUpdates(
		currentBlockedServices map[api.ServiceID]ServiceIDSet,
		newBlockedServices map[api.ServiceID]ServiceIDSet) (map[api.ServiceID]ServiceIDSet, error) {
	result := map[api.ServiceID]ServiceIDSet{}
	for serviceId, newBlockedServicesForId := range newBlockedServices {
		if newBlockedServicesForId.Contains(serviceId) {
			return nil, stacktrace.NewError("Requested for service ID '%v' to block itself!", serviceId)
		}

		// To avoid unnecessary Docker work, we won't update any iptables if the result would be the same
		//  as the current state
		currentBlockedServicesForId, found := currentBlockedServices[serviceId]
		if !found {
			currentBlockedServicesForId = *NewServiceIDSet()
		}

		noChangesNeeded := newBlockedServicesForId.Equal(currentBlockedServicesForId)
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
		toUpdate map[api.ServiceID]ServiceIDSet,
		serviceIps map[api.ServiceID]net.IP) (map[api.ServiceID][]string, error) {
	result := map[api.ServiceID][]string{}


	// TODO We build two separate commands - flush the Kurtosis iptables chain, and then populate it with new stuff
	//  This means that there's a (very small) window of time where the iptables aren't blocked
	//  To fix this, we should really have two Kurtosis chains, and while one is running build the other one and
	//  then switch over in one idempotent operation.
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
				ipToBlock, found := serviceIps[serviceIdToBlock]
				if !found {
					return nil, stacktrace.NewError("Service ID '%v' needs to block the IP of target service ID '%v', but " +
						"the target service doesn't have an IP associated to it",
						serviceId,
						serviceIdToBlock)
				}
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