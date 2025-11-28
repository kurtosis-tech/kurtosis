/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"net"
	"net/netip"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

const (
	enclaveName            = enclave.EnclaveUUID("test-enclave")
	testContainerImageName = "kurtosistech/test-container"

	localhostIPAddrStr       = "127.0.0.1"
	tcpNetworkName           = "tcp4"
	udpNetworkName           = "udp"
	localhostStr             = "localhost"
	availableFreePortKeyStr  = "0"
	availableFreePortAddress = localhostStr + ":" + availableFreePortKeyStr

	tcpPortId = "tcp"
	udpPortId = "udp"
)

var (
	apiContainerInfo = NewApiContainerInfo(
		testIpFromInt(0),
		uint16(1234),
		"0.0.0",
	)
	unusedEnclaveDataDir *enclave_data_directory.EnclaveDataDirectory

	portWaitForTest = port_spec.NewWait(5 * time.Second)
)

func TestAddService_Successful(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp, string(serviceName))
	serviceObj := service.NewService(serviceRegistration, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{}, container.NewContainer(container.ContainerStatus_Running, testContainerImageName, nil, nil, nil))
	serviceConfig := testServiceConfig(t, testContainerImageName)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)

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

	// DestroyUserServices is never being called as everything is successful for this test
	backend.EXPECT().DestroyUserServices(
		ctx,
		enclaveName,
		mock.Anything).Maybe().Times(0)

	startedService, err := network.AddService(ctx, serviceName, serviceConfig)
	require.Nil(t, err)
	require.NotNil(t, startedService)

	require.Equal(t, serviceRegistration, startedService.GetRegistration())

	allServiceRegistration, err := network.serviceRegistrationRepository.GetAll()
	require.NoError(t, err)

	require.Len(t, allServiceRegistration, 1)
	require.Contains(t, allServiceRegistration, serviceName)
	allExistingAndHistoricalIdentifiers, err := network.GetExistingAndHistoricalServiceIdentifiers()
	require.NoError(t, err)
	require.Len(t, allExistingAndHistoricalIdentifiers, 1)
}

func TestAddService_FailedToStart(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	serviceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, serviceIp, string(serviceName))
	serviceConfig := testServiceConfig(t, testContainerImageName)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)

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

	startedService, err := network.AddService(ctx, serviceName, serviceConfig)
	require.NotNil(t, err)
	require.Nil(t, startedService)

	allServiceRegistration, err := network.serviceRegistrationRepository.GetAll()
	require.NoError(t, err)
	require.Empty(t, allServiceRegistration)
	allExistingAndHistoricalIdentifiers, err := network.GetExistingAndHistoricalServiceIdentifiers()
	require.NoError(t, err)
	require.Empty(t, allExistingAndHistoricalIdentifiers)
}

func TestAddServices_Success(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	// One service will be started successfully
	successfulServiceIndex := 1
	successfulServiceName := testServiceNameFromInt(successfulServiceIndex)
	successfulServiceUuid := testServiceUuidFromInt(successfulServiceIndex)
	successfulServiceIp := testIpFromInt(successfulServiceIndex)
	successfulServiceRegistration := service.NewServiceRegistration(successfulServiceName, successfulServiceUuid, enclaveName, successfulServiceIp, string(successfulServiceName))
	successfulService := service.NewService(successfulServiceRegistration, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{}, container.NewContainer(container.ContainerStatus_Running, testContainerImageName, nil, nil, nil))
	successfulServiceConfig := testServiceConfig(t, testContainerImageName)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)

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

	// StartUserService will be called three times, with all the provided services
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		mock.MatchedBy(func(services map[service.ServiceUUID]*service.ServiceConfig) bool {
			// Matcher function returning true iff the services map arg contains exactly the following key:
			// {successfulServiceId}
			_, foundSuccessfulService := services[successfulServiceUuid]
			return len(services) == 1 && foundSuccessfulService
		})).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			successfulServiceUuid: successfulService,
		},
		map[service.ServiceUUID]error{},
		nil)

	success, failure, err := network.AddServices(
		ctx,
		map[service.ServiceName]*service.ServiceConfig{
			successfulServiceName: successfulServiceConfig,
		},
		2,
	)
	require.Nil(t, err)
	require.Len(t, success, 1)
	require.Contains(t, success, successfulServiceName)
	require.Empty(t, failure)

	allServiceRegistration, err := network.serviceRegistrationRepository.GetAll()
	require.NoError(t, err)
	require.Len(t, allServiceRegistration, 1)
	require.Contains(t, allServiceRegistration, successfulServiceName)

	allExistingAndHistoricalIdentifiers, err := network.GetExistingAndHistoricalServiceIdentifiers()
	require.NoError(t, err)
	require.Len(t, allExistingAndHistoricalIdentifiers, 1)
}

