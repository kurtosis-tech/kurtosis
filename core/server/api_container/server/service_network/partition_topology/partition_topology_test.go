/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package partition_topology

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"strconv"
	"testing"
)

var connectionWithSomeConstantLatency = NewPartitionConnection(ConnectionWithNoPacketLoss, NewUniformPacketDelayDistribution(500))
var connectionWithSoftPacketLoss = NewPacketLoss(50)

const (
	partition1 service_network_types.PartitionID = "partition1"
	partition2 service_network_types.PartitionID = "partition2"
	partition3 service_network_types.PartitionID = "partition3"

	service1 service.ServiceName = "service1"
	service2 service.ServiceName = "service2"
	service3 service.ServiceName = "service3"

	// How many nodes in a "huge" network, for benchmarking
	hugeNetworkNodeCount = 10000
)

var allTestServiceNames = map[service.ServiceName]bool{
	service1: true,
	service2: true,
	service3: true,
}

var emptyServiceSet = map[service.ServiceName]bool{}

var serviceSetWithService1 = map[service.ServiceName]bool{
	service1: true,
}
var serviceSetWithService2 = map[service.ServiceName]bool{
	service2: true,
}
var serviceSetWithService3 = map[service.ServiceName]bool{
	service3: true,
}
var serviceSetWithService1And2 = map[service.ServiceName]bool{
	service1: true,
	service2: true,
}
var serviceSetWithService2And3 = map[service.ServiceName]bool{
	service2: true,
	service3: true,
}

// ===========================================================================================
//
//	Benchmarks (execute with `go test -run=^$ -bench=.`)
//
// ===========================================================================================
func BenchmarkHugeNetworkSinglePartitionGetServicePacketConnectionConfigurationsByServiceID(b *testing.B) {
	topology, closerFunc := getHugeTestTopology(b, "service-", ConnectionBlocked)
	defer closerFunc()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := topology.GetServicePartitionConnectionConfigByServiceName(); err != nil {
			b.Fatal(stacktrace.Propagate(err, "An error occurred getting the sevice packet loss configuration by service ID"))
		}
	}

}

// 10k nodes, each in their own partition, partitioned into a line so each partition can only see the ones next to it
func BenchmarkHugeNetworkPathologicalRepartition(b *testing.B) {
	serviceNamePrefix := "service-"
	partitionIdPrefix := "partition-"
	topology, closerFunc := getHugeTestTopology(b, serviceNamePrefix, ConnectionBlocked)
	defer closerFunc()

	newPartitionServices := map[service_network_types.PartitionID]map[service.ServiceName]bool{}
	newPartitionConnections := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	for i := 0; i < hugeNetworkNodeCount; i++ {
		partitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i))
		serviceName := service.ServiceName(serviceNamePrefix + strconv.Itoa(i))
		serviceNameSet := map[service.ServiceName]bool{
			serviceName: true,
		}
		newPartitionServices[partitionId] = serviceNameSet

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

func BenchmarkHugeNetworkPathologicalPartitioningGetServicePacketConnectionConfigurationsByServiceID(b *testing.B) {
	serviceNamePrefix := "service-"
	partitionIdPrefix := "partition-"
	topology, closerFunc := getHugeTestTopology(b, serviceNamePrefix, ConnectionBlocked)
	defer closerFunc()

	newPartitionServices := map[service_network_types.PartitionID]map[service.ServiceName]bool{}
	newPartitionConnections := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	for i := 0; i < hugeNetworkNodeCount; i++ {
		partitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i))
		serviceName := service.ServiceName(serviceNamePrefix + strconv.Itoa(i))
		serviceNameSet := map[service.ServiceName]bool{
			serviceName: true,
		}
		newPartitionServices[partitionId] = serviceNameSet

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
		if _, err := topology.GetServicePartitionConnectionConfigByServiceName(); err != nil {
			b.Fatal(stacktrace.Propagate(err, "An error occurred getting the packet loss configuration map"))
		}
	}
}

