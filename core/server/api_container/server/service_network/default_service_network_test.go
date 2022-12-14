/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	lib_networking_sidecar "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net"
	"strconv"
	"testing"
)

const (
	packetLossConfigForBlockedPartition = float32(100)

	numServices = 10

	enclaveId               = enclave.EnclaveID("test-enclave")
	partitioningEnabled     = true
	partitioningDisabled    = false
	fakeApiContainerVersion = "0.0.0"
	apiContainerPort        = uint16(1234)
	testContainerImageName  = "kurtosistech/test-container"
)

var (
	ip                   = testIpFromInt(0)
	unusedEnclaveDataDir *enclave_data_directory.EnclaveDataDirectory
)

func TestStartService_Successful(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	// One service will be started successfully
	successfulServiceIndex := 1
	successfulServiceId := testServiceIdFromInt(successfulServiceIndex)
	successfulServiceGuid := testServiceGuidFromInt(successfulServiceIndex)
	successfulServiceIp := testIpFromInt(successfulServiceIndex)
	successfulServiceRegistration := service.NewServiceRegistration(successfulServiceId, successfulServiceGuid, enclaveId, successfulServiceIp)
	successfulService := service.NewService(successfulServiceRegistration, container_status.ContainerStatus_Running, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{})
	successfulServiceConfig := services.NewServiceConfigBuilder(testContainerImageName).Build()

	// One service will fail to be started
	failedServiceIndex := 2
	failedServiceId := testServiceIdFromInt(failedServiceIndex)
	failedServiceConfig := services.NewServiceConfigBuilder(testContainerImageName).Build()

	// One service will be successfully started but its sidecar will fail to start
	sidecarFailedServiceIndex := 3
	sidecarFailedServiceId := testServiceIdFromInt(sidecarFailedServiceIndex)
	sidecarFailedServiceGuid := testServiceGuidFromInt(sidecarFailedServiceIndex)
	sidecarFailedServiceIp := testIpFromInt(sidecarFailedServiceIndex)
	sidecarFailedServiceRegistration := service.NewServiceRegistration(sidecarFailedServiceId, sidecarFailedServiceGuid, enclaveId, sidecarFailedServiceIp)
	sidecarFailedService := service.NewService(sidecarFailedServiceRegistration, container_status.ContainerStatus_Running, map[string]*port_spec.PortSpec{}, sidecarFailedServiceIp, map[string]*port_spec.PortSpec{})
	sidecarFailedServiceConfig := services.NewServiceConfigBuilder(testContainerImageName).Build()

	network := NewDefaultServiceNetwork(
		enclaveId,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveId),
	)

	// Configre the mock to also be testing that the right functions are called along the way

	// StartUSerService will be called exactly once, with all the provided services
	backend.EXPECT().StartUserServices(
		ctx,
		enclaveId,
		mock.MatchedBy(func(services map[service.ServiceID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following three keys:
			// {successfulServiceId, failedServiceId, sidecarFailedServiceId}
			_, foundSuccessfulService := services[successfulServiceId]
			_, foundFailedService := services[failedServiceId]
			_, foundSidecarFailedService := services[sidecarFailedServiceId]
			return len(services) == 3 && foundSuccessfulService && foundFailedService && foundSidecarFailedService
		})).Times(1).Return(
		map[service.ServiceID]*service.Service{
			successfulServiceId:    successfulService,
			sidecarFailedServiceId: sidecarFailedService,
		},
		map[service.ServiceID]error{
			failedServiceId: stacktrace.NewError("Failed starting service"),
		},
		nil)

	// CreateNetworkingSidecar will be called exactly twice with the 2 successfully started services
	backend.EXPECT().CreateNetworkingSidecar(ctx, enclaveId, successfulServiceGuid).Times(1).Return(
		lib_networking_sidecar.NewNetworkingSidecar(successfulServiceGuid, enclaveId, container_status.ContainerStatus_Running),
		nil)
	backend.EXPECT().CreateNetworkingSidecar(ctx, enclaveId, sidecarFailedServiceGuid).Times(1).Return(
		nil,
		stacktrace.NewError("Failed starting sidecar for service"))

	// RunNetworkingSidecarExecCommands will be called only once for the successfully started sidecar
	backend.EXPECT().RunNetworkingSidecarExecCommands(
		ctx,
		enclaveId,
		mock.MatchedBy(func(commands map[service.ServiceGUID][]string) bool {
			// Matcher function returning true iff the commands map arg contains exactly the following key:
			// {successfulServiceGuid}
			_, foundSuccessfulService := commands[successfulServiceGuid]
			return len(commands) == 1 && foundSuccessfulService
		})).Times(2).Return(
		map[service.ServiceGUID]*exec_result.ExecResult{
			successfulServiceGuid: exec_result.NewExecResult(0, ""),
		},
		map[service.ServiceGUID]error{},
		nil)

	// DestroyUserServices is being called for sidecarFailedService only because the sidecar failed to be started
	backend.EXPECT().DestroyUserServices(
		ctx,
		enclaveId,
		mock.MatchedBy(func(filters *service.ServiceFilters) bool {
			// Matcher function returning true iff the filters map arg contains exactly the following key:
			// {sidecarFailedServiceGuid}
			_, foundSidecarFailedService := filters.GUIDs[sidecarFailedServiceGuid]
			return len(filters.Statuses) == 0 && len(filters.IDs) == 0 && len(filters.GUIDs) == 1 && foundSidecarFailedService
		})).Times(1).Return(
		map[service.ServiceGUID]bool{
			sidecarFailedServiceGuid: true,
		},
		map[service.ServiceGUID]error{},
		nil)

	success, failure, err := network.StartServices(
		ctx,
		map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig{
			successfulServiceId:    successfulServiceConfig,
			failedServiceId:        failedServiceConfig,
			sidecarFailedServiceId: sidecarFailedServiceConfig,
		},
		"",
	)
	require.Nil(t, err)
	require.Len(t, success, 1)
	require.Contains(t, success, successfulServiceId)
	require.Len(t, failure, 2)
	require.Contains(t, failure, failedServiceId)
	require.Contains(t, failure, sidecarFailedServiceId)
}

