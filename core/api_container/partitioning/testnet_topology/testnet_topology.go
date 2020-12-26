/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testnet_topology

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api"
	"github.com/kurtosis-tech/kurtosis/api_container/partitioning"
	"github.com/palantir/stacktrace"
	"strings"
)

/*
Represents two partitions, where order is unimportant
*/
type PartitionConnectionID struct {
	lexicalFirst  partitioning.PartitionID
	lexicalSecond partitioning.PartitionID
}

func NewPartitionConnectionID(partitionA partitioning.PartitionID, partitionB partitioning.PartitionID) *PartitionConnectionID {

	// We sort these upon creation so that this type can be used as a key in a map, and so that
	// 	this tuple is commutative: partitionConnectionID(A, B) == partitionConnectionID(B, A) as a map key
	first, second := partitionA, partitionB
	result := strings.Compare(string(first), string(second))
	if result > 0 {
		first, second = second, first
	}
	return &PartitionConnectionID{
		lexicalFirst:  first,
		lexicalSecond: second,
	}
}

type PartitionConnection struct {
	IsBlocked bool
}


// Provides an API for modifying the topology of
type TestnetTopology struct {
	defaultConnection PartitionConnection

	servicePartitions map[api.ServiceID]partitioning.PartitionID

	partitionConnections map[PartitionConnectionID]PartitionConnection

	// A service can be a part of exactly one partition at a time
	partitionServices map[partitioning.PartitionID]*partitioning.ServiceIDSet // partitionId -> set<serviceId>
}

func NewTestnetTopology(defaultPartition partitioning.PartitionID, defaultConnection PartitionConnection) *TestnetTopology {
	return &TestnetTopology{
		servicePartitions:    map[api.ServiceID]partitioning.PartitionID{},
		partitionConnections: map[PartitionConnectionID]PartitionConnection{},
		partitionServices: map[partitioning.PartitionID]*partitioning.ServiceIDSet{
			defaultPartition: partitioning.NewServiceIDSet(),
		},
	}
}

// ================================================================================================
//                                        Public Methods
// ================================================================================================
// TODO Add tests for this!
func (topology *TestnetTopology) Repartition(
		newPartitionServices map[partitioning.PartitionID]*partitioning.ServiceIDSet,
		newPartitionConnections map[PartitionConnectionID]PartitionConnection,
		newDefaultConnection PartitionConnection) error {

	// Validate that each existing service in the testnet gets exactly one partition allocation
	servicesNeedingAllocation := partitioning.NewServiceIDSet()
	for serviceId, _ := range topology.servicePartitions {
		servicesNeedingAllocation.AddElem(serviceId)
	}
	allocatedServices := partitioning.NewServiceIDSet()
	unknownServices := partitioning.NewServiceIDSet()
	duplicatedAllocations := partitioning.NewServiceIDSet()
	for _, servicesForPartition := range newPartitionServices {
		for _, serviceId := range servicesForPartition.Elems() {
			if allocatedServices.Contains(serviceId) {
				duplicatedAllocations.AddElem(serviceId)
			}
			if !servicesNeedingAllocation.Contains(serviceId) {
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
		if _, found := newPartitionServices[partitionConnectionId.lexicalFirst]; !found {
			return stacktrace.NewError(
				"Partition '%v' in partition connection '%v' <-> '%v' doesn't exist",
				partitionConnectionId.lexicalFirst,
				partitionConnectionId.lexicalFirst,
				partitionConnectionId.lexicalSecond)
		}
		if _, found := newPartitionServices[partitionConnectionId.lexicalSecond]; !found {
			return stacktrace.NewError("Partition '%v' in partition connection '%v' <-> '%v' doesn't exist",
				partitionConnectionId.lexicalSecond,
				partitionConnectionId.lexicalFirst,
				partitionConnectionId.lexicalSecond)
		}
	}

	// Defensive copies
	newPartitionServicesCopy := map[partitioning.PartitionID]*partitioning.ServiceIDSet{}
	newServicePartitionsCopy := map[api.ServiceID]partitioning.PartitionID{}
	for partitionId, servicesForPartition := range newPartitionServices {
		newPartitionServicesCopy[partitionId] = servicesForPartition.Copy()
		for _, serviceId := range servicesForPartition.Elems() {
			newServicePartitionsCopy[serviceId] = partitionId
		}
	}
	newPartitionConnectionsCopy := map[PartitionConnectionID]PartitionConnection{}
	for partitionConnectionId, connection := range newPartitionConnections {
		newPartitionConnectionsCopy[partitionConnectionId] = connection
	}

	topology.partitionServices = newPartitionServicesCopy
	topology.servicePartitions = newServicePartitionsCopy
	topology.partitionConnections = newPartitionConnectionsCopy
	topology.defaultConnection = newDefaultConnection

	return nil
}

func (topology *TestnetTopology) AddService(serviceId api.ServiceID, partitionId partitioning.PartitionID) error {
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

func (topology *TestnetTopology) RemoveService(serviceId api.ServiceID) {
	partitionId, found := topology.servicePartitions[serviceId]
	if !found {
		return
	}
	delete(topology.servicePartitions, serviceId)

	servicesForPartition, found := topology.partitionServices[partitionId]
	servicesForPartition.RemoveElem(serviceId)
}

func (topology TestnetTopology) GetPartitionServices() map[partitioning.PartitionID]*partitioning.ServiceIDSet {
	return topology.partitionServices
}


/*
func (topology TestnetTopology) GetPartitionServices(partitionId partitioning.PartitionID) (*partitioning.ServiceIDSet, bool) {
	services, found := topology.partitionServices[partitionId]
	return services, found
}

func (topology TestnetTopology) GetPartitions() []partitioning.PartitionID {
	// WHY doesn't Go have a stupid set type
	result := []partitioning.PartitionID{}
	for partitionId, _ := range topology.partitionServices {
		result = append(result, partitionId)
	}
	return result
}
*/

// TODO test me, including speed profiling!!
// Returns a map indicating, for each service, which services it should be blocking based on the current network topology
func (topology TestnetTopology) GetBlocklists() (map[api.ServiceID]*partitioning.ServiceIDSet, error) {
	// TODO to speed this method up, we can remove this method in favor of spitting out updated blocklists on each change operation (addservice, repartition, etc.)
	//  so that only the nodes that need updating will get updated incrementally
	result := map[api.ServiceID]*partitioning.ServiceIDSet{}
	for partitionId, servicesInPartition := range topology.partitionServices {
		for _, serviceId := range servicesInPartition.Elems() {
			blockedServices := partitioning.NewServiceIDSet()
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
func (topology TestnetTopology) getPartitionConnection(
	a partitioning.PartitionID,
	b partitioning.PartitionID) (PartitionConnection, error) {
	if _, found := topology.partitionServices[a]; !found {
		return PartitionConnection{}, stacktrace.NewError("Unrecognized partition '%v'", a)
	}
	if _, found := topology.partitionServices[b]; !found {
		return PartitionConnection{}, stacktrace.NewError("Unrecognized partition '%v'", b)
	}
	connectionId := NewPartitionConnectionID(a, b)
	connection, found := topology.partitionConnections[*connectionId]
	if !found {
		return topology.defaultConnection, nil
	}
	return connection, nil
}

func serviceIdSetToCommaStr(set *partitioning.ServiceIDSet) string {
	strSlice := []string{}
	for _, serviceId := range set.Elems() {
		strSlice = append(strSlice, string(serviceId))
	}
	return strings.Join(strSlice, ", ")
}