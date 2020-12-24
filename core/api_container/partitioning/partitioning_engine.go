/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package partitioning

import (
	"github.com/kurtosis-tech/kurtosis/api_container/partitioning/partition_connection"
	"github.com/kurtosis-tech/kurtosis/api_container/partitioning/service_id_set"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"strconv"
	"strings"
)

const (
	defaultPartitionId PartitionID = "default"
	defaultConnectionBlockStatus = false

	inputIpTablesChain = "INPUT"

	// We'll create a new chain at the 0th spot for every service
	kurtosisIpTablesChain = "KURTOSIS"

	ipTablesCommand = "iptables"
	ipTablesFlushChainFlag = "-F"
	ipTablesAppendRuleFlag  = "-A"
	ipTablesDropAction = "DROP"
)

/**
This is the engine
 */
type PartitioningEngine struct {
	// Whether partitioning has been enabled for this particular test
	isPartitioningEnabled bool

	dockerManager *commons.DockerManager

	serviceIps map[ServiceID]string

	// Mapping of Service ID -> the sidecar iproute container that we'll use to manipulate the
	//  container's networking
	sidecarContainerIds map[ServiceID]string

	defaultConnection partition_connection.PartitionConnection

	partitionConnections map[partition_connection.PartitionConnection]

	// A service can be a part of exactly one partition at a time
	partitionServices map[PartitionID]service_id_set.ServiceIDSet // partitionId -> set<serviceId>
	servicePartitions map[ServiceID]PartitionID

	// Mapping of serviceID -> set of serviceIDs that are currently being dropped in the INPUT chain of the service
	// An entry in this list means that the host has a Kurtosis entry as its first entry in the INPUT chain, set to the
	//  IPs of the services
	// A missing entry means that iptables doesn't have a Kurtosis entry present
	blockedServices map[ServiceID]service_id_set.ServiceIDSet
}

func NewPartitioningEngine(isPartitioningEnabled bool) *PartitioningEngine {
	return &PartitioningEngine{
		isPartitioningEnabled: isPartitioningEnabled,
		sidecarContainerIds:   map[ServiceID]string{},
		defaultConnection: partition_connection.PartitionConnection{
			IsBlocked: defaultConnectionBlockStatus,
		},
		partitionServices: map[PartitionID]service_id_set.ServiceIDSet{
			defaultPartitionId: *service_id_set.NewServiceIDSet(),
		},
		servicePartitions: map[ServiceID]PartitionID{},
	}
}

/*
Completely repartitions the network, throwing away the old topology
 */
func (engine *PartitioningEngine) Repartition(
		servicePartitions map[ServiceID]PartitionID,
		partitionConnections map[partition_connection.PartitionTuple]partition_connection.PartitionConnection) error {
	if !engine.isPartitioningEnabled {
		stacktrace.NewError("Cannot repartition; partitioning is not enabled")
	}


}



/*
Creates the service with the given ID in the given partition

If partitionId is empty string, the default partition ID is used
 */
func (engine *PartitioningEngine) CreateServiceInPartition(
		serviceId ServiceID,
		partitionId PartitionID) error {
	if partitionId == "" {
		partitionId = defaultPartitionId
	}

	if _, found := engine.partitionServices[partitionId]; !found {
		return stacktrace.NewError(
			"No partition with ID '%v' exists in the current partition topology",
			partitionId,
		)
	}

	newInputChainDrops := map[ServiceID]map[string]bool{}
	for otherPartitionId, servicesInPartition := range engine.partitionServices {
		// We'll never drop traffic from hosts inside our own partition
		if partitionId == otherPartitionId {
			continue
		}



	}




	// TODO Create the service
	if engine.isPartitioningEnabled {
		// TODO Create the sidecar immediately after
	}

	// TODO Create the KURTOSIS chain at the 0th spot in the INPUT chain

	return nil
}

func calcNewBlocklistWithPartitionHint() {

}

/*
Compares the newBlockedServices with the existing state and, if any services need to have their iptables updated,
	applies the changes before finally updating the engine's state with the new state
An empty blocklist will result in the Kurtosis rule being deleted entirely from IP tables
 */
func (engine *PartitioningEngine) updateIpTableBlocklists(newBlockedServices map[ServiceID]service_id_set.ServiceIDSet) error {
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
		containerId, found := engine.sidecarContainerIds[serviceId]
		if !found {
			// TODO maybe start one if we can't find it?
			return stacktrace.NewError(
				"Couldn't find a sidecar container for Service ID '%v', which means there's no way " +
					"to update its iptables",
				serviceId,
			)
		}

		// TODO issue a command to the sidecar container service, telling it to update the iptables
	}

	// Defensive copy when we store
	blockListToStore := map[ServiceID]service_id_set.ServiceIDSet{}
	for serviceId, newBlockedServicesForId := range newBlockedServices {
		blockListToStore[serviceId] = newBlockedServicesForId.Copy()
	}
	engine.blockedServices = blockListToStore
}

/*
Compares the desired new state of the world with the current state of the world, and returns only a list of
	the services that need to be updated.
 */
func getServicesNeedingIpTablesUpdates(
		currentBlockedServices map[ServiceID]service_id_set.ServiceIDSet,
		newBlockedServices map[ServiceID]service_id_set.ServiceIDSet) (map[ServiceID]service_id_set.ServiceIDSet, error) {
	result := map[ServiceID]service_id_set.ServiceIDSet{}
	for serviceId, newBlockedServicesForId := range newBlockedServices {
		if newBlockedServicesForId.Contains(serviceId) {
			return nil, stacktrace.NewError("Requested for service ID '%v' to block itself!", serviceId)
		}

		// To avoid unnecessary Docker work, we won't update any iptables if the result would be the same
		//  as the current state
		currentBlockedServicesForId, found := currentBlockedServices[serviceId]
		if !found {
			currentBlockedServicesForId = *service_id_set.NewServiceIDSet()
		}

		noChangesNeeded := newBlockedServicesForId.Equal(currentBlockedServicesForId)
		if noChangesNeeded {
			continue
		}

		result[serviceId] = newBlockedServicesForId
	}
	return result, nil
}

/*
Given a list of updates that need to happen to a service's iptables, a map of serviceID -> commands that
	will be executed on the sidecar Docker container for the service

Args:
	toUpdate: A mapping of serviceID -> set of serviceIDs to block
 */
func getSidecarContainerCommands(
		toUpdate map[ServiceID]service_id_set.ServiceIDSet,
		serviceIps map[ServiceID]string) (map[ServiceID][]string, error) {
	result := map[ServiceID][]string{}
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
			allIpsToBlock := []string{}
			for _, serviceIdToBlock := range newBlockedServicesForId.Elems() {
				ipToBlock, found := serviceIps[serviceIdToBlock]
				if !found {
					return nil, stacktrace.NewError("Service ID '%v' needs to block the IP of target service ID '%v', but " +
						"the target service doesn't have an IP associated to it",
						serviceId,
						serviceIdToBlock)
				}
				allIpsToBlock = append(allIpsToBlock, ipToBlock)
			}
			ipsToBlockCommaList := strings.Join(allIpsToBlock, ",")

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