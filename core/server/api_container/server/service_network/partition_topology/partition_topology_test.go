/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package partition_topology

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

const (
	partition1 service_network_types.PartitionID = "partition1"
	partition2 service_network_types.PartitionID = "partition2"
	partition3 service_network_types.PartitionID = "partition3"

	service1 service.ServiceID = "service1"
	service2 service.ServiceID = "service2"
	service3 service.ServiceID = "service3"

	// How many nodes in a "huge" network, for benchmarking
	hugeNetworkNodeCount = 10000
)

var allTestServiceIds = map[service.ServiceID]bool{
	service1: true,
	service2: true,
	service3: true,
}

var emptyServiceSet = map[service.ServiceID]bool{}

var serviceSetWithService1 = map[service.ServiceID]bool{
	service1: true,
}
var serviceSetWithService2 = map[service.ServiceID]bool{
	service2: true,
}
var serviceSetWithService3 = map[service.ServiceID]bool{
	service3: true,
}
var serviceSetWithService1And2 = map[service.ServiceID]bool{
	service1: true,
	service2: true,
}
var serviceSetWithService2And3 = map[service.ServiceID]bool{
	service2: true,
	service3: true,
}

// ===========================================================================================
//
//	Benchmarks (execute with `go test -run=^$ -bench=.`)
//
// ===========================================================================================
func BenchmarkHugeNetworkSinglePartitionGetServicePacketLossConfigurationsByServiceID(b *testing.B) {
	topology := getHugeTestTopology(b, "service-", ConnectionBlocked)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := topology.GetServicePacketLossConfigurationsByServiceID(); err != nil {
			b.Fatal(stacktrace.Propagate(err, "An error occurred getting the sevice packet loss configuration by service ID"))
		}
	}

}

// 10k nodes, each in their own partition, partitioned into a line so each partition can only see the ones next to it
func BenchmarkHugeNetworkPathologicalRepartition(b *testing.B) {
	serviceIdPrefix := "service-"
	partitionIdPrefix := "partition-"
	topology := getHugeTestTopology(b, serviceIdPrefix, ConnectionBlocked)

	newPartitionServices := map[service_network_types.PartitionID]map[service.ServiceID]bool{}
	newPartitionConnections := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	for i := 0; i < hugeNetworkNodeCount; i++ {
		partitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i))
		serviceId := service.ServiceID(serviceIdPrefix + strconv.Itoa(i))
		serviceIdSet := map[service.ServiceID]bool{
			serviceId: true,
		}
		newPartitionServices[partitionId] = serviceIdSet

		if i > 0 {
			previousPartitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i-1))
			partConnId := *service_network_types.NewPartitionConnectionID(partitionId, previousPartitionId)
			newPartitionConnections[partConnId] = ConnectionAllowed
		}
		if i < (hugeNetworkNodeCount - 1) {
			nextPartitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i+1))
			partConnId := *service_network_types.NewPartitionConnectionID(partitionId, nextPartitionId)
			newPartitionConnections[partConnId] = ConnectionAllowed
		}
	}
	defaultBlockedConnection := ConnectionBlocked

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := topology.Repartition(newPartitionServices, newPartitionConnections, defaultBlockedConnection); err != nil {
			b.Fatal(stacktrace.Propagate(err, "An error occurred repartitioning the network"))
		}
	}
}

func BenchmarkHugeNetworkPathologicalPartitioningGetServicePacketLossConfigurationsByServiceID(b *testing.B) {
	serviceIdPrefix := "service-"
	partitionIdPrefix := "partition-"
	topology := getHugeTestTopology(b, serviceIdPrefix, ConnectionBlocked)

	newPartitionServices := map[service_network_types.PartitionID]map[service.ServiceID]bool{}
	newPartitionConnections := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	for i := 0; i < hugeNetworkNodeCount; i++ {
		partitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i))
		serviceId := service.ServiceID(serviceIdPrefix + strconv.Itoa(i))
		serviceIdSet := map[service.ServiceID]bool{
			serviceId: true,
		}
		newPartitionServices[partitionId] = serviceIdSet

		if i > 0 {
			previousPartitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i-1))
			partConnId := *service_network_types.NewPartitionConnectionID(partitionId, previousPartitionId)
			newPartitionConnections[partConnId] = ConnectionAllowed
		}
		if i < (hugeNetworkNodeCount - 1) {
			nextPartitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i+1))
			partConnId := *service_network_types.NewPartitionConnectionID(partitionId, nextPartitionId)
			newPartitionConnections[partConnId] = ConnectionAllowed
		}
	}
	defaultBlockedConnection := ConnectionBlocked

	if err := topology.Repartition(newPartitionServices, newPartitionConnections, defaultBlockedConnection); err != nil {
		b.Fatal(stacktrace.Propagate(err, "An error occurred repartitioning the network"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := topology.GetServicePacketLossConfigurationsByServiceID(); err != nil {
			b.Fatal(stacktrace.Propagate(err, "An error occurred getting the packet loss configuration map"))
		}
	}
}

// ===========================================================================================
//
//	Repartition tests
//
// ===========================================================================================
func TestAllServicesAreAlwaysInServicePacketLossConfigMap(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)
	servicePacketLossConfigMapBeforeRepartition := getServicePacketLossConfigurationsByServiceIDMap(t, topology)
	require.Equal(t, len(servicePacketLossConfigMapBeforeRepartition), len(allTestServiceIds), "Service packet loss config map before repartition should contain all services")

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	servicePacketLossConfigMapAfterRepartition := getServicePacketLossConfigurationsByServiceIDMap(t, topology)
	require.Equal(t, len(servicePacketLossConfigMapAfterRepartition), len(allTestServiceIds), "Service packet loss config map after repartition should contain all services")
}

