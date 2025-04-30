/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/service_registration"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_identifiers"
	path_compression "github.com/kurtosis-tech/kurtosis/path-compression"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	minMemoryLimit              uint64 = 6 // Docker doesn't allow memory limits less than 6 megabytes
	defaultMemoryAllocMegabytes uint64 = 0

	tempDirForRenderedTemplatesPrefix = "temp-dir-for-rendered-templates-"

	enforceMaxFileSizeLimit = false

	exactlyOneShortenedUuidMatch = 1

	singleServiceStartupBatch = 1

	waitForPortsOpenRetriesDelayMilliseconds = 500

	shouldFollowLogs = false

	publicPortsSuffix = "-public"

	serviceLogsHeader = "== SERVICE '%s' LOGS ==================================="
	serviceLogsFooter = "== FINISHED SERVICE '%s' LOGS ==================================="

	scanPortTimeout = 200 * time.Millisecond
)

type storeFilesArtifactResult struct {
	err               error
	filesArtifactUuid enclave_data_directory.FilesArtifactUUID
}

// DefaultServiceNetwork is the in-memory representation of the service network that the API container will manipulate.
// To make any changes to the test network, this struct must be used.
type DefaultServiceNetwork struct {
	enclaveUuid enclave.EnclaveUUID

	apiContainerInfo *ApiContainerInfo

	mutex *sync.Mutex // VERY IMPORTANT TO CHECK AT THE START OF EVERY METHOD!

	kurtosisBackend backend_interface.KurtosisBackend

	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory

	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository

	// This contains all service identifiers ever successfully created
	serviceIdentifiersRepository *service_identifiers.ServiceIdentifiersRepository
}

func NewDefaultServiceNetwork(
	enclaveUuid enclave.EnclaveUUID,
	apiContainerInfo *ApiContainerInfo,
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory,
	enclaveDb *enclave_db.EnclaveDB,
) (*DefaultServiceNetwork, error) {
	serviceIdentifiersRepository, err := service_identifiers.GetOrCreateNewServiceIdentifiersRepository(enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service identifiers repository")
	}
	serviceRegistrationRepository, err := service_registration.GetOrCreateNewServiceRegistrationRepository(enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the service registration repository")
	}

	return &DefaultServiceNetwork{
		enclaveUuid:      enclaveUuid,
		apiContainerInfo: apiContainerInfo,
		mutex:            &sync.Mutex{},

		kurtosisBackend: kurtosisBackend,
		enclaveDataDir:  enclaveDataDir,

		serviceRegistrationRepository: serviceRegistrationRepository,
		serviceIdentifiersRepository:  serviceIdentifiersRepository,
	}, nil
}

// AddService creates and starts the service in their own container
func (network *DefaultServiceNetwork) AddService(
	ctx context.Context,
	serviceName service.ServiceName,
	serviceConfig *service.ServiceConfig,
) (
	*service.Service,
	error,
) {
	serviceConfigMap := map[service.ServiceName]*service.ServiceConfig{
		serviceName: serviceConfig,
	}

	startedServices, serviceFailed, err := network.AddServices(ctx, serviceConfigMap, singleServiceStartupBatch)
	if err != nil {
		return nil, err
	}
	if failure, found := serviceFailed[serviceName]; found {
		return nil, failure
	}
	if startedService, found := startedServices[serviceName]; found {
		return startedService, nil
	}
	return nil, stacktrace.NewError("Service '%s' could not be started properly, but its state is unknown. This is a Kurtosis internal bug", serviceName)
}

// AddServices creates and starts the services in their own containers. It is a bulk operation, if a
// single service fails to start, the entire batch is rolled back.
//
// This is a bulk operation that follows a sequential approach with no parallelization yet.
// This function returns:
//   - successfulService - mapping of successful service ids to service objects with info about that service when the
//     entire batch of service could be started
//   - failedServices - mapping of failed service ids to errors causing those failures. As noted above, successful
//     services will be rolled back.
//   - error - when a broad and unexpected error happened.
func (network *DefaultServiceNetwork) AddServices(
	ctx context.Context,
	serviceConfigs map[service.ServiceName]*service.ServiceConfig,
	batchSize int,
) (
	map[service.ServiceName]*service.Service,
	map[service.ServiceName]error,
	error,
) {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	batchSuccessfullyStarted := false
	startedServices := map[service.ServiceName]*service.Service{}
	failedServices := map[service.ServiceName]error{}

	if len(serviceConfigs) == 0 {
		return startedServices, failedServices, nil
	}

	// Save the services currently running in enclave for later
	currentlyRunningServicesInEnclave := map[service.ServiceName]bool{}
	allServiceNamesFromServiceRegistrations, err := network.serviceRegistrationRepository.GetAllServiceNames()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting all service names from service registration repository")
	}
	for serviceName := range allServiceNamesFromServiceRegistrations {
		currentlyRunningServicesInEnclave[serviceName] = true
	}

	// We register all the services one by one
	serviceSuccessfullyRegistered := map[service.ServiceName]*service.ServiceRegistration{}
	servicesToStart := map[service.ServiceUUID]*service.ServiceConfig{}
	for serviceName, serviceConfig := range serviceConfigs {
		serviceRegistration, err := network.registerService(ctx, serviceName)
		if err != nil {
			failedServices[serviceName] = stacktrace.Propagate(err, "Failed registering service with name: '%s'", serviceName)
			continue
		}
		serviceSuccessfullyRegistered[serviceName] = serviceRegistration
		servicesToStart[serviceRegistration.GetUUID()] = serviceConfig
	}
	defer func() {
		if batchSuccessfullyStarted {
			return
		}
		for serviceName := range serviceSuccessfullyRegistered {
			if err := network.unregisterService(ctx, serviceName); err != nil {
				logrus.Errorf("Error unregistering service '%s' from the service network. Error was: %v", serviceName, err)
			}
		}
	}()
	if len(failedServices) > 0 {
		return map[service.ServiceName]*service.Service{}, failedServices, nil
	}

	startedServicesPerUuid, failedServicePerUuid := network.startRegisteredServices(ctx, servicesToStart, batchSize)

	for serviceName, serviceRegistration := range serviceSuccessfullyRegistered {
		serviceUuid := serviceRegistration.GetUUID()
		if serviceStartFailure, found := failedServicePerUuid[serviceUuid]; found {
			failedServices[serviceName] = serviceStartFailure
			continue
		}
		if startedService, found := startedServicesPerUuid[serviceUuid]; found {
			startedServices[serviceName] = startedService
			continue
		}
		if len(failedServicePerUuid) == 0 {
			// it is expected that if a service failed to be started, Kurtosis did not even try to start the others
			// and stopped midway. However, if failedServicePerUuid is empty, this is an internal bug
			failedServices[serviceName] = stacktrace.NewError("State of the service is unknown: %s. This is a Kurtosis internal bug", serviceName)
		}
	}
	defer func() {
		if batchSuccessfullyStarted {
			return
		}
		for serviceName, startedService := range startedServices {
			if err := network.destroyService(ctx, serviceName, startedService.GetRegistration().GetUUID()); err != nil {
				logrus.Errorf("One or more services failed to be started for this batch. Kurtosis tries to"+
					"roll back the entire batch, but failed destroying some services. Error was: %v", err)
			}
		}
	}()
	if len(failedServices) > 0 {
		return map[service.ServiceName]*service.Service{}, failedServices, nil
	}

	if len(startedServices) != len(serviceConfigs) {
		var requested []service.ServiceName
		for serviceName := range serviceConfigs {
			requested = append(requested, serviceName)
		}
		var result []service.ServiceName
		for serviceName := range startedServices {
			result = append(result, serviceName)
		}
		return nil, nil, stacktrace.NewError("This is a Kurtosis internal bug. The batch of services being started does not fit the number of services that were requested. (service started: '%v', requested: '%v')", result, requested)
	}

	for _, startedService := range startedServices {
		serviceRegistration := startedService.GetRegistration()
		serviceIdentifier := service_identifiers.NewServiceIdentifier(serviceRegistration.GetUUID(), serviceRegistration.GetName())
		if err := network.serviceIdentifiersRepository.AddServiceIdentifier(serviceIdentifier); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred adding a new service identifier '%+v' into the repository", serviceIdentifier)
		}
		serviceName := serviceRegistration.GetName()
		serviceStatus := service.ServiceStatus_Started
		if err := network.serviceRegistrationRepository.UpdateStatus(serviceName, serviceStatus); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred while updating service status to '%s' in service registration for service '%s' after the service was started", serviceStatus, serviceName)
		}
	}

	batchSuccessfullyStarted = true
	return startedServices, map[service.ServiceName]error{}, nil
}