func TestAddServices_FailureRollsBackTheEntireBatch(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	// One service will be started successfully
	successfulServiceIndex := 1
	successfulServiceName := testServiceNameFromInt(successfulServiceIndex)
	successfulServiceUuid := testServiceUuidFromInt(successfulServiceIndex)
	successfulServiceIp := testIpFromInt(successfulServiceIndex)
	successfulServiceRegistration := service.NewServiceRegistration(successfulServiceName, successfulServiceUuid, enclaveName, successfulServiceIp, string(successfulServiceName))
	successfulService := service.NewService(successfulServiceRegistration, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{}, container.NewContainer(container.ContainerStatus_Running, testContainerImageName, nil, nil, nil))
	successfulServiceConfig := testServiceConfig(t, testContainerImageName)

	// One service will fail to be started
	failedServiceIndex := 2
	failedServiceName := testServiceNameFromInt(failedServiceIndex)
	failedServiceUuid := testServiceUuidFromInt(failedServiceIndex)
	failedServiceIp := testIpFromInt(failedServiceIndex)
	failedServiceRegistration := service.NewServiceRegistration(failedServiceName, failedServiceUuid, enclaveName, failedServiceIp, string(failedServiceName))
	failedServiceConfig := testServiceConfig(t, testContainerImageName)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)

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

	success, failure, err := network.AddServices(
		ctx,
		map[service.ServiceName]*service.ServiceConfig{
			successfulServiceName: successfulServiceConfig,
			failedServiceName:     failedServiceConfig,
		},
		2,
	)
	require.Nil(t, err)
	require.Empty(t, success) // as the full batch failed, the successful service should have been destroyed
	require.Len(t, failure, 1)
	require.Contains(t, failure, failedServiceName)

	allServiceRegistration, err := network.serviceRegistrationRepository.GetAll()
	require.NoError(t, err)
	require.Empty(t, allServiceRegistration)
	allExistingAndHistoricalIdentifiers, err := network.GetExistingAndHistoricalServiceIdentifiers()
	require.NoError(t, err)
	require.Empty(t, allExistingAndHistoricalIdentifiers)
}

func TestAddServices_FailedToRegisterService(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	// One service will fail to be started
	failedServiceIndex := 1
	failedServiceName := testServiceNameFromInt(failedServiceIndex)
	failedServiceConfig := testServiceConfig(t, testContainerImageName)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)

	backend.EXPECT().RegisterUserServices(
		ctx,
		enclaveName,
		map[service.ServiceName]bool{
			failedServiceName: true,
		},
	).Times(1).Return(
		map[service.ServiceName]*service.ServiceRegistration{},
		map[service.ServiceName]error{
			failedServiceName: errors.New("Service failed to register"),
		},
		nil,
	)

	success, failure, err := network.AddServices(
		ctx,
		map[service.ServiceName]*service.ServiceConfig{
			failedServiceName: failedServiceConfig,
		},
		1,
	)
	require.Nil(t, err)
	require.Empty(t, success) // as the full batch failed, the successful service should have been destroyed
	require.Len(t, failure, 1)
}

func TestStopService_Successful(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp, string(serviceName))
	serviceRegistration.SetStatus(service.ServiceStatus_Started)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)

	err = network.serviceRegistrationRepository.Save(serviceRegistration)
	require.NoError(t, err)

	// The service is registered before being started
	backend.EXPECT().StopUserServices(
		ctx,
		enclaveName,
		&service.ServiceFilters{
			Names: nil,
			UUIDs: map[service.ServiceUUID]bool{
				serviceUuid: true,
			},
			Statuses: nil,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil,
	)

	err = network.StopService(ctx, string(serviceName))
	require.Nil(t, err)

	serviceRegistrationAfterBeingStopped, err := network.serviceRegistrationRepository.Get(serviceName)
	require.NoError(t, err)
	require.Equal(t, serviceRegistrationAfterBeingStopped.GetStatus(), service.ServiceStatus_Stopped)
}