func TestServicesInSamePartitionAreNeverBlocked(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)
	servicePacketLossConfigMapBeforeRepartition := getServicePacketLossConfigurationsByServiceIDMap(t, topology)
	for _, blockedServices := range servicePacketLossConfigMapBeforeRepartition {
		require.Equal(t, len(blockedServices), 0, "No services should be blocked when all services are in the same partition")
	}

	repartition(
		t,
		topology,
		serviceSetWithService1And2,
		emptyServiceSet,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	servicePacketLossConfigMapAfterRepartition := getServicePacketLossConfigurationsByServiceIDMap(t, topology)

	service1Blocks := getServicePacketLossConfigForService(t, service1, servicePacketLossConfigMapAfterRepartition)
	require.Equal(t, len(service1Blocks), 1)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service1Blocks[service3])

	service2Blocks := getServicePacketLossConfigForService(t, service2, servicePacketLossConfigMapAfterRepartition)
	require.Equal(t, len(service2Blocks), 1)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service2Blocks[service3])

	service3Blocks := getServicePacketLossConfigForService(t, service3, servicePacketLossConfigMapAfterRepartition)
	require.Equal(t, len(service3Blocks), 2)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service3Blocks[service1])
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service3Blocks[service2])

}

func TestDefaultConnectionSettingsWork(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	// Default connection is blocked
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	expectedAmountOfServicesWithPacketLossConfig := 2

	servicePacketLossConfigMapWithClosedDefaultConn := getServicePacketLossConfigurationsByServiceIDMap(t, topology)
	for serviceId, otherServicesPacketLossConfig := range servicePacketLossConfigMapWithClosedDefaultConn {
		_, foundItself := otherServicesPacketLossConfig[serviceId]
		require.False(t, foundItself, "A service should never block itself")
		require.Equal(t, expectedAmountOfServicesWithPacketLossConfig, len(otherServicesPacketLossConfig), "Expected to have 2 other services with packet loss configurations for this service")
		for _, packetLossPercentage := range otherServicesPacketLossConfig {
			require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), packetLossPercentage, "Expected packet loss percentage value for other service were 100%")
		}
	}

	// Open default connection back up
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)

	servicePacketLossConfigMapWithOpenDefaultConn := getServicePacketLossConfigurationsByServiceIDMap(t, topology)

	for _, otherServicesPacketLossConfig := range servicePacketLossConfigMapWithOpenDefaultConn {
		require.Equal(t, expectedAmountOfServicesWithPacketLossConfig, len(otherServicesPacketLossConfig), "All connections should be open now that the default connection is unblocked")
		for _, packetLossPercentage := range otherServicesPacketLossConfig {
			require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), packetLossPercentage, "Expected packet loss percentage value for other service were 0%")
		}
	}
}

