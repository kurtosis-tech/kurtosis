package user_service_functions

import (
	"context"
	"strings"
	"sync"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/service_registration"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	unlimitedReplacements                = -1
	skipAddingUserServiceToBridgeNetwork = true
	emptyImageName                       = ""
)

func RegisterUserServices(
	enclaveUuid enclave.EnclaveUUID,
	servicesToRegister map[service.ServiceName]bool,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	freeIpProvidersForEnclave *free_ip_addr_tracker.FreeIpAddrTracker,
	serviceRegistrationMutex *sync.Mutex,
) (
	map[service.ServiceName]*service.ServiceRegistration,
	map[service.ServiceName]error,
	error,
) {
	serviceRegistrationMutex.Lock()
	defer serviceRegistrationMutex.Unlock()

	successfulServicesPool := map[service.ServiceName]*service.ServiceRegistration{}
	failedServicesPool := map[service.ServiceName]error{}

	// Check whether any services have been provided at all
	if len(servicesToRegister) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	successfulRegistrations, failedRegistrations, err := registerUserServices(enclaveUuid, servicesToRegister, serviceRegistrationRepository, freeIpProvidersForEnclave)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services with Names '%v'", servicesToRegister)
	}
	return successfulRegistrations, failedRegistrations, nil
}

// UnregisterUserServices unregisters all services currently registered for this enclave.
// If the service is not registered for this enclave, it no-ops and the service is returned as "successfully unregistered"
func UnregisterUserServices(
	enclaveUuid enclave.EnclaveUUID,
	serviceUUIDsToUnregister map[service.ServiceUUID]bool,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	freeIpAddrProviderForEnclave *free_ip_addr_tracker.FreeIpAddrTracker,
	serviceRegistrationMutex *sync.Mutex,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	serviceRegistrationMutex.Lock()
	defer serviceRegistrationMutex.Unlock()
	servicesSuccessfullyUnregistered := map[service.ServiceUUID]bool{}
	servicesFailed := map[service.ServiceUUID]error{}

	if len(serviceUUIDsToUnregister) == 0 {
		return servicesSuccessfullyUnregistered, servicesFailed, nil
	}

	enclaveServiceRegistrations, err := serviceRegistrationRepository.GetAllEnclaveServiceRegistrations(enclaveUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting all enclave service registrations from the repository for enclave with UUID '%v'", enclaveUuid)
	}

	for serviceUuid := range serviceUUIDsToUnregister {
		serviceRegistration, isServiceRegistered := enclaveServiceRegistrations[serviceUuid]
		if !isServiceRegistered {
			servicesSuccessfullyUnregistered[serviceUuid] = true
			continue
		}

		serviceIpAddr := serviceRegistration.GetPrivateIP()

		if err := freeIpAddrProviderForEnclave.ReleaseIpAddr(serviceIpAddr); err != nil {
			servicesFailed[serviceUuid] = err
			continue
		}

		if err := serviceRegistrationRepository.Delete(serviceRegistration.GetName()); err != nil {
			servicesFailed[serviceUuid] = err
			continue
		}

		servicesSuccessfullyUnregistered[serviceUuid] = true

	}
	return servicesSuccessfullyUnregistered, servicesFailed, nil
}

func StartRegisteredUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	services map[service.ServiceUUID]*service.ServiceConfig,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	logsCollector *logs_collector.LogsCollector,
	logsCollectorAvailabilityChecker logs_collector_functions.LogsCollectorAvailabilityChecker,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	freeIpProviderForEnclave *free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
	restartPolicy docker_manager.RestartPolicy,
) (
	map[service.ServiceUUID]*service.Service,
	map[service.ServiceUUID]error,
	error,
) {
	successfulServicesPool := map[service.ServiceUUID]*service.Service{}
	failedServicesPool := map[service.ServiceUUID]error{}

	serviceConfigsToStart := map[service.ServiceUUID]*service.ServiceConfig{}
	serviceRegistrationsToStart := map[service.ServiceUUID]*service.ServiceRegistration{}

	serviceRegistrations, err := serviceRegistrationRepository.GetAllEnclaveServiceRegistrations(enclaveUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting all enclave service registrations from the repository for enclave with UUID '%v'", enclaveUuid)
	}

	for serviceUuid, serviceConfig := range services {
		serviceRegistration, found := serviceRegistrations[serviceUuid]
		if !found {
			failedServicesPool[serviceUuid] = stacktrace.NewError("Attempted to start a service '%s' that is not registered to this enclave yet.", serviceUuid)
			continue
		}
		if serviceConfig.GetPrivateIPAddrPlaceholder() == "" {
			failedServicesPool[serviceUuid] = stacktrace.NewError("Service with UUID '%v' has an empty private IP Address placeholder. Expect this to be of length greater than zero.", serviceUuid)
			continue
		}
		serviceConfigsToStart[serviceUuid] = serviceConfig
		serviceRegistrationsToStart[serviceUuid] = serviceRegistration
	}

	// If no services had successful registrations, return immediately
	// This is to prevent an empty filter being used to query for matching objects and resources, returning all services
	// and causing logic to break (eg. check for duplicate service GUIDs)
	// Making this check allows us to eject early and maintain a guarantee that objects and resources returned
	// are 1:1 with serviceUUIDs
	if len(serviceConfigsToStart) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	for _, serviceRegistrationToStart := range serviceRegistrationsToStart {
		if serviceRegistrationToStart.GetStatus() == service.ServiceStatus_Stopped {
			// If the first service to start is stopped, we know that the other ones are too because
			// of a check done at the service network layer.
			// Restarting stopped services is a much lighter operation so we branch out to a simpler function
			return restartUserServices(ctx, enclaveUuid, serviceRegistrationsToStart, serviceRegistrationRepository, dockerManager)
		}
	}

	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	// Sanity check for port bindings on all services
	for serviceUuid, serviceConfig := range serviceConfigsToStart {
		publicPorts := serviceConfig.GetPublicPorts()
		if len(publicPorts) > 0 {
			privatePorts := serviceConfig.GetPrivatePorts()
			err := checkPrivateAndPublicPortsAreOneToOne(privatePorts, publicPorts)
			if err != nil {
				failedServicesPool[serviceUuid] = stacktrace.Propagate(err, "Private and public ports for service with UUID '%v' are not one to one.", serviceUuid)
				delete(serviceConfigsToStart, serviceUuid)
			}
		}
	}
	//TODO END huge hack to temporarily enable static ports for NEAR

	enclaveNetwork, err := shared_helpers.GetEnclaveNetworkByEnclaveUuid(ctx, enclaveUuid, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveUuid)
	}
	enclaveNetworkID := enclaveNetwork.GetId()

	enclaveObjAttrsProvider, err := objAttrsProvider.ForEnclave(enclaveUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveUuid)
	}

	// Check if the logs collector is available
	// As the container logs are sent asynchronously we'd not know whether they're being received by the collector and there would be no errors if the collector never comes up
	// The least we can do is check if the collector server is healthy before starting the user service, if in case it gets shut down later we can't do much about it anyway.
	if err = logsCollectorAvailabilityChecker.WaitForAvailability(); err != nil {
		return nil, nil,
			stacktrace.Propagate(err, "An error occurred while waiting to see if the logs collector was available.")
	}

	logsCollectorEnclaveAddr, err := logsCollector.GetEnclaveNetworkAddressString()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the private TCP address")
	}

	// The following docker labels will be added into the logs stream which is necessary for filtering, retrieving persisted logs
	logsCollectorLabels := logs_collector_functions.GetKurtosisTrackedLogsCollectorLabels()

	successfulStarts, failedStarts, err := runStartServiceOperationsInParallel(
		ctx,
		enclaveNetworkID,
		serviceConfigsToStart,
		serviceRegistrations,
		enclaveObjAttrsProvider,
		freeIpProviderForEnclave,
		dockerManager,
		restartPolicy,
		logsCollectorEnclaveAddr,
		logsCollectorLabels,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to start services in parallel.")
	}

	// Add operations to their respective pools
	for serviceUuid, successfullyStartedService := range successfulStarts {
		successfulServicesPool[serviceUuid] = successfullyStartedService
		serviceName := successfullyStartedService.GetRegistration().GetName()
		serviceStatus := service.ServiceStatus_Started
		if err := serviceRegistrationRepository.UpdateStatus(serviceName, serviceStatus); err != nil {
			failedServicesPool[serviceUuid] = stacktrace.Propagate(err, "An error occurred while updating service status to '%s' in service registration for service '%s'", serviceStatus, serviceName)
			delete(successfulStarts, serviceUuid)
			delete(successfulServicesPool, serviceUuid)
			continue
		}
	}

	for serviceUuid, serviceErr := range failedStarts {
		failedServicesPool[serviceUuid] = serviceErr
	}

	logrus.Debugf("Started services '%v' successfully while '%v' failed", successfulServicesPool, failedServicesPool)
	return successfulServicesPool, failedServicesPool, nil
}