func (network *DefaultServiceNetwork) UpdateService(ctx context.Context, serviceName service.ServiceName, updateServiceConfig *service.ServiceConfig) (*service.Service, error) {
	serviceConfigMap := map[service.ServiceName]*service.ServiceConfig{
		serviceName: updateServiceConfig,
	}

	startedServices, serviceFailed, err := network.UpdateServices(ctx, serviceConfigMap, singleServiceStartupBatch)
	if err != nil {
		return nil, err
	}
	if failure, found := serviceFailed[serviceName]; found {
		return nil, failure
	}
	if startedService, found := startedServices[serviceName]; found {
		return startedService, nil
	}
	return nil, stacktrace.NewError("Service '%s' could not be updated properly, and its state is unknown. This is a Kurtosis internal bug", serviceName)
}

// UpdateServices updates the service by removing the current container and re-creating it, keeping the registration
// identical. Note this function does not handle any kind of rollback if it fails halfway. This is because we have no
// way to do soft-delete for containers. Once it's deleted, it's gone, so if Kurtosis fails at re-creating it, it
// doesn't roll back to the state previous to calling this function
func (network *DefaultServiceNetwork) UpdateServices(ctx context.Context, updateServiceConfigs map[service.ServiceName]*service.ServiceConfig, batchSize int) (map[service.ServiceName]*service.Service, map[service.ServiceName]error, error) {
	network.mutex.Lock()
	defer network.mutex.Unlock()
	failedServicesPool := map[service.ServiceName]error{}
	successfullyUpdatedService := map[service.ServiceName]*service.Service{}

	if len(updateServiceConfigs) == 0 {
		return successfullyUpdatedService, failedServicesPool, nil
	}

	// First, remove the service
	serviceUuidToNameMap := map[service.ServiceUUID]service.ServiceName{}
	serviceUuidsToRemove := map[service.ServiceUUID]bool{}
	for serviceName := range updateServiceConfigs {
		serviceRegistration, err := network.serviceRegistrationRepository.Get(serviceName)
		if err != nil {
			failedServicesPool[serviceName] = stacktrace.Propagate(err, "Unable to update service that is not registered "+
				"inside this enclave: '%s'", serviceName)
		} else {
			serviceUuid := serviceRegistration.GetUUID()
			serviceUuidsToRemove[serviceUuid] = true
			serviceUuidToNameMap[serviceUuid] = serviceName
		}
	}
	successfullyRemovedServices, failedRemovedServices, err := network.kurtosisBackend.RemoveRegisteredUserServiceProcesses(ctx, network.enclaveUuid, serviceUuidsToRemove)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error happened updating services")
	}
	for serviceUuid, serviceErr := range failedRemovedServices {
		if serviceName, found := serviceUuidToNameMap[serviceUuid]; found {
			failedServicesPool[serviceName] = serviceErr
		} else {
			return nil, nil, stacktrace.NewError("Error mapping service UUID to service name. This is a bug in Kurtosis.\nserviceUuidsToRemove=%v\nfailedRemovedServices=%v\nsuccessfullyRemovedServices=%v\nserviceUuidToNameMap=%v", serviceUuidsToRemove, failedRemovedServices, successfullyRemovedServices, serviceUuidToNameMap)
		}
	}

	// Set service status back to registered and remove its currently saved service config
	successfullyRemovedServicesIncludingSidecars := map[service.ServiceUUID]bool{}
	for serviceUuid := range successfullyRemovedServices {
		if serviceName, found := serviceUuidToNameMap[serviceUuid]; found {
			serviceStatus := service.ServiceStatus_Registered
			if err := network.serviceRegistrationRepository.UpdateStatusAndConfig(serviceName, serviceStatus, nil); err != nil {
				failedServicesPool[serviceName] = stacktrace.Propagate(err, "An error occurred while cleaning the configuration and updating service status to '%s' into service registration fro service '%s' after this service was removed successfully", serviceStatus, serviceName)
				continue
			}
			successfullyRemovedServicesIncludingSidecars[serviceUuid] = true
		} else {
			return nil, nil, stacktrace.NewError("Error mapping service UUID to service name. This is a bug in Kurtosis")
		}
	}

	// Re-create service with the new service config
	serviceToRecreate := map[service.ServiceUUID]*service.ServiceConfig{}
	for serviceUuid := range successfullyRemovedServicesIncludingSidecars {
		serviceName, found := serviceUuidToNameMap[serviceUuid]
		if !found {
			failedServicesPool[serviceName] = stacktrace.NewError("Unable to update service that is not registered "+
				"inside this enclave: '%s'", serviceName)
			continue
		}
		newServiceConfig, found := updateServiceConfigs[serviceName]
		if !found {
			failedServicesPool[serviceName] = stacktrace.NewError("Unable to update service '%s' because no new "+
				"service config could be found. This is a bug in Kurtosis", serviceName)
			continue
		}
		serviceToRecreate[serviceUuid] = newServiceConfig
	}
	recreatedService, failedToRecreateService := network.startRegisteredServices(ctx, serviceToRecreate, batchSize)
	for serviceUuid, failedToRecreateServiceErr := range failedToRecreateService {
		serviceName, found := serviceUuidToNameMap[serviceUuid]
		if !found {
			failedServicesPool[serviceName] = stacktrace.NewError("Unable to update service that is not registered "+
				"inside this enclave: '%s'", serviceName)
			continue
		}
		failedServicesPool[serviceName] = failedToRecreateServiceErr
	}
	for serviceUuid, newServiceObj := range recreatedService {
		serviceName, found := serviceUuidToNameMap[serviceUuid]
		if !found {
			failedServicesPool[serviceName] = stacktrace.NewError("Unable to update service that is not registered "+
				"inside this enclave: '%s'", serviceName)
			continue
		}
		serviceStatus := service.ServiceStatus_Started
		if err := network.serviceRegistrationRepository.UpdateStatus(serviceName, serviceStatus); err != nil {
			failedServicesPool[serviceName] = stacktrace.Propagate(err, "An error occurred while updating service status to '%s' in service registration for service '%s' after the service was updated", serviceStatus, serviceName)
			continue
		}
		successfullyUpdatedService[serviceName] = newServiceObj
	}
	return successfullyUpdatedService, failedServicesPool, nil
}