func TestExplicitConnectionBlocksWork(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			// Partition 2 can access 1 and 3

			*service_network_types.NewPartitionConnectionID(partition1, partition2): ConnectionAllowed,
			*service_network_types.NewPartitionConnectionID(partition2, partition3): ConnectionAllowed,
			// Access between 1 and 3 is blocked
			*service_network_types.NewPartitionConnectionID(partition1, partition3): ConnectionBlocked,
		},
		ConnectionBlocked)

	servicePacketLossConfigurationsByServiceIDMap := getServicePacketLossConfigurationsByServiceIDMap(t, topology)

	service1OtherServicesPacketLossConfig := getServicePacketLossConfigForService(t, service1, servicePacketLossConfigurationsByServiceIDMap)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service1OtherServicesPacketLossConfig[service3])
	require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), service1OtherServicesPacketLossConfig[service2])

	service2OtherServicesPacketLossConfig := getServicePacketLossConfigForService(t, service2, servicePacketLossConfigurationsByServiceIDMap)
	require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), service2OtherServicesPacketLossConfig[service1])
	require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), service2OtherServicesPacketLossConfig[service3])

	service3OtherServicesPacketLossConfig := getServicePacketLossConfigForService(t, service3, servicePacketLossConfigurationsByServiceIDMap)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service3OtherServicesPacketLossConfig[service1])
	require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), service3OtherServicesPacketLossConfig[service2])
}

func TestDuplicateServicesError(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionAllowed)

	err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceID]bool{
			partition1: serviceSetWithService1,
			partition2: serviceSetWithService1And2, // Should cause error
			partition3: serviceSetWithService3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)
	require.Error(t, err, "Expected an error due to duplicate service IDs, but none was thrown")
}

func TestUnknownServicesError(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionAllowed)

	err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceID]bool{
			partition1: {
				service1:          true,
				"unknown-service": true,
			}, // Should error
			partition2: serviceSetWithService2,
			partition3: serviceSetWithService3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)
	require.Error(t, err, "Expected an error due to unknown service IDs, but none was thrown")
}

func TestNotAllServicesAllocatedError(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionAllowed)

	err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceID]bool{
			partition1: emptyServiceSet,
			partition2: serviceSetWithService2,
			partition3: serviceSetWithService3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)
	require.Error(t, err, "Expected an error due to not all services being allocated, but none was thrown")
}

func TestEmptyPartitionsError(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionAllowed)

	err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceID]bool{},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)
	require.Error(t, err, "Expected an error due to no partitions beign defined, but none was thrown")
}

func TestUnknownPartitionsError(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionAllowed)

	firstPartErr := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceID]bool{
			partition1: serviceSetWithService1,
			partition2: serviceSetWithService2And3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID("aa-unknown-partition", partition2): ConnectionBlocked,
		},
		ConnectionAllowed)

	require.Error(t, firstPartErr, "Expected an error due to an unknown partition in the first slot, but none was thrown")

	secondPartErr := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceID]bool{
			partition1: serviceSetWithService1,
			partition2: serviceSetWithService2And3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID("zz-unknown-partition", partition2): ConnectionBlocked,
		},
		ConnectionBlocked)

	require.Error(t, secondPartErr, "Expected an error due to an unknown partition in the second slot, but none was thrown")
}

// ===========================================================================================
//
//	Add service tests
//
// ===========================================================================================
func TestRegularAddServiceFlow(t *testing.T) {
	defaultConnection := ConnectionBlocked
	topology := NewPartitionTopology(DefaultPartitionId, defaultConnection)
	if err := topology.AddService(service1, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
	}
	if err := topology.AddService(service2, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 2"))
	}

	// Default connection is blocked
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		emptyServiceSet,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	if err := topology.AddService(service3, partition3); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 3 to the network"))
	}

	servicePacketLossConfigurationsByServiceIDMap := getServicePacketLossConfigurationsByServiceIDMap(t, topology)

	expectedAmountOfServicesWithPacketLossConfig := 2
	// All services should be blocking each other, even though Service 3 came late to the party
	for serviceId, otherServicesPacketLossConfig := range servicePacketLossConfigurationsByServiceIDMap {
		_, foundItself := otherServicesPacketLossConfig[serviceId]
		require.False(t, foundItself, "A service should never block itself")
		require.Equal(t, expectedAmountOfServicesWithPacketLossConfig, len(otherServicesPacketLossConfig), "Expected to have 2 other services with packet loss configurations for this service")
	}
}

func TestAddDuplicateServiceError(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	err := topology.AddService(service1, DefaultPartitionId)
	require.Error(t, err, "Expected an error when trying to add a service ID that already exists, but none was thrown")
}

func TestAddServiceToNonexistentPartitionError(t *testing.T) {
	defaultConnection := ConnectionBlocked
	topology := NewPartitionTopology(DefaultPartitionId, defaultConnection)
	if err := topology.AddService(service1, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
	}
	if err := topology.AddService(service2, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 2"))
	}

	err := topology.AddService(service3, "nonexistent-partition")
	require.Error(t, err, "Expected an error when trying to add a service to a nonexistent partition, but none was thrown")
}

