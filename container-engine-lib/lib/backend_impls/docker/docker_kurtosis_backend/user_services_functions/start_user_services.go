package user_service_functions

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_components"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/free_ip_addr_tracker"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const (
	unlimitedReplacements = -1
)

func StartUserServices(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	services map[service.ServiceID]*service.ServiceConfig,
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration,
	serviceRegistrationMutex *sync.Mutex,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	enclaveFreeIpProviders map[enclave.EnclaveID]*free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceID]*service.Service,
	map[service.ServiceID]error,
	error,
) {
	serviceRegistrationMutex.Lock()
	defer serviceRegistrationMutex.Unlock()
	successfulServicesPool := map[service.ServiceID]*service.Service{}
	failedServicesPool := map[service.ServiceID]error{}

	// Check whether any services have been provided at all
	if len(services) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	freeIpAddrProvider, found := enclaveFreeIpProviders[enclaveID]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to start services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveID,
		)
	}

	var serviceIDsToRegister []service.ServiceID
	for serviceID, config := range services {
		if config.GetPrivateIPAddrPlaceholder() == "" {
			failedServicesPool[serviceID] = stacktrace.NewError("Service with ID '%v' has an empty private IP Address placeholder. Expect this to be of length greater than zero.", serviceID)
			continue
		}
		serviceIDsToRegister = append(serviceIDsToRegister, serviceID)
	}
	if len(serviceIDsToRegister) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	successfulRegistrations, failedRegistrations, err := registerUserServices(enclaveID, serviceIDsToRegister, serviceRegistrations, freeIpAddrProvider)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services with IDs '%v'", serviceIDsToRegister)
	}
	// Defer an undo to all the successful registrations in case an error occurs in future phases
	serviceGUIDsToRemove := map[service.ServiceGUID]bool{}
	for _, registration := range successfulRegistrations {
		serviceGUIDsToRemove[registration.GetGUID()] = true
	}
	defer func() {
		if len(serviceGUIDsToRemove) == 0 {
			return
		}
		userServiceFilters := &service.ServiceFilters{
			GUIDs: serviceGUIDsToRemove,
		}
		_, failedToDestroyGUIDs, err := destroyUserServicesUnlocked(ctx, enclaveID, userServiceFilters, serviceRegistrations, enclaveFreeIpProviders, dockerManager)
		if err != nil {
			logrus.Errorf("Attempted to destroy all services with GUIDs '%v' together but had no success. You must manually destroy the services! The following error had occurred:\n'%v'", serviceGUIDsToRemove, err)
			return
		}
		if len(failedToDestroyGUIDs) == 0 {
			return
		}
		for serviceID, registration := range successfulRegistrations {
			destroyErr, found := failedToDestroyGUIDs[registration.GetGUID()]
			if !found {
				continue
			}
			logrus.Errorf("Failed to destroy the service '%v' after it failed to start. You must manually destroy the service! The following error had occurred:\n'%v'", serviceID, destroyErr)
		}
	}()
	for serviceID, registrationError := range failedRegistrations {
		failedServicesPool[serviceID] = stacktrace.Propagate(registrationError, "Failed to register service with ID '%v'", serviceID)
	}

	serviceConfigsToStart := map[service.ServiceID]*service.ServiceConfig{}
	for serviceID, serviceConfig := range services {
		if _, found := successfulRegistrations[serviceID]; !found {
			continue
		}
		serviceConfigsToStart[serviceID] = serviceConfig
	}

	// If no services had successful registrations, return immediately
	// This is to prevent an empty filter being used to query for matching objects and resources, returning all services
	// and causing logic to break (eg. check for duplicate service GUIDs)
	// Making this check allows us to eject early and maintain a guarantee that objects and resources returned
	// are 1:1 with serviceGUIDs
	if len(serviceConfigsToStart) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	// Sanity check for port bindings on all services
	for serviceID, serviceConfig := range serviceConfigsToStart {
		publicPorts := serviceConfig.GetPublicPorts()
		if publicPorts != nil && len(publicPorts) > 0 {
			privatePorts := serviceConfig.GetPrivatePorts()
			err := checkPrivateAndPublicPortsAreOneToOne(privatePorts, publicPorts)
			if err != nil {
				failedServicesPool[serviceID] = stacktrace.Propagate(err, "Private and public ports for service with ID '%v' are not one to one.", serviceID)
				delete(serviceConfigsToStart, serviceID)
			}
		}
	}
	//TODO END huge hack to temporarily enable static ports for NEAR

	enclaveNetwork, err := shared_helpers.GetEnclaveNetworkByEnclaveId(ctx, enclaveID, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveID)
	}
	enclaveNetworkID := enclaveNetwork.GetId()

	enclaveObjAttrsProvider, err := objAttrsProvider.ForEnclave(enclaveID)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveID)
	}

	logsCollectorServiceAddress, err := shared_helpers.GetLogsCollectorServiceAddress(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs collector service address")
	}

	//The following docker labels will be added into the logs stream which is necessary for creating new tags
	//in the logs database and then use it for querying them to get the specific user service's logs
	logsCollectorLabels := logs_components.LogsCollectorLabels{
		label_key_consts.EnclaveIDDockerLabelKey.GetString(),
		label_key_consts.GUIDDockerLabelKey.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(),
	}

	successfulStarts, failedStarts, err := runStartServiceOperationsInParallel(
		ctx,
		enclaveNetworkID,
		serviceConfigsToStart,
		successfulRegistrations,
		enclaveObjAttrsProvider,
		freeIpAddrProvider,
		dockerManager,
		logsCollectorServiceAddress,
		logsCollectorLabels,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to start services in parallel.")
	}

	// Add operations to their respective pools
	for serviceID, service := range successfulStarts {
		successfulServicesPool[serviceID] = service
	}

	for serviceID, serviceErr := range failedStarts {
		failedServicesPool[serviceID] = serviceErr
	}

	// Do not remove services that were started successfully
	for _, service := range successfulServicesPool {
		guid := service.GetRegistration().GetGUID()
		delete(serviceGUIDsToRemove, guid)
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
	serviceConfigs map[service.ServiceID]*service.ServiceConfig,
	serviceRegistrations map[service.ServiceID]*service.ServiceRegistration,
	enclaveObjAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	freeIpAddrProvider *free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
	logsCollectorAddress logs_components.LogsCollectorAddress,
	logsCollectorLabels logs_components.LogsCollectorLabels,
) (
	map[service.ServiceID]*service.Service,
	map[service.ServiceID]error,
	error,
) {
	successfulServices := map[service.ServiceID]*service.Service{}
	failedServices := map[service.ServiceID]error{}

	startServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for serviceID, config := range serviceConfigs {
		serviceRegistration, found := serviceRegistrations[serviceID]
		if !found {
			failedServices[serviceID] = stacktrace.NewError("Failed to get service registration for service ID '%v' while creating start service operation. This should never happen. This is a Kurtosis bug.", serviceID)
			continue
		}
		startServiceOperations[operation_parallelizer.OperationID(serviceID)] = createStartServiceOperation(
			ctx,
			serviceRegistration.GetGUID(),
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

	for id, data := range successfulServicesObjs {
		serviceID := service.ServiceID(id)
		serviceObj, ok := data.(*service.Service)
		if !ok {
			return nil, nil, stacktrace.NewError(
				"An error occurred downcasting data returned from the start user service operation for service with ID: '%v'. "+
					"This is a Kurtosis bug. Make sure the desired type is actually being returned in the corresponding Operation.", serviceID)
		}
		successfulServices[serviceID] = serviceObj
	}

	for id, err := range failedOperations {
		serviceID := service.ServiceID(id)
		failedServices[serviceID] = err
	}

	return successfulServices, failedServices, nil
}

func createStartServiceOperation(
	ctx context.Context,
	serviceGUID service.ServiceGUID,
	serviceConfig *service.ServiceConfig,
	serviceRegistration *service.ServiceRegistration,
	enclaveNetworkId string,
	enclaveObjAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	freeIpAddrProvider *free_ip_addr_tracker.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
	logsCollectorAddress logs_components.LogsCollectorAddress,
	logsCollectorLabels logs_components.LogsCollectorLabels,
) operation_parallelizer.Operation {
	id := serviceRegistration.GetID()
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
		for index, _ := range entrypointArgs {
			entrypointArgs[index] = strings.Replace(entrypointArgs[index], privateIPAddrPlaceholder, privateIPAddrStr, unlimitedReplacements)
		}
		for index, _ := range cmdArgs {
			cmdArgs[index] = strings.Replace(cmdArgs[index], privateIPAddrPlaceholder, privateIPAddrStr, unlimitedReplacements)
		}
		for key, _ := range envVars {
			envVars[key] = strings.Replace(envVars[key], privateIPAddrPlaceholder, privateIPAddrStr, unlimitedReplacements)
		}

		volumeMounts := map[string]string{}
		shouldDeleteVolumes := true
		if filesArtifactsExpansion != nil {
			candidateVolumeMounts, err := doFilesArtifactExpansionAndGetUserServiceVolumes(
				ctx,
				serviceGUID,
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
			serviceGUID,
			privateIpAddr,
			privatePorts,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while trying to get the user service container attributes for user service with GUID '%v'", serviceGUID)
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

		containerId, hostMachinePortBindings, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred starting the user service container for user service with GUID '%v'", serviceGUID)
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

// Registers a user service for each given serviceID, allocating each an IP and ServiceGUID
func registerUserServices(
	enclaveId enclave.EnclaveID,
	serviceIDs []service.ServiceID,
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration,
	freeIpAddrProvider *free_ip_addr_tracker.FreeIpAddrTracker) (map[service.ServiceID]*service.ServiceRegistration, map[service.ServiceID]error, error) {
	successfulServicesPool := map[service.ServiceID]*service.ServiceRegistration{}
	failedServicesPool := map[service.ServiceID]error{}

	// If no service IDs passed in, return empty maps
	if len(serviceIDs) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	registrationsForEnclave, found := serviceRegistrations[enclaveId]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the registration request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveId,
		)
	}

	successfulRegistrations := map[service.ServiceID]*service.ServiceRegistration{}
	failedRegistrations := map[service.ServiceID]error{}
	for _, serviceID := range serviceIDs {
		ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
		if err != nil {
			failedRegistrations[serviceID] = stacktrace.Propagate(err, "An error occurred getting a free IP address to give to service '%v' in enclave '%v'", serviceID, enclaveId)
			continue
		}
		shouldFreeIp := true
		defer func() {
			if shouldFreeIp {
				freeIpAddrProvider.ReleaseIpAddr(ipAddr)
			}
		}()

		guid := service.ServiceGUID(fmt.Sprintf(
			"%v-%v",
			serviceID,
			time.Now().Unix(),
		))
		registration := service.NewServiceRegistration(
			serviceID,
			guid,
			enclaveId,
			ipAddr,
		)

		registrationsForEnclave[guid] = registration
		shouldRemoveRegistration := true
		defer func() {
			if shouldRemoveRegistration {
				delete(registrationsForEnclave, guid)

			}
		}()

		shouldFreeIp = false
		shouldRemoveRegistration = false
		successfulRegistrations[serviceID] = registration
	}

	// Add operations to their respective pools
	for serviceID, serviceRegistration := range successfulRegistrations {
		successfulServicesPool[serviceID] = serviceRegistration
	}

	for serviceID, serviceErr := range failedRegistrations {
		failedServicesPool[serviceID] = serviceErr
	}

	return successfulRegistrations, failedRegistrations, nil
}
