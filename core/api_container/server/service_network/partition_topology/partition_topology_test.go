/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package partition_topology

import (
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/service_network_types"
	"github.com/palantir/stacktrace"
	"gotest.tools/assert"
	"strconv"
	"testing"
)

const (
	defaultPartitionId service_network_types.PartitionID = "default"

	partition1 service_network_types.PartitionID = "partition1"
	partition2 service_network_types.PartitionID = "partition2"
	partition3 service_network_types.PartitionID = "partition3"

	service1 service_network_types.ServiceID = "service1"
	service2 service_network_types.ServiceID = "service2"
	service3 service_network_types.ServiceID = "service3"

	// How many nodes in a "huge" network, for benchmarking
	hugeNetworkNodeCount = 10000
)

var allTestServiceIds = map[service_network_types.ServiceID]bool{
	service1: true,
	service2: true,
	service3: true,
}

// ===========================================================================================
//               Benchmarks (execute with `go test -run=^$ -bench=.`)
// ===========================================================================================
func BenchmarkHugeNetworkSinglePartitionGetBlocklists(b *testing.B) {
	topology := getHugeTestTopology(b, "service-", true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := topology.GetBlocklists(); err != nil {
			b.Fatal(stacktrace.Propagate(err, "An error occurred getting the blocklists map"))
		}
	}

}

// 10k nodes, each in their own partition, partitioned into a line so each partition can only see the ones next to it
func BenchmarkHugeNetworkPathologicalRepartition(b *testing.B) {
	serviceIdPrefix := "service-"
	partitionIdPrefix := "partition-"
	topology := getHugeTestTopology(b, serviceIdPrefix, true)

	newPartitionServices := map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{}
	newPartitionConnections := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	for i := 0; i < hugeNetworkNodeCount; i++ {
		partitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i))
		serviceId := service_network_types.ServiceID(serviceIdPrefix + strconv.Itoa(i))
		newPartitionServices[partitionId] = service_network_types.NewServiceIDSet(serviceId)

		if i > 0 {
			previousPartitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i - 1))
			partConnId := *service_network_types.NewPartitionConnectionID(partitionId, previousPartitionId)
			newPartitionConnections[partConnId] = PartitionConnection{IsBlocked: false}
		}
		if i < (hugeNetworkNodeCount - 1) {
			nextPartitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i + 1))
			partConnId := *service_network_types.NewPartitionConnectionID(partitionId, nextPartitionId)
			newPartitionConnections[partConnId] = PartitionConnection{IsBlocked: false}
		}
	}
	defaultBlockedConnection := PartitionConnection{IsBlocked: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := topology.Repartition(newPartitionServices, newPartitionConnections, defaultBlockedConnection); err != nil {
			b.Fatal(stacktrace.Propagate(err, "An error occurred repartitioning the network"))
		}
	}
}