func TestSetDefaultConnection(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveId,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveId),
	)

	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection("test-partition"))
	require.Nil(t, network.topology.AddService("test-service", "test-partition"))
	network.networkingSidecars["test-service"] = networking_sidecar.NewMockNetworkingSidecarWrapper()

	newDefaultConnection := partition_topology.NewPartitionConnection(50)
	err := network.SetDefaultConnection(ctx, newDefaultConnection)
	require.Nil(t, err)
	require.Equal(t, network.topology.GetDefaultConnection(), newDefaultConnection)
}

func TestSetDefaultConnection_FailureRollbackDefaultConnection(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveId,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveId),
	)

	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection("test-partition"))
	require.Nil(t, network.topology.AddService("test-service", "test-partition"))
	// not add the sidecar such that it won't be able to update the networwing rule

	newDefaultConnection := partition_topology.NewPartitionConnection(50)
	err := network.SetDefaultConnection(ctx, newDefaultConnection)
	require.Contains(t, err.Error(), "Unable to update connections between the different partitions of the topology")
	// check connection was rolled back to ConnectionAllowed in the topology
	require.Equal(t, network.topology.GetDefaultConnection(), partition_topology.ConnectionAllowed)
}

func TestSetConnection(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveId,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveId),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	service1Index := 1
	service1 := service.NewServiceRegistration(
		testServiceIdFromInt(service1Index),
		testServiceGuidFromInt(service1Index),
		enclaveId,
		testIpFromInt(service1Index))
	service2Index := 2
	service2 := service.NewServiceRegistration(
		testServiceIdFromInt(service2Index),
		testServiceGuidFromInt(service2Index),
		enclaveId,
		testIpFromInt(service2Index))

	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition2))
	require.Nil(t, network.topology.AddService(service1.GetID(), partition1))
	require.Nil(t, network.topology.AddService(service2.GetID(), partition2))

	network.registeredServiceInfo[service1.GetID()] = service1
	network.registeredServiceInfo[service2.GetID()] = service2

	network.networkingSidecars[service1.GetID()] = networking_sidecar.NewMockNetworkingSidecarWrapper()
	network.networkingSidecars[service2.GetID()] = networking_sidecar.NewMockNetworkingSidecarWrapper()

	connectionOverride := partition_topology.NewPartitionConnection(50)
	err := network.SetConnection(ctx, partition1, partition2, connectionOverride)
	require.Nil(t, err)

	// check that connection override was successfully set to connectionOverride
	_, currentConnectionOverride, err := network.topology.GetPartitionConnection(partition1, partition2)
	require.Nil(t, err)
	require.Equal(t, connectionOverride, currentConnectionOverride)
}

