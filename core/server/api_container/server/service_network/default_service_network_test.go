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
