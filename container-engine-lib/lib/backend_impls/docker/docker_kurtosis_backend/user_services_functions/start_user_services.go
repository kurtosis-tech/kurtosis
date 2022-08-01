package user_service_functions

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_functions"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

const (
	shouldGetStoppedContainersWhenGettingServiceInfo = true
)

// NOTE: Normally we'd have a "canonical" resource here, where that resource is always guaranteed to exist. For Kurtosis services,
// we want this to be the container engine's representation of a user service registration. Unfortunately, Docker has no way
// of representing a user service registration, so we store them in an in-memory map on the DockerKurtosisBackend. Therefore, an
// entry in that map is actually the canonical representation, which means that any of these fields could be nil!
type userServiceDockerResources struct {
	serviceContainer *types.Container

	// Will never be nil but may be empty if no expander volumes exist
	expanderVolumeNames []string
}

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
	filesArtifactsExpansion *backend_interface.FilesArtifactsExpansion,
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

	enclaveNetwork, err := shared_functions.GetEnclaveNetworkByEnclaveId(ctx, enclaveId, dockerManager)
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
	preexistingServices, _, err := getMatchingUserServiceObjsAndDockerResourcesNoMutex(ctx, enclaveId, preexistingServicesFilters, dockerManager)
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
		dockerPort, err := shared_functions.TransformPortSpecToDockerPort(privatePortSpec)
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

	_, _, maybePublicIp, maybePublicPortSpecs, err := getIpAndPortInfoFromContainer(
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

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets the service objects & Docker resources for services matching the given filters
// NOTE: Does not use registration information so does not need the mutex!
func getMatchingUserServiceObjsAndDockerResourcesNoMutex(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceGUID]*service.Service,
	map[service.ServiceGUID]*userServiceDockerResources,
	error,
) {
	matchingDockerResources, err := getMatchingUserServiceDockerResources(ctx, enclaveId, filters.GUIDs, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting matching user service resources")
	}

	matchingServiceObjs, err := getUserServiceObjsFromDockerResources(enclaveId, matchingDockerResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis service objects from user service Docker resources")
	}

	resultServiceObjs := map[service.ServiceGUID]*service.Service{}
	resultDockerResources := map[service.ServiceGUID]*userServiceDockerResources{}
	for guid, serviceObj := range matchingServiceObjs {
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[serviceObj.GetRegistration().GetGUID()]; !found {
				continue
			}
		}

		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[serviceObj.GetRegistration().GetID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[serviceObj.GetStatus()]; !found {
				continue
			}
		}

		dockerResources, found := matchingDockerResources[guid]
		if !found {
			// This should never happen; the Services map and the Docker resources maps should have the same GUIDs
			return nil, nil, stacktrace.Propagate(
				err,
				"Needed to return Docker resources for service with GUID '%v', but none was "+
					"found; this is a bug in Kurtosis",
				guid,
			)
		}

		resultServiceObjs[guid] = serviceObj
		resultDockerResources[guid] = dockerResources
	}
	return resultServiceObjs, resultDockerResources, nil
}

func getMatchingUserServiceDockerResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	maybeGuidsToMatch map[service.ServiceGUID]bool,
	dockerManager *docker_manager.DockerManager,
) (map[service.ServiceGUID]*userServiceDockerResources, error) {
	result := map[service.ServiceGUID]*userServiceDockerResources{}

	// Grab services, INDEPENDENT OF volumes
	userServiceContainerSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.EnclaveIDDockerLabelKey.GetString():     string(enclaveId),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.UserServiceContainerTypeDockerLabelValue.GetString(),
	}
	userServiceContainers, err := dockerManager.GetContainersByLabels(ctx, userServiceContainerSearchLabels, shouldGetStoppedContainersWhenGettingServiceInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service containers in enclave '%v' by labels: %+v", enclaveId, userServiceContainerSearchLabels)
	}

	for _, container := range userServiceContainers {
		serviceGuidStr, found := container.GetLabels()[label_key_consts.GUIDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Found user service container '%v' that didn't have expected GUID label '%v'", container.GetId(), label_key_consts.GUIDDockerLabelKey.GetString())
		}
		serviceGuid := service.ServiceGUID(serviceGuidStr)

		if maybeGuidsToMatch != nil && len(maybeGuidsToMatch) > 0 {
			if _, found := maybeGuidsToMatch[serviceGuid]; !found {
				continue
			}
		}

		resourceObj, found := result[serviceGuid]
		if !found {
			resourceObj = &userServiceDockerResources{}
		}
		resourceObj.serviceContainer = container
		result[serviceGuid] = resourceObj
	}

	// Grab volumes, INDEPENDENT OF whether there any containers
	filesArtifactExpansionVolumeSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():      label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.EnclaveIDDockerLabelKey.GetString():  string(enclaveId),
		label_key_consts.VolumeTypeDockerLabelKey.GetString(): label_value_consts.FilesArtifactExpansionVolumeTypeDockerLabelValue.GetString(),
	}
	matchingFilesArtifactExpansionVolumes, err := dockerManager.GetVolumesByLabels(ctx, filesArtifactExpansionVolumeSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact expansion volumes in enclave '%v' by labels: %+v", enclaveId, filesArtifactExpansionVolumeSearchLabels)
	}

	for _, volume := range matchingFilesArtifactExpansionVolumes {
		serviceGuidStr, found := volume.Labels[label_key_consts.UserServiceGUIDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Found files artifact expansion volume '%v' that didn't have expected service GUID label '%v'", volume.Name, label_key_consts.UserServiceGUIDDockerLabelKey.GetString())
		}
		serviceGuid := service.ServiceGUID(serviceGuidStr)

		if maybeGuidsToMatch != nil && len(maybeGuidsToMatch) > 0 {
			if _, found := maybeGuidsToMatch[serviceGuid]; !found {
				continue
			}
		}

		resourceObj, found := result[serviceGuid]
		if !found {
			resourceObj = &userServiceDockerResources{}
		}
		resourceObj.expanderVolumeNames = append(resourceObj.expanderVolumeNames, volume.Name)
		result[serviceGuid] = resourceObj
	}

	return result, nil
}