// ===========================================================================================
//
//	Repartition tests
//
// ===========================================================================================
func TestAllServicesAreAlwaysInServicePacketConnectionConfigMap(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()
	servicePacketConnectionConfigMapBeforeRepartition := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)
	require.Equal(t, len(servicePacketConnectionConfigMapBeforeRepartition), len(allTestServiceNames), "Service packet loss config map before repartition should contain all services")

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	servicePacketConnectionConfigMapAfterRepartition := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)
	require.Equal(t, len(servicePacketConnectionConfigMapAfterRepartition), len(allTestServiceNames), "Service packet loss config map after repartition should contain all services")
}

func TestServicesInSamePartitionAreNeverBlocked(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()
	servicePacketConnectionConfigMapBeforeRepartition := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)
	for _, blockedServices := range servicePacketConnectionConfigMapBeforeRepartition {
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

	servicePacketConnectionConfigMapAfterRepartition := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)

	service1Blocks := getServicePacketConnectionConfigForService(t, service1, servicePacketConnectionConfigMapAfterRepartition)
	require.Equal(t, len(service1Blocks), 1)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service1Blocks[service3].GetPacketLossPercentage())

	service2Blocks := getServicePacketConnectionConfigForService(t, service2, servicePacketConnectionConfigMapAfterRepartition)
	require.Equal(t, len(service2Blocks), 1)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service2Blocks[service3].GetPacketLossPercentage())

	service3Blocks := getServicePacketConnectionConfigForService(t, service3, servicePacketConnectionConfigMapAfterRepartition)
	require.Equal(t, len(service3Blocks), 2)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service3Blocks[service1].GetPacketLossPercentage())
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service3Blocks[service2].GetPacketLossPercentage())

}

func TestDefaultConnectionSettingsWork(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()
	// Default connection is blocked
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	expectedAmountOfServicesWithPacketConnectionConfig := 2

	servicePacketConnectionConfigMapWithClosedDefaultConn := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)
	for serviceId, otherServicesPacketConnectionConfig := range servicePacketConnectionConfigMapWithClosedDefaultConn {
		_, foundItself := otherServicesPacketConnectionConfig[serviceId]
		require.False(t, foundItself, "A service should never block itself")
		require.Equal(t, expectedAmountOfServicesWithPacketConnectionConfig, len(otherServicesPacketConnectionConfig), "Expected to have 2 other services with packet loss configurations for this service")
		for _, PacketConnectionPercentage := range otherServicesPacketConnectionConfig {
			require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), PacketConnectionPercentage.GetPacketLossPercentage(), "Expected packet loss percentage value for other service were 100%")
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

	servicePacketConnectionConfigMapWithOpenDefaultConn := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)

	for _, otherServicesPacketConnectionConfig := range servicePacketConnectionConfigMapWithOpenDefaultConn {
		require.Equal(t, expectedAmountOfServicesWithPacketConnectionConfig, len(otherServicesPacketConnectionConfig), "All connections should be open now that the default connection is unblocked")
		for _, PacketConnectionPercentage := range otherServicesPacketConnectionConfig {
			require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), PacketConnectionPercentage.GetPacketLossPercentage(), "Expected packet loss percentage value for other service were 0%")
		}
	}

	// Open default connection with latency
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		connectionWithSomeConstantLatency)

	servicePacketConnectionConfigMapWithOpenDefaultConnWithLatency := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)

	for _, otherServicesPacketConnectionConfig := range servicePacketConnectionConfigMapWithOpenDefaultConnWithLatency {
		require.Equal(t, expectedAmountOfServicesWithPacketConnectionConfig, len(otherServicesPacketConnectionConfig), "All connections should be open now that the default connection is unblocked")
		for _, PacketConnectionPercentage := range otherServicesPacketConnectionConfig {
			require.Equal(t, connectionWithSomeConstantLatency.GetPacketLossPercentage(), PacketConnectionPercentage.GetPacketLossPercentage(), "Expected packet loss percentage value for other service were 0%")
			require.Equal(t, connectionWithSomeConstantLatency.GetPacketDelay(), PacketConnectionPercentage.GetPacketDelay(), "Expected packet delay for other service to be non zero")
		}
	}
}

