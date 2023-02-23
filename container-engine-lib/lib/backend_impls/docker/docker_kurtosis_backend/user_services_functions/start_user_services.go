package user_service_functions

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

const (
	unlimitedReplacements                = -1
	skipAddingUserServiceToBridgeNetwork = true
)

func RegisterUserServices(
	enclaveUuid enclave.EnclaveUUID,
	servicesToRegister map[service.ServiceName]bool,
	serviceRegistrationsForEnclave map[service.ServiceUUID]*service.ServiceRegistration,
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

	successfulRegistrations, failedRegistrations, err := registerUserServices(enclaveUuid, servicesToRegister, serviceRegistrationsForEnclave, freeIpProvidersForEnclave)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services with Names '%v'", servicesToRegister)
	}
	return successfulRegistrations, failedRegistrations, nil
}

// UnregisterUserServices unregisters all services currently registered for this enclave.
// If the service is not registered for this enclave, it no-ops and the service is returned as "successfully unregistered"
func UnregisterUserServices(
	serviceUUIDsToUnregister map[service.ServiceUUID]bool,
	enclaveServiceRegistrations map[service.ServiceUUID]*service.ServiceRegistration,
	freeIpAddrProviderForEnclave *free_ip_addr_tracker.FreeIpAddrTracker,
	serviceRegistrationMutex *sync.Mutex,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
) {
	serviceRegistrationMutex.Lock()
	defer serviceRegistrationMutex.Unlock()
	servicesSuccessfullyUnregistered := map[service.ServiceUUID]bool{}
	servicesFailed := map[service.ServiceUUID]error{}

	if len(serviceUUIDsToUnregister) == 0 {
		return servicesSuccessfullyUnregistered, servicesFailed
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
		} else {
			delete(enclaveServiceRegistrations, serviceUuid)
			servicesSuccessfullyUnregistered[serviceUuid] = true
		}
	}
	return servicesSuccessfullyUnregistered, servicesFailed
}

