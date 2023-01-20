/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package partition_topology

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

const (
	DefaultPartitionId = service_network_types.PartitionID("default")
)

// Stores the partition topology of the network, and exposes an API for modifying it
type PartitionTopology struct {
	lock *sync.RWMutex

	defaultConnection PartitionConnection

	servicePartitions map[service.ServiceName]service_network_types.PartitionID

	// By default, connection between 2 partitions is set to defaultConnection. This map contains overrides
	partitionConnectionOverrides map[service_network_types.PartitionConnectionID]PartitionConnection

	// A service can be a part of exactly one partition at a time
	partitionServices map[service_network_types.PartitionID]map[service.ServiceName]bool // partitionId -> set<serviceId>
}

func NewPartitionTopology(defaultPartition service_network_types.PartitionID, defaultConnection PartitionConnection) *PartitionTopology {
	return &PartitionTopology{
		lock:              &sync.RWMutex{},
		servicePartitions: map[service.ServiceName]service_network_types.PartitionID{},
		partitionServices: map[service_network_types.PartitionID]map[service.ServiceName]bool{
			defaultPartition: {},
		},
		partitionConnectionOverrides: map[service_network_types.PartitionConnectionID]PartitionConnection{},
		defaultConnection:            defaultConnection,
	}
}

// ParsePartitionId returns the partition ID form the provided strings.
// As partition ID is optional in most places, it falls back to DefaultPartitionID is the argument is nil or empty
func ParsePartitionId(partitionIdMaybe *string) service_network_types.PartitionID {
	if partitionIdMaybe == nil || *partitionIdMaybe == "" {
		return DefaultPartitionId
	}
	return service_network_types.PartitionID(*partitionIdMaybe)
}

// ================================================================================================
//
//	Public Methods
//
// ================================================================================================
func (topology *PartitionTopology) Repartition(
	newPartitionServices map[service_network_types.PartitionID]map[service.ServiceName]bool,
	newPartitionConnectionOverrides map[service_network_types.PartitionConnectionID]PartitionConnection,
	newDefaultConnection PartitionConnection) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	// Validate we have at least one partition
	if len(newPartitionServices) == 0 {
		return stacktrace.NewError("Cannot repartition with no partitions")
	}

	// Validate that each existing service in the testnet gets exactly one partition allocation
	allServicesInNetwork := map[service.ServiceName]bool{}
	servicesNeedingAllocation := map[service.ServiceName]bool{}
	for serviceId := range topology.servicePartitions {
		allServicesInNetwork[serviceId] = true
		servicesNeedingAllocation[serviceId] = true
	}
	allocatedServices := map[service.ServiceName]bool{}
	unknownServices := map[service.ServiceName]bool{}
	duplicatedAllocations := map[service.ServiceName]bool{}
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
	for partitionConnectionId := range newPartitionConnectionOverrides {
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
	newPartitionServicesCopy := map[service_network_types.PartitionID]map[service.ServiceName]bool{}
	newServicePartitionsCopy := map[service.ServiceName]service_network_types.PartitionID{}
	for partitionId, servicesForPartition := range newPartitionServices {
		newPartitionServicesCopy[partitionId] = copyServiceSet(servicesForPartition)
		for serviceId := range servicesForPartition {
			newServicePartitionsCopy[serviceId] = partitionId
		}
	}
	newPartitionConnectionOverridesCopy := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	for partitionConnectionId, connection := range newPartitionConnectionOverrides {
		newPartitionConnectionOverridesCopy[partitionConnectionId] = connection
	}

	topology.partitionServices = newPartitionServicesCopy
	topology.servicePartitions = newServicePartitionsCopy
	topology.partitionConnectionOverrides = newPartitionConnectionOverridesCopy
	topology.defaultConnection = newDefaultConnection
	return nil
}

// CreateEmptyPartitionWithDefaultConnection creates an empty connection with no connection overrides (i.e. all
// connections to this partition will inherit the defaultConnection)
// It returns an error if the partition ID already exists
func (topology *PartitionTopology) CreateEmptyPartitionWithDefaultConnection(newPartitionId service_network_types.PartitionID) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	if _, found := topology.partitionServices[newPartitionId]; found {
		return stacktrace.NewError("Partition with ID '%v' can't be created empty because it already exists in the topology", newPartitionId)
	}
	// servicePartitions remains unchanged as the new partition is empty
	// partitionConnections remains unchanged as default connection is being used for this new partition

	// update partitionServices. As the new partition is empty, it is mapped to an empty set
	topology.partitionServices[newPartitionId] = map[service.ServiceName]bool{}
	return nil
}