func TestExplicitConnectionBlocksWork(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			// Partition 2 can access 1 with no latency
			*service_network_types.NewPartitionConnectionID(partition1, partition2): ConnectionAllowed,
			// Partition 2 can access 3 with latency
			*service_network_types.NewPartitionConnectionID(partition2, partition3): connectionWithSomeConstantLatency,
			// Access between 1 and 3 is blocked
			*service_network_types.NewPartitionConnectionID(partition1, partition3): ConnectionBlocked,
		},
		ConnectionBlocked)

	servicePacketConnectionConfigurationsByServiceIDMap := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)

	service1andOtherServicesPacketConnectionConfig := getServicePacketConnectionConfigForService(t, service1, servicePacketConnectionConfigurationsByServiceIDMap)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service1andOtherServicesPacketConnectionConfig[service3].GetPacketLossPercentage())
	require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), service1andOtherServicesPacketConnectionConfig[service2].GetPacketLossPercentage())

	service2andOtherServicesPacketConnectionConfig := getServicePacketConnectionConfigForService(t, service2, servicePacketConnectionConfigurationsByServiceIDMap)
	require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), service2andOtherServicesPacketConnectionConfig[service1].GetPacketLossPercentage())
	require.Equal(t, ConnectionAllowed.GetPacketLossPercentage(), service2andOtherServicesPacketConnectionConfig[service3].GetPacketLossPercentage())
	require.Equal(t, connectionWithSomeConstantLatency.GetPacketLossPercentage(), service2andOtherServicesPacketConnectionConfig[service3].GetPacketLossPercentage())
	require.Equal(t, connectionWithSomeConstantLatency.GetPacketDelay(), service2andOtherServicesPacketConnectionConfig[service3].GetPacketDelay())

	service3andOtherServicesPacketConnectionConfig := getServicePacketConnectionConfigForService(t, service3, servicePacketConnectionConfigurationsByServiceIDMap)
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service3andOtherServicesPacketConnectionConfig[service1].GetPacketLossPercentage())
	require.Equal(t, connectionWithSomeConstantLatency.GetPacketLossPercentage(), service3andOtherServicesPacketConnectionConfig[service2].GetPacketLossPercentage())
	require.Equal(t, connectionWithSomeConstantLatency.GetPacketDelay(), service3andOtherServicesPacketConnectionConfig[service2].GetPacketDelay())
}

func TestDuplicateServicesError(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceName]bool{
			partition1: serviceSetWithService1,
			partition2: serviceSetWithService1And2, // Should cause error
			partition3: serviceSetWithService3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)
	require.Error(t, err, "Expected an error due to duplicate service Names, but none was thrown")
}

func TestUnknownServicesError(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceName]bool{
			partition1: {
				service1:          true,
				"unknown-service": true,
			}, // Should error
			partition2: serviceSetWithService2,
			partition3: serviceSetWithService3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)
	require.Error(t, err, "Expected an error due to unknown service Names, but none was thrown")
}

func TestNotAllServicesAllocatedError(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceName]bool{
			partition1: emptyServiceSet,
			partition2: serviceSetWithService2,
			partition3: serviceSetWithService3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)
	require.Error(t, err, "Expected an error due to not all services being allocated, but none was thrown")
}

func TestEmptyPartitionsError(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceName]bool{},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionAllowed)
	require.Error(t, err, "Expected an error due to no partitions beign defined, but none was thrown")
}