func StartUserServices(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	services map[service.ServiceUUID]*service.ServiceConfig,
	serviceRegistrations map[service.ServiceUUID]*service.ServiceRegistration,
	logsCollector *logs_collector.LogsCollector,
	logsCollectorAvailabilityChecker logs_collector_functions.LogsCollectorAvailabilityChecker,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	freeIpProviderForEnclave *free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceUUID]*service.Service,
	map[service.ServiceUUID]error,
	error,
) {
	successfulServicesPool := map[service.ServiceUUID]*service.Service{}
	failedServicesPool := map[service.ServiceUUID]error{}

	serviceConfigsToStart := map[service.ServiceUUID]*service.ServiceConfig{}
	for serviceUuid, serviceConfig := range services {
		if _, found := serviceRegistrations[serviceUuid]; !found {
			failedServicesPool[serviceUuid] = stacktrace.NewError("Attempted to start a service '%s' that is not registered to this enclave yet.", serviceUuid)
			continue
		}
		if serviceConfig.GetPrivateIPAddrPlaceholder() == "" {
			failedServicesPool[serviceUuid] = stacktrace.NewError("Service with UUID '%v' has an empty private IP Address placeholder. Expect this to be of length greater than zero.", serviceUuid)
			continue
		}
		serviceConfigsToStart[serviceUuid] = serviceConfig
	}

	// If no services had successful registrations, return immediately
	// This is to prevent an empty filter being used to query for matching objects and resources, returning all services
	// and causing logic to break (eg. check for duplicate service GUIDs)
	// Making this check allows us to eject early and maintain a guarantee that objects and resources returned
	// are 1:1 with serviceUUIDs
	if len(serviceConfigsToStart) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	// Sanity check for port bindings on all services
	for serviceUuid, serviceConfig := range serviceConfigsToStart {
		publicPorts := serviceConfig.GetPublicPorts()
		if publicPorts != nil && len(publicPorts) > 0 {
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
			stacktrace.Propagate(err, "An error occurred while waiting for the log container to become available")
	}

	//We use the public TCP address because the logging driver connection link is from the Docker demon to the logs collector container
	//so the direction is from the host machine to the container inside the Docker cluster
	logsCollectorServiceAddress, err := logsCollector.GetEnclaveNetworkTcpAddress()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the private TCP address")
	}

	//The following docker labels will be added into the logs stream which is necessary for creating new tags
	//in the logs database and then use it for querying them to get the specific user service's logs
	//even the 'enclaveUuid' value is used for Fluentbit to send it to Loki as the "X-Scope-OrgID" request's header
	//due Loki is now configured to use multi tenancy, and we established this relation: enclaveUuid = tenantID
	logsCollectorLabels := logs_collector_functions.GetKurtosisTrackedLogsCollectorLabels()

	successfulStarts, failedStarts, err := runStartServiceOperationsInParallel(
		ctx,
		enclaveNetworkID,
		serviceConfigsToStart,
		serviceRegistrations,
		enclaveObjAttrsProvider,
		freeIpProviderForEnclave,
		dockerManager,
		logsCollectorServiceAddress,
		logsCollectorLabels,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to start services in parallel.")
	}

	// Add operations to their respective pools
	for serviceUuid, successfullyStartedService := range successfulStarts {
		successfulServicesPool[serviceUuid] = successfullyStartedService
	}

	for serviceUuid, serviceErr := range failedStarts {
		failedServicesPool[serviceUuid] = serviceErr
	}

	logrus.Debugf("Started services '%v' succesfully while '%v' failed", successfulServicesPool, failedServicesPool)
	return successfulServicesPool, failedServicesPool, nil
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func runStartServiceOperationsInParallel(
	ctx context.Context,
	enclaveNetworkId string,
	serviceConfigs map[service.ServiceUUID]*service.ServiceConfig,
	serviceRegistrations map[service.ServiceUUID]*service.ServiceRegistration,
	enclaveObjAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	freeIpAddrProvider *free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
	logsCollectorAddress logs_collector.LogsCollectorAddress,
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
	logsCollectorAddress logs_collector.LogsCollectorAddress,
	logsCollectorLabels logs_collector_functions.LogsCollectorLabels,
) operation_parallelizer.Operation {
	id := serviceRegistration.GetName()
	privateIpAddr := serviceRegistration.GetPrivateIP()

	return func() (interface{}, error) {
		filesArtifactsExpansion := serviceConfig.GetFilesArtifactsExpansion()
		containerImageName := serviceConfig.GetContainerImageName()
		privatePorts := serviceConfig.GetPrivatePorts()
		publicPorts := serviceConfig.GetPublicPorts()
		entrypointArgs := serviceConfig.GetEntrypointArgs()
		cmdArgs := serviceConfig.GetCmdArgs()
		envVars := serviceConfig.GetEnvVars()
		cpuAllocationMillicpus := serviceConfig.GetCPUAllocationMillicpus()
		memoryAllocationMegabytes := serviceConfig.GetMemoryAllocationMegabytes()
		privateIPAddrPlaceholder := serviceConfig.GetPrivateIPAddrPlaceholder()

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
						// Use background context so we delete these even if input context was cancelled
						if err := dockerManager.RemoveVolume(context.Background(), volumeName); err != nil {
							logrus.Errorf("Starting the service failed so we tried to delete files artifact expansion volume '%v' that we created, but doing so threw an error:\n%v", volumeName, err)
							logrus.Errorf("You'll need to delete volume '%v' manually!", volumeName)
						}
					}
				}
			}()
			volumeMounts = candidateVolumeMounts
		}

		containerAttrs, err := enclaveObjAttrsProvider.ForUserServiceContainer(
			id,
			serviceUUID,
			privateIpAddr,
			privatePorts,
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
			if publicPorts != nil && len(publicPorts) > 0 {
				publicPortSpec, found := publicPorts[portId]
				if !found {
					return nil, stacktrace.NewError("Expected to receive public port with ID '%v' bound to private port number '%v', but it was not found", portId, privatePortSpec.GetNumber())
				}
				dockerUsedPorts[dockerPort] = docker_manager.NewManualPublishingSpec(publicPortSpec.GetNumber())
			} else {
				dockerUsedPorts[dockerPort] = docker_manager.NewAutomaticPublishingSpec()
			}
		}

		if logsCollectorAddress == "" {
			return nil, stacktrace.NewError("Expected to have a logs collector server address value to send the user service logs, but it is empty")
		}

		logsCollectorAddressStr := string(logsCollectorAddress)
		//The container will be configured to send the logs to the Fluentbit logs collector server
		fluentdLoggingDriverCnfg := docker_manager.NewFluentdLoggingDriver(
			logsCollectorAddressStr,
			logsCollectorLabels,
		)

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
		).WithLoggingDriver(
			fluentdLoggingDriverCnfg,
		).WithSkipAddingToBridgeNetworkIfStaticIpIsSet(
			skipAddingUserServiceToBridgeNetwork,
		)

		if entrypointArgs != nil {
			createAndStartArgsBuilder.WithEntrypointArgs(entrypointArgs)
		}
		if cmdArgs != nil {
			createAndStartArgsBuilder.WithCmdArgs(cmdArgs)
		}
		if volumeMounts != nil {
			createAndStartArgsBuilder.WithVolumeMounts(volumeMounts)
		}

		createAndStartArgs := createAndStartArgsBuilder.Build()

		// Best-effort pull attempt
		if err = dockerManager.PullImage(ctx, containerImageName); err != nil {
			logrus.Warnf("Failed to pull the latest version of user service container image '%v'; you may be running an out-of-date version", containerImageName)
		}

		containerId, hostMachinePortBindings, err := dockerManager.CreateAndStartContainer(ctx, false, createAndStartArgs)
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
			container_status.ContainerStatus_Running,
			privatePorts,
			maybePublicIp,
			maybePublicPortSpecs)

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
	serviceRegistrationsForEnclave map[service.ServiceUUID]*service.ServiceRegistration,
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

		serviceRegistrationsForEnclave[serviceUuid] = registration
		shouldRemoveRegistration := true
		defer func() {
			if shouldRemoveRegistration {
				delete(serviceRegistrationsForEnclave, serviceUuid)
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