// RemovePartition removes the partition from the topology if it is present and empty.
// If it is not present, it returns successfully and does nothing
// If the partition is present and not empty, it throws an error as the partition cannot be removed from the topology
// Note that the default partition cannot be removed. It will throw an error is an attempt is being made to remove the
// default partition
func (topology *PartitionTopology) RemovePartition(partitionId service_network_types.PartitionID) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	if partitionId == DefaultPartitionId {
		return stacktrace.NewError("Default partition cannot be removed")
	}

	servicesInPartition, found := topology.partitionServices[partitionId]
	if !found {
		logrus.Debugf("Removing partition '%s' that does not exist. Not doing anything", partitionId)
		// partition does not seem to exist. No-op
		return nil
	}

	numServicesInPartition := len(servicesInPartition)
	if numServicesInPartition > 0 {
		// partition is not empty. No-op
		return stacktrace.NewError("Partition '%s' cannot be removed as it currently contains '%d' services", partitionId, numServicesInPartition)
	}

	// delete the entry in partitionServices
	delete(topology.partitionServices, partitionId)

	// update partition connections dropping all potential entries referencing the deleted partition
	for partitionConnectionId := range topology.partitionConnectionOverrides {
		if partitionConnectionId.GetFirst() == partitionId || partitionConnectionId.GetSecond() == partitionId {
			// drop this partition connection
			delete(topology.partitionConnectionOverrides, partitionConnectionId)
		}
	}
	return nil
}

// SetDefaultConnection sets the default connection by updating its value.
// Note that all connections between 2 partitions inheriting from defaultConnection will be affected
func (topology *PartitionTopology) SetDefaultConnection(connection PartitionConnection) {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	topology.defaultConnection = connection
}

// GetDefaultConnection returns a safe-copy of the current defaultConnection
// Use SetDefaultConnection to update the default connection of this topology
func (topology *PartitionTopology) GetDefaultConnection() PartitionConnection {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	return topology.defaultConnection
}