func TestUnknownPartitionsError(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	firstPartErr := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceName]bool{
			partition1: serviceSetWithService1,
			partition2: serviceSetWithService2And3,
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID("aa-unknown-partition", partition2): ConnectionBlocked,
		},
		ConnectionAllowed)

	require.Error(t, firstPartErr, "Expected an error due to an unknown partition in the first slot, but none was thrown")

	secondPartErr := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceName]bool{
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
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}
	defaultConnection := ConnectionBlocked
	topology, err := NewPartitionTopology(DefaultPartitionId, defaultConnection, enclaveDb)
	require.Nil(t, err)
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

	servicePacketConnectionConfigurationsByServiceIDMap := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)

	expectedAmountOfServicesWithPacketConnectionConfig := 2
	// All services should be blocking each other, even though Service 3 came late to the party
	for serviceId, otherServicesPacketConnectionConfig := range servicePacketConnectionConfigurationsByServiceIDMap {
		_, foundItself := otherServicesPacketConnectionConfig[serviceId]
		require.False(t, foundItself, "A service should never block itself")
		require.Equal(t, expectedAmountOfServicesWithPacketConnectionConfig, len(otherServicesPacketConnectionConfig), "Expected to have 2 other services with packet loss configurations for this service")
	}
}

func TestAddDuplicateServiceError(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()
	err := topology.AddService(service1, DefaultPartitionId)
	require.Error(t, err, "Expected an error when trying to add a service ID that already exists, but none was thrown")
}

func TestAddServiceToNonexistentPartitionError(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}
	defaultConnection := ConnectionBlocked
	topology, err := NewPartitionTopology(DefaultPartitionId, defaultConnection, enclaveDb)
	require.Nil(t, err)
	if err := topology.AddService(service1, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
	}
	if err := topology.AddService(service2, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 2"))
	}

	err = topology.AddService(service3, "nonexistent-partition")
	require.Error(t, err, "Expected an error when trying to add a service to a nonexistent partition, but none was thrown")
}

// ===========================================================================================
//
//	Remove service tests
//
// ===========================================================================================
func TestRegularRemoveServiceFlow(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()
	// Default connection is blocked
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	err := topology.RemoveService(service2)
	require.Nil(t, err)

	servicePacketConnectionConfigurationsByServiceNameMap := getServicePacketConnectionConfigurationsByServiceIDMap(t, topology)

	require.Equal(t, len(servicePacketConnectionConfigurationsByServiceNameMap), 2, "Service packet loss configuration by service id map should only have 2 entries after we removed a service")

	expectedAmountOfServicesWithPacketConnectionConfig := 1

	service1Blocks := getServicePacketConnectionConfigForService(t, service1, servicePacketConnectionConfigurationsByServiceNameMap)
	require.Equal(t, expectedAmountOfServicesWithPacketConnectionConfig, len(service1Blocks), "Network should have only one other node, so service1Blocks should be of size 1")
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service1Blocks[service3].GetPacketLossPercentage(), "Service 1 should be blocking the only other node in the network, Service 3")

	service3Blocks := getServicePacketConnectionConfigForService(t, service3, servicePacketConnectionConfigurationsByServiceNameMap)
	require.Equal(t, expectedAmountOfServicesWithPacketConnectionConfig, len(service3Blocks), "Network should have only one other node, so service3Blocks should be of size 1")
	require.Equal(t, ConnectionBlocked.GetPacketLossPercentage(), service3Blocks[service1].GetPacketLossPercentage(), "Service 3 should be blocking the only other node in the network, Service 1")
}

func TestSetConnection(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()
	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	connectionOverride := NewPartitionConnection(connectionWithSoftPacketLoss, ConnectionWithNoPacketDelay)
	err := topology.SetConnection(partition1, partition2, connectionOverride)
	require.Nil(t, err)

	require.Equal(t, ConnectionBlocked, topology.GetDefaultConnection())

	currentServicePartitions, err := topology.GetServicePartitions()
	require.Nil(t, err)
	require.Equal(t, map[service.ServiceName]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, currentServicePartitions)

	partitionServices, err := topology.GetPartitionServices()
	require.Nil(t, err)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceName]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
	}, partitionServices)

	expectedConnectionOverrides := map[service_network_types.PartitionConnectionID]PartitionConnection{
		*service_network_types.NewPartitionConnectionID(partition1, partition2): connectionOverride,
	}
	require.Equal(t, expectedConnectionOverrides, topology.partitionConnectionOverrides)
}