func getUserServiceObjsFromDockerResources(
	enclaveId enclave.EnclaveID,
	allDockerResources map[service.ServiceGUID]*userServiceDockerResources,
) (map[service.ServiceGUID]*service.Service, error) {
	result := map[service.ServiceGUID]*service.Service{}

	// If we have an entry in the map, it means there's at least one Docker resource
	for serviceGuid, resources := range allDockerResources {
		container := resources.serviceContainer

		// If we don't have a container, we don't have the service ID label which means we can't actually construct a Service object
		// The only case where this would happen is if, during deletion, we delete the container but an error occurred deleting the volumes
		if container == nil {
			return nil, stacktrace.NewError(
				"Service '%v' has Docker resources but not a container; this indicates that there the service's "+
					"container was deleted but errors occurred deleting the rest of the resources",
				serviceGuid,
			)
		}
		containerName := container.GetName()
		containerLabels := container.GetLabels()

		serviceIdStr, found := containerLabels[label_key_consts.IDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find label '%v' on container '%v' but label was missing", label_key_consts.IDDockerLabelKey.GetString(), containerName)
		}
		serviceId := service.ServiceID(serviceIdStr)

		privateIp, privatePorts, maybePublicIp, maybePublicPorts, err := getIpAndPortInfoFromContainer(
			containerName,
			containerLabels,
			container.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting IP & port info from container '%v'", container.GetName())
		}

		registration := service.NewServiceRegistration(
			serviceId,
			serviceGuid,
			enclaveId,
			privateIp,
		)

		containerStatus := container.GetStatus()
		isContainerRunning, found := shared_functions.IsContainerRunningDeterminer[containerStatus]
		if !found {
			return nil, stacktrace.NewError("No is-running determination found for status '%v' for container '%v'", containerStatus.String(), containerName)
		}
		serviceStatus := container_status.ContainerStatus_Stopped
		if isContainerRunning {
			serviceStatus = container_status.ContainerStatus_Running
		}

		result[serviceGuid] = service.NewService(
			registration,
			serviceStatus,
			privatePorts,
			maybePublicIp,
			maybePublicPorts,
		)
	}
	return result, nil
}

// TODO Extract this to DockerKurtosisBackend and use it everywhere, for Engines, Modules, and API containers?
func getIpAndPortInfoFromContainer(
	containerName string,
	labels map[string]string,
	hostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (
	resultPrivateIp net.IP,
	resultPrivatePortSpecs map[string]*port_spec.PortSpec,
	resultPublicIp net.IP,
	resultPublicPortSpecs map[string]*port_spec.PortSpec,
	resultErr error,
) {
	privateIpAddrStr, found := labels[label_key_consts.PrivateIPDockerLabelKey.GetString()]
	if !found {
		return nil, nil, nil, nil, stacktrace.NewError("Expected to find label '%v' on container '%v' but label was missing", label_key_consts.PrivateIPDockerLabelKey.GetString(), containerName)
	}
	privateIp := net.ParseIP(privateIpAddrStr)
	if privateIp == nil {
		return nil, nil, nil, nil, stacktrace.NewError("Couldn't parse private IP string '%v' on container '%v' to an IP address", privateIpAddrStr, containerName)
	}

	serializedPortSpecs, found := labels[label_key_consts.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, nil, nil, nil, stacktrace.NewError(
			"Expected to find port specs label '%v' on container '%v' but none was found",
			containerName,
			label_key_consts.PortSpecsDockerLabelKey.GetString(),
		)
	}

	privatePortSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		if err != nil {
			return nil, nil, nil, nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
		}
	}

	var containerPublicIp net.IP
	var publicPortSpecs map[string]*port_spec.PortSpec
	if hostMachinePortBindings == nil || len(hostMachinePortBindings) == 0 {
		return privateIp, privatePortSpecs, containerPublicIp, publicPortSpecs, nil
	}

	for portId, privatePortSpec := range privatePortSpecs {
		portPublicIp, publicPortSpec, err := shared_functions.GetPublicPortBindingFromPrivatePortSpec(privatePortSpec, hostMachinePortBindings)
		if err != nil {
			return nil, nil, nil, nil, stacktrace.Propagate(
				err,
				"An error occurred getting public port spec for private port '%v' with spec '%v/%v' on container '%v'",
				portId,
				privatePortSpec.GetNumber(),
				privatePortSpec.GetProtocol().String(),
				containerName,
			)
		}

		if containerPublicIp == nil {
			containerPublicIp = portPublicIp
		} else {
			if !containerPublicIp.Equal(portPublicIp) {
				return nil, nil, nil, nil, stacktrace.NewError(
					"Private port '%v' on container '%v' yielded a public IP '%v', which doesn't agree with "+
						"previously-seen public IP '%v'",
					portId,
					containerName,
					portPublicIp.String(),
					containerPublicIp.String(),
				)
			}
		}

		if publicPortSpecs == nil {
			publicPortSpecs = map[string]*port_spec.PortSpec{}
		}
		publicPortSpecs[portId] = publicPortSpec
	}

	return privateIp, privatePortSpecs, containerPublicIp, publicPortSpecs, nil
}
