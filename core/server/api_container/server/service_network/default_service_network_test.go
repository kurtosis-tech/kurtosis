/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net"
	"strconv"
	"testing"
)

const (
	packetLossConfigForBlockedPartition = float32(100)

	numServices = 10

	enclaveName             = enclave.EnclaveUUID("test-enclave")
	partitioningEnabled     = true
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

	serviceInternalTestId := 1
	servicePartitionId := testPartitionIdFromInt(serviceInternalTestId)
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp)
	serviceObj := service.NewService(serviceRegistration, container_status.ContainerStatus_Running, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{})
	serviceConfig := services.NewServiceConfigBuilder(testContainerImageName).WithSubnetwork(string(servicePartitionId)).Build()

	network := NewDefaultServiceNetwork(
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	// The service is registered before being started
	backend.EXPECT().RegisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceName]bool{
			serviceName: true,
		},
	).Times(1).Return(
		map[service.ServiceName]*service.ServiceRegistration{
			serviceName: serviceRegistration,
		},
		map[service.ServiceName]error{},
		nil,
	)

	// Then the service is started
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(services map[service.ServiceUUID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following key:
			// {serviceName}
			_, foundService := services[serviceUuid]
			return len(services) == 1 && foundService
		})).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			serviceUuid: serviceObj,
		},
		map[service.ServiceUUID]error{},
		nil)

	// CreateNetworkingSidecar will be called for this service
	backend.EXPECT().CreateNetworkingSidecar(ctx, enclaveName, serviceUuid).Times(1).Return(
		lib_networking_sidecar.NewNetworkingSidecar(serviceUuid, enclaveName, container_status.ContainerStatus_Running),
		nil)

	// RunNetworkingSidecarExecCommands will be called for this service
	backend.EXPECT().RunNetworkingSidecarExecCommands(
		ctx,
		enclaveName,
		mock.MatchedBy(func(commands map[service.ServiceUUID][]string) bool {
			// Matcher function returning true iff the commands map arg contains exactly the following key:
			// {serviceUuid}
			_, foundService := commands[serviceUuid]
			return len(commands) == 1 && foundService
		})).Times(2).Return(
		map[service.ServiceUUID]*exec_result.ExecResult{
			serviceUuid: exec_result.NewExecResult(0, ""),
		},
		map[service.ServiceUUID]error{},
		nil)

	// DestroyUserServices is never being called as everything is successful for this test
	backend.EXPECT().DestroyUserServices(
		ctx,
		enclaveName,
		mock.Anything).Maybe().Times(0)

	startedService, err := network.StartService(ctx, serviceName, serviceConfig)
	require.Nil(t, err)
	require.NotNil(t, startedService)

	require.Equal(t, serviceRegistration, startedService.GetRegistration())

	require.Len(t, network.registeredServiceInfo, 1)
	require.Contains(t, network.registeredServiceInfo, serviceName)

	require.Len(t, network.networkingSidecars, 1)
	require.Contains(t, network.networkingSidecars, serviceName)

	expectedPartitionsInTopolody := map[service_network_types.PartitionID]map[service.ServiceName]bool{
		partition_topology.DefaultPartitionId: {},
		servicePartitionId: {
			serviceName: true,
		},
		// partitions with services that failed to start were removed from the topology
	}
	require.Equal(t, expectedPartitionsInTopolody, network.topology.GetPartitionServices())
}