func (network *DefaultServiceNetwork) RemoveService(
	ctx context.Context,
	serviceIdentifier string,
) (service.ServiceUUID, error) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while fetching name for service identifier '%v'", serviceIdentifier)
	}

	serviceToRemove, err := network.serviceRegistrationRepository.Get(serviceName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the service registration for service '%s'", serviceName)
	}
	serviceUuid := serviceToRemove.GetUUID()

	// We stop the service, rather than destroying it, so that we can keep logs around
	stopServiceFilters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		Statuses: nil,
	}
	_, erroredUuids, err := network.kurtosisBackend.StopUserServices(ctx, network.enclaveUuid, stopServiceFilters)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred during the call to stop service '%v'", serviceUuid)
	}
	if err, found := erroredUuids[serviceUuid]; found {
		return "", stacktrace.Propagate(err, "An error occurred stopping service '%v'", serviceUuid)
	}

	if err := network.serviceRegistrationRepository.Delete(serviceName); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred deleting the service registration for service '%v' from the repository", serviceName)
	}

	return serviceUuid, nil
}

func (network *DefaultServiceNetwork) StartService(
	ctx context.Context,
	serviceIdentifier string,
) error {
	serviceIdentifiers := []string{serviceIdentifier}
	_, erroredUuids, err := network.StartServices(ctx, serviceIdentifiers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while starting services")
	}

	for serviceUuid, erroredUuid := range erroredUuids {
		return stacktrace.Propagate(erroredUuid, "An error occurred while starting service '%v'", serviceUuid)
	}

	return nil
}

func (network *DefaultServiceNetwork) StartServices(
	ctx context.Context,
	serviceIdentifiers []string,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	successfulUuids := map[service.ServiceUUID]bool{}
	erroredUuids := map[service.ServiceUUID]error{}
	serviceConfigs := map[service.ServiceUUID]*service.ServiceConfig{}
	serviceRegistrations := map[service.ServiceUUID]*service.ServiceRegistration{}

	for _, serviceIdentifier := range serviceIdentifiers {
		serviceRegistration, err := network.getServiceRegistrationForIdentifierUnlocked(serviceIdentifier)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred while getting service registration for identifier '%v'", serviceIdentifier)
		}
		serviceRegistrations[serviceRegistration.GetUUID()] = serviceRegistration
	}

	for serviceUuid, serviceRegistration := range serviceRegistrations {
		serviceConfigs[serviceUuid] = serviceRegistration.GetConfig()
	}

	successfulServices, failedServices, err := network.kurtosisBackend.StartRegisteredUserServices(ctx, network.enclaveUuid, serviceConfigs)
	if err != nil {
		return nil, nil, err
	}

	for successfulUuid, successfulService := range successfulServices {
		serviceName := successfulService.GetRegistration().GetName()
		serviceStatus := service.ServiceStatus_Started
		if err := network.serviceRegistrationRepository.UpdateStatus(serviceName, serviceStatus); err != nil {
			failedServices[successfulUuid] = stacktrace.Propagate(err, "An error occurred while updating status to '%v' for service '%v' after it was successfully started", serviceStatus, serviceName)
			continue
		}
		successfulUuids[successfulUuid] = true
	}

	for erroredUuid, err := range failedServices {
		erroredUuids[erroredUuid] = err
	}

	return successfulUuids, erroredUuids, nil
}

func (network *DefaultServiceNetwork) StopService(
	ctx context.Context,
	serviceIdentifier string,
) error {
	serviceIdentifiers := []string{serviceIdentifier}
	_, erroredUuids, err := network.StopServices(ctx, serviceIdentifiers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while stopping services")
	}

	for serviceUuid, erroredUuid := range erroredUuids {
		return stacktrace.Propagate(erroredUuid, "An error occurred while stopping service '%v'", serviceUuid)
	}

	return nil
}

func (network *DefaultServiceNetwork) StopServices(
	ctx context.Context,
	serviceIdentifiers []string,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceUuids := map[service.ServiceUUID]bool{}
	serviceNamesByUuid := map[service.ServiceUUID]service.ServiceName{}

	for _, serviceIdentifier := range serviceIdentifiers {
		serviceRegistration, err := network.getServiceRegistrationForIdentifierUnlocked(serviceIdentifier)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred while getting service registration for identifier '%v'", serviceIdentifier)
		}
		serviceUuids[serviceRegistration.GetUUID()] = true
		serviceNamesByUuid[serviceRegistration.GetUUID()] = serviceRegistration.GetName()
	}

	stopServiceFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    serviceUuids,
		Statuses: nil,
	}
	successfulUuids, erroredUuids, err := network.kurtosisBackend.StopUserServices(ctx, network.enclaveUuid, stopServiceFilters)
	if err != nil {
		return successfulUuids, erroredUuids, stacktrace.Propagate(err, "An error occurred during the call to stop services")
	}

	for successfulUuid := range successfulUuids {
		serviceName, found := serviceNamesByUuid[successfulUuid]
		if !found {
			erroredUuids[successfulUuid] = stacktrace.NewError("Expected to find service UUID '%v' in map '%+v' after the service was successfully stopped, but it was not found; this is a bug in Kurtosis", successfulUuid, serviceNamesByUuid)
			delete(successfulUuids, successfulUuid)
			continue
		}
		serviceStatus := service.ServiceStatus_Stopped
		if err := network.serviceRegistrationRepository.UpdateStatus(serviceName, serviceStatus); err != nil {
			erroredUuids[successfulUuid] = stacktrace.Propagate(err, "An error occurred while updating status to '%v' for service '%v' after it was successfully stopped", serviceStatus, serviceName)
			delete(successfulUuids, successfulUuid)
			continue
		}
	}

	return successfulUuids, erroredUuids, nil
}

func (network *DefaultServiceNetwork) RunExec(ctx context.Context, serviceIdentifier string, userServiceCommand []string) (*exec_result.ExecResult, error) {
	// NOTE: This will block all other operations while this command is running!!!! We might need to change this so it's
	// asynchronous
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceRegistration, err := network.getServiceRegistrationForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting service registration for identifier '%v'", serviceIdentifier)
	}

	serviceUuid := serviceRegistration.GetUUID()
	userServiceCommands := map[service.ServiceUUID][]string{
		serviceUuid: userServiceCommand,
	}

	successfulExecCommands, failedExecCommands, err := network.kurtosisBackend.RunUserServiceExecCommands(
		ctx, network.enclaveUuid, "", userServiceCommands)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v'",
			userServiceCommand,
			serviceIdentifier)
	}

	if execResult, found := successfulExecCommands[serviceUuid]; found {
		if len(failedExecCommands) > 0 {
			return nil, stacktrace.NewError("An error was returned even though the exec command was successful. "+
				"This is a Kurtosis internal bug. The exec result was: '%s' (exit code %d) and the error(s) were:\n%v",
				execResult.GetOutput(), execResult.GetExitCode(), failedExecCommands)
		}
		return execResult, nil
	}

	if err, found := failedExecCommands[serviceUuid]; found {
		return nil, stacktrace.Propagate(err, "An error occurred running exec command '%v' on service '%s' "+
			"(uuid '%s')", userServiceCommand, serviceIdentifier, serviceUuid)
	}
	return nil, stacktrace.NewError("The status of the exec command '%v' on service '%s' (uuid '%s') is unknown. "+
		"It did not return as a success nor as a failure. This is a Kurtosis internal bug.",
		userServiceCommand, serviceIdentifier, serviceUuid)
}