func TestStopService_StopUserServicesFailed(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp, string(serviceName))
	serviceRegistration.SetStatus(service.ServiceStatus_Started)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)
	err = network.serviceRegistrationRepository.Save(serviceRegistration)
	require.NoError(t, err)

	// The service is registered before being started
	backend.EXPECT().StopUserServices(
		ctx,
		enclaveName,
		&service.ServiceFilters{
			Names: nil,
			UUIDs: map[service.ServiceUUID]bool{
				serviceUuid: true,
			},
			Statuses: nil,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{},
		map[service.ServiceUUID]error{
			serviceUuid: stacktrace.NewError("Failed stopping service"),
		},
		nil,
	)

	err = network.StopService(ctx, string(serviceName))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Failed stopping service")
	require.Equal(t, serviceRegistration.GetStatus(), service.ServiceStatus_Started)
}

func TestStopService_ServiceAlreadyStopped(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp, string(serviceName))
	serviceRegistration.SetStatus(service.ServiceStatus_Stopped)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)
	err = network.serviceRegistrationRepository.Save(serviceRegistration)
	require.NoError(t, err)

	// The service is registered before being started
	backend.EXPECT().StopUserServices(
		ctx,
		enclaveName,
		&service.ServiceFilters{
			Names: nil,
			UUIDs: map[service.ServiceUUID]bool{
				serviceUuid: true,
			},
			Statuses: nil,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		map[service.ServiceUUID]error{},
		nil,
	)

	err = network.StopService(ctx, string(serviceName))
	require.Nil(t, err)
	serviceRegistrationAfterBeingStopped, err := network.serviceRegistrationRepository.Get(serviceName)
	require.NoError(t, err)
	require.Equal(t, serviceRegistrationAfterBeingStopped.GetStatus(), service.ServiceStatus_Stopped)
}

func TestStartService_Successful(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp, string(serviceName))
	serviceRegistration.SetStatus(service.ServiceStatus_Stopped)
	serviceConfig := testServiceConfig(t, testContainerImageName)
	serviceRegistration.SetConfig(serviceConfig)
	serviceObj := service.NewService(serviceRegistration, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{}, container.NewContainer(container.ContainerStatus_Running, testContainerImageName, nil, nil, nil))

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)
	err = network.serviceRegistrationRepository.Save(serviceRegistration)
	require.NoError(t, err)

	// The service is registered before being started
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]*service.ServiceConfig{
			serviceUuid: serviceConfig,
		},
	).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			serviceUuid: serviceObj,
		},
		map[service.ServiceUUID]error{},
		nil,
	)

	err = network.StartService(ctx, string(serviceName))
	require.Nil(t, err)
	serviceRegistrationAfterBeingStarted, err := network.serviceRegistrationRepository.Get(serviceName)
	require.NoError(t, err)
	require.Equal(t, serviceRegistrationAfterBeingStarted.GetStatus(), service.ServiceStatus_Started)
}

func TestStartService_StartRegisteredUserServicesFailed(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp, string(serviceName))
	serviceRegistration.SetStatus(service.ServiceStatus_Stopped)
	serviceConfig := testServiceConfig(t, testContainerImageName)
	serviceRegistration.SetConfig(serviceConfig)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)
	err = network.serviceRegistrationRepository.Save(serviceRegistration)
	require.NoError(t, err)

	// The service is registered before being started
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]*service.ServiceConfig{
			serviceUuid: serviceConfig,
		},
	).Times(1).Return(
		map[service.ServiceUUID]*service.Service{},
		map[service.ServiceUUID]error{
			serviceUuid: stacktrace.NewError("Failed starting service"),
		},
		nil,
	)

	err = network.StartService(ctx, string(serviceName))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Failed starting service")
	require.Equal(t, serviceRegistration.GetStatus(), service.ServiceStatus_Stopped)
}