// ===========================================================================================
//
//	Remove service tests
//
// ===========================================================================================
func TestRegularRemoveServiceFlow(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	// Default connection is blocked
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	topology.RemoveService(service2)

	servicePacketLossConfigurationsByServiceIDMap := getServicePacketLossConfigurationsByServiceIDMap(t, topology)

	require.Equal(t, len(servicePacketLossConfigurationsByServiceIDMap), 2, "Service paccket los configuration by service id map should only have 2 entries after we removed a service")

	expectedAmountOfServicesWithPacketLossConfig := 1

	service1Blocks := getServicePacketLossConfigForService(t, service1, servicePacketLossConfigurationsByServiceIDMap)
	require.Equal(t, expectedAmountOfServicesWithPacketLossConfig, len(service1Blocks), "Network should have only one other node, so service1Blocks should be of size 1")
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service1Blocks[service3], "Service 1 should be blocking the only other node in the network, Service 3")

	service3Blocks := getServicePacketLossConfigForService(t, service3, servicePacketLossConfigurationsByServiceIDMap)
	require.Equal(t, expectedAmountOfServicesWithPacketLossConfig, len(service3Blocks), "Network should have only one other node, so service3Blocks should be of size 1")
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service3Blocks[service1], "Service 3 should be blocking the only other node in the network, Service 1")
}

func TestSetConnection(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	connectionOverride := NewPartitionConnection(55)
	err := topology.SetConnection(partition1, partition2, connectionOverride)
	require.Nil(t, err)

	require.Equal(t, ConnectionBlocked, topology.GetDefaultConnection())

	require.Equal(t, map[service.ServiceID]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, topology.servicePartitions)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceID]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
	}, topology.partitionServices)

	expectedConnectionOverrides := map[service_network_types.PartitionConnectionID]PartitionConnection{
		*service_network_types.NewPartitionConnectionID(partition1, partition2): connectionOverride,
	}
	require.Equal(t, expectedConnectionOverrides, topology.partitionConnectionOverrides)
}

func TestSetConnection_FailureUnknownPartition(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	connectionOverride := NewPartitionConnection(55)
	err := topology.SetConnection(partition1, "unknownPartition", connectionOverride)
	require.Contains(t, err.Error(), "About to set a connection between 'partition1' and 'unknownPartition' but 'unknownPartition' does not exist")
}

func TestUnsetConnection(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)
	connectionOverride := NewPartitionConnection(55)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID(partition1, partition2): connectionOverride,
		},
		ConnectionBlocked)

	err := topology.UnsetConnection(partition1, partition2)
	require.Nil(t, err)

	require.Equal(t, ConnectionBlocked, topology.GetDefaultConnection())

	require.Equal(t, map[service.ServiceID]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, topology.servicePartitions)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceID]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
	}, topology.partitionServices)

	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestUnsetConnection_FailurePartitionNotFound(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)
	connectionOverride := NewPartitionConnection(55)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID(partition1, partition2): connectionOverride,
		},
		ConnectionBlocked)

	err := topology.UnsetConnection(partition1, "unknownPartition")
	require.Contains(t, err.Error(), "About to unset a connection between 'partition1' and 'unknownPartition' but 'unknownPartition' does not exist")
}

func TestGetConnection(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	connectionOverride := NewPartitionConnection(50)
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID(partition1, partition2): connectionOverride,
		},
		ConnectionBlocked)

	isDefault, partition1Partition2Connection, err := topology.GetPartitionConnection(partition1, partition2)
	require.Nil(t, err)
	require.False(t, isDefault)
	require.Equal(t, connectionOverride, partition1Partition2Connection)

	isDefault, partition2Partition3Connection, err := topology.GetPartitionConnection(partition2, partition3)
	require.Nil(t, err)
	require.True(t, isDefault)
	require.Equal(t, ConnectionBlocked, partition2Partition3Connection)

	isDefault, partition1Partition3Connection, err := topology.GetPartitionConnection(partition1, partition3)
	require.Nil(t, err)
	require.True(t, isDefault)
	require.Equal(t, ConnectionBlocked, partition1Partition3Connection)
}

func TestGetConnection_FailurePartitionNotFound(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	connectionOverride := NewPartitionConnection(50)
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID(partition1, partition2): connectionOverride,
		},
		ConnectionBlocked)

	isDefault, partition1Partition2Connection, err := topology.GetPartitionConnection(partition1, "unknownPartition")
	require.Contains(t, err.Error(), "About to get a connection between 'partition1' and 'unknownPartition' but 'unknownPartition' does not exist")
	require.False(t, isDefault)
	require.Equal(t, ConnectionAllowed, partition1Partition2Connection)
}