func (network *DefaultServiceNetwork) RunExecs(ctx context.Context, userServiceCommands map[string][]string) (map[service.ServiceUUID]*exec_result.ExecResult, map[service.ServiceUUID]error, error) {
	// NOTE: This will block all other operations while this command is running!!!! We might need to change this so it's
	// asynchronous
	network.mutex.Lock()
	defer network.mutex.Unlock()

	userServiceCommandsByServiceUuid := map[service.ServiceUUID][]string{}
	for serviceIdentifier, userServiceCommand := range userServiceCommands {
		serviceRegistration, err := network.getServiceRegistrationForIdentifierUnlocked(serviceIdentifier)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred while getting service registration for identifier '%v'", serviceIdentifier)
		}
		serviceUuid := serviceRegistration.GetUUID()
		userServiceCommandsByServiceUuid[serviceUuid] = userServiceCommand
	}

	successfulExecs, failedExecs, err := network.kurtosisBackend.RunUserServiceExecCommands(ctx, network.enclaveUuid, "", userServiceCommandsByServiceUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An unexpected error occurred running multiple exec commands "+
			"on user services:\n%v", userServiceCommands)
	}
	return successfulExecs, failedExecs, nil
}

func (network *DefaultServiceNetwork) HttpRequestService(ctx context.Context, serviceIdentifier string, portId string, method string, contentType string, endpoint string, body string, headers map[string]string) (*http.Response, error) {
	logrus.Debugf("Making a request '%v' '%v' '%v' '%v' '%v' '%v'", serviceIdentifier, portId, method, contentType, endpoint, body)
	userService, getServiceErr := network.GetService(ctx, serviceIdentifier)
	if getServiceErr != nil {
		return nil, stacktrace.Propagate(getServiceErr, "An error occurred when getting service '%v' for HTTP request", serviceIdentifier)
	}
	port, found := userService.GetPrivatePorts()[portId]
	if !found {
		return nil, stacktrace.NewError("An error occurred when getting port '%v' from service '%v' for HTTP request", serviceIdentifier, portId)
	}
	url := fmt.Sprintf("http://%v:%v%v", userService.GetRegistration().GetPrivateIP(), port.GetNumber(), endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	if err != nil {
		return nil, stacktrace.NewError("An error occurred building HTTP request on service '%v', URL '%v'", userService, url)
	}
	for header, headerValue := range headers {
		req.Header.Set(header, headerValue)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred on HTTP request on service '%v', URL '%v'", userService, url)
	}
	return resp, nil
}

func (network *DefaultServiceNetwork) GetServices(ctx context.Context) (map[service.ServiceUUID]*service.Service, error) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	registeredServices, err := network.serviceRegistrationRepository.GetAll()
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred getting registered services from the repository")
	}
	registeredServiceNames := map[service.ServiceName]bool{}
	for name := range registeredServices {
		registeredServiceNames[name] = true
	}

	registeredServiceUuidsFilters := &service.ServiceFilters{
		Names:    registeredServiceNames,
		UUIDs:    nil,
		Statuses: nil,
	}

	allServices, err := network.kurtosisBackend.GetUserServices(ctx, network.enclaveUuid, registeredServiceUuidsFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while fetching services from the backend")
	}

	filteredServicesToRegisteredServices := map[service.ServiceUUID]*service.Service{}

	for name, registration := range registeredServices {
		uuid := registration.GetUUID()
		serviceObj, found := allServices[uuid]
		if !found {
			return nil, stacktrace.NewError("couldn't find service with uuid '%v' and name '%v' in backend", uuid, name)
		}
		serviceObj.GetRegistration().SetStatus(registration.GetStatus())
		filteredServicesToRegisteredServices[uuid] = serviceObj
	}

	return filteredServicesToRegisteredServices, nil
}

func (network *DefaultServiceNetwork) GetService(ctx context.Context, serviceIdentifier string) (*service.Service, error) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	serviceRegistration, err := network.getServiceRegistrationForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching registration for service identifier '%v'", serviceIdentifier)
	}

	serviceUuid := serviceRegistration.GetUUID()

	getServiceFilters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			serviceRegistration.GetUUID(): true,
		},
		Statuses: nil,
	}
	matchingServices, err := network.kurtosisBackend.GetUserServices(ctx, network.enclaveUuid, getServiceFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service '%v'", serviceUuid)
	}
	if len(matchingServices) == 0 {
		return nil, stacktrace.NewError(
			"A registration exists for service UUID '%v' but no service objects were found; this indicates that the service was "+
				"registered but not started",
			serviceUuid,
		)
	}
	if len(matchingServices) > 1 {
		return nil, stacktrace.NewError("Found multiple service objects matching UUID '%v'", serviceUuid)
	}
	serviceObj, found := matchingServices[serviceUuid]
	if !found {
		return nil, stacktrace.NewError("Found exactly one service object, but it didn't match expected UUID '%v'", serviceUuid)
	}

	// The service status is managed at the service network layer so we copy it to the response
	serviceObj.GetRegistration().SetStatus(serviceRegistration.GetStatus())
	return serviceObj, nil
}

func (network *DefaultServiceNetwork) GetServiceNames() (map[service.ServiceName]bool, error) {

	serviceNames, err := network.serviceRegistrationRepository.GetAllServiceNames()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting all service names from service registration repository")
	}

	return serviceNames, nil
}

func (network *DefaultServiceNetwork) CopyFilesFromService(ctx context.Context, serviceIdentifier string, srcPath string, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while fetching name for service identifier '%v'", serviceIdentifier)
	}

	filesArtifactUuid, err := network.copyFilesFromServiceUnlocked(ctx, serviceName, srcPath, artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error in copying files over to disk")
	}
	return filesArtifactUuid, nil
}

func (network *DefaultServiceNetwork) ExistServiceRegistration(serviceName service.ServiceName) (bool, error) {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	exist, err := network.serviceRegistrationRepository.Exist(serviceName)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting service registration for service '%s'", serviceName)
	}

	return exist, nil
}

func (network *DefaultServiceNetwork) RenderTemplates(templatesAndDataByDestinationRelFilepath map[string]*render_templates.TemplateData, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	filesArtifactUuid, err := network.renderTemplatesUnlocked(templatesAndDataByDestinationRelFilepath, artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error in rendering templates to disk")
	}
	return filesArtifactUuid, nil
}

func (network *DefaultServiceNetwork) UploadFilesArtifact(data io.Reader, contentMd5 []byte, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	filesArtifactStore, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting files artifact store")
	}

	filesArtifactUuid, err := filesArtifactStore.StoreFile(data, contentMd5, artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while trying to store files.")
	}
	return filesArtifactUuid, nil
}