func TestSetConnection_FailureRollsBackChanges(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveId,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveId),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	service1Index := 1
	service1 := service.NewServiceRegistration(
		testServiceIdFromInt(service1Index),
		testServiceGuidFromInt(service1Index),
		enclaveId,
		testIpFromInt(service1Index))
	service2Index := 2
	service2 := service.NewServiceRegistration(
		testServiceIdFromInt(service2Index),
		testServiceGuidFromInt(service2Index),
		enclaveId,
		testIpFromInt(service2Index))

	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	// don't create partition 2 as it will be created on the fly by SetConnection
	require.Nil(t, network.topology.AddService(service1.GetID(), partition1))
	require.Nil(t, network.topology.AddService(service2.GetID(), partition1))

	network.registeredServiceInfo[service1.GetID()] = service1
	network.registeredServiceInfo[service2.GetID()] = service2

	// do not add any sidecar such that updating network traffic will throw an exception

	connectionOverride := partition_topology.NewPartitionConnection(50)
	err := network.SetConnection(ctx, partition1, partition2, connectionOverride)
	require.Contains(t, err.Error(), "Unable to update connections between the different partitions of the topology")

	// check that partition2 was successfully cleaned up
	require.NotContains(t, network.topology.GetPartitionServices(), partition2)

}

func TestUnsetConnection(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveId,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveId),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	service1Index := 1
	service1 := service.NewServiceRegistration(
		testServiceIdFromInt(service1Index),
		testServiceGuidFromInt(service1Index),
		enclaveId,
		testIpFromInt(service1Index))
	service2Index := 2
	service2 := service.NewServiceRegistration(
		testServiceIdFromInt(service2Index),
		testServiceGuidFromInt(service2Index),
		enclaveId,
		testIpFromInt(service2Index))

	connectionOverride := partition_topology.NewPartitionConnection(50)
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition2))
	require.Nil(t, network.topology.SetConnection(partition1, partition2, connectionOverride))
	require.Nil(t, network.topology.AddService(service1.GetID(), partition1))
	require.Nil(t, network.topology.AddService(service2.GetID(), partition2))

	network.registeredServiceInfo[service1.GetID()] = service1
	network.registeredServiceInfo[service2.GetID()] = service2

	network.networkingSidecars[service1.GetID()] = networking_sidecar.NewMockNetworkingSidecarWrapper()
	network.networkingSidecars[service2.GetID()] = networking_sidecar.NewMockNetworkingSidecarWrapper()

	err := network.UnsetConnection(ctx, partition1, partition2)
	require.Nil(t, err)
	// test connection was successfully unset back to default
	_, currentConnectionOverride, err := network.topology.GetPartitionConnection(partition1, partition2)
	require.Nil(t, err)
	require.Equal(t, partition_topology.ConnectionAllowed, currentConnectionOverride)
}