func RemoveRegisteredUserServiceProcesses(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	services map[service.ServiceUUID]bool,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	successfullyRemovedService := map[service.ServiceUUID]bool{}
	failedServicesPool := map[service.ServiceUUID]error{}

	serviceRegistrations, err := serviceRegistrationRepository.GetAllEnclaveServiceRegistrations(enclaveUuid)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting all enclave service registrations from the repository for enclave with UUID '%v'", enclaveUuid)
	}

	serviceUuidsToUpdate := map[service.ServiceUUID]bool{}
	for serviceUuid := range services {
		if _, found := serviceRegistrations[serviceUuid]; found {
			serviceUuidsToUpdate[serviceUuid] = true
		} else {
			failedServicesPool[serviceUuid] = stacktrace.NewError("Unable to update service '%s' that is not "+
				"registered inside this enclave", serviceUuid)
		}
	}

	removeServiceFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    serviceUuidsToUpdate,
		Statuses: nil,
	}

	allServiceObjs, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveUuid, removeServiceFilters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", removeServiceFilters)
	}

	servicesToStartByContainerId := map[string]*service.Service{}
	for uuid, serviceResources := range allDockerResources {
		serviceObj, found := allServiceObjs[uuid]
		if !found {
			// Should never happen; there should be a 1:1 mapping between service_objects:docker_resources by GUID
			return nil, nil, stacktrace.NewError("No service object found for service '%v' that had Docker resources; this is a bug in Kurtosis", uuid)
		}
		servicesToStartByContainerId[serviceResources.ServiceContainer.GetId()] = serviceObj
	}

	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing user service processes container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulUuidStrs, erroredUuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		servicesToStartByContainerId,
		dockerManager,
		extractServiceUUIDFromService,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing user service processes containers matching filters '%+v'", removeServiceFilters)
	}

	for failedServiceUuid, failedServiceErr := range erroredUuidStrs {
		failedServicesPool[service.ServiceUUID(failedServiceUuid)] = failedServiceErr
	}

	for serviceUuidStr := range successfulUuidStrs {
		serviceUuid := service.ServiceUUID(serviceUuidStr)
		serviceConfig, found := services[serviceUuid]
		if !found {
			failedServicesPool[serviceUuid] = stacktrace.NewError("An error occurred removing user service processes for service with UUID '%s'", serviceUuid)
		}
		successfullyRemovedService[serviceUuid] = serviceConfig
	}
	return successfullyRemovedService, failedServicesPool, nil
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================