func (network *DefaultServiceNetwork) GetFilesArtifactMd5(artifactName string) (enclave_data_directory.FilesArtifactUUID, []byte, bool, error) {
	filesArtifactStore, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", nil, false, stacktrace.Propagate(err, "An error occurred while getting files artifact store")
	}
	filesArtifactUuid, _, currentlyStoredFileContentHash, found, err := filesArtifactStore.GetFile(artifactName)
	if err != nil {
		return "", nil, false, stacktrace.Propagate(err, "An error occurred retrieving the files artifact from the store")
	}
	return filesArtifactUuid, currentlyStoredFileContentHash, found, nil
}

func (network *DefaultServiceNetwork) UpdateFilesArtifact(fileArtifactUuid enclave_data_directory.FilesArtifactUUID, updatedContent io.Reader, contentMd5 []byte) error {
	filesArtifactStore, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting files artifact store")
	}

	if err = filesArtifactStore.UpdateFile(fileArtifactUuid, updatedContent, contentMd5); err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to update a files artifact.")
	}
	return nil
}

func (network *DefaultServiceNetwork) GetExistingAndHistoricalServiceIdentifiers() (service_identifiers.ServiceIdentifiers, error) {
	serviceIdentifiers, err := network.serviceIdentifiersRepository.GetServiceIdentifiers()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service identifiers list from the repository")
	}
	return serviceIdentifiers, nil
}

// GetUniqueNameForFileArtifact : this will return unique artifact name after 5 retries, same as enclave id generator
func (network *DefaultServiceNetwork) GetUniqueNameForFileArtifact() (string, error) {
	filesArtifactStore, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting files artifact store")
	}
	return filesArtifactStore.GenerateUniqueNameForFileArtifact(), nil
}

func (network *DefaultServiceNetwork) GetApiContainerInfo() *ApiContainerInfo {
	return network.apiContainerInfo
}

func (network *DefaultServiceNetwork) GetEnclaveUuid() enclave.EnclaveUUID {
	return network.enclaveUuid
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
// registerService handles all the operations necessary to register a service before is can be started with
// startRegisteredService. If something fails along the way, the function takes care of rolling back the previous
// changes such that the enclave remains in the state before the call
// TODO: The approach is naive here as we register a single service, so it needs to be called within a loop
//
//	to register multiple services. We could leverage the fact that the BE handles registering multiple services
//	with a single call. For now, as registering a service is fairly low lift, it's fine this way
func (network *DefaultServiceNetwork) registerService(
	ctx context.Context,
	serviceName service.ServiceName,
) (
	*service.ServiceRegistration,
	error,
) {
	serviceSuccessfullyRegistered := false

	serviceToRegister := map[service.ServiceName]bool{
		serviceName: true,
	}
	servicesSuccessfullyRegistered, serviceFailedRegistration, err := network.kurtosisBackend.RegisterUserServices(ctx, network.enclaveUuid, serviceToRegister)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unexpected error happened registering service '%s'", serviceName)
	}
	if serviceRegistrationErr, found := serviceFailedRegistration[serviceName]; found {
		return nil, stacktrace.Propagate(serviceRegistrationErr, "Error registering service '%s'", serviceName)
	}
	serviceRegistration, found := servicesSuccessfullyRegistered[serviceName]
	if !found {
		return nil, stacktrace.NewError("Unexpected error while registering service '%s'. It was not flagged as neither failed nor successfully registered. This is a Kurtosis internal bug.", serviceName)
	}
	defer func() {
		if serviceSuccessfullyRegistered {
			return
		}
		serviceUuid := serviceRegistration.GetUUID()
		serviceToUnregister := map[service.ServiceUUID]bool{
			serviceUuid: true,
		}
		_, failedService, unexpectedErr := network.kurtosisBackend.UnregisterUserServices(ctx, network.enclaveUuid, serviceToUnregister)
		if unexpectedErr != nil {
			logrus.Errorf("An unexpected error happened unregistering service '%s' after it failed starting. It"+
				"is possible the service is still registered to the enclave.", serviceName)
			return
		}
		if unregisteringErr, found := failedService[serviceUuid]; found {
			logrus.Errorf("An error happened unregistering service '%s' after it failed starting. It"+
				"is possible the service is still registered to the enclave. The error was\n%v",
				serviceName, unregisteringErr.Error())
			return
		}
	}()

	if err := network.serviceRegistrationRepository.Save(serviceRegistration); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred saving service registration '%+v' for service '%s'", serviceRegistration, serviceName)
	}
	// remove service from the registered service repository is something fails downstream
	defer func() {
		if serviceSuccessfullyRegistered {
			return
		}
		if err := network.serviceRegistrationRepository.Delete(serviceName); err != nil {
			logrus.Errorf("We tried to delete the service registration for service  '%s' we had stored but failed:\n%v", serviceName, err)
		}
	}()

	serviceSuccessfullyRegistered = true
	return serviceRegistration, nil
}

// unregisterService is the opposite of register service. It cleans up everything is can to property unregister a
// service. It is expected that the service was properly registered.
// As registerService rolls back things if a failure happens halfway, we should never end up with a service
// half-registered, but it's worth calling out that this method with throw if called with such a service
func (network *DefaultServiceNetwork) unregisterService(ctx context.Context, serviceName service.ServiceName) error {

	serviceRegistration, err := network.serviceRegistrationRepository.Get(serviceName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service registration for service '%s'", serviceName)
	}

	if err := network.serviceRegistrationRepository.Delete(serviceName); err != nil {
		return stacktrace.Propagate(err, "An error occurred deleting the service registration for service '%v' from the repository", serviceName)
	}
	serviceUuid := serviceRegistration.GetUUID()
	serviceToUnregister := map[service.ServiceUUID]bool{
		serviceUuid: true,
	}
	_, failedService, unexpectedErr := network.kurtosisBackend.UnregisterUserServices(ctx, network.enclaveUuid, serviceToUnregister)
	if unexpectedErr != nil {
		return stacktrace.Propagate(unexpectedErr, "An unexpected error happened unregistering service '%s'. It "+
			"is possible the service is still registered to the enclave.", serviceName)
	}
	if unregisteringErr, found := failedService[serviceUuid]; found {
		return stacktrace.Propagate(unregisteringErr, "An error happened unregistering service '%s'. It"+
			"is possible the service is still registered to the enclave.",
			serviceName)
	}
	return nil
}