func TestUnsetConnection_FailureRollsBackChanges(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveId,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveId),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	service1Index := 1
	service1 := service.NewServiceRegistration(
		testServiceIdFromInt(service1Index),
		testServiceGuidFromInt(service1Index),
		enclaveId,
		testIpFromInt(service1Index))
	service2Index := 2
	service2 := service.NewServiceRegistration(
		testServiceIdFromInt(service2Index),
		testServiceGuidFromInt(service2Index),
		enclaveId,
		testIpFromInt(service2Index))

	connectionOverride := partition_topology.NewPartitionConnection(50)
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition2))
	require.Nil(t, network.topology.SetConnection(partition1, partition2, connectionOverride))
	require.Nil(t, network.topology.AddService(service1.GetID(), partition1))
	require.Nil(t, network.topology.AddService(service2.GetID(), partition2))

	network.registeredServiceInfo[service1.GetID()] = service1
	network.registeredServiceInfo[service2.GetID()] = service2

	// do not add any sidecar such that updating network traffic will throw an exception

	err := network.UnsetConnection(ctx, partition1, partition2)
	require.Contains(t, err.Error(), "Unable to update connections between the different partitions of the topology")

	// check that connection was rolled back to the previous override
	_, currentConnectionOverride, err := network.topology.GetPartitionConnection(partition1, partition2)
	require.Nil(t, err)
	require.Equal(t, connectionOverride, currentConnectionOverride)
}

func TestUpdateTrafficControl(t *testing.T) {
	ctx := context.Background()

	enclaveId := enclave.EnclaveID("test")

	sidecars := map[service.ServiceID]networking_sidecar.NetworkingSidecarWrapper{}
	registrations := map[service.ServiceID]*service.ServiceRegistration{}
	mockSidecars := map[service.ServiceID]*networking_sidecar.MockNetworkingSidecarWrapper{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)

		sidecar := networking_sidecar.NewMockNetworkingSidecarWrapper()
		sidecars[serviceId] = sidecar
		mockSidecars[serviceId] = sidecar

		ip := testIpFromInt(i)
		serviceGuid := testServiceGuidFromInt(i)
		registrations[serviceId] = service.NewServiceRegistration(
			serviceId,
			serviceGuid,
			enclaveId,
			ip,
		)
	}

	// Creates the pathological "line" of connections, where each service can only see the services adjacent
	targetServicePacketLossConfigs := map[service.ServiceID]map[service.ServiceID]float32{}
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)
		otherServicesPacketLossConfig := map[service.ServiceID]float32{}
		for j := 0; j < numServices; j++ {
			if j < i-1 || j > i+1 {
				blockedServiceId := testServiceIdFromInt(j)
				otherServicesPacketLossConfig[blockedServiceId] = packetLossConfigForBlockedPartition
			}
		}
		targetServicePacketLossConfigs[serviceId] = otherServicesPacketLossConfig
	}

	require.Nil(t, updateTrafficControlConfiguration(ctx, targetServicePacketLossConfigs, registrations, sidecars))

	// Verify that each service got told to block exactly the right things
	for i := 0; i < numServices; i++ {
		serviceId := testServiceIdFromInt(i)

		expected := map[string]float32{}
		for j := 0; j < numServices; j++ {
			if j < i-1 || j > i+1 {
				ip := testIpFromInt(j)
				expected[ip.String()] = packetLossConfigForBlockedPartition
			}
		}

		mockSidecar := mockSidecars[serviceId]
		recordedPacketLossConfig := mockSidecar.GetRecordedUpdatePacketLossConfig()
		require.Equal(t, 1, len(recordedPacketLossConfig), "Expected sidecar for service ID '%v' to have recorded exactly one call to update")

		actualPacketLossConfigForService := recordedPacketLossConfig[0]

		require.Equal(t, expected, actualPacketLossConfigForService)
	}
}

func testIpFromInt(i int) net.IP {
	return []byte{1, 1, 1, byte(i)}
}

func testServiceIdFromInt(i int) service.ServiceID {
	return service.ServiceID("service-" + strconv.Itoa(i))
}

func testServiceGuidFromInt(i int) service.ServiceGUID {
	return service.ServiceGUID(fmt.Sprintf(
		"service-%v-%v",
		i,
		i,
	))
}