// SetConnection overrides the connection between partition1 and partition2.
// It throws an error if either of the two partitions does not exist yet
func (topology *PartitionTopology) SetConnection(partition1 service_network_types.PartitionID, partition2 service_network_types.PartitionID, connection PartitionConnection) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	if _, found := topology.partitionServices[partition1]; !found {
		return stacktrace.NewError("About to set a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition1)
	}
	if _, found := topology.partitionServices[partition2]; !found {
		return stacktrace.NewError("About to set a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition2)
	}
	partitionConnectionId := service_network_types.NewPartitionConnectionID(partition1, partition2)
	topology.partitionConnectionOverrides[*partitionConnectionId] = connection
	return nil
}

// UnsetConnection unsets the connection override between partition1 and partition2. It will therefore fallback to
// defaultConnection
// It throws an error if either of the two partitions does not exist yet
// It no-ops if there was no override for this partition connection yet
func (topology *PartitionTopology) UnsetConnection(partition1 service_network_types.PartitionID, partition2 service_network_types.PartitionID) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	if _, found := topology.partitionServices[partition1]; !found {
		return stacktrace.NewError("About to unset a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition1)
	}
	if _, found := topology.partitionServices[partition2]; !found {
		return stacktrace.NewError("About to unset a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition2)
	}
	partitionConnectionId := service_network_types.NewPartitionConnectionID(partition1, partition2)
	delete(topology.partitionConnectionOverrides, *partitionConnectionId)
	return nil
}

func (topology *PartitionTopology) AddService(serviceName service.ServiceName, partitionId service_network_types.PartitionID) error {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	if existingPartition, found := topology.servicePartitions[serviceName]; found {
		return stacktrace.NewError(
			"Cannot add service '%v' to partition '%v' because the service is already assigned to partition '%v'",
			serviceName,
			partitionId,
			existingPartition)
	}

	servicesForPartition, found := topology.partitionServices[partitionId]
	if !found {
		return stacktrace.NewError(
			"Cannot assign service '%v' to partition '%v'; the partition doesn't exist",
			serviceName,
			partitionId)
	}
	servicesForPartition[serviceName] = true
	topology.servicePartitions[serviceName] = partitionId
	topology.partitionServices[partitionId][serviceName] = true
	return nil
}

// RemoveService removes the given service from the topology, if it exists. If it doesn't exist, this is a no-op.
// Note that RemoveService leaves the partition in the topology even if it is empty after the service has been removed
func (topology *PartitionTopology) RemoveService(serviceName service.ServiceName) {
	topology.lock.Lock()
	defer topology.lock.Unlock()
	partitionId, found := topology.servicePartitions[serviceName]
	if !found {
		return
	}
	delete(topology.servicePartitions, serviceName)

	servicesForPartition, found := topology.partitionServices[partitionId]
	if !found {
		return
	}
	delete(servicesForPartition, serviceName)
}

func (topology *PartitionTopology) GetPartitionServices() map[service_network_types.PartitionID]map[service.ServiceName]bool {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	return topology.partitionServices
}

// GetPartitionConnection returns a clone of the partition connection between the 2 partitions
// It also returns a boolean indicating whether the connection was the default connection or not
// It throws an error if the one of the partition does not exist.
func (topology *PartitionTopology) GetPartitionConnection(partition1 service_network_types.PartitionID, partition2 service_network_types.PartitionID) (bool, PartitionConnection, error) {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	if _, found := topology.partitionServices[partition1]; !found {
		return false, ConnectionAllowed, stacktrace.NewError("About to get a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition1)
	}
	if _, found := topology.partitionServices[partition2]; !found {
		return false, ConnectionAllowed, stacktrace.NewError("About to get a connection between '%s' and '%s' but '%s' does not exist", partition1, partition2, partition2)
	}

	partitionConnectionId := service_network_types.NewPartitionConnectionID(partition1, partition2)
	currentPartitionConnection, found := topology.partitionConnectionOverrides[*partitionConnectionId]
	if !found {
		return true, topology.GetDefaultConnection(), nil
	}
	return false, currentPartitionConnection, nil
}

func (topology *PartitionTopology) GetServicePartitions() map[service.ServiceName]service_network_types.PartitionID {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	return topology.servicePartitions
}

// GetServicePacketLossConfigurationsByServiceName this method returns a partition config map
// containing information a structure similar to adjacency graph hashmap data structure between services
// where nodes are services, and edges are partition connection object
func (topology *PartitionTopology) GetServicePacketLossConfigurationsByServiceName() (map[service.ServiceName]map[service.ServiceName]*PartitionConnection, error) {
	topology.lock.RLock()
	defer topology.lock.RUnlock()
	result := map[service.ServiceName]map[service.ServiceName]*PartitionConnection{}
	for partitionId, servicesInPartition := range topology.partitionServices {
		for serviceName := range servicesInPartition {
			partitionConnectionConfigBetweenServices := map[service.ServiceName]*PartitionConnection{}
			for otherPartitionId, servicesInOtherPartition := range topology.partitionServices {
				if partitionId == otherPartitionId {
					// Two services in the same partition will never block each other
					continue
				}
				connection, err := topology.getPartitionConnectionUnlocked(partitionId, otherPartitionId)
				if err != nil {
					return nil, stacktrace.NewError("Couldn't get connection between partitions '%v' and '%v'", partitionId, otherPartitionId)
				}
				for otherServiceId := range servicesInOtherPartition {
					partitionConnectionConfigBetweenServices[otherServiceId] = &connection
				}
			}
			result[serviceName] = partitionConnectionConfigBetweenServices
		}
	}
	return result, nil
}

// ================================================================================================
//
//	Private Helper Methods
//
// ================================================================================================
func (topology *PartitionTopology) getPartitionConnectionUnlocked(
	a service_network_types.PartitionID,
	b service_network_types.PartitionID) (PartitionConnection, error) {
	if _, found := topology.partitionServices[a]; !found {
		return ConnectionAllowed, stacktrace.NewError("Unrecognized partition '%v'", a)
	}
	if _, found := topology.partitionServices[b]; !found {
		return ConnectionAllowed, stacktrace.NewError("Unrecognized partition '%v'", b)
	}
	connectionId := service_network_types.NewPartitionConnectionID(a, b)
	connection, found := topology.partitionConnectionOverrides[*connectionId]
	if !found {
		return topology.defaultConnection, nil
	}
	return connection, nil
}

func serviceIdSetToCommaStr(serviceSet map[service.ServiceName]bool) string {
	strSlice := []string{}
	for serviceId := range serviceSet {
		strSlice = append(strSlice, string(serviceId))
	}
	return strings.Join(strSlice, ", ")
}

func doesServiceSetContainsElement(serviceSet map[service.ServiceName]bool, element service.ServiceName) bool {
	if _, found := serviceSet[element]; found {
		return true
	}
	return false
}

func copyServiceSet(serviceSet map[service.ServiceName]bool) map[service.ServiceName]bool {
	newServiceSet := map[service.ServiceName]bool{}
	for serviceUuid := range serviceSet {
		newServiceSet[serviceUuid] = true
	}
	return newServiceSet
}