func TestSetConnection_FailureUnknownPartition(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	connectionOverride := NewPartitionConnection(connectionWithSoftPacketLoss, ConnectionWithNoPacketDelay)
	err := topology.SetConnection(partition1, "unknownPartition", connectionOverride)
	require.Contains(t, err.Error(), "About to set a connection between 'partition1' and 'unknownPartition' but 'unknownPartition' does not exist")
}

func TestUnsetConnection(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	connectionOverride := NewPartitionConnection(ConnectionWithEntirePacketLoss, ConnectionWithNoPacketDelay)

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

	currentServicePartitions, err := topology.GetServicePartitions()
	require.Nil(t, err)
	require.Equal(t, map[service.ServiceName]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, currentServicePartitions)
	partitionServices, err := topology.GetPartitionServices()
	require.Nil(t, err)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceName]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
	}, partitionServices)

	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestUnsetConnection_FailurePartitionNotFound(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	connectionOverride := NewPartitionConnection(ConnectionWithEntirePacketLoss, ConnectionWithNoPacketDelay)

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
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	connectionOverride := NewPartitionConnection(ConnectionWithEntirePacketLoss, ConnectionWithNoPacketDelay)
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
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	connectionOverride := NewPartitionConnection(ConnectionWithEntirePacketLoss, ConnectionWithNoPacketDelay)
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
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	newDefaultConnection := NewPartitionConnection(ConnectionWithEntirePacketLoss, ConnectionWithNoPacketDelay)
	topology.SetDefaultConnection(newDefaultConnection)

	require.Equal(t, newDefaultConnection, topology.GetDefaultConnection())

	currentServicePartitions, err := topology.GetServicePartitions()
	require.Nil(t, err)
	require.Equal(t, map[service.ServiceName]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, currentServicePartitions)
	partitionServices, err := topology.GetPartitionServices()
	require.Nil(t, err)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceName]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
	}, partitionServices)
	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestCreateEmptyPartitionWithDefaultConnection(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

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

	currentServicePartitions, err := topology.GetServicePartitions()
	require.Nil(t, err)
	require.Equal(t, map[service.ServiceName]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, currentServicePartitions)
	partitionServices, err := topology.GetPartitionServices()
	require.Nil(t, err)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceName]bool{
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
	}, partitionServices)

	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestCreateEmptyPartitionWithDefaultConnection_FailurePartitionAlreadyExists(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2,
		serviceSetWithService3,
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	err := topology.CreateEmptyPartitionWithDefaultConnection(partition1)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Partition with ID 'partition1' can't be created empty because it already exists in the topology")
}

func TestRemovePartition(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

	repartition(
		t,
		topology,
		serviceSetWithService1,
		serviceSetWithService2And3,
		map[service.ServiceName]bool{},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		ConnectionBlocked)

	err := topology.RemovePartition(partition3)
	require.Nil(t, err)

	require.Equal(t, ConnectionBlocked, topology.GetDefaultConnection())

	currentServicePartitions, err := topology.GetServicePartitions()
	require.Nil(t, err)
	require.Equal(t, map[service.ServiceName]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition2",
	}, currentServicePartitions)
	partitionServices, err := topology.GetPartitionServices()
	require.Nil(t, err)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceName]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
			"service3": true,
		},
	}, partitionServices)

	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestRemovePartition_NoopDoesNotExist(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

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

	currentServicePartitions, err := topology.GetServicePartitions()
	require.Nil(t, err)
	require.Equal(t, map[service.ServiceName]service_network_types.PartitionID{
		"service1": "partition1",
		"service2": "partition2",
		"service3": "partition3",
	}, currentServicePartitions)
	partitionServices, err := topology.GetPartitionServices()
	require.Nil(t, err)
	require.Equal(t, map[service_network_types.PartitionID]map[service.ServiceName]bool{
		"partition1": {
			"service1": true,
		},
		"partition2": {
			"service2": true,
		},
		"partition3": {
			"service3": true,
		},
	}, partitionServices)

	noConnectionOverride := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	require.Equal(t, noConnectionOverride, topology.partitionConnectionOverrides)
}