// Restart a stopped user service
func restartUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	serviceRegistrations map[service.ServiceUUID]*service.ServiceRegistration,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceUUID]*service.Service,
	map[service.ServiceUUID]error,
	error,
) {

	serviceUuids := map[service.ServiceUUID]bool{}
	serviceNamesByUuids := map[service.ServiceUUID]service.ServiceName{}
	for serviceUuid, serviceRegistration := range serviceRegistrations {
		serviceUuids[serviceUuid] = true
		serviceNamesByUuids[serviceUuid] = serviceRegistration.GetName()
	}
	startServiceFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    serviceUuids,
		Statuses: nil,
	}

	allServiceObjs, allDockerResources, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveUuid, startServiceFilters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", startServiceFilters)
	}

	servicesToStartByContainerId := map[string]*service.Service{}
	for uuid, serviceResources := range allDockerResources {
		serviceObj, found := allServiceObjs[uuid]
		if !found {
			// Should never happen; there should be a 1:1 mapping between service_objects:docker_resources by GUID
			return nil, nil, stacktrace.NewError("No service object found for service '%v' that had Docker resources", uuid)
		}
		servicesToStartByContainerId[serviceResources.ServiceContainer.GetId()] = serviceObj
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.StartContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred starting user service container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulUuidStrs, erroredUuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		servicesToStartByContainerId,
		dockerManager,
		extractServiceUUIDFromService,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting user service containers matching filters '%+v'", startServiceFilters)
	}

	successfulServices := map[service.ServiceUUID]*service.Service{}
	for uuidStr := range successfulUuidStrs {
		serviceUuid := service.ServiceUUID(uuidStr)
		serviceName, found := serviceNamesByUuids[serviceUuid]
		if !found {
			erroredUuidStrs[uuidStr] = stacktrace.NewError("Expected to find service name by UUID '%v' in map '%+v', but none was found; this is a bug in Kurtosis", serviceUuid, serviceNamesByUuids)
			continue
		}

		successfulServices[serviceUuid] = service.NewService(
			serviceRegistrations[serviceUuid],
			nil,
			nil,
			nil,
			container.NewContainer(
				container.ContainerStatus_Running,
				emptyImageName,
				nil,
				nil,
				nil),
		)

		serviceStatus := service.ServiceStatus_Started
		if err := serviceRegistrationRepository.UpdateStatus(serviceName, serviceStatus); err != nil {
			erroredUuidStrs[uuidStr] = stacktrace.Propagate(err, "An error occurred while updating service status to '%s' in service registration for service '%s'", serviceStatus, serviceName)
			delete(successfulServices, serviceUuid)
			continue
		}
	}

	erroredUuids := map[service.ServiceUUID]error{}
	for uuidStr, err := range erroredUuidStrs {
		erroredUuids[service.ServiceUUID(uuidStr)] = stacktrace.Propagate(
			err,
			"An error occurred starting service '%v'",
			uuidStr,
		)
	}

	return successfulServices, erroredUuids, nil
}

func runStartServiceOperationsInParallel(
	ctx context.Context,
	enclaveNetworkId string,
	serviceConfigs map[service.ServiceUUID]*service.ServiceConfig,
	serviceRegistrations map[service.ServiceUUID]*service.ServiceRegistration,
	enclaveObjAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	freeIpAddrProvider *free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
	restartPolicy docker_manager.RestartPolicy,
	logsCollectorAddress string,
	logsCollectorLabels logs_collector_functions.LogsCollectorLabels,
) (
	map[service.ServiceUUID]*service.Service,
	map[service.ServiceUUID]error,
	error,
) {
	successfulServices := map[service.ServiceUUID]*service.Service{}
	failedServices := map[service.ServiceUUID]error{}

	startServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for serviceUuid, config := range serviceConfigs {
		serviceRegistration, found := serviceRegistrations[serviceUuid]
		if !found {
			failedServices[serviceUuid] = stacktrace.NewError("Failed to get service registration for service UUID '%v' while creating start service operation. This should never happen. This is a Kurtosis bug.", serviceUuid)
			continue
		}
		startServiceOperations[operation_parallelizer.OperationID(serviceUuid)] = createStartServiceOperation(
			ctx,
			serviceUuid,
			config,
			serviceRegistration,
			enclaveNetworkId,
			enclaveObjAttrsProvider,
			freeIpAddrProvider,
			dockerManager,
			restartPolicy,
			logsCollectorAddress,
			logsCollectorLabels,
		)
	}

	successfulServicesObjs, failedOperations := operation_parallelizer.RunOperationsInParallel(startServiceOperations)

	for uuid, data := range successfulServicesObjs {
		serviceUuid := service.ServiceUUID(uuid)
		serviceObj, ok := data.(*service.Service)
		if !ok {
			return nil, nil, stacktrace.NewError(
				"An error occurred downcasting data returned from the start user service operation for service with UUID: '%v'. "+
					"This is a Kurtosis bug. Make sure the desired type is actually being returned in the corresponding Operation.", serviceUuid)
		}
		successfulServices[serviceUuid] = serviceObj
	}

	for uuid, err := range failedOperations {
		serviceUuid := service.ServiceUUID(uuid)
		failedServices[serviceUuid] = err
	}

	return successfulServices, failedServices, nil
}

