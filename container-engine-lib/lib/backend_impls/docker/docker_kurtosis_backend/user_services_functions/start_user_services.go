package user_service_functions

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sync"
)

func StartUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	containerImageName string,
	privatePorts map[string]*port_spec.PortSpec,
	publicPorts map[string]*port_spec.PortSpec, //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	filesArtifactsExpansion *files_artifacts_expansion.FilesArtifactsExpansion,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration,
	serviceRegistrationMutex *sync.Mutex,
	enclaveFreeIpProviders map[enclave.EnclaveID]*lib.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (*service.Service, error) {

	//Sanity check for port bindings
	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	if publicPorts != nil && len(publicPorts) > 0 {

		if len(privatePorts) != len(publicPorts) {
			return nil, stacktrace.NewError("The received private ports length and the public ports length are not equal, received '%v' private ports and '%v' public ports", len(privatePorts), len(publicPorts))
		}

		for portId, privatePortSpec := range privatePorts {
			if _, found := publicPorts[portId]; !found {
				return nil, stacktrace.NewError("Expected to receive public port with ID '%v' bound to private port number '%v', but it was not found", portId, privatePortSpec.GetNumber())
			}
		}
	}

	serviceRegistrationMutex.Lock()
	defer serviceRegistrationMutex.Unlock()

	enclaveNetwork, err := shared_helpers.GetEnclaveNetworkByEnclaveId(ctx, enclaveId, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}
	enclaveNetworkId := enclaveNetwork.GetId()

	// Needed for files artifacts expansion container
	freeIpAddrProvider, found := enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"Received a request to start service with GUID '%v' in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			serviceGuid,
			enclaveId,
		)
	}

	// Find the registration
	registrationsForEnclave, found := serviceRegistrations[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the start service request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveId,
		)
	}
	serviceRegistration, found := registrationsForEnclave[serviceGuid]
	if !found {
		return nil, stacktrace.NewError(
			"Cannot start service '%v' because no preexisting registration has been made for the service",
			serviceGuid,
		)
	}
	serviceId := serviceRegistration.GetID()
	privateIpAddr := serviceRegistration.GetPrivateIP()

	// Find if a container has been associated with the registration yet
	preexistingServicesFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}
	preexistingServices, _, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, preexistingServicesFilters, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting preexisting containers for service '%v'", serviceGuid)
	}
	if len(preexistingServices) > 0 {
		return nil, stacktrace.Propagate(err, "Cannot start service '%v'; a container already exists for the service", serviceGuid)
	}

	enclaveObjAttrsProvider, err := objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	volumeMounts := map[string]string{}
	shouldDeleteVolumes := true
	if filesArtifactsExpansion != nil {
		candidateVolumeMounts, err := doFilesArtifactExpansionAndGetUserServiceVolumes(
			ctx,
			serviceGuid,
			enclaveObjAttrsProvider,
			freeIpAddrProvider,
			enclaveNetworkId,
			filesArtifactsExpansion.ExpanderImage,
			filesArtifactsExpansion.ExpanderEnvVars,
			filesArtifactsExpansion.ExpanderDirpathsToServiceDirpaths,
			dockerManager,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred doing files artifacts expansion to get user service volumes",
			)
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
		serviceId,
		serviceGuid,
		privateIpAddr,
		privatePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the user service container attributes for user service with GUID '%v'", serviceGuid)
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
		string(serviceId),
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
		return nil, stacktrace.Propagate(err, "An error occurred starting the user service container for user service with GUID '%v'", serviceGuid)
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

	result := service.NewService(
		serviceRegistration,
		container_status.ContainerStatus_Running,
		privatePorts,
		maybePublicIp,
		maybePublicPortSpecs,
	)

	shouldDeleteVolumes = false
	shouldKillContainer = false
	return result, nil
}