func BenchmarkHugeNetworkPathologicalPartitioningGetBlocklists(b *testing.B) {
	serviceIdPrefix := "service-"
	partitionIdPrefix := "partition-"
	topology := getHugeTestTopology(b, serviceIdPrefix, true)

	newPartitionServices := map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{}
	newPartitionConnections := map[service_network_types.PartitionConnectionID]PartitionConnection{}
	for i := 0; i < hugeNetworkNodeCount; i++ {
		partitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i))
		serviceId := service_network_types.ServiceID(serviceIdPrefix + strconv.Itoa(i))
		newPartitionServices[partitionId] = service_network_types.NewServiceIDSet(serviceId)

		if i > 0 {
			previousPartitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i - 1))
			partConnId := *service_network_types.NewPartitionConnectionID(partitionId, previousPartitionId)
			newPartitionConnections[partConnId] = PartitionConnection{IsBlocked: false}
		}
		if i < (hugeNetworkNodeCount - 1) {
			nextPartitionId := service_network_types.PartitionID(partitionIdPrefix + strconv.Itoa(i + 1))
			partConnId := *service_network_types.NewPartitionConnectionID(partitionId, nextPartitionId)
			newPartitionConnections[partConnId] = PartitionConnection{IsBlocked: false}
		}
	}
	defaultBlockedConnection := PartitionConnection{IsBlocked: true}

	if err := topology.Repartition(newPartitionServices, newPartitionConnections, defaultBlockedConnection); err != nil {
		b.Fatal(stacktrace.Propagate(err, "An error occurred repartitioning the network"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := topology.GetBlocklists(); err != nil {
			b.Fatal(stacktrace.Propagate(err, "An error occurred getting the blocklists map"))
		}
	}
}

// ===========================================================================================
//                                   Repartition tests
// ===========================================================================================
func TestAllServicesAreAlwaysInBlocklist(t *testing.T) {
	topology := get3NodeTestTopology(t, true)
	blocklistsBeforeRepartition := getBlocklistsMap(t, topology)
	assert.Equal(t, len(blocklistsBeforeRepartition), len(allTestServiceIds), "Blocklists map before repartition should contain all services")

	repartition(
		t,
		topology,
		service_network_types.NewServiceIDSet(service1),
		service_network_types.NewServiceIDSet(service2),
		service_network_types.NewServiceIDSet(service3),
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		true)

	blocklistsAfterRepartition := getBlocklistsMap(t, topology)
	assert.Equal(t, len(blocklistsAfterRepartition), len(allTestServiceIds), "Blocklists map after repartition should contain all services")
}

func TestServicesInSamePartitionAreNeverBlocked(t *testing.T) {
	topology := get3NodeTestTopology(t, true)
	blocklistsBeforeRepartition := getBlocklistsMap(t, topology)
	for _, blockedServices := range blocklistsBeforeRepartition {
		assert.Equal(t, blockedServices.Size(), 0, "No services should be blocked when all services are in the same partition")
	}

	repartition(
		t,
		topology,
		service_network_types.NewServiceIDSet(service1, service2),
		service_network_types.NewServiceIDSet(),
		service_network_types.NewServiceIDSet(service3),
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		true)

	blocklistsAfterRepartition := getBlocklistsMap(t, topology)

	service1Blocks := getBlocklistForService(t, service1, blocklistsAfterRepartition)
	assert.Equal(t, service1Blocks.Size(), 1)
	assert.Assert(t, service1Blocks.Contains(service3))

	service2Blocks := getBlocklistForService(t, service2, blocklistsAfterRepartition)
	assert.Equal(t, service2Blocks.Size(), 1)
	assert.Assert(t, service2Blocks.Contains(service3))

	service3Blocks := getBlocklistForService(t, service3, blocklistsAfterRepartition)
	assert.Equal(t, service3Blocks.Size(), 2)
	assert.Assert(t, service3Blocks.Contains(service1))
	assert.Assert(t, service3Blocks.Contains(service2))
}

func TestDefaultConnectionSettingsWork(t *testing.T) {
	topology := get3NodeTestTopology(t, true)

	// Default connection is blocked
	repartition(
		t,
		topology,
		service_network_types.NewServiceIDSet(service1),
		service_network_types.NewServiceIDSet(service2),
		service_network_types.NewServiceIDSet(service3),
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		true)

	blocklistsWithClosedDefaultConn := getBlocklistsMap(t, topology)
	for serviceId, blockedServiceIds := range blocklistsWithClosedDefaultConn {
		assert.Assert(t, !blockedServiceIds.Contains(serviceId), "A service should never block itself")
		assert.Equal(t, blockedServiceIds.Size(), 2, "Expected the other services to be in the service's blocklist")
	}

	// Open default connection back up
	repartition(
		t,
		topology,
		service_network_types.NewServiceIDSet(service1),
		service_network_types.NewServiceIDSet(service2),
		service_network_types.NewServiceIDSet(service3),
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		false)

	blocklistsWithOpenDefaultConn := getBlocklistsMap(t, topology)
	for _, blockedServiceIds := range blocklistsWithOpenDefaultConn {
		assert.Equal(t, blockedServiceIds.Size(), 0, "All connections should be open now that the default connection is unblocked")
	}
}

func TestExplicitConnectionBlocksWork(t *testing.T) {
	topology := get3NodeTestTopology(t, true)

	repartition(
		t,
		topology,
		service_network_types.NewServiceIDSet(service1),
		service_network_types.NewServiceIDSet(service2),
		service_network_types.NewServiceIDSet(service3),
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			// Partition 2 can access 1 and 3
			*service_network_types.NewPartitionConnectionID(partition1, partition2): {
				IsBlocked: false,
			},
			*service_network_types.NewPartitionConnectionID(partition2, partition3): {
				IsBlocked: false,
			},
			// Access between 1 and 3 is blocked
			*service_network_types.NewPartitionConnectionID(partition1, partition3): {
				IsBlocked: true,
			},
		},
		true)

	blocklists := getBlocklistsMap(t, topology)

	service1Blocks := getBlocklistForService(t, service1, blocklists)
	assert.Assert(t, service1Blocks.Size() == 1)
	assert.Assert(t, service1Blocks.Contains(service3))

	service2Blocks := getBlocklistForService(t, service2, blocklists)
	assert.Assert(t, service2Blocks.Size() == 0)

	service3Blocks := getBlocklistForService(t, service3, blocklists)
	assert.Assert(t, service3Blocks.Size() == 1)
	assert.Assert(t, service3Blocks.Contains(service1))
}