func TestStartService_FailedToStart(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	servicePartitionId := testPartitionIdFromInt(serviceInternalTestId)
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	serviceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, serviceIp)
	serviceConfig := services.NewServiceConfigBuilder(testContainerImageName).WithSubnetwork(string(servicePartitionId)).Build()

	network := NewDefaultServiceNetwork(
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	// The service is registered before being started
	backend.EXPECT().RegisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceName]bool{
			serviceName: true,
		},
	).Times(1).Return(
		map[service.ServiceName]*service.ServiceRegistration{
			serviceName: serviceRegistration,
		},
		map[service.ServiceName]error{},
		nil,
	)

	// StartRegisteredUserServices will be called for this service
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(services map[service.ServiceUUID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following key:
			// {serviceName}
			_, foundService := services[serviceUuid]
			return len(services) == 1 && foundService
		})).Times(1).Return(
		map[service.ServiceUUID]*service.Service{},
		map[service.ServiceUUID]error{
			serviceUuid: stacktrace.NewError("Failed starting service"),
		},
		nil)

	// CreateNetworkingSidecar will be called for this service
	backend.EXPECT().CreateNetworkingSidecar(ctx, enclaveName, mock.Anything).Maybe().Times(0)

	// RunNetworkingSidecarExecCommands will never be called
	backend.EXPECT().RunNetworkingSidecarExecCommands(
		ctx,
		enclaveName,
		mock.Anything).Maybe().Times(0)

	// DestroyUserServices is never being called as the service fails to start for this test
	backend.EXPECT().DestroyUserServices(
		ctx,
		enclaveName,
		mock.Anything).Maybe().Times(0)

	// Since the service fails to start, it is unregistered in a deferred function
	backend.EXPECT().UnregisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil,
	)

	startedService, err := network.StartService(ctx, serviceName, serviceConfig)
	require.NotNil(t, err)
	require.Nil(t, startedService)

	require.Empty(t, network.registeredServiceInfo)

	require.Empty(t, network.networkingSidecars)

	expectedPartitionsInTopolody := map[service_network_types.PartitionID]map[service.ServiceName]bool{
		partition_topology.DefaultPartitionId: {},
	}
	require.Equal(t, expectedPartitionsInTopolody, network.topology.GetPartitionServices())
}

func TestStartService_SidecarFailedToStart(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	servicePartitionId := testPartitionIdFromInt(serviceInternalTestId)
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp)
	serviceObj := service.NewService(serviceRegistration, container_status.ContainerStatus_Running, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{})
	serviceConfig := services.NewServiceConfigBuilder(testContainerImageName).WithSubnetwork(string(servicePartitionId)).Build()

	network := NewDefaultServiceNetwork(
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	// The service is registered before being started
	backend.EXPECT().RegisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceName]bool{
			serviceName: true,
		},
	).Times(1).Return(
		map[service.ServiceName]*service.ServiceRegistration{
			serviceName: serviceRegistration,
		},
		map[service.ServiceName]error{},
		nil,
	)

	// StartRegisteredUserServices will be called for this service
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(services map[service.ServiceUUID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following key:
			// {serviceName}
			_, foundService := services[serviceUuid]
			return len(services) == 1 && foundService
		})).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			serviceUuid: serviceObj,
		},
		map[service.ServiceUUID]error{},
		nil)

	// CreateNetworkingSidecar will be called for this service
	backend.EXPECT().CreateNetworkingSidecar(ctx, enclaveName, serviceUuid).Times(1).Return(
		nil,
		errors.New("failed creating sidecar"))

	// RunNetworkingSidecarExecCommands will never be called
	backend.EXPECT().RunNetworkingSidecarExecCommands(
		ctx,
		enclaveName,
		mock.Anything).Maybe().Times(0)

	// DestroyUserServices is being called for sidecarFailedService only because the sidecar failed to be started
	backend.EXPECT().DestroyUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(filters *service.ServiceFilters) bool {
			// Matcher function returning true iff the filters map arg contains exactly the following key:
			// {serviceUuid}
			_, foundService := filters.UUIDs[serviceUuid]
			return len(filters.Statuses) == 0 && len(filters.Names) == 0 && len(filters.UUIDs) == 1 && foundService
		})).Times(1).Return(
		map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil)

	// Since the service sidecar fails to start, the service is destroyed and then unregistered
	backend.EXPECT().UnregisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil,
	)

	startedService, err := network.StartService(ctx, serviceName, serviceConfig)
	require.NotNil(t, err)
	require.Nil(t, startedService)

	require.Empty(t, network.registeredServiceInfo)

	require.Empty(t, network.networkingSidecars, 1)

	expectedPartitionsInTopolody := map[service_network_types.PartitionID]map[service.ServiceName]bool{
		partition_topology.DefaultPartitionId: {},
	}
	require.Equal(t, expectedPartitionsInTopolody, network.topology.GetPartitionServices())
}