func StartUserServices(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	services map[service.ServiceGUID]*service.ServiceConfig,
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration,
	serviceRegistrationMutex *sync.Mutex,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	enclaveFreeIpProviders map[enclave.EnclaveID]*lib.FreeIpAddrTracker,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceGUID]service.Service,
	map[service.ServiceGUID]error,
	error,
) {
	serviceRegistrationMutex.Lock()
	defer serviceRegistrationMutex.Unlock()

	// Sanity check for port bindings on all services
	for guid, config := range services {
	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
		publicPorts := config.GetPublicPorts()
		privatePorts := config.GetPrivatePorts()
		if publicPorts != nil && len(publicPorts) > 0 {
			if len(privatePorts) != len(publicPorts) {
				return nil, nil, stacktrace.NewError("The received private ports length and the public ports length are not equal for service with guid `%v`. Received '%v' private ports and '%v' public ports", guid, len(privatePorts), len(publicPorts))
			}

			for portId, privatePortSpec := range privatePorts {
				if _, found := publicPorts[portId]; !found {
					return nil, nil, stacktrace.NewError("Expected to receive public port with ID '%v' bound to private port number '%v' for service with guid `%v`, but it was not found", portId, privatePortSpec.GetNumber(), guid)
				}
			}
		}
	}

	enclaveNetwork, err := shared_helpers.GetEnclaveNetworkByEnclaveId(ctx, enclaveID, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveID)
	}
	enclaveNetworkId := enclaveNetwork.GetId()

	freeIpAddrProvider, found := enclaveFreeIpProviders[enclaveID]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to start services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the start request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveID,
		)
	}

	// Find the registrations
	registrationsForEnclave, found := serviceRegistrations[enclaveID]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the start service request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveID,
		)
	}

	// Check that all services have valid registrations attached
	for serviceGUID, _ := range services {
		_, found := registrationsForEnclave[serviceGUID]
		if !found {
			return nil, nil, stacktrace.NewError(
				"Cannot start service '%v' because no preexisting registration has been made for the service",
				serviceGUID,
			)
		}
	}

	// Find if a container has already been associated with any of the registrations yet
	serviceGUIDs := map[service.ServiceGUID]bool{}
	for guid := range services{
		serviceGUIDs[guid] = true
	}
	preexistingServicesFilters := &service.ServiceFilters{
		GUIDs: serviceGUIDs,
	}
	preexistingServices, _, err := shared_helpers.GetMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveID, preexistingServicesFilters, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting preexisting containers for the services.")
	}
	if len(preexistingServices) > 0 {
		var preexistingServiceGUIDs []service.ServiceGUID
		for guid := range preexistingServices {
			preexistingServiceGUIDs = append(preexistingServiceGUIDs, guid)
		}
		return nil, nil, stacktrace.Propagate(err, "Cannot start services '%v'; because containers already exists for those services.", preexistingServiceGUIDs)
	}


	enclaveObjAttrsProvider, err := objAttrsProvider.ForEnclave(enclaveID)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveID)
	}

	successfulStarts, failedStarts, err := runStartServiceOperationsInParallel(
		ctx,
		enclaveNetworkId,
		services,
		registrationsForEnclave,
		enclaveObjAttrsProvider,
		freeIpAddrProvider,
		dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to start services in parallel.")
	}

	return successfulStarts, failedStarts, nil
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
	map[service.ServiceGUID]service.Service,
	map[service.ServiceGUID]error,
	error,
) {
	operations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for guid, config := range serviceConfigs {
		operations[operation_parallelizer.OperationID(guid)] = createStartServiceOperation(
			ctx,
			guid,
			config,
			serviceRegistrations[guid],
			enclaveNetworkId,
			enclaveObjAttrsProvider,
			freeIpAddrProvider,
			dockerManager)
	}

	successfulServicesObjs, failedOps := operation_parallelizer.RunOperationsInParallel(operations)

	successfulServices := map[service.ServiceGUID]service.Service{}
	failedServices := map[service.ServiceGUID]error{}

	for id, data := range successfulServicesObjs {
		serviceGUID := service.ServiceGUID(id)
		serviceObj, ok := data.(service.Service)
		if !ok {
			return nil, nil, stacktrace.NewError("Casting to a service object on data returned from the operation to start service with guid `%v` failed.", serviceGUID)
		}
		successfulServices[serviceGUID] = serviceObj
	}

	for id, err := range failedOps {
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
	dockerManager *docker_manager.DockerManager) func() (interface{}, error) {
	id := serviceRegistration.GetID()
	privateIpAddr := serviceRegistration.GetPrivateIP()

	var startServiceOp operation_parallelizer.Operation = func() (interface{}, error) {
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

		serviceObj := service.NewService(
			serviceRegistration,
			container_status.ContainerStatus_Running,
			privatePorts,
			maybePublicIp,
			maybePublicPortSpecs)

		shouldDeleteVolumes = false
		shouldKillContainer = false
		return serviceObj, nil
	}
	return startServiceOp
}