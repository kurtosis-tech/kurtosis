/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package partition_topology

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

type PartitionConnection struct {
	PacketLossPercentage float32
}

// TODO Add a RWMutex to this class to keep it threadsafe: https://github.com/kurtosis-tech/kurtosis-core/issues/582https://github.com/kurtosis-tech/kurtosis-core/issues/582
// Stores the partition topology of the network, and exposes an API for modifying it
type PartitionTopology struct {
	defaultConnection PartitionConnection

	servicePartitions map[service.ServiceID]service_network_types.PartitionID

	partitionConnections map[service_network_types.PartitionConnectionID]PartitionConnection

	// A service can be a part of exactly one partition at a time
	partitionServices map[service_network_types.PartitionID]map[service.ServiceID]bool // partitionId -> set<serviceId>
}

func NewPartitionTopology(defaultPartition service_network_types.PartitionID, defaultConnection PartitionConnection) *PartitionTopology {
	return &PartitionTopology{
		servicePartitions:    map[service.ServiceID]service_network_types.PartitionID{},
		partitionConnections: map[service_network_types.PartitionConnectionID]PartitionConnection{},
		partitionServices: map[service_network_types.PartitionID]map[service.ServiceID]bool{
			defaultPartition: map[service.ServiceID]bool{},
		},
		defaultConnection: defaultConnection,
	}
}

// ================================================================================================
//
//	Public Methods
//
// ================================================================================================
func (topology *PartitionTopology) Repartition(
	newPartitionServices map[service_network_types.PartitionID]map[service.ServiceID]bool,
	newPartitionConnections map[service_network_types.PartitionConnectionID]PartitionConnection,
	newDefaultConnection PartitionConnection) error {
	// Validate we have at least one partition
	if len(newPartitionServices) == 0 {
		return stacktrace.NewError("Cannot repartition with no partitions")
	}

	// Validate that each existing service in the testnet gets exactly one partition allocation
	allServicesInNetwork := map[service.ServiceID]bool{}
	servicesNeedingAllocation := map[service.ServiceID]bool{}
	for serviceId := range topology.servicePartitions {
		allServicesInNetwork[serviceId] = true
		servicesNeedingAllocation[serviceId] = true
	}
	allocatedServices := map[service.ServiceID]bool{}
	unknownServices := map[service.ServiceID]bool{}
	duplicatedAllocations := map[service.ServiceID]bool{}
	for _, servicesForPartition := range newPartitionServices {
		for serviceId := range servicesForPartition {
			if doesServiceSetContainsElement(allocatedServices, serviceId) {
				duplicatedAllocations[serviceId] = true
			}
			if !doesServiceSetContainsElement(allServicesInNetwork, serviceId) {
				unknownServices[serviceId] = true
			}
			allocatedServices[serviceId] = true
			delete(servicesNeedingAllocation, serviceId)
		}
	}
	if len(servicesNeedingAllocation) > 0 {
		return stacktrace.NewError(
			"All services must be allocated to a partition when repartitioning, but the following weren't "+
				"accounted for: %v",
			serviceIdSetToCommaStr(servicesNeedingAllocation),
		)
	}
	if len(unknownServices) > 0 {
		return stacktrace.NewError(
			"The following services are unkonwn, but have partition definitions: %v",
			serviceIdSetToCommaStr(unknownServices),
		)
	}
	if len(duplicatedAllocations) > 0 {
		return stacktrace.NewError(
			"The following services have partitions defined twice: %v",
			serviceIdSetToCommaStr(duplicatedAllocations),
		)
	}

	// Validate the connections point to defined partitions
	for partitionConnectionId := range newPartitionConnections {
		firstPartition := partitionConnectionId.GetFirst()
		secondPartition := partitionConnectionId.GetSecond()
		if _, found := newPartitionServices[firstPartition]; !found {
			return stacktrace.NewError(
				"Partition '%v' in partition connection '%v' <-> '%v' doesn't exist",
				firstPartition,
				firstPartition,
				secondPartition)
		}
		if _, found := newPartitionServices[secondPartition]; !found {
			return stacktrace.NewError("Partition '%v' in partition connection '%v' <-> '%v' doesn't exist",
				secondPartition,
				firstPartition,
				secondPartition)
		}
	}

	// Defensive copies
	newPartitionServicesCopy := map[service_network_types.PartitionID]map[service.ServiceID]bool{}
	newServicePartitionsCopy := map[service.ServiceID]service_network_types.PartitionID{}
	for partitionId, servicesForPartition := range newPartitionServices {
		newPartitionServicesCopy[partitionId] = copyServiceSet(servicesForPartition)
		for serviceId := range servicesForPartition {
			newServicePartitionsCopy[serviceId] = partitionId
		}
	}
	newPartitionConnectionsCopy := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	for partitionConnectionId, connection := range newPartitionConnections {
		newPartitionConnectionsCopy[partitionConnectionId] = connection
	}

	topology.partitionServices = newPartitionServicesCopy
	topology.servicePartitions = newServicePartitionsCopy
	topology.partitionConnections = newPartitionConnectionsCopy
	topology.defaultConnection = newDefaultConnection

	return nil
}