func TestStartServices_Success(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	// One service will be started successfully
	successfulServiceIndex := 1
	successfulServicePartitionId := testPartitionIdFromInt(successfulServiceIndex)
	successfulServiceId := testServiceNameFromInt(successfulServiceIndex)
	successfulServiceGuid := testServiceUuidFromInt(successfulServiceIndex)
	successfulServiceIp := testIpFromInt(successfulServiceIndex)
	successfulServiceRegistration := service.NewServiceRegistration(successfulServiceId, successfulServiceGuid, enclaveName, successfulServiceIp)
	successfulService := service.NewService(successfulServiceRegistration, container_status.ContainerStatus_Running, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{})
	successfulServiceConfig := services.NewServiceConfigBuilder(testContainerImageName).WithSubnetwork(string(successfulServicePartitionId)).Build()

	network := NewDefaultServiceNetwork(
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	// Configure the mock to also be testing that the right functions are called along the way

	// The services are registered one by one before being started
	backend.EXPECT().RegisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceName]bool{
			successfulServiceId: true,
		},
	).Times(1).Return(
		map[service.ServiceName]*service.ServiceRegistration{
			successfulServiceId: successfulServiceRegistration,
		},
		map[service.ServiceName]error{},
		nil,
	)

	// StartUserService will be called three times, with all the provided services
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(services map[service.ServiceUUID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following key:
			// {successfulServiceId}
			_, foundSuccessfulService := services[successfulServiceGuid]
			return len(services) == 1 && foundSuccessfulService
		})).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			successfulServiceGuid: successfulService,
		},
		map[service.ServiceUUID]error{},
		nil)

	// CreateNetworkingSidecar will be called exactly twice with the 2 successfully started services
	backend.EXPECT().CreateNetworkingSidecar(ctx, enclaveName, successfulServiceGuid).Times(1).Return(
		lib_networking_sidecar.NewNetworkingSidecar(successfulServiceGuid, enclaveName, container_status.ContainerStatus_Running),
		nil)

	// RunNetworkingSidecarExecCommands will be called only once for the successfully started sidecar
	backend.EXPECT().RunNetworkingSidecarExecCommands(
		ctx,
		enclaveName,
		mock.MatchedBy(func(commands map[service.ServiceUUID][]string) bool {
			// Matcher function returning true iff the commands map arg contains exactly the following key:
			// {successfulServiceGuid}
			_, foundSuccessfulService := commands[successfulServiceGuid]
			return len(commands) == 1 && foundSuccessfulService
		})).Times(2).Return(
		map[service.ServiceUUID]*exec_result.ExecResult{
			successfulServiceGuid: exec_result.NewExecResult(0, ""),
		},
		map[service.ServiceUUID]error{},
		nil)

	success, failure, err := network.StartServices(
		ctx,
		map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig{
			successfulServiceId: successfulServiceConfig,
		},
	)
	require.Nil(t, err)
	require.Len(t, success, 1)
	require.Contains(t, success, successfulServiceId)
	require.Empty(t, failure)

	require.Len(t, network.registeredServiceInfo, 1)
	require.Contains(t, network.registeredServiceInfo, successfulServiceId)

	require.Len(t, network.networkingSidecars, 1)
	require.Contains(t, network.networkingSidecars, successfulServiceId)

	expectedPartitionsInTopolody := map[service_network_types.PartitionID]map[service.ServiceName]bool{
		partition_topology.DefaultPartitionId: {},
		successfulServicePartitionId: {
			successfulServiceId: true,
		},
	}
	require.Equal(t, expectedPartitionsInTopolody, network.topology.GetPartitionServices())
}