func TestDuplicateServicesError(t *testing.T) {
	topology := get3NodeTestTopology(t, false)

	err := topology.Repartition(
		map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{
			partition1: service_network_types.NewServiceIDSet(service1),
			partition2: service_network_types.NewServiceIDSet(service1, service2), // Should cause error
			partition3: service_network_types.NewServiceIDSet(service3),
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		PartitionConnection{IsBlocked: false})
	assert.Assert(t, err != nil, "Expected an error due to duplicate service IDs, but none was thrown")
}

func TestUnknownServicesError(t *testing.T) {
	topology := get3NodeTestTopology(t, false)

	err := topology.Repartition(
		map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{
			partition1: service_network_types.NewServiceIDSet(service1, "unknown-service"), // Should error
			partition2: service_network_types.NewServiceIDSet(service2),
			partition3: service_network_types.NewServiceIDSet(service3),
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		PartitionConnection{IsBlocked: false})
	assert.Assert(t, err != nil, "Expected an error due to unknown service IDs, but none was thrown")
}

func TestNotAllServicesAllocatedError(t *testing.T) {
	topology := get3NodeTestTopology(t, false)

	err := topology.Repartition(
		map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{
			partition1: service_network_types.NewServiceIDSet(),
			partition2: service_network_types.NewServiceIDSet(service2),
			partition3: service_network_types.NewServiceIDSet(service3),
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		PartitionConnection{IsBlocked: false})
	assert.Assert(t, err != nil, "Expected an error due to not all services being allocated, but none was thrown")
}

func TestEmptyPartitionsError(t *testing.T) {
	topology := get3NodeTestTopology(t, false)

	err := topology.Repartition(
		map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{},
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		PartitionConnection{IsBlocked: false})
	assert.Assert(t, err != nil, "Expected an error due to no partitions beign defined, but none was thrown")
}

func TestUnknownPartitionsError(t *testing.T) {
	topology := get3NodeTestTopology(t, false)

	firstPartErr := topology.Repartition(
		map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{
			partition1: service_network_types.NewServiceIDSet(service1),
			partition2: service_network_types.NewServiceIDSet(service2, service3),
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID("aa-unknown-partition", partition2): {
				IsBlocked: true,
			},
		},
		PartitionConnection{IsBlocked: false})
	assert.Assert(t, firstPartErr != nil, "Expected an error due to an unknown partition in the first slot, but none was thrown")

	secondPartErr := topology.Repartition(
		map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{
			partition1: service_network_types.NewServiceIDSet(service1),
			partition2: service_network_types.NewServiceIDSet(service2, service3),
		},
		map[service_network_types.PartitionConnectionID]PartitionConnection{
			*service_network_types.NewPartitionConnectionID("zz-unknown-partition", partition2): {
				IsBlocked: true,
			},
		},
		PartitionConnection{IsBlocked: false})
	assert.Assert(t, secondPartErr != nil, "Expected an error due to an unknown partition in the second slot, but none was thrown")
}

// ===========================================================================================
//                                 Add service tests
// ===========================================================================================
func TestRegularAddServiceFlow(t *testing.T) {
	defaultConnection := PartitionConnection{IsBlocked: true}
	topology := NewPartitionTopology(defaultPartitionId, defaultConnection)
	if err := topology.AddService(service1, defaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
	}
	if err := topology.AddService(service2, defaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 2"))
	}

	// Default connection is blocked
	repartition(
		t,
		topology,
		service_network_types.NewServiceIDSet(service1),
		service_network_types.NewServiceIDSet(service2),
		service_network_types.NewServiceIDSet(),
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		true)

	if err := topology.AddService(service3, partition3); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 3 to the network"))
	}

	blocklists := getBlocklistsMap(t, topology)

	// All services should be blocking each other, even though Service 3 came late to the party
	for serviceId, blockedServiceIds := range blocklists {
		assert.Assert(t, !blockedServiceIds.Contains(serviceId), "A service should never block itself")
		assert.Equal(t, blockedServiceIds.Size(), 2, "Expected the other services to be in the service's blocklist")
	}
}

func TestAddDuplicateServiceError(t *testing.T) {
	topology := get3NodeTestTopology(t, true)

	err := topology.AddService(service1, defaultPartitionId)
	assert.Assert(t, err != nil, "Expected an error when trying to add a service ID that already exists, but none was thrown")
}

func TestAddServiceToNonexistentPartitionError(t *testing.T) {
	defaultConnection := PartitionConnection{IsBlocked: true}
	topology := NewPartitionTopology(defaultPartitionId, defaultConnection)
	if err := topology.AddService(service1, defaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
	}
	if err := topology.AddService(service2, defaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 2"))
	}

	err := topology.AddService(service3, "nonexistent-partition")
	assert.Assert(t, err != nil, "Expected an error when trying to add a service to a nonexistent partition, but none was thrown")
}

// ===========================================================================================
//                                Remove service tests
// ===========================================================================================
func TestRegularRemoveServiceFlow(t *testing.T) {
	topology := get3NodeTestTopology(t, true)

	// Default connection is blocked
	repartition(
		t,
		topology,
		service_network_types.NewServiceIDSet(service1),
		service_network_types.NewServiceIDSet(service2),
		service_network_types.NewServiceIDSet(service3),
		map[service_network_types.PartitionConnectionID]PartitionConnection{},
		true)

	topology.RemoveService(service2)

	blocklists := getBlocklistsMap(t, topology)
	assert.Equal(t, len(blocklists), 2, "Blocklists map should only have 2 entries after we removed a service")

	service1Blocks := getBlocklistForService(t, service1, blocklists)
	assert.Equal(t, service1Blocks.Size(), 1, "Network should have only one other node, so blocklist should be of size 1")
	assert.Assert(t, service1Blocks.Contains(service3), "Service 1 should be blocking the only other node in the network, Service 3")

	service3Blocks := getBlocklistForService(t, service3, blocklists)
	assert.Equal(t, service3Blocks.Size(), 1, "Network should have only one other node, so blocklist should be of size 1")
	assert.Assert(t, service3Blocks.Contains(service1), "Service 3 should be blocking the only other node in the network, Service 1")
}

// ===========================================================================================
//                               Private helper methods
// ===========================================================================================
func get3NodeTestTopology(t *testing.T, isDefaultConnectionBlocked bool) *PartitionTopology {
	defaultConnection := PartitionConnection{IsBlocked: isDefaultConnectionBlocked}
	topology := NewPartitionTopology(defaultPartitionId, defaultConnection)
	if err := topology.AddService(service1, defaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
	}
	if err := topology.AddService(service2, defaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 2"))
	}
	if err := topology.AddService(service3, defaultPartitionId); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 3"))
	}
	return topology
}

// Used for benchmarking
func getHugeTestTopology(t *testing.B, serviceIdPrefx string, isDefaultConnBlocked bool) *PartitionTopology {
	defaultConnection := PartitionConnection{IsBlocked: isDefaultConnBlocked}
	topology := NewPartitionTopology(defaultPartitionId, defaultConnection)

	for i := 0; i < hugeNetworkNodeCount; i++ {
		serviceId := service_network_types.ServiceID(serviceIdPrefx + strconv.Itoa(i))
		if err := topology.AddService(serviceId, defaultPartitionId); err != nil {
			t.Fatal(stacktrace.Propagate(err, "An error occurred adding service 1"))
		}
	}
	return topology
}

func repartition(
		t *testing.T,
		topology *PartitionTopology,
		partition1Services *service_network_types.ServiceIDSet,
		partition2Services *service_network_types.ServiceIDSet,
		partition3Services *service_network_types.ServiceIDSet,
		connections map[service_network_types.PartitionConnectionID]PartitionConnection,
		isDefaultConnBlocked bool) {
	if err := topology.Repartition(
			map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{
				partition1: partition1Services,
				partition2: partition2Services,
				partition3: partition3Services,
			},
			connections,
			PartitionConnection{IsBlocked: isDefaultConnBlocked}); err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred repartitioning the network"))
	}
}