// startRegisteredService handles the logistic of starting a service in the relevant Kurtosis backend:
func (network *DefaultServiceNetwork) startRegisteredService(
	ctx context.Context,
	serviceUuid service.ServiceUUID,
	serviceConfig *service.ServiceConfig,
) (
	*service.Service,
	error,
) {
	serviceStartedSuccessfully := false

	// Docker and K8s requires the minimum memory limit to be 6 megabytes to we make sure the allocation is at least that amount
	// But first, we check that it's not the default value, meaning the user potentially didn't even set it
	if serviceConfig.GetMemoryAllocationMegabytes() != defaultMemoryAllocMegabytes && serviceConfig.GetMemoryAllocationMegabytes() < minMemoryLimit {
		return nil, stacktrace.NewError("Memory allocation, `%d`, is too low. Kurtosis requires the memory limit to be at least `%d` megabytes for service with UUID '%v'.", serviceConfig.GetMemoryAllocationMegabytes(), minMemoryLimit, serviceUuid)
	}

	// TODO(gb): make the backend also handle starting service sequentially to simplify the logic there as well
	serviceConfigMap := map[service.ServiceUUID]*service.ServiceConfig{
		serviceUuid: serviceConfig,
	}
	successfulServices, failedServices, err := network.kurtosisBackend.StartRegisteredUserServices(ctx, network.enclaveUuid, serviceConfigMap)
	if err != nil {
		return nil, err
	}
	if failedServiceErr, isFailed := failedServices[serviceUuid]; isFailed {
		return nil, failedServiceErr
	}

	startedService, isSuccessful := successfulServices[serviceUuid]
	if !isSuccessful {
		return nil, stacktrace.NewError("Service '%s' did not start properly but no error was thrown. This is a Kurtosis internal bug", serviceUuid)
	}
	defer func() {
		if serviceStartedSuccessfully {
			return
		}
		serviceToDestroyUuid := startedService.GetRegistration().GetUUID()
		userServiceFilters := &service.ServiceFilters{
			Names: nil,
			UUIDs: map[service.ServiceUUID]bool{
				serviceToDestroyUuid: true,
			},
			Statuses: nil,
		}
		_, failedToDestroyUuids, err := network.kurtosisBackend.DestroyUserServices(context.Background(), network.enclaveUuid, userServiceFilters)
		if err != nil {
			logrus.Errorf("Attempted to destroy the services with UUIDs '%v' but had no success. You must manually destroy the services! The following error had occurred:\n'%v'", serviceToDestroyUuid, err)
			return
		}
		if failedToDestroyErr, found := failedToDestroyUuids[serviceToDestroyUuid]; found {
			logrus.Errorf("Attempted to destroy the services with UUIDs '%v' but had no success. You must manually destroy the services! The following error had occurred:\n'%v'", serviceToDestroyUuid, failedToDestroyErr)
		}
	}()

	allPrivateAndPublicPorts := mergeAndGetAllPrivateAndPublicServicePorts(startedService)

	if err := waitUntilAllTCPAndUDPPortsAreOpen(
		startedService.GetRegistration().GetPrivateIP(),
		allPrivateAndPublicPorts,
	); err != nil {
		serviceLogs, getServiceLogsErr := network.getServiceLogs(ctx, startedService, shouldFollowLogs)
		if getServiceLogsErr != nil {
			serviceLogs = fmt.Sprintf("An error occurred while getting the service logs.\n Error:%v", getServiceLogsErr)
		}
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for all TCP and UDP ports to be open for service '%v' with private IP '%v'; "+
				"this is usually due to a misconfiguration in the service itself, so here are the logs:\n%s",
			startedService.GetRegistration().GetName(),
			startedService.GetRegistration().GetPrivateIP(),
			serviceLogs,
		)
	}

	serviceStartedSuccessfully = true
	serviceName := startedService.GetRegistration().GetName()
	if err := network.serviceRegistrationRepository.UpdateConfig(serviceName, serviceConfig); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while updating service config to '%+v' in service registration for service '%s' after the service was started", serviceConfig, serviceName)
	}

	return startedService, nil
}

// destroyService is the opposite of startRegisteredService. It removes a started service from the enclave. Note that it does not
// take care of unregistering the service. For this, unregisterService should be called
// Similar to unregisterService, it is expected that the service passed to destroyService has been properly started.
// the function might fail if the service is half-started
// Note: the function also takes care of destroying any networking sidecar associated with the service
func (network *DefaultServiceNetwork) destroyService(ctx context.Context, serviceName service.ServiceName, serviceUuid service.ServiceUUID) error {
	// deleting the service first
	userServiceFilters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			serviceUuid: true,
		},
		Statuses: nil,
	}
	successfullyDestroyedUuids, failedToDestroyUuids, err := network.kurtosisBackend.DestroyUserServices(context.Background(), network.enclaveUuid, userServiceFilters)
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to destroy the service with UUID '%v' but had no success. You must manually destroy the service as well as its sidecar if any", serviceUuid)
	}
	if _, found := successfullyDestroyedUuids[serviceUuid]; found {
		if len(failedToDestroyUuids) != 0 {
			return stacktrace.NewError("Service '%s' was successfully destroyed but unexpected error were returned by Kurtosis backend. This is unexpected and should be investigated. Here are the errors:\n%v", serviceUuid, failedToDestroyUuids)
		}
	} else {
		if failedToDestroyErr, found := failedToDestroyUuids[serviceUuid]; found {
			return stacktrace.Propagate(failedToDestroyErr, "Attempted to destroy the service with UUID '%v' but had no success. You must manually destroy the service as well as its sidecar if any", serviceUuid)
		} else {
			return stacktrace.NewError("Attempted to destroy service '%s' but it was neither marked as successfully destroyed nor errored in the result. This is a Kurtosis bug", serviceUuid)
		}
	}

	return nil
}

// startRegisteredServices starts multiple services in parallel
//
// It iterates over all the services to start and kicks off a go subroutine for each of them.
// The subroutine will block until it can write to concurrencyControlChan. concurrencyControlChan is a simple buffered
// channel that can contain a max of 4 values. It's a common way in go to control concurrency to make sure no more than
// X subroutine is running at the same time.
//
// Once the for loops has started all the subroutine, we use a WaitGroup for this method to block until all subroutines
// have completed
//
// The subroutine accounts for its result populating the startedServices and failedServices maps (which are be accessed
// behind a mutex as those are not concurrent maps) and finishes by release a permit from the WaitGroup
func (network *DefaultServiceNetwork) startRegisteredServices(
	ctx context.Context,
	serviceConfigs map[service.ServiceUUID]*service.ServiceConfig,
	batchSize int,
) (map[service.ServiceUUID]*service.Service, map[service.ServiceUUID]error) {
	wg := sync.WaitGroup{}

	concurrencyControlChan := make(chan bool, batchSize)
	defer close(concurrencyControlChan)

	startedServices := map[service.ServiceUUID]*service.Service{}
	failedServices := map[service.ServiceUUID]error{}
	mapWriteMutex := sync.Mutex{}

	// async kick off all the routines one by one
	for serviceUuid, serviceConfig := range serviceConfigs {
		serviceToStartUuid := serviceUuid
		serviceToStartConfig := serviceConfig

		if len(failedServices) > 0 {
			// stop scheduling more service start
			// as one already failed, the full batch will be reverted anyway so no need to continue any further
			break
		}
		// The concurrencyControlChan will block if the buffer is currently full, i.e. if maxConcurrentServiceStart
		// subroutines are already running in the background
		concurrencyControlChan <- true
		wg.Add(1)
		go func() {
			defer func() {
				// at the end, make sure the subroutine releases one permit from the WaitGroup, and make sure to
				// also pop a value from the concurrencyControlChan to allow any potentially waiting subroutine to
				// start
				wg.Done()
				<-concurrencyControlChan
			}()
			logrus.Debugf("Starting service '%s'", serviceToStartUuid)
			startedService, err := network.startRegisteredService(ctx, serviceToStartUuid, serviceToStartConfig)
			mapWriteMutex.Lock()
			defer mapWriteMutex.Unlock()
			if err != nil {
				failedServices[serviceToStartUuid] = err
				logrus.Debugf("Service '%s' could not start due to some errors", serviceToStartUuid)
			} else {
				startedServices[serviceToStartUuid] = startedService
				logrus.Debugf("Service '%s' started successfully", serviceToStartUuid)
			}
		}()
	}

	// wait for all subroutines to complete and return
	wg.Wait()
	return startedServices, failedServices
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *DefaultServiceNetwork) copyFilesFromServiceUnlocked(ctx context.Context, serviceName service.ServiceName, srcPath string, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {

	serviceRegistrationObj, err := network.serviceRegistrationRepository.Get(serviceName)
	if err != nil {
		return "", stacktrace.Propagate(err, "Cannot copy files from service '%v' because it does not exist in the network", serviceName)
	}
	serviceUuid := serviceRegistrationObj.GetUUID()

	store, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the files artifact store")
	}

	existingFilesArtifactUuid, _, _, filesArtifactAlreadyExists, err := store.GetFile(artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An unexpected error occurred checking for file artifact '%s' existence in the store", artifactName)
	}

	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	storeFilesArtifactResultChan := make(chan storeFilesArtifactResult)
	go func() {
		defer pipeReader.Close()
		defer close(storeFilesArtifactResultChan)
		//And finally pass it the .tgz file to the artifact file store
		// It's hard to compute the content hash here. For files stored from services, hash will be empty and they will
		// be re-stored everytime regardless of their content

		var storeOrUpdateErr error
		var filesArtifactUuid enclave_data_directory.FilesArtifactUUID
		if filesArtifactAlreadyExists {
			filesArtifactUuid = existingFilesArtifactUuid
			// we're not able to check file hash here b/c we don't get it from the CompressPath helper method
			// TODO: Maybe compute the hash ad-hoc here, but it would require us to consume the reader
			storeOrUpdateErr = store.UpdateFile(filesArtifactUuid, pipeReader, []byte{})
		} else {
			filesArtifactUuid, storeOrUpdateErr = store.StoreFile(pipeReader, []byte{}, artifactName)
		}
		storeFilesArtifactResultChan <- storeFilesArtifactResult{
			err:               storeOrUpdateErr,
			filesArtifactUuid: filesArtifactUuid,
		}
	}()

	if err := network.gzipAndPushTarredFileBytesToOutput(ctx, pipeWriter, serviceUuid, srcPath); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred gzip'ing and pushing tar'd file bytes to the pipe")
	}

	storeFileResult := <-storeFilesArtifactResultChan
	if storeFileResult.err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred storing files from path '%v' on service '%v' in in the files artifact store",
			srcPath,
			serviceUuid,
		)
	}

	return storeFileResult.filesArtifactUuid, nil
}