func TestStartServices_FailureRollsBackTheEntireBatch(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	// One service will be started successfully
	successfulServiceIndex := 1
	successfulServicePartitionId := testPartitionIdFromInt(successfulServiceIndex)
	successfulServiceName := testServiceNameFromInt(successfulServiceIndex)
	successfulServiceUuid := testServiceUuidFromInt(successfulServiceIndex)
	successfulServiceIp := testIpFromInt(successfulServiceIndex)
	successfulServiceRegistration := service.NewServiceRegistration(successfulServiceName, successfulServiceUuid, enclaveName, successfulServiceIp)
	successfulService := service.NewService(successfulServiceRegistration, container_status.ContainerStatus_Running, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{})
	successfulServiceConfig := services.NewServiceConfigBuilder(testContainerImageName).WithSubnetwork(string(successfulServicePartitionId)).Build()

	// One service will fail to be started
	failedServiceIndex := 2
	failedServicePartitionId := testPartitionIdFromInt(failedServiceIndex)
	failedServiceName := testServiceNameFromInt(failedServiceIndex)
	failedServiceUuid := testServiceUuidFromInt(failedServiceIndex)
	failedServiceIp := testIpFromInt(failedServiceIndex)
	failedServiceRegistration := service.NewServiceRegistration(failedServiceName, failedServiceUuid, enclaveName, failedServiceIp)
	failedServiceConfig := services.NewServiceConfigBuilder(testContainerImageName).WithSubnetwork(string(failedServicePartitionId)).Build()

	// One service will be successfully started but its sidecar will fail to start
	sidecarFailedServiceIndex := 3
	sidecarFailedServicePartitionId := testPartitionIdFromInt(sidecarFailedServiceIndex)
	sidecarFailedServiceName := testServiceNameFromInt(sidecarFailedServiceIndex)
	sidecarFailedServiceUuid := testServiceUuidFromInt(sidecarFailedServiceIndex)
	sidecarFailedServiceIp := testIpFromInt(sidecarFailedServiceIndex)
	sidecarFailedServiceRegistration := service.NewServiceRegistration(sidecarFailedServiceName, sidecarFailedServiceUuid, enclaveName, sidecarFailedServiceIp)
	sidecarFailedService := service.NewService(sidecarFailedServiceRegistration, container_status.ContainerStatus_Running, map[string]*port_spec.PortSpec{}, sidecarFailedServiceIp, map[string]*port_spec.PortSpec{})
	sidecarFailedServiceConfig := services.NewServiceConfigBuilder(testContainerImageName).WithSubnetwork(string(sidecarFailedServicePartitionId)).Build()

	network := NewDefaultServiceNetwork(
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	// Configure the mock to also be testing that the right functions are called along the way

	// The services are registered one by one before being started
	backend.EXPECT().RegisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceName]bool{
			successfulServiceName: true,
		},
	).Times(1).Return(
		map[service.ServiceName]*service.ServiceRegistration{
			successfulServiceName: successfulServiceRegistration,
		},
		map[service.ServiceName]error{},
		nil,
	)
	backend.EXPECT().RegisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceName]bool{
			failedServiceName: true,
		},
	).Times(1).Return(
		map[service.ServiceName]*service.ServiceRegistration{
			failedServiceName: failedServiceRegistration,
		},
		map[service.ServiceName]error{},
		nil,
	)
	backend.EXPECT().RegisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceName]bool{
			sidecarFailedServiceName: true,
		},
	).Times(1).Return(
		map[service.ServiceName]*service.ServiceRegistration{
			sidecarFailedServiceName: sidecarFailedServiceRegistration,
		},
		map[service.ServiceName]error{},
		nil,
	)

	// StartUserService will be called three times, with all the provided services
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(services map[service.ServiceUUID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following key:
			// {successfulServiceName}
			_, foundSuccessfulService := services[successfulServiceUuid]
			return len(services) == 1 && foundSuccessfulService
		})).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			successfulServiceUuid: successfulService,
		},
		map[service.ServiceUUID]error{},
		nil)
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(services map[service.ServiceUUID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following key:
			// {failedServiceName}
			_, foundFailedService := services[failedServiceUuid]
			return len(services) == 1 && foundFailedService
		})).Times(1).Return(
		map[service.ServiceUUID]*service.Service{},
		map[service.ServiceUUID]error{
			failedServiceUuid: stacktrace.NewError("Failed starting service"),
		},
		nil)
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(services map[service.ServiceUUID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following key:
			// {sidecarFailedServiceName}
			_, foundSidecarFailedService := services[sidecarFailedServiceUuid]
			return len(services) == 1 && foundSidecarFailedService
		})).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			sidecarFailedServiceUuid: sidecarFailedService,
		},
		map[service.ServiceUUID]error{},
		nil)

	// CreateNetworkingSidecar will be called exactly twice with the 2 successfully started services
	backend.EXPECT().CreateNetworkingSidecar(ctx, enclaveName, successfulServiceUuid).Times(1).Return(
		lib_networking_sidecar.NewNetworkingSidecar(successfulServiceUuid, enclaveName, container_status.ContainerStatus_Running),
		nil)
	backend.EXPECT().CreateNetworkingSidecar(ctx, enclaveName, sidecarFailedServiceUuid).Times(1).Return(
		nil,
		stacktrace.NewError("Failed starting sidecar for service"))

	// RunNetworkingSidecarExecCommands will be called only once for the successfully started sidecar
	backend.EXPECT().RunNetworkingSidecarExecCommands(
		ctx,
		enclaveName,
		mock.MatchedBy(func(commands map[service.ServiceUUID][]string) bool {
			// Matcher function returning true iff the commands map arg contains exactly the following key:
			// {successfulServiceUuid}
			_, foundSuccessfulService := commands[successfulServiceUuid]
			return len(commands) == 1 && foundSuccessfulService
		})).Times(1).Return(
		map[service.ServiceUUID]*exec_result.ExecResult{
			successfulServiceUuid: exec_result.NewExecResult(0, ""),
		},
		map[service.ServiceUUID]error{},
		nil)

	// DestroyUserServices is being called for sidecarFailedService only because the sidecar failed to be started
	backend.EXPECT().DestroyUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(filters *service.ServiceFilters) bool {
			// Matcher function returning true iff the filters map arg contains exactly the following key:
			// {sidecarFailedServiceGuid}
			_, foundSuccessfulService := filters.UUIDs[successfulServiceUuid]
			return len(filters.Statuses) == 0 && len(filters.Names) == 0 && len(filters.UUIDs) == 1 && foundSuccessfulService
		})).Times(1).Return(
		map[service.ServiceUUID]bool{
			successfulServiceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil)
	backend.EXPECT().DestroyUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(filters *service.ServiceFilters) bool {
			// Matcher function returning true iff the filters map arg contains exactly the following key:
			// {sidecarFailedServiceGuid}
			_, foundSidecarFailedService := filters.UUIDs[sidecarFailedServiceUuid]
			return len(filters.Statuses) == 0 && len(filters.Names) == 0 && len(filters.UUIDs) == 1 && foundSidecarFailedService
		})).Times(1).Return(
		map[service.ServiceUUID]bool{
			sidecarFailedServiceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil)

	// Both failedService and sidecarFailedService are unregistered in the deferred functions
	backend.EXPECT().UnregisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]bool{
			successfulServiceUuid: true,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			successfulServiceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil,
	)
	backend.EXPECT().UnregisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]bool{
			failedServiceUuid: true,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			failedServiceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil,
	)
	backend.EXPECT().UnregisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]bool{
			sidecarFailedServiceUuid: true,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			sidecarFailedServiceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil,
	)

	backend.EXPECT().StopNetworkingSidecars(
		ctx,
		&lib_networking_sidecar.NetworkingSidecarFilters{
			EnclaveUUIDs: nil,
			UserServiceUUIDs: map[service.ServiceUUID]bool{
				successfulServiceUuid: true,
			},
			Statuses: nil,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			sidecarFailedServiceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil,
	)

	success, failure, err := network.StartServices(
		ctx,
		map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig{
			successfulServiceName:    successfulServiceConfig,
			failedServiceName:        failedServiceConfig,
			sidecarFailedServiceName: sidecarFailedServiceConfig,
		},
	)
	require.Nil(t, err)
	require.Empty(t, success) // as the full batch failed, the successful service should have been destroyed
	require.Len(t, failure, 2)
	require.Contains(t, failure, failedServiceName)
	require.Contains(t, failure, sidecarFailedServiceName)

	require.Empty(t, network.registeredServiceInfo)

	require.Empty(t, network.networkingSidecars)

	expectedPartitionsInTopolody := map[service_network_types.PartitionID]map[service.ServiceName]bool{
		partition_topology.DefaultPartitionId: {},
	}
	require.Equal(t, expectedPartitionsInTopolody, network.topology.GetPartitionServices())
}