func createStartServiceOperation(
	ctx context.Context,
	serviceUUID service.ServiceUUID,
	serviceConfig *service.ServiceConfig,
	serviceRegistration *service.ServiceRegistration,
	enclaveNetworkId string,
	enclaveObjAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	freeIpAddrProvider *free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
	restartPolicy docker_manager.RestartPolicy,
	logsCollectorAddress string,
	logsCollectorLabels logs_collector_functions.LogsCollectorLabels,
) operation_parallelizer.Operation {
	id := serviceRegistration.GetName()
	privateIpAddr := serviceRegistration.GetPrivateIP()

	return func() (interface{}, error) {
		filesArtifactsExpansion := serviceConfig.GetFilesArtifactsExpansion()
		persistentDirectories := serviceConfig.GetPersistentDirectories()
		containerImageName := serviceConfig.GetContainerImageName()
		privatePorts := serviceConfig.GetPrivatePorts()
		publicPorts := serviceConfig.GetPublicPorts()
		entrypointArgs := serviceConfig.GetEntrypointArgs()
		cmdArgs := serviceConfig.GetCmdArgs()
		envVars := serviceConfig.GetEnvVars()
		cpuAllocationMillicpus := serviceConfig.GetCPUAllocationMillicpus()
		memoryAllocationMegabytes := serviceConfig.GetMemoryAllocationMegabytes()
		privateIPAddrPlaceholder := serviceConfig.GetPrivateIPAddrPlaceholder()
		user := serviceConfig.GetUser()

		// We replace the placeholder value with the actual private IP address
		privateIPAddrStr := privateIpAddr.String()
		for index := range entrypointArgs {
			entrypointArgs[index] = strings.Replace(entrypointArgs[index], privateIPAddrPlaceholder, privateIPAddrStr, unlimitedReplacements)
		}
		for index := range cmdArgs {
			cmdArgs[index] = strings.Replace(cmdArgs[index], privateIPAddrPlaceholder, privateIPAddrStr, unlimitedReplacements)
		}
		for key := range envVars {
			envVars[key] = strings.Replace(envVars[key], privateIPAddrPlaceholder, privateIPAddrStr, unlimitedReplacements)
		}

		volumeMounts := map[string]string{}
		shouldDeleteVolumes := true
		if filesArtifactsExpansion != nil {
			candidateVolumeMounts, err := doFilesArtifactExpansionAndGetUserServiceVolumes(
				ctx,
				serviceUUID,
				enclaveObjAttrsProvider,
				freeIpAddrProvider,
				enclaveNetworkId,
				filesArtifactsExpansion.ExpanderImage,
				filesArtifactsExpansion.ExpanderEnvVars,
				filesArtifactsExpansion.ExpanderDirpathsToServiceDirpaths,
				dockerManager,
			)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred doing files artifacts expansion to get user service volumes")
			}
			defer func() {
				if shouldDeleteVolumes {
					for volumeName := range candidateVolumeMounts {
						// Use background context, so we delete these even if input context was cancelled
						if err := dockerManager.RemoveVolume(context.Background(), volumeName); err != nil {
							logrus.Errorf("Starting the service failed so we tried to delete files artifact expansion volume '%v' that we created, but doing so threw an error:\n%v", volumeName, err)
							logrus.Errorf("You'll need to delete volume '%v' manually!", volumeName)
						}
					}
				}
			}()

			for dirpath, volumeName := range candidateVolumeMounts {
				if _, found := volumeMounts[dirpath]; found {
					return nil, stacktrace.NewError("An error occurred doing files artifacts expansion. Multiple volumes were mounted on the same path.")
				}
				volumeMounts[dirpath] = volumeName
			}
		}

		if persistentDirectories != nil {
			candidateVolumeMounts, err := getOrCreatePersistentDirectories(
				ctx,
				serviceUUID,
				enclaveObjAttrsProvider,
				persistentDirectories.ServiceDirpathToPersistentDirectory,
				dockerManager,
			)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating persistent directory volumes")
			}
			defer func() {
				if shouldDeleteVolumes {
					for volumeName := range candidateVolumeMounts {
						// Use background context, so we delete these even if input context was cancelled
						if err := dockerManager.RemoveVolume(context.Background(), volumeName); err != nil {
							logrus.Errorf("Starting the service failed so we tried to delete persistent directory volume '%v' that we created, but doing so threw an error:\n%v", volumeName, err)
							logrus.Errorf("You'll need to delete volume '%v' manually!", volumeName)
						}
					}
				}
			}()
			for dirpath, volumeName := range candidateVolumeMounts {
				if _, found := volumeMounts[dirpath]; found {
					return nil, stacktrace.NewError("An error occurred creating persistent directory volumes. Multiple volumes were mounted on the same path.")
				}
				volumeMounts[dirpath] = volumeName
			}
		}

		containerAttrs, err := enclaveObjAttrsProvider.ForUserServiceContainer(
			id,
			serviceUUID,
			privateIpAddr,
			privatePorts,
			serviceConfig.GetLabels(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while trying to get the user service container attributes for user service with UUID '%v'", serviceUUID)
		}
		containerName := containerAttrs.GetName()

		labelStrs := map[string]string{}
		for labelKey, labelValue := range containerAttrs.GetLabels() {
			labelStrs[labelKey.GetString()] = labelValue.GetString()
		}

		dockerUsedPorts := map[nat.Port]docker_manager.PortPublishSpec{}
		for portId, privatePortSpec := range privatePorts {
			dockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privatePortSpec)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred converting private port spec '%v' to a Docker port", portId)
			}
			//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
			if len(publicPorts) > 0 {
				publicPortSpec, found := publicPorts[portId]
				if !found {
					return nil, stacktrace.NewError("Expected to receive public port with ID '%v' bound to private port number '%v', but it was not found", portId, privatePortSpec.GetNumber())
				}
				dockerUsedPorts[dockerPort] = docker_manager.NewManualPublishingSpec(publicPortSpec.GetNumber())
			} else {
				dockerUsedPorts[dockerPort] = docker_manager.NewAutomaticPublishingSpec()
			}
		}

		// if logsCollectorAddress == "" {
		// 	return nil, stacktrace.NewError("Expected to have a logs collector server address value to send the user service logs, but it is empty")
		// }

		// The container will be configured to send the logs to the Fluentbit logs collector server
		// fluentdLoggingDriverCnfg := docker_manager.NewFluentdLoggingDriver(
		// 	logsCollectorAddress,
		// 	logsCollectorLabels,
		// )

		createAndStartArgsBuilder := docker_manager.NewCreateAndStartContainerArgsBuilder(
			containerImageName,
			containerName.GetString(),
			enclaveNetworkId,
		).WithStaticIP(
			privateIpAddr,
		).WithUsedPorts(
			dockerUsedPorts,
		).WithEnvironmentVariables(
			envVars,
		).WithLabels(
			labelStrs,
		).WithAlias(
			string(id),
		).WithCPUAllocationMillicpus(
			cpuAllocationMillicpus,
		).WithMemoryAllocationMegabytes(
			memoryAllocationMegabytes,
		).WithSkipAddingToBridgeNetworkIfStaticIpIsSet(
			skipAddingUserServiceToBridgeNetwork,
		).WithContainerInitEnabled(
			true,
		).WithVolumeMounts(
			volumeMounts,
		).WithRestartPolicy(
			restartPolicy,
		).WithUser(user)

		if entrypointArgs != nil {
			createAndStartArgsBuilder.WithEntrypointArgs(entrypointArgs)
		}
		if cmdArgs != nil {
			createAndStartArgsBuilder.WithCmdArgs(cmdArgs)
		}

		createAndStartArgs := createAndStartArgsBuilder.Build()

		containerId, hostMachinePortBindings, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting the user service container for user service with UUID '%v'", serviceUUID)
		}
		shouldKillContainer := true
		defer func() {
			if shouldKillContainer {
				// TODO switch to removing the container, so that the service registration is still viable?
				// NOTE: We use the background context here so that the kill will still go off even if the reason for
				// the failure was the original context being cancelled
				if err := dockerManager.KillContainer(context.Background(), containerId); err != nil {
					logrus.Errorf(
						"Launching user service container '%v' with container ID '%v' didn't complete successfully so we "+
							"tried to kill the container we started, but doing so exited with an error:\n%v",
						containerName.GetString(),
						containerId,
						err)
					logrus.Errorf("ACTION REQUIRED: You'll need to manually stop user service container with ID '%v'!!!!!!", containerId)
				}
			}
		}()

		_, _, maybePublicIp, maybePublicPortSpecs, err := shared_helpers.GetIpAndPortInfoFromContainer(
			containerName.GetString(),
			labelStrs,
			hostMachinePortBindings,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the public IP and ports from container '%v'", containerName)
		}

		serviceObjectPtr := service.NewService(
			serviceRegistration,
			privatePorts,
			maybePublicIp,
			maybePublicPortSpecs,
			container.NewContainer(
				container.ContainerStatus_Running,
				containerImageName,
				entrypointArgs,
				cmdArgs,
				envVars),
		)

		shouldDeleteVolumes = false
		shouldKillContainer = false
		return serviceObjectPtr, nil
	}
}