func TestStartService_ServiceAlreadyStarted(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	serviceInternalTestId := 1
	serviceName := testServiceNameFromInt(serviceInternalTestId)
	serviceUuid := testServiceUuidFromInt(serviceInternalTestId)
	successfulServiceIp := testIpFromInt(serviceInternalTestId)
	serviceRegistration := service.NewServiceRegistration(serviceName, serviceUuid, enclaveName, successfulServiceIp, string(serviceName))
	serviceRegistration.SetStatus(service.ServiceStatus_Started)
	serviceConfig := testServiceConfig(t, testContainerImageName)
	serviceRegistration.SetConfig(serviceConfig)
	serviceObj := service.NewService(serviceRegistration, map[string]*port_spec.PortSpec{}, successfulServiceIp, map[string]*port_spec.PortSpec{}, container.NewContainer(container.ContainerStatus_Running, testContainerImageName, nil, nil, nil))

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)
	err = network.serviceRegistrationRepository.Save(serviceRegistration)
	require.NoError(t, err)

	// The service is registered before being started
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]*service.ServiceConfig{
			serviceUuid: serviceConfig,
		},
	).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			serviceUuid: serviceObj,
		},
		map[service.ServiceUUID]error{},
		nil,
	)

	err = network.StartService(ctx, string(serviceName))
	require.Nil(t, err)
	serviceRegistrationAfterBeingStarted, err := network.serviceRegistrationRepository.Get(serviceName)
	require.NoError(t, err)
	require.Equal(t, serviceRegistrationAfterBeingStarted.GetStatus(), service.ServiceStatus_Started)
}