func (topology *PartitionTopology) AddService(serviceId service.ServiceID, partitionId service_network_types.PartitionID) error {
	if existingPartition, found := topology.servicePartitions[serviceId]; found {
		return stacktrace.NewError(
			"Cannot add service '%v' to partition '%v' because the service is already assigned to partition '%v'",
			serviceId,
			partitionId,
			existingPartition)
	}

	servicesForPartition, found := topology.partitionServices[partitionId]
	if !found {
		return stacktrace.NewError(
			"Cannot assign service '%v' to partition '%v'; the partition doesn't exist",
			serviceId,
			partitionId)
	}
	servicesForPartition[serviceId] = true
	topology.servicePartitions[serviceId] = partitionId
	return nil
}

/*
Removes the given service from the toplogy, if it exists. If it doesn't exist, this is a no-op.
*/
func (topology *PartitionTopology) RemoveService(serviceId service.ServiceID) {
	partitionId, found := topology.servicePartitions[serviceId]
	if !found {
		return
	}
	delete(topology.servicePartitions, serviceId)

	servicesForPartition, found := topology.partitionServices[partitionId]
	if !found {
		return
	}
	delete(servicesForPartition, serviceId)
}

func (topology PartitionTopology) GetPartitionServices() map[service_network_types.PartitionID]map[service.ServiceID]bool {
	return topology.partitionServices
}

func (topology PartitionTopology) GetServicePacketLossConfigurationsByServiceID() (map[service.ServiceID]map[service.ServiceID]float32, error) {
	result := map[service.ServiceID]map[service.ServiceID]float32{}
	for partitionId, servicesInPartition := range topology.partitionServices {
		for serviceId := range servicesInPartition {
			otherServicesPacketLossConfigMap := map[service.ServiceID]float32{}
			for otherPartitionId, servicesInOtherPartition := range topology.partitionServices {
				if partitionId == otherPartitionId {
					// Two services in the same partition will never block each other
					continue
				}
				connection, err := topology.getPartitionConnection(partitionId, otherPartitionId)
				if err != nil {
					return nil, stacktrace.NewError("Couldn't get connection between partitions '%v' and '%v'", partitionId, otherPartitionId)
				}
				for otherServiceId := range servicesInOtherPartition {
					otherServicesPacketLossConfigMap[otherServiceId] = connection.PacketLossPercentage
				}
			}
			result[serviceId] = otherServicesPacketLossConfigMap
		}
	}
	return result, nil
}

// ================================================================================================
//
//	Private Helper Methods
//
// ================================================================================================
func (topology PartitionTopology) getPartitionConnection(
	a service_network_types.PartitionID,
	b service_network_types.PartitionID) (PartitionConnection, error) {
	if _, found := topology.partitionServices[a]; !found {
		return PartitionConnection{}, stacktrace.NewError("Unrecognized partition '%v'", a)
	}
	if _, found := topology.partitionServices[b]; !found {
		return PartitionConnection{}, stacktrace.NewError("Unrecognized partition '%v'", b)
	}
	connectionId := service_network_types.NewPartitionConnectionID(a, b)
	connection, found := topology.partitionConnections[*connectionId]
	if !found {
		return topology.defaultConnection, nil
	}
	return connection, nil
}

func serviceIdSetToCommaStr(serviceSet map[service.ServiceID]bool) string {
	strSlice := []string{}
	for serviceId := range serviceSet {
		strSlice = append(strSlice, string(serviceId))
	}
	return strings.Join(strSlice, ", ")
}

func doesServiceSetContainsElement(serviceSet map[service.ServiceID]bool, element service.ServiceID) bool {
	if _, found := serviceSet[element]; found {
		return true
	}
	return false
}

func copyServiceSet(serviceSet map[service.ServiceID]bool) map[service.ServiceID]bool {
	newServiceSet := map[service.ServiceID]bool{}
	for serviceGuid := range serviceSet {
		newServiceSet[serviceGuid] = true
	}
	return newServiceSet
}