func TestSetDefaultConnection(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	newDefaultConnection := NewPartitionConnection(50)
	topology.SetDefaultConnection(newDefaultConnection)

	require.Equal(t, newDefaultConnection, topology.GetDefaultConnection())

	require.Equal(t, map[service.ServiceID]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, topology.servicePartitions)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceID]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
	}, topology.partitionServices)
	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestCreateEmptyPartitionWithDefaultConnection(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	partition4 := service_network_types.PartitionID("partition4")
	err := topology.CreateEmptyPartitionWithDefaultConnection(partition4)
	require.Nil(t, err)

	require.Equal(t, ConnectionBlocked, topology.GetDefaultConnection())

	require.Equal(t, map[service.ServiceID]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, topology.servicePartitions)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceID]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
		"partition4": {},
	}, topology.partitionServices)

	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestCreateEmptyPartitionWithDefaultConnection_FailurePartitionAlreadyExists(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	err := topology.CreateEmptyPartitionWithDefaultConnection(partition1)
	require.Contains(t, err.Error(), "Partition with ID 'partition1' can't be created empty because it already exists in the topology")
}

func TestRemovePartition(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2And3,
		map[service.ServiceID]bool{},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	err := topology.RemovePartition(partition3)
	require.Nil(t, err)

	require.Equal(t, ConnectionBlocked, topology.GetDefaultConnection())

	require.Equal(t, map[service.ServiceID]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition2",
	}, topology.servicePartitions)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceID]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
			"service3": true,
		},
	}, topology.partitionServices)

	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestRemovePartition_NoopDoesNotExist(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	err := topology.RemovePartition("unknown-partition")
	require.Nil(t, err)

	require.Equal(t, ConnectionBlocked, topology.GetDefaultConnection())

	require.Equal(t, map[service.ServiceID]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, topology.servicePartitions)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceID]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
	}, topology.partitionServices)

	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestRemovePartition_FailureRemovingDefaultDisallowed(t *testing.T) {
	topology := get3NodeTestTopology(t, ConnectionBlocked)

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	err := topology.RemovePartition(DefaultPartitionId)
	require.Contains(t, err.Error(), "Default partition cannot be removed")
}

// ===========================================================================================
//
//	Private helper methods
//
// ===========================================================================================
func get3NodeTestTopology(t *testing.T, defaultConnection PartitionConnection) *PartitionTopology {
	topology := NewPartitionTopology(DefaultPartitionId, defaultConnection)
	if err := topology.AddService(service1, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
	}
	if err := topology.AddService(service2, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 2"))
	}
	if err := topology.AddService(service3, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 3"))
	}
	return topology
}

// Used for benchmarking
func getHugeTestTopology(t *testing.B, serviceIdPrefix string, defaultConnection PartitionConnection) *PartitionTopology {
	topology := NewPartitionTopology(DefaultPartitionId, defaultConnection)

	for i := 0; i < hugeNetworkNodeCount; i++ {
		serviceId := service.ServiceID(serviceIdPrefix + strconv.Itoa(i))
		if err := topology.AddService(serviceId, DefaultPartitionId); err != nil {
			t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
		}
	}
	return topology
}

func repartition(
	t *testing.T,
	topology *PartitionTopology,
	partition1Services map[service.ServiceID]bool,
	partition2Services map[service.ServiceID]bool,
	partition3Services map[service.ServiceID]bool,
	connections map[service_network_types.PartitionConnectionID]PartitionConnection,
	defaultConnection PartitionConnection) {
	if err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceID]bool{
			partition1: partition1Services,
			partition2: partition2Services,
			partition3: partition3Services,
		},
		connections,
		defaultConnection); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred repartitioning the network"))
	}
}

func getServicePacketLossConfigurationsByServiceIDMap(t *testing.T, topology *PartitionTopology) map[service.ServiceID]map[service.ServiceID]float32 {
	servicePacketLossConfigurationByServiceID, err := topology.GetServicePacketLossConfigurationsByServiceID()
	if err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred getting service packet loss configuration by service id"))
	}
	return servicePacketLossConfigurationByServiceID
}

func getServicePacketLossConfigForService(
	t *testing.T,
	serviceId service.ServiceID,
	servicePacketLossConfigMap map[service.ServiceID]map[service.ServiceID]float32,
) map[service.ServiceID]float32 {
	result, found := servicePacketLossConfigMap[serviceId]
	if !found {
		t.Fatal(stacktrace.NewError("Expected to find service '%v' in service packet loss config map but didn't", serviceId))
	}
	return result
}