func TestUpdateService(t *testing.T) {
	ctx := context.Background()
	backend := backend_interface.NewMockKurtosisBackend(t)

	file, err := os.CreateTemp("/tmp", "*.db")
	defer os.Remove(file.Name())
	require.Nil(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.Nil(t, err)
	defer db.Close()
	enclaveDb := &enclave_db.EnclaveDB{DB: db}

	network, err := NewDefaultServiceNetwork(
		enclaveName,
		apiContainerInfo,
		backend,
		unusedEnclaveDataDir,
		enclaveDb,
	)
	require.Nil(t, err)

	initialServiceConfig := testServiceConfig(t, testContainerImageName)
	updatedServiceConfig := testServiceConfig(t, "kurtosistech/new-service-image")

	// service that will be successfully updated
	existingServiceIndex := 1
	existingServiceIp := testIpFromInt(existingServiceIndex)
	existingServiceRegistration := service.NewServiceRegistration(
		testServiceNameFromInt(existingServiceIndex),
		testServiceUuidFromInt(existingServiceIndex),
		enclaveName,
		existingServiceIp,
		testServiceHostnameFromInt(existingServiceIndex))
	existingServiceRegistration.SetConfig(initialServiceConfig)
	existingServiceRegistration.SetStatus(service.ServiceStatus_Started)
	err = network.serviceRegistrationRepository.Save(existingServiceRegistration)
	require.NoError(t, err)

	// service that will fail because it's not registered in the enclave prior to update
	unknownServiceIndex := 2
	unknownServiceIp := testIpFromInt(unknownServiceIndex)
	unknownServiceRegistration := service.NewServiceRegistration(
		testServiceNameFromInt(unknownServiceIndex),
		testServiceUuidFromInt(unknownServiceIndex),
		enclaveName,
		unknownServiceIp,
		testServiceHostnameFromInt(unknownServiceIndex))

	// service that will fail to be removed from the enclave
	failedToBeRemovedServiceIndex := 3
	failedToBeRemovedServiceIp := testIpFromInt(unknownServiceIndex)
	failedToBeRemovedServiceRegistration := service.NewServiceRegistration(
		testServiceNameFromInt(failedToBeRemovedServiceIndex),
		testServiceUuidFromInt(failedToBeRemovedServiceIndex),
		enclaveName,
		failedToBeRemovedServiceIp,
		testServiceHostnameFromInt(failedToBeRemovedServiceIndex))
	failedToBeRemovedServiceRegistration.SetConfig(initialServiceConfig)
	failedToBeRemovedServiceRegistration.SetStatus(service.ServiceStatus_Started)
	err = network.serviceRegistrationRepository.Save(failedToBeRemovedServiceRegistration)
	require.NoError(t, err)

	// service that will fail to be re-created once it has been removed
	failedToBeRecreatedServiceIndex := 4
	failedToBeRecreatedServiceIp := testIpFromInt(unknownServiceIndex)
	failedToBeRecreatedServiceRegistration := service.NewServiceRegistration(
		testServiceNameFromInt(failedToBeRecreatedServiceIndex),
		testServiceUuidFromInt(failedToBeRecreatedServiceIndex),
		enclaveName,
		failedToBeRecreatedServiceIp,
		testServiceHostnameFromInt(failedToBeRecreatedServiceIndex))
	failedToBeRecreatedServiceRegistration.SetConfig(initialServiceConfig)
	failedToBeRecreatedServiceRegistration.SetStatus(service.ServiceStatus_Started)
	err = network.serviceRegistrationRepository.Save(failedToBeRecreatedServiceRegistration)
	require.NoError(t, err)

	// The service will be removed first
	backend.EXPECT().RemoveRegisteredUserServiceProcesses(
		ctx,
		enclaveName,
		map[service.ServiceUUID]bool{
			existingServiceRegistration.GetUUID():            true,
			failedToBeRemovedServiceRegistration.GetUUID():   true,
			failedToBeRecreatedServiceRegistration.GetUUID(): true,
		},
	).Times(1).Return(
		map[service.ServiceUUID]bool{
			existingServiceRegistration.GetUUID():            true,
			failedToBeRecreatedServiceRegistration.GetUUID(): true,
		},
		map[service.ServiceUUID]error{
			failedToBeRemovedServiceRegistration.GetUUID(): stacktrace.NewError("Unable to remove service"),
		},
		nil,
	)

	// The services will then be re-created
	serviceObj := service.NewService(existingServiceRegistration, map[string]*port_spec.PortSpec{}, existingServiceIp, map[string]*port_spec.PortSpec{}, container.NewContainer(container.ContainerStatus_Running, "", nil, nil, nil))
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]*service.ServiceConfig{
			existingServiceRegistration.GetUUID(): updatedServiceConfig,
		},
	).Times(1).Return(
		map[service.ServiceUUID]*service.Service{
			existingServiceRegistration.GetUUID(): serviceObj,
		},
		map[service.ServiceUUID]error{},
		nil,
	)
	backend.EXPECT().StartRegisteredUserServices(
		ctx,
		enclaveName,
		map[service.ServiceUUID]*service.ServiceConfig{
			failedToBeRecreatedServiceRegistration.GetUUID(): updatedServiceConfig,
		},
	).Times(1).Return(
		map[service.ServiceUUID]*service.Service{},
		map[service.ServiceUUID]error{
			failedToBeRecreatedServiceRegistration.GetUUID(): stacktrace.NewError("Unable to re-create service"),
		},
		nil,
	)

	success, failure, err := network.UpdateServices(ctx, map[service.ServiceName]*service.ServiceConfig{
		existingServiceRegistration.GetName():            updatedServiceConfig,
		unknownServiceRegistration.GetName():             updatedServiceConfig,
		failedToBeRemovedServiceRegistration.GetName():   updatedServiceConfig,
		failedToBeRecreatedServiceRegistration.GetName(): updatedServiceConfig,
	}, 1)
	require.Nil(t, err)
	require.Len(t, success, 1)
	require.Contains(t, success, existingServiceRegistration.GetName())

	require.Len(t, failure, 3)
	require.Contains(t, failure, unknownServiceRegistration.GetName())
	require.Contains(t, failure, failedToBeRemovedServiceRegistration.GetName())
	require.Contains(t, failure, failedToBeRecreatedServiceRegistration.GetName())

	newExistingServiceRegistration, err := network.serviceRegistrationRepository.Get(existingServiceRegistration.GetName())
	require.NoError(t, err)
	require.NotNil(t, newExistingServiceRegistration)

	require.Equal(t, newExistingServiceRegistration.GetStatus(), service.ServiceStatus_Started)
	require.Equal(t, newExistingServiceRegistration.GetConfig(), updatedServiceConfig)

	exist, err := network.serviceRegistrationRepository.Exist(unknownServiceRegistration.GetName())
	require.NoError(t, err)
	require.False(t, exist)

	newFailedToBeRemovedServiceRegistration, err := network.serviceRegistrationRepository.Get(failedToBeRemovedServiceRegistration.GetName())
	require.NoError(t, err)
	require.NotNil(t, newFailedToBeRemovedServiceRegistration)
	require.Equal(t, newFailedToBeRemovedServiceRegistration.GetStatus(), service.ServiceStatus_Started)
	require.Equal(t, newFailedToBeRemovedServiceRegistration.GetConfig(), initialServiceConfig)

	newFailedToBeRecreatedServiceRegistration, err := network.serviceRegistrationRepository.Get(failedToBeRecreatedServiceRegistration.GetName())
	require.NoError(t, err)
	require.NotNil(t, newFailedToBeRecreatedServiceRegistration)
	require.Equal(t, newFailedToBeRecreatedServiceRegistration.GetStatus(), service.ServiceStatus_Registered)
	require.Nil(t, newFailedToBeRecreatedServiceRegistration.GetConfig())
}

