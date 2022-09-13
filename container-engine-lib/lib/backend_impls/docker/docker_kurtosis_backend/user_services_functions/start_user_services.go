package user_service_functions

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func StartUserServices(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	services map[service.ServiceID]*service.ServiceConfig,
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration,
	serviceRegistrationMutex *sync.Mutex,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	enclaveFreeIpProviders map[enclave.EnclaveID]*lib.FreeIpAddrTracker,
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

	serviceIDs := make([]service.ServiceID, len(services))
	for serviceID := range services {
		serviceIDs = append(serviceIDs, serviceID)
	}
	successfulRegistrations, failedRegistrations, err := registerUserServices(enclaveID, serviceIDs, serviceRegistrations, freeIpAddrProvider)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services with IDs '%v'", serviceIDs)
	}
	// Defer an undo to all the successful registrations in case an error occurs in future phases
	serviceIDsToRemove := map[service.ServiceID]bool{}
	for serviceID, _ := range successfulRegistrations {
		serviceIDsToRemove[serviceID] = true
	}
	defer func() {
		userServiceFilters := &service.ServiceFilters{
			IDs: serviceIDsToRemove,
		}
		_, failedToDestroyGUIDs, err := destroyUserServicesUnlocked(ctx, enclaveID, userServiceFilters, serviceRegistrations, enclaveFreeIpProviders, dockerManager)
		if err != nil {
			for serviceID, _ := range serviceIDsToRemove {
				failedServicesPool[serviceID] = stacktrace.Propagate(err, "Attempted to destroy all services with serviceIDs '%v' together but had no success. You must manually destroy the service '%v'.", serviceIDsToRemove, serviceID)
			}
			return
		}
		if len(failedToDestroyGUIDs) == 0 {
			return
		}
		for serviceID, registration := range successfulRegistrations {
			if err, found := failedToDestroyGUIDs[registration.GetGUID()]; found {
				failedServicesPool[serviceID] = stacktrace.Propagate(err, "Failed to destroy the service '%v' after it failed to start. You must manually destroy the service!", serviceID)
			}
		}
	}()
	for serviceID, registrationError := range failedRegistrations {
		failedServicesPool[serviceID] = stacktrace.Propagate(registrationError, "Failed to register service with ID '%v'", serviceID)
	}

	serviceConfigsToStart := map[service.ServiceGUID]*service.ServiceConfig{}
	successfulRegistrationsByGUID := map[service.ServiceGUID]*service.ServiceRegistration{}
	for serviceID, successfulRegistration := range successfulRegistrations {
		guid := successfulRegistration.GetGUID()
		config := services[serviceID]
		config.ReplacePlaceholderWithPrivateIPAddr(successfulRegistration.GetPrivateIP().String())
		serviceConfigsToStart[guid] = config
		successfulRegistrationsByGUID[guid] = successfulRegistration
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
	for serviceGUID, serviceConfig := range serviceConfigsToStart {
		publicPorts := serviceConfig.GetPublicPorts()
		if publicPorts != nil && len(publicPorts) > 0 {
			privatePorts := serviceConfig.GetPrivatePorts()
			err := checkPrivateAndPublicPortsAreOneToOne(privatePorts, publicPorts)
			if err != nil {
				serviceID := successfulRegistrationsByGUID[serviceGUID].GetID()
				failedServicesPool[serviceID] = stacktrace.Propagate(err, "Private and public ports for service with ID '%v' are not one to one.", serviceID)
				delete(serviceConfigsToStart, serviceGUID)
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

	successfulStarts, failedStarts, err := runStartServiceOperationsInParallel(
		ctx,
		enclaveNetworkID,
		serviceConfigsToStart,
		successfulRegistrationsByGUID,
		enclaveObjAttrsProvider,
		freeIpAddrProvider,
		dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to start services in parallel.")
	}

	// Add operations to their respective pools
	for serviceGUID, service := range successfulStarts {
		serviceID := successfulRegistrationsByGUID[serviceGUID].GetID()
		successfulServicesPool[serviceID] = service
	}

	for serviceGUID, serviceErr := range failedStarts {
		serviceID := successfulRegistrationsByGUID[serviceGUID].GetID()
		failedServicesPool[serviceID] = serviceErr
	}

	// Do not remove services that were started successfully
	for serviceID, _ := range successfulServicesPool {
		delete(serviceIDsToRemove, serviceID)
	}
	return successfulServicesPool, failedServicesPool, nil
}

// ====================================================================================================
//                       Private helper functions
// ====================================================================================================
func runStartServiceOperationsInParallel(
	ctx context.Context,
	enclaveNetworkId string,
	serviceConfigs map[service.ServiceGUID]*service.ServiceConfig,
	serviceRegistrations map[service.ServiceGUID]*service.ServiceRegistration,
	enclaveObjAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	freeIpAddrProvider *lib.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceGUID]*service.Service,
	map[service.ServiceGUID]error,
	error,
) {
	startServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for guid, config := range serviceConfigs {
		startServiceOperations[operation_parallelizer.OperationID(guid)] = createStartServiceOperation(
			ctx,
			guid,
			config,
			serviceRegistrations[guid],
			enclaveNetworkId,
			enclaveObjAttrsProvider,
			freeIpAddrProvider,
			dockerManager)
	}

	successfulServicesObjs, failedOperations := operation_parallelizer.RunOperationsInParallel(startServiceOperations)

	successfulServices := map[service.ServiceGUID]*service.Service{}
	failedServices := map[service.ServiceGUID]error{}

	for id, data := range successfulServicesObjs {
		serviceGUID := service.ServiceGUID(id)
		serviceObj, ok := data.(*service.Service)
		if !ok {
			return nil, nil, stacktrace.NewError(
				"An error occurred downcasting data returned from the start user service operation for service with GUID: '%v'. "+
					"This is a Kurtosis bug. Make sure the desired type is actually being returned in the corresponding Operation.", serviceGUID)
		}
		successfulServices[serviceGUID] = serviceObj
	}

	for id, err := range failedOperations {
		serviceGUID := service.ServiceGUID(id)
		failedServices[serviceGUID] = err
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
	freeIpAddrProvider *lib.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager) operation_parallelizer.Operation {
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
	freeIpAddrProvider *lib.FreeIpAddrTracker) (map[service.ServiceID]*service.ServiceRegistration, map[service.ServiceID]error, error) {
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

func destroyServiceAfterFailure(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration,
	enclaveFreeIpProviders map[enclave.EnclaveID]*lib.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager) error {
	destroyServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGUID: true,
		},
	}
	// Use background context in case the input one is cancelled
	_, erroredRegistrations, err := destroyUserServicesUnlocked(ctx, enclaveId, destroyServiceFilters, serviceRegistrations, enclaveFreeIpProviders, dockerManager)
	var errToReturn error
	if err != nil {
		errToReturn = err
	} else if destroyErr, found := erroredRegistrations[serviceGUID]; found {
		errToReturn = destroyErr
	}
	if errToReturn != nil {
		return stacktrace.NewError("Attempted to destroy the service with GUID'%v', but doing so threw an error:\n%v", serviceGUID, errToReturn)
	}
	return nil
}