func getBlocklistsMap(t *testing.T, topology *PartitionTopology) map[service_network_types.ServiceID]*service_network_types.ServiceIDSet {
	blocklists, err := topology.GetBlocklists()
	if err != nil {
		t.Fatal(stacktrace.Propagate(err, "An error occurred getting the blocklists map before repartition"))
	}
	return blocklists
}

func getBlocklistForService(
		t *testing.T,
		serviceId service_network_types.ServiceID,
		blocklists map[service_network_types.ServiceID]*service_network_types.ServiceIDSet) *service_network_types.ServiceIDSet {
	result, found := blocklists[serviceId]
	if !found {
		t.Fatal(stacktrace.NewError("Expected to find service '%v' in blocklists map but didn't", serviceId))
	}
	return result
}

func assertBlocklistsEqual(
		t *testing.T,
		expected map[service_network_types.ServiceID]*service_network_types.ServiceIDSet,
		actual map[service_network_types.ServiceID]*service_network_types.ServiceIDSet) {
	assert.Equal(
		t,
		len(expected),
		len(actual),
		"Blocklists expected length %v != actual length %v",
		len(expected),
		len(actual))

	for expectedServiceId, expectedBlockedServiceIds := range expected {
		actualBlockedServiceIds, foundInActual := actual[expectedServiceId]
		assert.Assert(t, foundInActual, "Expected service ID %v not found in actual blocklists map", expectedServiceId)
		assert.Assert(
			t,
			expectedBlockedServiceIds.Equals(actualBlockedServiceIds),
			"For service ID '%v', expected blocked service IDs '%v' don't match actual blocked service IDs '%v'",
			expectedServiceId,
			expectedBlockedServiceIds.Elems(),
			actualBlockedServiceIds.Elems(),
		)
	}
}