func TestScanPort(t *testing.T) {
	localhost := net.ParseIP(localhostIPAddrStr)

	tcpAddrPort, udpAddrPort, closeOpenedPortsFunc, err := openFreeTCPAndUDPLocalHostPortAddressesForTesting()
	require.NoError(t, err)
	defer func() {
		err = closeOpenedPortsFunc()
		require.NoError(t, err)
	}()

	tcpPortSpec, err := port_spec.NewPortSpec(tcpAddrPort.Port(), port_spec.TransportProtocol_TCP, "", portWaitForTest, "")
	require.NoError(t, err)

	scanPortTimeout := 5 * time.Second

	err = scanPort(localhost, tcpPortSpec, scanPortTimeout)
	require.NoError(t, err)

	udpPortSpec, err := port_spec.NewPortSpec(udpAddrPort.Port(), port_spec.TransportProtocol_UDP, "", portWaitForTest, "")
	require.NoError(t, err)

	err = scanPort(localhost, udpPortSpec, scanPortTimeout)
	require.NoError(t, err)
}

func TestWaitUntilAllTCPAndUDPPortsAreOpen_Success(t *testing.T) {
	localhost := net.ParseIP(localhostIPAddrStr)

	tcpAddrPort, udpAddrPort, closeOpenedPortsFunc, err := openFreeTCPAndUDPLocalHostPortAddressesForTesting()
	require.NoError(t, err)
	defer func() {
		err = closeOpenedPortsFunc()
		require.NoError(t, err)
	}()

	tcpPortSpec, err := port_spec.NewPortSpec(tcpAddrPort.Port(), port_spec.TransportProtocol_TCP, "", portWaitForTest, "")
	require.NoError(t, err)

	udpPortSpec, err := port_spec.NewPortSpec(udpAddrPort.Port(), port_spec.TransportProtocol_UDP, "", portWaitForTest, "")
	require.NoError(t, err)

	ports := map[string]*port_spec.PortSpec{
		tcpPortId: tcpPortSpec,
		udpPortId: udpPortSpec,
	}

	err = waitUntilAllTCPAndUDPPortsAreOpen(localhost, ports)
	require.NoError(t, err)
}

func TestWaitUntilAllTCPAndUDPPortsAreOpen_Fails(t *testing.T) {
	localhost := net.ParseIP(localhostIPAddrStr)

	closedPortId := "closed-port-for-testing"
	closedPortNumber := uint16(42821)

	tcpAddrPort, udpAddrPort, closeOpenedPortsFunc, err := openFreeTCPAndUDPLocalHostPortAddressesForTesting()
	require.NoError(t, err)
	defer func() {
		err = closeOpenedPortsFunc()
		require.NoError(t, err)
	}()

	tcpPortSpec, err := port_spec.NewPortSpec(tcpAddrPort.Port(), port_spec.TransportProtocol_TCP, "", portWaitForTest, "")
	require.NoError(t, err)

	udpPortSpec, err := port_spec.NewPortSpec(udpAddrPort.Port(), port_spec.TransportProtocol_UDP, "", portWaitForTest, "")
	require.NoError(t, err)

	closedPortSpec, err := port_spec.NewPortSpec(closedPortNumber, port_spec.TransportProtocol_TCP, "", portWaitForTest, "")
	require.NoError(t, err)

	ports := map[string]*port_spec.PortSpec{
		tcpPortId:    tcpPortSpec,
		udpPortId:    udpPortSpec,
		closedPortId: closedPortSpec,
	}

	expectedErrorMsg := "An error occurred while waiting for all TCP and UDP ports to be open"

	err = waitUntilAllTCPAndUDPPortsAreOpen(localhost, ports)
	require.Error(t, err)
	require.ErrorContains(t, err, expectedErrorMsg)
}