func TestUpdateService(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	partition0 := service_network_types.PartitionID("partition0")
	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	// service that will be successfully moved from partition1 to partition2
	successfulServiceIndex := 1
	successfulService := service.NewServiceRegistration(
		testServiceNameFromInt(successfulServiceIndex),
		testServiceUuidFromInt(successfulServiceIndex),
		enclaveName,
		testIpFromInt(successfulServiceIndex))

	// service that will be in partition0 from the start and be updated to partition0 (i.e. it will no-op)
	serviceAlreadyInPartitionIndex := 2
	serviceAlreadyInPartition := service.NewServiceRegistration(
		testServiceNameFromInt(serviceAlreadyInPartitionIndex),
		testServiceUuidFromInt(serviceAlreadyInPartitionIndex),
		enclaveName,
		testIpFromInt(serviceAlreadyInPartitionIndex))

	// service that does not exist, and yet we will try to update it. It should fail
	nonExistentServiceIndex := 3
	nonExistentService := service.NewServiceRegistration(
		testServiceNameFromInt(nonExistentServiceIndex),
		testServiceUuidFromInt(nonExistentServiceIndex),
		enclaveName,
		testIpFromInt(nonExistentServiceIndex))

	// service that will be moved from default partition to partition2
	serviceToMoveOutOfDefaultPartitionIndex := 4
	serviceToMoveOutOfDefaultPartition := service.NewServiceRegistration(
		testServiceNameFromInt(serviceToMoveOutOfDefaultPartitionIndex),
		testServiceUuidFromInt(serviceToMoveOutOfDefaultPartitionIndex),
		enclaveName,
		testIpFromInt(serviceToMoveOutOfDefaultPartitionIndex))

	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition0))
	require.Nil(t, network.topology.AddService(successfulService.GetName(), partition1))
	require.Nil(t, network.topology.AddService(serviceAlreadyInPartition.GetName(), partition0))
	require.Nil(t, network.topology.AddService(serviceToMoveOutOfDefaultPartition.GetName(), partition_topology.DefaultPartitionId))

	network.registeredServiceInfo[successfulService.GetName()] = successfulService
	network.registeredServiceInfo[serviceAlreadyInPartition.GetName()] = serviceAlreadyInPartition
	network.registeredServiceInfo[serviceToMoveOutOfDefaultPartition.GetName()] = serviceToMoveOutOfDefaultPartition
	// nonExistentService don't get added here as it is supposed to be an unknown service

	network.networkingSidecars[successfulService.GetName()] = networking_sidecar.NewMockNetworkingSidecarWrapper()
	network.networkingSidecars[serviceAlreadyInPartition.GetName()] = networking_sidecar.NewMockNetworkingSidecarWrapper()
	network.networkingSidecars[serviceToMoveOutOfDefaultPartition.GetName()] = networking_sidecar.NewMockNetworkingSidecarWrapper()

	success, failure, err := network.UpdateService(ctx, map[service.ServiceName]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig{
		successfulService.GetName():                  binding_constructors.NewUpdateServiceConfig(string(partition2)),
		serviceAlreadyInPartition.GetName():          binding_constructors.NewUpdateServiceConfig(string(partition0)),
		nonExistentService.GetName():                 binding_constructors.NewUpdateServiceConfig(string(partition2)),
		serviceToMoveOutOfDefaultPartition.GetName(): binding_constructors.NewUpdateServiceConfig(string(partition2)),
	})
	require.Nil(t, err)
	require.Len(t, success, 3)
	require.Contains(t, success, successfulService.GetName())
	require.Contains(t, success, serviceAlreadyInPartition.GetName())
	require.Contains(t, success, serviceToMoveOutOfDefaultPartition.GetName())

	require.Len(t, failure, 1)
	require.Contains(t, failure, nonExistentService.GetName())

	expectedPartitions := map[service_network_types.PartitionID]map[service.ServiceName]bool{
		partition_topology.DefaultPartitionId: {},
		partition0: {
			serviceAlreadyInPartition.GetName(): true,
		},
		partition1: {},
		partition2: {
			successfulService.GetName():                  true,
			serviceToMoveOutOfDefaultPartition.GetName(): true,
		},
	}
	require.Equal(t, expectedPartitions, network.topology.GetPartitionServices())
}

