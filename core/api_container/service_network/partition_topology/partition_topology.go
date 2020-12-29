/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package partition_topology

import (
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/palantir/stacktrace"
	"strings"
)

type PartitionConnection struct {
	IsBlocked bool
}


// Stores the partition topology of the network, and exposes an API for modifying it
type PartitionTopology struct {
	defaultConnection PartitionConnection

	servicePartitions map[topology_types.ServiceID]topology_types.PartitionID

	partitionConnections map[topology_types.PartitionConnectionID]PartitionConnection

	// A service can be a part of exactly one partition at a time
	partitionServices map[topology_types.PartitionID]*topology_types.ServiceIDSet // partitionId -> set<serviceId>
}

func NewPartitionTopology(defaultPartition topology_types.PartitionID, defaultConnection PartitionConnection) *PartitionTopology {
	return &PartitionTopology{
		servicePartitions:    map[topology_types.ServiceID]topology_types.PartitionID{},
		partitionConnections: map[topology_types.PartitionConnectionID]PartitionConnection{},
		partitionServices: map[topology_types.PartitionID]*topology_types.ServiceIDSet{
			defaultPartition: topology_types.NewServiceIDSet(),
		},
		defaultConnection: defaultConnection,
	}
}

// ================================================================================================
//                                        Public Methods
// ================================================================================================
func (topology *PartitionTopology) Repartition(
		newPartitionServices map[topology_types.PartitionID]*topology_types.ServiceIDSet,
		newPartitionConnections map[topology_types.PartitionConnectionID]PartitionConnection,
		newDefaultConnection PartitionConnection) error {
	// Validate we have at least one partition
	if len(newPartitionServices) == 0 {
		return stacktrace.NewError("Cannot repartition with no partitions")
	}

	// Validate that each existing service in the testnet gets exactly one partition allocation
	allServicesInNetwork := topology_types.NewServiceIDSet()
	for serviceId, _ := range topology.servicePartitions {
		allServicesInNetwork.AddElem(serviceId)
	}
	servicesNeedingAllocation := allServicesInNetwork.Copy()
	allocatedServices := topology_types.NewServiceIDSet()
	unknownServices := topology_types.NewServiceIDSet()
	duplicatedAllocations := topology_types.NewServiceIDSet()
	for _, servicesForPartition := range newPartitionServices {
		for _, serviceId := range servicesForPartition.Elems() {
			if allocatedServices.Contains(serviceId) {
				duplicatedAllocations.AddElem(serviceId)
			}
			if !allServicesInNetwork.Contains(serviceId) {
				unknownServices.AddElem(serviceId)
			}
			allocatedServices.AddElem(serviceId)
			servicesNeedingAllocation.RemoveElem(serviceId)
		}
	}
	if servicesNeedingAllocation.Size() > 0 {
		return stacktrace.NewError(
			"All services must be allocated to a partition when repartitioning, but the following weren't " +
				"accounted for: %v",
			serviceIdSetToCommaStr(servicesNeedingAllocation),
		)
	}
	if unknownServices.Size() > 0 {
		return stacktrace.NewError(
			"The following services are unkonwn, but have partition definitions: %v",
			serviceIdSetToCommaStr(unknownServices),
		)
	}
	if duplicatedAllocations.Size() > 0 {
		return stacktrace.NewError(
			"The following services have partitions defined twice: %v",
			serviceIdSetToCommaStr(duplicatedAllocations),
		)
	}

	// Validate the connections point to defined partitions
	for partitionConnectionId, _ := range newPartitionConnections {
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
	newPartitionServicesCopy := map[topology_types.PartitionID]*topology_types.ServiceIDSet{}
	newServicePartitionsCopy := map[topology_types.ServiceID]topology_types.PartitionID{}
	for partitionId, servicesForPartition := range newPartitionServices {
		newPartitionServicesCopy[partitionId] = servicesForPartition.Copy()
		for _, serviceId := range servicesForPartition.Elems() {
			newServicePartitionsCopy[serviceId] = partitionId
		}
	}
	newPartitionConnectionsCopy := map[topology_types.PartitionConnectionID]PartitionConnection{}
	for partitionConnectionId, connection := range newPartitionConnections {
		newPartitionConnectionsCopy[partitionConnectionId] = connection
	}

	topology.partitionServices = newPartitionServicesCopy
	topology.servicePartitions = newServicePartitionsCopy
	topology.partitionConnections = newPartitionConnectionsCopy
	topology.defaultConnection = newDefaultConnection

	return nil
}

func (topology *PartitionTopology) AddService(serviceId topology_types.ServiceID, partitionId topology_types.PartitionID) error {
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
	servicesForPartition.AddElem(serviceId)
	topology.servicePartitions[serviceId] = partitionId
	return nil
}

/*
Removes the given service from the toplogy, if it exists. If it doesn't exist, this is a no-op.
 */
func (topology *PartitionTopology) RemoveService(serviceId topology_types.ServiceID) {
	partitionId, found := topology.servicePartitions[serviceId]
	if !found {
		return
	}
	delete(topology.servicePartitions, serviceId)

	servicesForPartition, found := topology.partitionServices[partitionId]
	servicesForPartition.RemoveElem(serviceId)
}

func (topology PartitionTopology) GetPartitionServices() map[topology_types.PartitionID]*topology_types.ServiceIDSet {
	return topology.partitionServices
}

// TODO test me, including speed profiling!!
// Returns a map indicating, for each service, which services it should be blocking based on the current network topology
func (topology PartitionTopology) GetBlocklists() (map[topology_types.ServiceID]*topology_types.ServiceIDSet, error) {
	// TODO to speed this method up, we can remove this method in favor of spitting out updated blocklists on each change operation (addservice, repartition, etc.)
	//  so that only the nodes that need updating will get updated incrementally
	result := map[topology_types.ServiceID]*topology_types.ServiceIDSet{}
	for partitionId, servicesInPartition := range topology.partitionServices {
		for _, serviceId := range servicesInPartition.Elems() {
			blockedServices := topology_types.NewServiceIDSet()
			for otherPartitionId, servicesInOtherPartition := range topology.partitionServices {
				if partitionId == otherPartitionId {
					// Two services in the same partition will never block each other
					continue
				}
				connection, err := topology.getPartitionConnection(partitionId, otherPartitionId)
				if err != nil {
					return nil, stacktrace.NewError("Couldn't get connection between partitions '%v' and '%v'", partitionId, otherPartitionId)
				}
				for _, otherServiceId := range servicesInOtherPartition.Elems() {
					if connection.IsBlocked {
						blockedServices.AddElem(otherServiceId)
					}
				}
			}
			result[serviceId] = blockedServices
		}
	}
	return result, nil
}

// ================================================================================================
//                                  Private Helper Methods
// ================================================================================================
func (topology PartitionTopology) getPartitionConnection(
	a topology_types.PartitionID,
	b topology_types.PartitionID) (PartitionConnection, error) {
	if _, found := topology.partitionServices[a]; !found {
		return PartitionConnection{}, stacktrace.NewError("Unrecognized partition '%v'", a)
	}
	if _, found := topology.partitionServices[b]; !found {
		return PartitionConnection{}, stacktrace.NewError("Unrecognized partition '%v'", b)
	}
	connectionId := topology_types.NewPartitionConnectionID(a, b)
	connection, found := topology.partitionConnections[*connectionId]
	if !found {
		return topology.defaultConnection, nil
	}
	return connection, nil
}

func serviceIdSetToCommaStr(set *topology_types.ServiceIDSet) string {
	strSlice := []string{}
	for _, serviceId := range set.Elems() {
		strSlice = append(strSlice, string(serviceId))
	}
	return strings.Join(strSlice, ", ")
}