func openFreeTCPAndUDPLocalHostPortAddressesForTesting() (*netip.AddrPort, *netip.AddrPort, func() error, error) {
	availableTCPAddress, err := net.ResolveTCPAddr(tcpNetworkName, availableFreePortAddress)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred resolving TCP address '%s' in network '%s'",
			availableFreePortAddress,
			tcpNetworkName,
		)
	}

	tcpListener, err := net.ListenTCP(tcpNetworkName, availableTCPAddress)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred listening TCP address '%s' in network '%s'",
			availableTCPAddress,
			tcpNetworkName,
		)
	}
	shouldCloseTCPListener := true
	defer func() {
		if shouldCloseTCPListener {
			if err := tcpListener.Close(); err != nil {
				logrus.Warnf("We tried to close TCP address '%v' we opened for testing purpose but something fails, you should manually close it", availableTCPAddress)
			}
		}
	}()

	tcpAddressPort, err := netip.ParseAddrPort(tcpListener.Addr().String())
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred parsing TCP address '%s'",
			tcpListener.Addr(),
		)
	}

	availableUDPAddress, err := net.ResolveUDPAddr(udpNetworkName, availableFreePortAddress)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred resolving UDP address '%s' in network '%s'",
			availableFreePortAddress,
			udpNetworkName,
		)
	}

	udpListener, err := net.ListenUDP(udpNetworkName, availableUDPAddress)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred listening UDP address '%s' in network '%s'",
			availableUDPAddress,
			udpNetworkName,
		)
	}
	shouldCloseUDPListener := true
	defer func() {
		if shouldCloseUDPListener {
			if err := udpListener.Close(); err != nil {
				logrus.Warnf("We tried to close UDP address '%v' we opened for testing purpose but something fails, you should manually close it", availableUDPAddress)
			}
		}
	}()

	udpAddressPort, err := netip.ParseAddrPort(udpListener.LocalAddr().String())
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred parsing UDP address '%s'",
			udpListener.LocalAddr(),
		)
	}

	closeBothListenersFunc := func() error {
		if err := tcpListener.Close(); err != nil {
			logrus.Warnf("We tried to close TCP address '%v' we opened for testing purpose but something fails, you should manually close it", availableTCPAddress)
		}

		if err := udpListener.Close(); err != nil {
			logrus.Warnf("We tried to close UDP address '%v' we opened for testing purpose but something fails, you should manually close it", availableUDPAddress)
		}
		return nil
	}

	shouldCloseTCPListener = false
	shouldCloseUDPListener = false
	return &tcpAddressPort, &udpAddressPort, closeBothListenersFunc, nil
}

func testServiceConfig(t *testing.T, imageName string) *service.ServiceConfig {
	serviceConfig, err := service.CreateServiceConfig(imageName, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, "", 0, 0, map[string]string{}, nil, nil, map[string]string{}, image_download_mode.ImageDownloadMode_Missing, true, false, []string{})
	require.NoError(t, err)
	return serviceConfig
}

func testIpFromInt(i int) net.IP {
	return []byte{1, 1, 1, byte(i)}
}

func testServiceNameFromInt(i int) service.ServiceName {
	return service.ServiceName("service-" + strconv.Itoa(i))
}

func testServiceHostnameFromInt(i int) string {
	return "service-" + strconv.Itoa(i)
}

func testServiceUuidFromInt(i int) service.ServiceUUID {
	return service.ServiceUUID(fmt.Sprintf(
		"massive-uuid-with-32-req-chars-%v",
		i,
	))
}