func TestRemovePartition_FailureRemovingDefaultDisallowed(t *testing.T) {
	topology, closerFunc := get3NodeTestTopology(t, ConnectionBlocked)
	defer closerFunc()

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
func get3NodeTestTopology(t *testing.T, defaultConnection PartitionConnection) (*PartitionTopology, func()) {
	file, err := os.CreateTemp("/tmp", "*.db")
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	enclaveDb := &enclave_db.EnclaveDB{DB: db}
	topology, err := NewPartitionTopology(DefaultPartitionId, defaultConnection, enclaveDb)
	require.Nil(t, err)
	if err := topology.AddService(service1, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
	}
	if err := topology.AddService(service2, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 2"))
	}
	if err := topology.AddService(service3, DefaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 3"))
	}
	return topology, func() {
		err = db.Close()
		if err != nil {
			logrus.Warn("Tried closing DB but failed")
		}
		err = os.Remove(file.Name())
		if err != nil {
			logrus.Warnf("Tried deleting the db from disk at '%v' but failed", file.Name())
		}
	}
}

// Used for benchmarking
func getHugeTestTopology(t *testing.B, serviceNamePrefix string, defaultConnection PartitionConnection) (*PartitionTopology, func()) {
	file, err := os.CreateTemp("/tmp", "*.db")
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	enclaveDb := &enclave_db.EnclaveDB{DB: db}
	topology, err := NewPartitionTopology(DefaultPartitionId, defaultConnection, enclaveDb)
	require.Nil(t, err)

	for i := 0; i < hugeNetworkNodeCount; i++ {
		serviceName := service.ServiceName(serviceNamePrefix + strconv.Itoa(i))
		if err := topology.AddService(serviceName, DefaultPartitionId); err != nil {
			t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
		}
	}
	return topology, func() {
		err = db.Close()
		if err != nil {
			logrus.Warn("Tried closing DB but failed")
		}
		err = os.Remove(file.Name())
		if err != nil {
			logrus.Warnf("Tried deleting the db from disk at '%v' but failed", file.Name())
		}
	}
}

func repartition(
	t *testing.T,
	topology *PartitionTopology,
	partition1Services map[service.ServiceName]bool,
	partition2Services map[service.ServiceName]bool,
	partition3Services map[service.ServiceName]bool,
	connections map[service_network_types.PartitionConnectionID]PartitionConnection,
	defaultConnection PartitionConnection) {
	if err := topology.Repartition(
		map[service_network_types.PartitionID]map[service.ServiceName]bool{
			partition1: partition1Services,
			partition2: partition2Services,
			partition3: partition3Services,
		},
		connections,
		defaultConnection); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred repartitioning the network"))
	}
}

func getServicePacketConnectionConfigurationsByServiceIDMap(t *testing.T, topology *PartitionTopology) map[service.ServiceName]map[service.ServiceName]*PartitionConnection {
	servicePacketConnectionConfigurationByServiceID, err := topology.GetServicePartitionConnectionConfigByServiceName()
	if err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred getting service packet loss configuration by service id"))
	}
	return servicePacketConnectionConfigurationByServiceID
}

func getServicePacketConnectionConfigForService(
	t *testing.T,
	serviceName service.ServiceName,
	servicePacketConnectionConfigMap map[service.ServiceName]map[service.ServiceName]*PartitionConnection,
) map[service.ServiceName]*PartitionConnection {
	result, found := servicePacketConnectionConfigMap[serviceName]
	if !found {
		t.Fatal(stacktrace.NewError("Expected to find service '%v' in service packet loss config map but didn't", serviceName))
	}
	return result
}