func TestUpdateService_FullBatchFailureRollBack(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	failingServiceIndex := 1
	failingService := service.NewServiceRegistration(
		testServiceNameFromInt(failingServiceIndex),
		testServiceUuidFromInt(failingServiceIndex),
		enclaveName,
		testIpFromInt(failingServiceIndex))
	successfulServiceIndex := 2
	successfulService := service.NewServiceRegistration(
		testServiceNameFromInt(successfulServiceIndex),
		testServiceUuidFromInt(successfulServiceIndex),
		enclaveName,
		testIpFromInt(successfulServiceIndex))

	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	require.Nil(t, network.topology.AddService(failingService.GetName(), partition1))
	require.Nil(t, network.topology.AddService(successfulService.GetName(), partition1))

	network.registeredServiceInfo[failingService.GetName()] = failingService
	network.registeredServiceInfo[successfulService.GetName()] = successfulService

	// do not add sidecar for failingService so that it fails updating the connections
	network.networkingSidecars[successfulService.GetName()] = networking_sidecar.NewMockNetworkingSidecarWrapper()

	success, failure, err := network.UpdateService(ctx, map[service.ServiceName]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig{
		failingService.GetName():    binding_constructors.NewUpdateServiceConfig(string(partition2)),
		successfulService.GetName(): binding_constructors.NewUpdateServiceConfig(string(partition2)),
	})
	require.Contains(t, err.Error(), "Unable to update connections between the different partitions of the topology")

	require.Nil(t, success)
	require.Nil(t, failure)

	expectedPartitions := map[service_network_types.PartitionID]map[service.ServiceName]bool{
		partition_topology.DefaultPartitionId: {},
		// partition2 was removed as it was a new partition that remained empty
		// both services were left into partition1
		partition1: {
			failingService.GetName():    true,
			successfulService.GetName(): true,
		},
	}
	require.Equal(t, expectedPartitions, network.topology.GetPartitionServices())
}