func (network *DefaultServiceNetwork) gzipAndPushTarredFileBytesToOutput(
	ctx context.Context,
	output io.WriteCloser,
	serviceUuid service.ServiceUUID,
	srcPathOnContainer string,
) error {
	defer output.Close()

	// Need to compress the TAR bytes on our side, since we're not guaranteedj
	gzippingOutput := gzip.NewWriter(output)
	defer gzippingOutput.Close()

	if err := network.kurtosisBackend.CopyFilesFromUserService(ctx, network.enclaveUuid, serviceUuid, srcPathOnContainer, gzippingOutput); err != nil {
		return stacktrace.Propagate(err, "An error occurred copying source '%v' from user service with UUID '%v' in enclave with UUID '%v'", srcPathOnContainer, serviceUuid, network.enclaveUuid)
	}

	return nil
}

// This method is not thread safe. Only call this from a method where there is a mutex lock on the network.
func (network *DefaultServiceNetwork) renderTemplatesUnlocked(templatesAndDataByDestinationRelFilepath map[string]*render_templates.TemplateData, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	tempDirForRenderedTemplates, err := os.MkdirTemp("", tempDirForRenderedTemplatesPrefix)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while creating a temp dir for rendered templates '%v'", tempDirForRenderedTemplates)
	}
	defer os.RemoveAll(tempDirForRenderedTemplates)

	for destinationRelFilepath, templateAndData := range templatesAndDataByDestinationRelFilepath {
		destinationAbsoluteFilePath := path.Join(tempDirForRenderedTemplates, destinationRelFilepath)
		if err := templateAndData.RenderToFile(destinationAbsoluteFilePath); err != nil {
			return "", stacktrace.Propagate(err, "There was an error in rendering template for file '%s'", destinationRelFilepath)
		}
	}

	compressedFile, _, compressedFileMd5, err := path_compression.CompressPath(tempDirForRenderedTemplates, enforceMaxFileSizeLimit)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error compressing dir '%v'", tempDirForRenderedTemplates)
	}
	defer compressedFile.Close()

	store, err := network.enclaveDataDir.GetFilesArtifactStore()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting files artifact store")
	}
	existingFilesArtifactUuid, _, existingFilesArtifactMd5, found, err := store.GetFile(artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An unexpected error occurred checking for file artifact '%s' existence in the store", artifactName)
	}

	var storeOrUpdateErr error
	var filesArtifactUuid enclave_data_directory.FilesArtifactUUID
	if found {
		filesArtifactUuid = existingFilesArtifactUuid
		// If one of the md5 is empty, we can't really assume the files are equal so we fallback to updating them
		// Otherwise we check the equality of their MD5 and if they are different we update them
		if len(existingFilesArtifactMd5) == 0 || len(compressedFileMd5) == 0 || !bytes.Equal(existingFilesArtifactMd5, compressedFileMd5) {
			storeOrUpdateErr = store.UpdateFile(existingFilesArtifactUuid, compressedFile, compressedFileMd5)
		}
	} else {
		filesArtifactUuid, storeOrUpdateErr = store.StoreFile(compressedFile, compressedFileMd5, artifactName)
	}
	if storeOrUpdateErr != nil {
		return "", stacktrace.Propagate(err, "An error occurred while storing the file '%v' in the files artifact store", compressedFile)
	}
	shouldDeleteFilesArtifact := true
	defer func() {
		if shouldDeleteFilesArtifact {
			if err = store.RemoveFile(string(filesArtifactUuid)); err != nil {
				logrus.Errorf("We tried to clean up the files artifact '%v' we had stored but failed:\n%v", artifactName, err)
			}
		}
	}()

	shouldDeleteFilesArtifact = false
	return filesArtifactUuid, nil
}

// This isn't thread safe and must be called from a thread safe context
func (network *DefaultServiceNetwork) getServiceNameForIdentifierUnlocked(serviceIdentifier string) (service.ServiceName, error) {
	maybeServiceUuid := service.ServiceUUID(serviceIdentifier)
	serviceUuidToServiceName := map[service.ServiceUUID]service.ServiceName{}
	serviceShortenedUuidToServiceName := map[string][]service.ServiceName{}

	allServiceRegistrationsByServiceName, err := network.serviceRegistrationRepository.GetAll()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting all service registrations from the repository")
	}

	for serviceName, registration := range allServiceRegistrationsByServiceName {
		serviceUuid := registration.GetUUID()
		serviceShortenedUuid := uuid_generator.ShortenedUUIDString(string(serviceUuid))
		serviceUuidToServiceName[serviceUuid] = serviceName
		serviceShortenedUuidToServiceName[serviceShortenedUuid] = append(serviceShortenedUuidToServiceName[serviceShortenedUuid], serviceName)
	}

	if serviceName, found := serviceUuidToServiceName[maybeServiceUuid]; found {
		return serviceName, nil
	}

	maybeShortenedUuid := serviceIdentifier
	if serviceNames, found := serviceShortenedUuidToServiceName[maybeShortenedUuid]; found {
		if len(serviceNames) == exactlyOneShortenedUuidMatch {
			return serviceNames[0], nil
		} else {
			return "", stacktrace.NewError("Found multiple matching service names '%v' for shortened uuid '%v'. Please be more specific", serviceNames, maybeShortenedUuid)
		}
	}

	maybeServiceName := service.ServiceName(serviceIdentifier)

	maybeServiceNameExist, err := network.serviceRegistrationRepository.Exist(maybeServiceName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred checking if service name '%s' exist in the service registration repository", maybeServiceName)
	}
	if maybeServiceNameExist {
		return maybeServiceName, nil
	}

	return "", stacktrace.NewError("Couldn't find a matching service name for identifier '%v'", serviceIdentifier)
}