// Ensure that provided [privatePorts] and [publicPorts] are one to one by checking:
// - There is a matching publicPort for every portID in privatePorts
// - There are the same amount of private and public ports
// If error is nil, the public and private ports are one to one.
func checkPrivateAndPublicPortsAreOneToOne(privatePorts map[string]*port_spec.PortSpec, publicPorts map[string]*port_spec.PortSpec) error {
	if len(privatePorts) != len(publicPorts) {
		return stacktrace.NewError("The received private ports length and the public ports length are not equal. Received '%v' private ports and '%v' public ports", len(privatePorts), len(publicPorts))
	}

	for portID, privatePortSpec := range privatePorts {
		if _, found := publicPorts[portID]; !found {
			return stacktrace.NewError("Expected to receive public port with ID '%v' bound to private port number '%v', but it was not found", portID, privatePortSpec.GetNumber())
		}
	}
	return nil
}

// Registers a user service for each given serviceName, allocating each an IP and ServiceUUID
func registerUserServices(
	enclaveUuid enclave.EnclaveUUID,
	serviceNames map[service.ServiceName]bool,
	serviceRegistrationRepository *service_registration.ServiceRegistrationRepository,
	freeIpAddrProvider *free_ip_addr_tracker.FreeIpAddrTracker) (map[service.ServiceName]*service.ServiceRegistration, map[service.ServiceName]error, error) {
	successfulServicesPool := map[service.ServiceName]*service.ServiceRegistration{}
	failedServicesPool := map[service.ServiceName]error{}

	// If no service Names passed in, return empty maps
	if len(serviceNames) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	successfulRegistrations := map[service.ServiceName]*service.ServiceRegistration{}
	failedRegistrations := map[service.ServiceName]error{}
	for serviceName := range serviceNames {
		ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
		if err != nil {
			failedRegistrations[serviceName] = stacktrace.Propagate(err, "An error occurred getting a free IP address to give to service '%v' in enclave '%v'", serviceName, enclaveUuid)
			continue
		}
		shouldFreeIp := true
		defer func() {
			if shouldFreeIp {
				if err = freeIpAddrProvider.ReleaseIpAddr(ipAddr); err != nil {
					logrus.Errorf("Error releasing IP address '%v'", ipAddr)
				}
			}
		}()

		uuidStr, err := uuid_generator.GenerateUUIDString()
		if err != nil {
			failedRegistrations[serviceName] = stacktrace.Propagate(err, "An error occurred generating a UUID to use for the service UUID")
			continue
		}

		serviceUuid := service.ServiceUUID(uuidStr)
		registration := service.NewServiceRegistration(
			serviceName,
			serviceUuid,
			enclaveUuid,
			ipAddr,
			string(serviceName), // in Docker, hostname = serviceName because we're setting the "alias" of the container to serviceName
		)

		if err := serviceRegistrationRepository.Save(registration); err != nil {
			failedRegistrations[serviceName] = stacktrace.Propagate(err, "An error occurred saving service registration '%+v' for service '%s'", registration, serviceName)
		}

		shouldRemoveRegistration := true
		defer func() {
			if shouldRemoveRegistration {
				if err := serviceRegistrationRepository.Delete(serviceName); err != nil {
					logrus.Errorf("We tried to delete the service registration for service  '%s' we had stored but failed:\n%v", serviceName, err)
				}
			}
		}()

		shouldFreeIp = false
		shouldRemoveRegistration = false
		successfulRegistrations[serviceName] = registration
	}

	// Add operations to their respective pools
	for serviceName, serviceRegistration := range successfulRegistrations {
		successfulServicesPool[serviceName] = serviceRegistration
	}

	for serviceName, serviceErr := range failedRegistrations {
		failedServicesPool[serviceName] = serviceErr
	}

	return successfulRegistrations, failedRegistrations, nil
}