func TestSetDefaultConnection(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	network := NewDefaultServiceNetwork(
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
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
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
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
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	service1Index := 1
	service1 := service.NewServiceRegistration(
		testServiceNameFromInt(service1Index),
		testServiceUuidFromInt(service1Index),
		enclaveName,
		testIpFromInt(service1Index))
	service2Index := 2
	service2 := service.NewServiceRegistration(
		testServiceNameFromInt(service2Index),
		testServiceUuidFromInt(service2Index),
		enclaveName,
		testIpFromInt(service2Index))

	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition2))
	require.Nil(t, network.topology.AddService(service1.GetName(), partition1))
	require.Nil(t, network.topology.AddService(service2.GetName(), partition2))

	network.registeredServiceInfo[service1.GetName()] = service1
	network.registeredServiceInfo[service2.GetName()] = service2

	network.networkingSidecars[service1.GetName()] = networking_sidecar.NewMockNetworkingSidecarWrapper()
	network.networkingSidecars[service2.GetName()] = networking_sidecar.NewMockNetworkingSidecarWrapper()

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
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	service1Index := 1
	service1 := service.NewServiceRegistration(
		testServiceNameFromInt(service1Index),
		testServiceUuidFromInt(service1Index),
		enclaveName,
		testIpFromInt(service1Index))
	service2Index := 2
	service2 := service.NewServiceRegistration(
		testServiceNameFromInt(service2Index),
		testServiceUuidFromInt(service2Index),
		enclaveName,
		testIpFromInt(service2Index))

	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	// don't create partition 2 as it will be created on the fly by SetConnection
	require.Nil(t, network.topology.AddService(service1.GetName(), partition1))
	require.Nil(t, network.topology.AddService(service2.GetName(), partition1))

	network.registeredServiceInfo[service1.GetName()] = service1
	network.registeredServiceInfo[service2.GetName()] = service2

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
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	service1Index := 1
	service1 := service.NewServiceRegistration(
		testServiceNameFromInt(service1Index),
		testServiceUuidFromInt(service1Index),
		enclaveName,
		testIpFromInt(service1Index))
	service2Index := 2
	service2 := service.NewServiceRegistration(
		testServiceNameFromInt(service2Index),
		testServiceUuidFromInt(service2Index),
		enclaveName,
		testIpFromInt(service2Index))

	connectionOverride := partition_topology.NewPartitionConnection(50)
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition2))
	require.Nil(t, network.topology.SetConnection(partition1, partition2, connectionOverride))
	require.Nil(t, network.topology.AddService(service1.GetName(), partition1))
	require.Nil(t, network.topology.AddService(service2.GetName(), partition2))

	network.registeredServiceInfo[service1.GetName()] = service1
	network.registeredServiceInfo[service2.GetName()] = service2

	network.networkingSidecars[service1.GetName()] = networking_sidecar.NewMockNetworkingSidecarWrapper()
	network.networkingSidecars[service2.GetName()] = networking_sidecar.NewMockNetworkingSidecarWrapper()

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
		enclaveName,
		ip,
		apiContainerPort,
		fakeApiContainerVersion,
		partitioningEnabled,
		backend,
		unusedEnclaveDataDir,
		networking_sidecar.NewStandardNetworkingSidecarManager(backend, enclaveName),
	)

	partition1 := service_network_types.PartitionID("partition1")
	partition2 := service_network_types.PartitionID("partition2")

	service1Index := 1
	service1 := service.NewServiceRegistration(
		testServiceNameFromInt(service1Index),
		testServiceUuidFromInt(service1Index),
		enclaveName,
		testIpFromInt(service1Index))
	service2Index := 2
	service2 := service.NewServiceRegistration(
		testServiceNameFromInt(service2Index),
		testServiceUuidFromInt(service2Index),
		enclaveName,
		testIpFromInt(service2Index))

	connectionOverride := partition_topology.NewPartitionConnection(50)
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition1))
	require.Nil(t, network.topology.CreateEmptyPartitionWithDefaultConnection(partition2))
	require.Nil(t, network.topology.SetConnection(partition1, partition2, connectionOverride))
	require.Nil(t, network.topology.AddService(service1.GetName(), partition1))
	require.Nil(t, network.topology.AddService(service2.GetName(), partition2))

	network.registeredServiceInfo[service1.GetName()] = service1
	network.registeredServiceInfo[service2.GetName()] = service2

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

	enclaveId := enclave.EnclaveUUID("test")

	sidecars := map[service.ServiceName]networking_sidecar.NetworkingSidecarWrapper{}
	registrations := map[service.ServiceName]*service.ServiceRegistration{}
	mockSidecars := map[service.ServiceName]*networking_sidecar.MockNetworkingSidecarWrapper{}
	for i := 0; i < numServices; i++ {
		serviceName := testServiceNameFromInt(i)

		sidecar := networking_sidecar.NewMockNetworkingSidecarWrapper()
		sidecars[serviceName] = sidecar
		mockSidecars[serviceName] = sidecar

		ip := testIpFromInt(i)
		serviceUuid := testServiceUuidFromInt(i)
		registrations[serviceName] = service.NewServiceRegistration(
			serviceName,
			serviceUuid,
			enclaveId,
			ip,
		)
	}

	// Creates the pathological "line" of connections, where each service can only see the services adjacent
	targetServicePacketLossConfigs := map[service.ServiceName]map[service.ServiceName]*partition_topology.PartitionConnection{}
	for i := 0; i < numServices; i++ {
		serviceName := testServiceNameFromInt(i)
		otherServicesPacketLossConfig := map[service.ServiceName]*partition_topology.PartitionConnection{}
		for j := 0; j < numServices; j++ {
			if j < i-1 || j > i+1 {
				blockedServiceId := testServiceNameFromInt(j)
				connectionConfig := partition_topology.NewPartitionConnection(packetLossConfigForBlockedPartition)
				otherServicesPacketLossConfig[blockedServiceId] = &connectionConfig
			}
		}
		targetServicePacketLossConfigs[serviceName] = otherServicesPacketLossConfig
	}

	require.Nil(t, updateTrafficControlConfiguration(ctx, targetServicePacketLossConfigs, registrations, sidecars))

	// Verify that each service got told to block exactly the right things
	for i := 0; i < numServices; i++ {
		serviceName := testServiceNameFromInt(i)

		expected := map[string]*partition_topology.PartitionConnection{}
		for j := 0; j < numServices; j++ {
			if j < i-1 || j > i+1 {
				ip := testIpFromInt(j)
				connectionConfig := partition_topology.NewPartitionConnection(packetLossConfigForBlockedPartition)
				expected[ip.String()] = &connectionConfig
			}
		}

		mockSidecar := mockSidecars[serviceName]
		recordedPacketLossConfig := mockSidecar.GetRecordedUpdatePacketLossConfig()
		require.Equal(t, 1, len(recordedPacketLossConfig), "Expected sidecar for service ID '%v' to have recorded exactly one call to update")

		actualPacketLossConfigForService := recordedPacketLossConfig[0]

		require.Equal(t, expected, actualPacketLossConfigForService)
	}
}

func testIpFromInt(i int) net.IP {
	return []byte{1, 1, 1, byte(i)}
}

func testServiceNameFromInt(i int) service.ServiceName {
	return service.ServiceName("service-" + strconv.Itoa(i))
}

func testPartitionIdFromInt(i int) service_network_types.PartitionID {
	return service_network_types.PartitionID("partition-" + strconv.Itoa(i))
}

func testServiceUuidFromInt(i int) service.ServiceUUID {
	return service.ServiceUUID(fmt.Sprintf(
		"massive-uuid-with-32-req-chars-%v",
		i,
	))
}