func (network *DefaultServiceNetwork) getServiceLogs(
	ctx context.Context,
	serviceObj *service.Service,
	shouldFollowLogs bool,
) (string, error) {
	enclaveUuid := serviceObj.GetRegistration().GetEnclaveID()
	serviceUUID := serviceObj.GetRegistration().GetUUID()
	userServiceFilters := &service.ServiceFilters{
		Names: nil,
		UUIDs: map[service.ServiceUUID]bool{
			serviceUUID: true,
		},
		Statuses: nil,
	}

	successfulUserServiceLogs, erroredUserServiceUuids, err := network.kurtosisBackend.GetUserServiceLogs(ctx, enclaveUuid, userServiceFilters, shouldFollowLogs)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
	}
	defer func() {
		for _, userServiceLogsReadCloser := range successfulUserServiceLogs {
			if err := userServiceLogsReadCloser.Close(); err != nil {
				logrus.Warnf("We tried to close the user service logs read-closer-objects after we're done using it, but doing so threw an error:\n%v", err)
			}
		}
	}()

	err, found := erroredUserServiceUuids[serviceUUID]
	if found && err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred getting user service logs for user service with UUID '%v'", serviceUUID)
	}

	userServiceReadCloserLog, found := successfulUserServiceLogs[serviceUUID]
	if !found {
		return "", stacktrace.NewError(
			"Expected to find logs for user service with UUID '%v' on user service logs map '%+v' "+
				"but was not found; this should never happen, and is a bug "+
				"in Kurtosis", serviceUUID, userServiceReadCloserLog)
	}
	defer func() {
		if err := userServiceReadCloserLog.Close(); err != nil {
			logrus.Warnf("Something fails when we tried to close the read closer logs for service with UUID %v", serviceUUID)
		}
	}()

	copyBuf := bytes.NewBuffer(nil)
	if _, err := io.Copy(copyBuf, userServiceReadCloserLog); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred copying the service logs to a buffer")
	}

	header := fmt.Sprintf(serviceLogsHeader, serviceObj.GetRegistration().GetName())
	footer := fmt.Sprintf(serviceLogsFooter, serviceObj.GetRegistration().GetName())

	serviceLogs := fmt.Sprintf("%s\n%s\n%s", header, copyBuf.String(), footer)

	return serviceLogs, nil
}

func waitUntilAllTCPAndUDPPortsAreOpen(
	ipAddr net.IP,
	ports map[string]*port_spec.PortSpec,
) error {
	var portCheckErrorGroup errgroup.Group

	for _, portSpec := range ports {
		// we are using value here because we are using it inside the closure
		portSpecValue := *portSpec
		wrappedWaitFunc := func() error {
			return waitUntilPortIsOpenWithTimeout(ipAddr, portSpecValue)
		}
		portCheckErrorGroup.Go(wrappedWaitFunc)
	}
	//This error group pattern allow us to reject early if at least one of the ports check fails
	//in this opportunity we want to reject early because the main user pain that we want to handle
	//is a wrong service configuration
	if err := portCheckErrorGroup.Wait(); err != nil {
		return stacktrace.Propagate(err, "An error occurred while waiting for all TCP and UDP ports to be open")
	}

	return nil
}

func waitUntilPortIsOpenWithTimeout(
	ipAddr net.IP,
	portSpec port_spec.PortSpec,
) error {
	// reject early if it's disable
	if portSpec.GetWait() == nil {
		return nil
	}
	timeout := portSpec.GetWait().GetTimeout()

	var (
		startTime  = time.Now()
		finishTime = startTime.Add(timeout)
		retries    = 0
		err        error
	)

	ticker := time.NewTicker(waitForPortsOpenRetriesDelayMilliseconds * time.Millisecond)
	defer ticker.Stop()

	logrus.Debugf("Checking if port '%+v' in '%v' is open...", portSpec, ipAddr)

	for {
		if time.Now().After(finishTime) {
			return stacktrace.Propagate(err, "Unsuccessful ports check for IP '%s' and port spec '%+v', "+
				"even after '%v' retries with '%v' milliseconds in between retries. Timeout '%v' has been reached",
				ipAddr,
				portSpec,
				retries,
				waitForPortsOpenRetriesDelayMilliseconds,
				timeout.String(),
			)
		}
		if err = scanPort(ipAddr, &portSpec, scanPortTimeout); err == nil {
			logrus.Debugf(
				"Successful port open check for IP '%s' and port spec '%+v' after retry number '%v', "+
					"with '%v' milliseconds between retries and it took '%v'",
				ipAddr,
				portSpec,
				retries,
				waitForPortsOpenRetriesDelayMilliseconds,
				time.Since(startTime),
			)
			return nil
		}
		retries++
		<-ticker.C // block until the next tick
	}
}

func scanPort(ipAddr net.IP, portSpec *port_spec.PortSpec, timeout time.Duration) error {
	portNumberStr := fmt.Sprintf("%v", portSpec.GetNumber())
	networkAddress := net.JoinHostPort(ipAddr.String(), portNumberStr)
	networkStr := strings.ToLower(portSpec.GetTransportProtocol().String())
	conn, err := net.DialTimeout(networkStr, networkAddress, timeout)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred while calling network address '%s' with port protocol '%s' and using time out '%v'",
			networkAddress,
			portSpec.GetTransportProtocol().String(),
			timeout,
		)
	}
	defer conn.Close()
	return nil
}

func mergeAndGetAllPrivateAndPublicServicePorts(service *service.Service) map[string]*port_spec.PortSpec {
	allPrivateAndPublicPorts := map[string]*port_spec.PortSpec{}

	for portId, portSpec := range service.GetPrivatePorts() {
		allPrivateAndPublicPorts[portId] = portSpec
	}

	for portId, portSpec := range service.GetMaybePublicPorts() {
		newPortId := portId
		if _, found := allPrivateAndPublicPorts[portId]; found {
			newPortId = portId + publicPortsSuffix
		}
		allPrivateAndPublicPorts[newPortId] = portSpec
	}
	return allPrivateAndPublicPorts
}

// This isn't thread safe and must be called from a thread safe context
func (network *DefaultServiceNetwork) getServiceRegistrationForIdentifierUnlocked(
	serviceIdentifier string,
) (*service.ServiceRegistration, error) {
	serviceName, err := network.getServiceNameForIdentifierUnlocked(serviceIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching name for service identifier '%v'", serviceIdentifier)
	}

	serviceRegistration, err := network.serviceRegistrationRepository.Get(serviceName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service registration for service '%s'", serviceName)
	}

	return serviceRegistration, nil
}
