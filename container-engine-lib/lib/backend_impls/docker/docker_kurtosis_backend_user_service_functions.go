package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	// The location where the enclave data volume will be mounted
	//  on the user service container
	enclaveDataVolumeDirpathOnServiceContainer = "/kurtosis-data"

	shouldGetStoppedContainersWhenGettingServiceInfo = true
)

/*
DOCKER SERVICE LIFECYCLE EXPLANATION:

Kurtosis services are uniquely identified by a ServiceGUID and can have four states:
1. Registered = a GUID and an IP address in the enclave has been allocated for the service, but no user container is running
1. Running = user's container is running
1. Paused = user's container is paused
1. Stopped = user's container is stopped *and will not run again*
1. Destroyed = the service no longer exists

In Docker, we implement this like so:
- Registration: the DockerKurtosisBackend will keep an in-memory map of the registration info (IP & ServiceGUID), because there's no Docker
	object that corresponds to a registration
- Running: the user's container is running with the IP that was generated during registration
- Paused: the user's container is paused
- Stopped: the user's container is stopped (rather than deleted) so that logs are still accessible, and the IP that was
	allocated to the container has been freed (so that we don't consume the entire IP pool if a bunch of services are started
	and stopped). Because stopped Kurtosis services can never be restarted as of 2022-05-14, the releasing of the IP is fine
	to do because the container will never be restarted unless the user starts messing around with Docker directly.
- Destroyed: any container that was started is destroyed, and the IP address is freed (if not already freed because the
	service was previously stopped).

The benefits of this implementation:
- We can get the IP address before the service is started, which is crucial because certain user containers actually need
	to know their own IP when they start (e.g. Ethereum and Avalanche nodes require a flag to be passed in with their own IP)
- We can stop a service and free its memory/CPU resources while still preserving the logs for users
 */

// We'll try to use the nicer-to-use shells first before we drop down to the lower shells
var commandToRunWhenCreatingUserServiceShell = []string{
	"sh",
	"-c",
	"if command -v 'bash' > /dev/null; then echo \"Found bash on container; creating bash shell...\"; bash; else echo \"No bash found on container; dropping down to sh shell...\"; sh; fi",
}

type userServiceDockerResources struct {
	// Nil if the service is purely registered and has no container started
	container *types.Container
}

// Its completeness is enforced via unit test
var userServiceStatusDeterminer = map[types.ContainerStatus]service.UserServiceStatus{
	types.ContainerStatus_Paused:     service.UserServiceStatus_Paused,
	types.ContainerStatus_Restarting: service.UserServiceStatus_Running,
	types.ContainerStatus_Running:    service.UserServiceStatus_Running,
	types.ContainerStatus_Removing:   service.UserServiceStatus_Stopped,
	types.ContainerStatus_Dead:       service.UserServiceStatus_Stopped,
	types.ContainerStatus_Created:    service.UserServiceStatus_Stopped,
	types.ContainerStatus_Exited:     service.UserServiceStatus_Stopped,
}

func (backend *DockerKurtosisBackend) RegisterService(
	_ context.Context,
	enclaveId enclave.EnclaveID,
	serviceId service.ServiceID,
) (*service.Service, error) {
	// Write mutex locked; modification of the service registration map is allowed
	backend.serviceRegistrationMutex.Lock()
	defer backend.serviceRegistrationMutex.Unlock()

	freeIpAddrProvider, found := backend.enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"Received a request to register service with ID '%v' in enclave '%v', but no free IP address provider was " +
				"defined for this enclave; this likely means that the registration request is being called where it shouldn't " +
				"be (i.e. outside the API container)",
			serviceId,
			enclaveId,
		)
	}

	enclaveServices, found := backend.serviceRegistrations[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"No services are being tracked for enclave '%v'; this likely means that the registration request is being called where it shouldn't " +
				"be (i.e. outside the API container)",
			enclaveId,
		)
	}

	ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a free IP address to give to service '%v' in enclave '%v'", serviceId, enclaveId)
	}
	shouldFreeIp := true
	defer func() {
		if shouldFreeIp {
			freeIpAddrProvider.ReleaseIpAddr(ipAddr)
		}
	}()

	// TODO Switch this, and all other GUIDs, to a UUID??
	guid := service.ServiceGUID(fmt.Sprintf(
		"%v-%v",
		serviceId,
		time.Now().Unix(),
	))
	registrationInfo := &registeredServiceInfo{
		enclaveId: enclaveId,
		id:        serviceId,
		guid:      guid,
		ip:        ipAddr,
	}

	enclaveServices[guid] = registrationInfo
	shouldRemoveRegistration := true
	defer func() {
		if shouldRemoveRegistration {
			delete(enclaveServices, guid)
		}
	}()

	// Registered-but-not-started services don't have any private ports
	var privatePorts map[string]*port_spec.PortSpec = nil
	// Registered-but-not-started services don't have public info
	var publicIp net.IP = nil
	var publicPorts map[string]*port_spec.PortSpec = nil
	result := service.NewService(
		serviceId,
		guid,
		service.UserServiceStatus_Registered,
		enclaveId,
		ipAddr,
		privatePorts,
		publicIp,
		publicPorts,
	)

	shouldFreeIp = false
	shouldRemoveRegistration = false
	return result, nil
}

func (backend *DockerKurtosisBackend) StartUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	guid service.ServiceGUID,
	containerImageName string,
	privatePorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	// Volume -> mountpoint on container
	filesArtifactVolumeMountDirpaths map[string]string,
) (
	*service.Service,
	error,
) {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	serviceBeforeStartingContainer, _, err := backend.getSingleServiceObjWithResourcesWithoutMutex(ctx, enclaveId, guid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service with GUID '%v' that should exist when starting the service", guid)
	}
	serviceId := serviceBeforeStartingContainer.GetID()
	serviceGuid := serviceBeforeStartingContainer.GetGUID()
	privateIp := serviceBeforeStartingContainer.GetPrivateIP()

	currentServiceStatus := serviceBeforeStartingContainer.GetStatus()
	if currentServiceStatus != service.UserServiceStatus_Registered {
		return nil, stacktrace.NewError(
			"Cannot start service '%v'; expected it to be in status '%v' but was '%v'",
			guid,
			service.UserServiceStatus_Registered.String(),
			currentServiceStatus.String(),
		)
	}

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForUserServiceContainer(
		serviceId,
		serviceGuid,
		privateIp,
		privatePorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the user service container attributes for user service with GUID '%v'", guid)
	}
	containerName := containerAttrs.GetName()

	labelStrs := map[string]string{}
	for labelKey, labelValue := range containerAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	enclaveDataVolumeName, err := backend.getEnclaveDataVolumeByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave data volume for enclave '%v'", enclaveId)
	}

	dockerUsedPorts := map[nat.Port]docker_manager.PortPublishSpec{}
	for portId, portSpec := range privatePorts {
		dockerPort, err := transformPortSpecToDockerPort(portSpec)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting private port spec '%v' to a Docker port", portId)
		}
		dockerUsedPorts[dockerPort] = docker_manager.NewAutomaticPublishingSpec()
	}

	volumeMounts := map[string]string{
		enclaveDataVolumeName: enclaveDataVolumeDirpathOnServiceContainer,
	}

	createAndStartArgsBuilder := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImageName,
		containerName.GetString(),
		enclaveNetwork.GetId(),
	).WithStaticIP(
		serviceBeforeStartingContainer.GetPrivateIP(),
	).WithUsedPorts(
		dockerUsedPorts,
	).WithEnvironmentVariables(
		envVars,
	).WithVolumeMounts(
		volumeMounts,
	).WithLabels(
		labelStrs,
	).WithAlias(
		string(serviceBeforeStartingContainer.GetID()),
	)
	if entrypointArgs != nil {
		createAndStartArgsBuilder.WithEntrypointArgs(entrypointArgs)
	}
	if cmdArgs != nil {
		createAndStartArgsBuilder.WithCmdArgs(cmdArgs)
	}
	if filesArtifactVolumeMountDirpaths != nil {
		createAndStartArgsBuilder.WithVolumeMounts(filesArtifactVolumeMountDirpaths)
	}
	createAndStartArgs := createAndStartArgsBuilder.Build()

	// Best-effort pull attempt
	if err = backend.dockerManager.PullImage(ctx, containerImageName); err != nil {
		logrus.Warnf("Failed to pull the latest version of user service container image '%v'; you may be running an out-of-date version", containerImageName)
	}

	containerId, hostMachinePortBindings, err := backend.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the user service container for user service with GUID '%v'", guid)
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			// NOTE: We use the background context here so that the kill will still go off even if the reason for
			// the failure was the original context being cancelled
			if err := backend.dockerManager.KillContainer(context.Background(), containerId); err != nil {
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

	_, maybePublicIp, maybePublicPortSpecs, err := getIpAndPortInfoFromContainer(
		containerName.GetString(),
		labelStrs,
		hostMachinePortBindings,
	)

	result := service.NewService(
		serviceId,
		serviceGuid,
		service.UserServiceStatus_Running,
		enclaveId,
		privateIp,
		privatePorts,
		maybePublicIp,
		maybePublicPortSpecs,
	)

	shouldKillContainer = false
	return result, nil
}

func (backend *DockerKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.Service,
	error,
) {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	userServices, _, err := backend.getMatchingUserServiceObjsAndDockerResourcesWithoutMutex(ctx, enclaveId, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}
	return userServices, nil
}

func (backend *DockerKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
) {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	_, allDockerResources, err := backend.getMatchingUserServiceObjsAndDockerResourcesWithoutMutex(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	//TODO use concurrency to improve perf
	successfulUserServicesLogs := map[service.ServiceGUID]io.ReadCloser{}
	erroredUserServices := map[service.ServiceGUID]error{}
	for guid, resourcesForService := range allDockerResources {
		container := resourcesForService.container
		if container == nil {
			erroredUserServices[guid] = stacktrace.NewError("Cannot get logs for service '%v' as it has no container")
			continue
		}

		readCloserLogs, err := backend.dockerManager.GetContainerLogs(ctx, container.GetId(), shouldFollowLogs)
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting logs for container '%v' for user service with GUID '%v'", container.GetName(), guid)
			erroredUserServices[guid] = serviceError
			continue
		}
		successfulUserServicesLogs[guid] = readCloserLogs
	}

	return successfulUserServicesLogs, erroredUserServices, nil
}

func (backend *DockerKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
) error {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	serviceObj, dockerResources, err := backend.getSingleServiceObjWithResourcesWithoutMutex(ctx, enclaveId, serviceGuid)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get information about service '%v' from Kurtosis backend.", serviceGuid)
	}
	if serviceObj.GetStatus() != service.UserServiceStatus_Running {
		return stacktrace.NewError("Cannot pause service '%v' because it is not in state '%v'", service.UserServiceStatus_Running.String())
	}
	container := dockerResources.container
	if container == nil {
		return stacktrace.NewError("Cannot pause service '%v' as it doesn't have a container to pause", serviceGuid)
	}
	if err = backend.dockerManager.PauseContainer(ctx, container.GetId()); err != nil {
		return stacktrace.Propagate(err, "Failed to pause container '%v' for service '%v' ", container.GetName(), serviceGuid)
	}
	return nil
}

func (backend *DockerKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
) error {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	serviceObj, dockerResources, err := backend.getSingleServiceObjWithResourcesWithoutMutex(ctx, enclaveId, serviceGuid)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get information about service '%v' from Kurtosis backend.", serviceGuid)
	}
	if serviceObj.GetStatus() != service.UserServiceStatus_Paused {
		return stacktrace.NewError("Cannot unpause service '%v' because it is not in state '%v'", service.UserServiceStatus_Paused.String())
	}
	container := dockerResources.container
	if container == nil {
		return stacktrace.NewError("Cannot unpause service '%v' as it doesn't have a container to pause", serviceGuid)
	}
	if err = backend.dockerManager.PauseContainer(ctx, container.GetId()); err != nil {
		return stacktrace.Propagate(err, "Failed to pause container '%v' for service '%v' ", container.GetName(), serviceGuid)
	}
	return nil
}

// NOTE: This is a blocking task!!!! If we need more perf we can make it async
func (backend *DockerKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
) (
	map[service.ServiceGUID]*exec_result.ExecResult,
	map[service.ServiceGUID]error,
	error,
) {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	userServiceGuids := map[service.ServiceGUID]bool{}
	for userServiceGuid := range userServiceCommands {
		userServiceGuids[userServiceGuid] = true
	}

	filters := &service.ServiceFilters{
		GUIDs: userServiceGuids,
	}
	_, dockerResources, err := backend.getMatchingUserServiceObjsAndDockerResourcesWithoutMutex(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	if len(userServiceCommands) != len(dockerResources) {
		return nil, nil, stacktrace.NewError(
			"The number of user services found '%v' is not equal to the number of user service to run exec commands on, '%v'",
			len(dockerResources),
			len(userServiceCommands),
		)
	}

	// TODO Parallelize to increase perf
	succesfulUserServiceExecResults := map[service.ServiceGUID]*exec_result.ExecResult{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}
	for guid, commandArgs := range userServiceCommands {
		dockerResources, found := dockerResources[guid]
		if !found {
			// Should never happen if our logic for pulling back resources is correct because we always return a Docker resources object, even if
			// no container exists
			erroredUserServiceGuids[guid] = stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because no Docker resources were found for it",
				commandArgs,
				guid,
			)
			continue
		}

		container := dockerResources.container
		if container == nil {
			erroredUserServiceGuids[guid] = stacktrace.NewError(
				"Cannot execute command '%+v' on service '%v' because it doesn't have a Docker container",
				commandArgs,
				guid,
			)
			continue
		}

		execOutputBuf := &bytes.Buffer{}
		exitCode, err := backend.dockerManager.RunExecCommand(
			ctx,
			container.GetId(),
			commandArgs,
			execOutputBuf,
		)
		if err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred executing command '%+v' on container '%v' for user service '%v'",
				commandArgs,
				container.GetName(),
				guid,
			)
			erroredUserServiceGuids[guid] = wrappedErr
			continue
		}
		newExecResult := exec_result.NewExecResult(exitCode, execOutputBuf.String())
		succesfulUserServiceExecResults[guid] = newExecResult
	}

	return succesfulUserServiceExecResults, erroredUserServiceGuids, nil
}

/*
func (backend *DockerKurtosisBackend) WaitForUserServiceHttpEndpointAvailability(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
	httpMethod wait_for_availability_http_methods.WaitForAvailabilityHttpMethod,
	port uint32,
	path string,
	requestBody string,
	expectedResponseBody string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
) (
	resultErr error,
) {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	_, userService, err := backend.getSingleUserService(ctx, enclaveId, serviceGUID)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting container ID and user service object for enclave ID '%v' and user service GUID '%v'", enclaveId, serviceGUID)
	}

	url := fmt.Sprintf("http://%v:%v/%v", userService.GetPrivateIP(), port, path)

	time.Sleep(time.Duration(initialDelayMilliseconds) * time.Millisecond)

	httpMethodStr := httpMethod.String()
	for i := uint32(0); i < retries; i++ {
		resp, err := makeHttpRequest(httpMethodStr, url, requestBody)
		if err == nil && doesHttpResponseMatchExpected(expectedResponseBody, resp.Body, url) {
			break
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(
			err,
			"The HTTP endpoint '%v' didn't return a success code, even after %v retries with %v milliseconds in between retries",
			url,
			retries,
			retriesDelayMilliseconds,
		)
	}

	return nil
}

 */

func (backend *DockerKurtosisBackend) GetConnectionWithUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
) (
	net.Conn,
	error,
) {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	_, serviceDockerResources, err := backend.getSingleServiceObjWithResourcesWithoutMutex(ctx, enclaveId, serviceGUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service object and Docker resources for service '%v' in enclave '%v'", serviceGUID, enclaveId)
	}
	container := serviceDockerResources.container
	if container == nil {
		return nil, stacktrace.NewError("Cannot get a connection to user service '%v' in enclave '%v' because no container exists for the service", serviceGUID, enclaveId)
	}

	hijackedResponse, err := backend.dockerManager.CreateContainerExec(ctx, container.GetId(), commandToRunWhenCreatingUserServiceShell)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a shell on user service with GUID '%v' in enclave '%v'", serviceGUID, enclaveId)
	}

	newConnection := hijackedResponse.Conn

	return newConnection, nil
}

// It returns io.ReadCloser which is a tar stream. It's up to the caller to close the reader.
func (backend *DockerKurtosisBackend) CopyFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	srcPath string,
)(
	io.ReadCloser,
	error,
) {
	// !!!!! THIS IS JUST A READ LOCK; YOU MAY NOT WRITE TO REGISTRATION INFO IN THIS METHOD !!!!!!!!!!
	backend.serviceRegistrationMutex.RLock()
	defer backend.serviceRegistrationMutex.RUnlock()

	_, serviceDockerResources, err := backend.getSingleServiceObjWithResourcesWithoutMutex(ctx, enclaveId, serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service with GUID '%v' in enclave with ID '%v'", serviceGuid, enclaveId)
	}
	container := serviceDockerResources.container
	if container == nil {
		return nil, stacktrace.NewError("Cannot copy files from user service '%v' in enclave '%v' because it doesn't have a container", serviceGuid, enclaveId)
	}

	tarStreamReadCloser, err := backend.dockerManager.CopyFromContainer(ctx, container.GetId(), srcPath)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred copying content from sourcepath '%v' in container '%v' for user service '%v' in enclave '%v'",
			srcPath,
			container.GetName(),
			serviceGuid,
			enclaveId,
		)
	}

	return tarStreamReadCloser, nil
}

func (backend *DockerKurtosisBackend) StopUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	resultSuccessfulServiceGUIDs map[service.ServiceGUID]bool,
	resultErroredServiceGUIDs map[service.ServiceGUID]error,
	resultErr error,
) {
	matchingUserServicesByContainerId, err := backend.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedObjectsByContainerId := map[string]interface{}{}
	for containerId, object := range matchingUserServicesByContainerId {
		matchingUncastedObjectsByContainerId[containerId] = interface{}(object)
	}

	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.KillContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred killing user service container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulServiceGuidStrs, erroredServiceGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedObjectsByContainerId,
		backend.dockerManager,
		extractServiceGUIDFromServiceObj,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred killing user service containers matching filters '%+v'", filters)
	}

	successfulServiceGuids := map[service.ServiceGUID]bool{}
	for serviceGuidStr := range successfulServiceGuidStrs {
		successfulServiceGuids[service.ServiceGUID(serviceGuidStr)] = true
	}
	erroredGuids := map[service.ServiceGUID]error{}
	for serviceGuidStr, removalErr := range erroredServiceGuidStrs {
		erroredGuids[service.ServiceGUID(serviceGuidStr)] = removalErr
	}

	return successfulServiceGuids, erroredGuids, nil
}

func (backend *DockerKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	userServices, err := backend.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedObjectsByContainerId := map[string]interface{}{}
	for containerId, object := range userServices {
		matchingUncastedObjectsByContainerId[containerId] = interface{}(object)
	}

	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing user service container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulServiceGuidStrs, erroredServiceGuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedObjectsByContainerId,
		backend.dockerManager,
		extractServiceGUIDFromServiceObj,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing user service containers matching filters '%+v'", filters)
	}

	successfulServiceGuids := map[service.ServiceGUID]bool{}
	for serviceGuidStr := range successfulServiceGuidStrs {
		successfulServiceGuids[service.ServiceGUID(serviceGuidStr)] = true
	}
	erroredGuids := map[service.ServiceGUID]error{}
	for serviceGuidStr, removalErr := range erroredServiceGuidStrs {
		erroredGuids[service.ServiceGUID(serviceGuidStr)] = removalErr
	}

	return successfulServiceGuids, erroredGuids, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets the service objects & Docker resources for services matching the given filters
// !!!!!!!!!! It's VERY important that the service registration mutex is locked before this is called !!!!!!!!
func (backend *DockerKurtosisBackend) getMatchingUserServiceObjsAndDockerResourcesWithoutMutex(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.Service,
	map[service.ServiceGUID]*userServiceDockerResources,
	error,
) {
	allEnclaveServices, found := backend.serviceRegistrations[enclaveId]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to find services in enclave '%v', but this enclave isn't listed as being tracked " +
				"by the backend; this likely means that the request is originating from somewhere it shouldn't (i.e. " +
				"outside the API container)",
			enclaveId,
		)
	}

	// Filter on GUID & ID first, so that we don't pull back unnecessary containers from Docker
	matchingServiceRegistrations := map[service.ServiceGUID]*registeredServiceInfo{}
	serviceGuidsToGetContainersFor := map[service.ServiceGUID]bool{}
	for serviceGuid, registrationInfo := range allEnclaveServices {
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[serviceGuid]; !found {
				continue
			}
			matchingServiceRegistrations[serviceGuid] = registrationInfo
		}

		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[registrationInfo.id]; !found {
				continue
			}
		}

		matchingServiceRegistrations[serviceGuid] = registrationInfo
		serviceGuidsToGetContainersFor[serviceGuid] = true
	}

	matchingDockerResources, err := backend.getMatchingUserServiceDockerResources(ctx, enclaveId, serviceGuidsToGetContainersFor)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting matching user service resources")
	}

	matchingServiceObjs, err := getServiceObjectsFromRegistrationsAndDockerResources(matchingServiceRegistrations, matchingDockerResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis service objects from user service registrations & Docker resources")
	}

	// We've already filtered by service ID & GUID, so all that remains is by status
	resultServiceObjs := map[service.ServiceGUID]*service.Service{}
	resultDockerResources := map[service.ServiceGUID]*userServiceDockerResources{}
	for guid, serviceObj := range matchingServiceObjs {
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[serviceObj.GetStatus()]; !found {
				continue
			}
		}

		dockerResources, found := matchingDockerResources[guid]
		if !found {
			// This should never happen; the Services map and the Docker resources maps should have the same GUIDs
			return nil, nil, stacktrace.Propagate(err, "Needed to return Docker resources for service with GUID '%v', but none was found; this is a bug in Kurtosis")
		}

		resultServiceObjs[guid] = serviceObj
		resultDockerResources[guid] = dockerResources
	}
	return resultServiceObjs, resultDockerResources, nil
}

func (backend *DockerKurtosisBackend) getMatchingUserServiceDockerResources(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuidsToFind map[service.ServiceGUID]bool,
) (map[service.ServiceGUID]*userServiceDockerResources, error) {
	// For the matching values, get the containers to check the status
	userServiceContainerSearchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.EnclaveIDLabelKey.GetString(): string(enclaveId),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.UserServiceContainerTypeLabelValue.GetString(),
	}
	userServiceContainers, err := backend.dockerManager.GetContainersByLabels(ctx, userServiceContainerSearchLabels, shouldGetStoppedContainersWhenGettingServiceInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service containers in enclave '%v' by labels: %+v", enclaveId, userServiceContainerSearchLabels)
	}
	containersByServiceGuid := map[service.ServiceGUID]*types.Container{}
	for _, container := range userServiceContainers {
		containerGuidStr, found := container.GetLabels()[label_key_consts.GUIDLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Found user service container '%v' that didn't have expected GUID label '%v'", container.GetId(), label_key_consts.GUIDLabelKey.GetString())
		}
		containersByServiceGuid[service.ServiceGUID(containerGuidStr)] = container
	}

	result := map[service.ServiceGUID]*userServiceDockerResources{}
	for guid := range serviceGuidsToFind {
		var serviceContainer *types.Container = nil
		if container, found := containersByServiceGuid[guid]; found {
			serviceContainer = container
		}

		result[guid] = &userServiceDockerResources{container: serviceContainer}
	}

	return result, nil
}

func getServiceObjectsFromRegistrationsAndDockerResources(
	registrationInfo map[service.ServiceGUID]*registeredServiceInfo,
	allDockerResources map[service.ServiceGUID]*userServiceDockerResources,
) (map[service.ServiceGUID]*service.Service, error) {
	result := map[service.ServiceGUID]*service.Service{}
	for guid, registration := range registrationInfo {
		enclaveId := registration.enclaveId
		serviceId := registration.id
		serviceGuid := registration.guid
		status := service.UserServiceStatus_Registered
		privateIpAddr := registration.ip
		var maybePrivatePortsSpecs map[string]*port_spec.PortSpec
		var maybePublicIpAddr net.IP
		var maybePublicPortsSpecs map[string]*port_spec.PortSpec

		dockerResourcesForService, found := allDockerResources[guid]
		if found && dockerResourcesForService.container != nil {
			container := dockerResourcesForService.container
			containerStatus := container.GetStatus()

			serviceStatus, found := userServiceStatusDeterminer[containerStatus]
			if !found {
				// This should never happen because we enforce the completeness in a unit test
				return nil, stacktrace.NewError("Expected to find a user service status for container status '%v'; this is a bug in Kurtosis", containerStatus.String())
			}
			status = serviceStatus

			parsedPrivatePorts, maybeParsedPublicIp, maybeParsedPublicPorts, err := getIpAndPortInfoFromContainer(
				container.GetName(),
				container.GetLabels(),
				container.GetHostPortBindings(),
			)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred getting IP & port info from container '%v'", container.GetName())
			}

			maybePrivatePortsSpecs = parsedPrivatePorts
			maybePublicIpAddr = maybeParsedPublicIp
			maybePublicPortsSpecs = maybeParsedPublicPorts
		}

		result[guid] = service.NewService(
			serviceId,
			serviceGuid,
			status,
			enclaveId,
			privateIpAddr,
			maybePrivatePortsSpecs,
			maybePublicIpAddr,
			maybePublicPortsSpecs,
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
	resultPrivatePortSpecs map[string]*port_spec.PortSpec,
	resultPublicIp net.IP,
	resultPublicPortSpecs map[string]*port_spec.PortSpec,
	resultErr error,
){
	serializedPortSpecs, found := labels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return nil, nil, nil, stacktrace.NewError(
			"Expected to find port specs label '%v' on container '%v' but none was found",
			containerName,
			label_key_consts.PortSpecsLabelKey.GetString(),
		)
	}

	privatePortSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
		}
	}

	var containerPublicIp net.IP
	var publicPortSpecs map[string]*port_spec.PortSpec
	for portId, privatePortSpec := range privatePortSpecs {
		portPublicIp, publicPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privatePortSpec, hostMachinePortBindings)
		if err != nil {
			return nil, nil, nil, stacktrace.Propagate(
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
				return nil, nil, nil, stacktrace.NewError(
					"Private port '%v' on container '%v' yielded a public IP '%v', which doesn't agree with " +
						"previously-seen public IPs",
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

	return privatePortSpecs, containerPublicIp, publicPortSpecs, nil
}

func makeHttpRequest(httpMethod string, url string, body string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	if body != "" && httpMethod != http.MethodPost {
		return nil, stacktrace.NewError("Is not possible to execute the http request with body '%v' using the http '%v' method", body, httpMethod)
	}

	if httpMethod == http.MethodPost {
		var bodyByte = []byte(body)
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(bodyByte))
	} else if httpMethod == http.MethodGet {
		resp, err = http.Get(url)
	} else if httpMethod == http.MethodHead {
		resp, err = http.Head(url)
	} else {
		return nil, stacktrace.NewError("HTTP method '%v' not allowed", httpMethod)
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "An HTTP error occurred sending a request to endpoint '%v' using http method '%v'", url, httpMethod)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v' when calling http url '%v'", resp.StatusCode, url)
	}
	return resp, nil
}

func (backend *DockerKurtosisBackend) getSingleServiceObjWithResourcesWithoutMutex(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceGuid service.ServiceGUID,
) (
	*service.Service,
	*userServiceDockerResources,
	error,
) {
	filters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			userServiceGuid: true,
		},
	}
	userServices, dockerResources, err := backend.getMatchingUserServiceObjsAndDockerResourcesWithoutMutex(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services using filters '%v'", filters)
	}
	numOfUserServices := len(userServices)
	if numOfUserServices == 0 {
		return nil, nil, stacktrace.NewError("No user service with GUID '%v' in enclave with ID '%v' was found", userServiceGuid, enclaveId)
	}
	if numOfUserServices > 1 {
		return nil, nil, stacktrace.NewError("Expected to find only one user service with GUID '%v' in enclave with ID '%v', but '%v' was found", userServiceGuid, enclaveId, numOfUserServices)
	}

	var resultService *service.Service
	for _, resultService = range userServices {}

	var resultDockerResources *userServiceDockerResources
	for _, resultDockerResources = range dockerResources {}

	return resultService, resultDockerResources, nil
}

/*
func (backend *DockerKurtosisBackend) getMatchingUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (map[string]*service.Service, error) {

	searchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():         label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.UserServiceContainerTypeLabelValue.GetString(),
	}
	matchingContainers, err := backend.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching containers using labels: %+v", searchLabels)
	}

	matchingObjects := map[string]*service.Service{}
	for _, container := range matchingContainers {
		containerId := container.GetId()
		object, err := getUserServiceObjectFromContainerInfo(
			containerId,
			container.GetLabels(),
			container.GetStatus(),
			container.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with ID '%v' into a user service object", container.GetId())
		}

		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[object.GetEnclaveID()]; !found {
				continue
			}
		}

		if filters.RegistrationGUIDs != nil && len(filters.RegistrationGUIDs) > 0 {
			if _, found := filters.RegistrationGUIDs[object.GetRegistrationGUID()]; !found {
				continue
			}
		}

		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[object.GetGUID()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[object.GetStatus()]; !found {
				continue
			}
		}

		matchingObjects[containerId] = object
	}

	return matchingObjects, nil
}

 */

/*
func getUserServiceObjectFromContainerInfo(
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (*service.Service, error) {

	enclaveId, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the user service's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDLabelKey.GetString())
	}

	registrationGuid, found := labels[label_key_consts.UserServiceRegistrationGUIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the user service's registration GUID to be found under label '%v' but the label wasn't present", label_key_consts.UserServiceRegistrationGUIDLabelKey.GetString())
	}

	guid, found := labels[label_key_consts.GUIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find user service GUID label key '%v' but none was found", label_key_consts.GUIDLabelKey.GetString())
	}

	privatePorts, err := getUserServicePrivatePortsFromContainerLabels(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting port specs from container '%v' with labels '%+v'", containerId, labels)
	}

	isContainerRunning, found := isContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for user service container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var status container_status.ContainerStatus
	if isContainerRunning {
		status = container_status.ContainerStatus_Running
	} else {
		status = container_status.ContainerStatus_Stopped
	}

	// TODO Replace with the (simpler) way that's currently done when creating API container/engine container
	_, portIdsForDockerPortObjs, err := getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting used ports from private ports spec '%+v'", privatePorts)
	}

	var maybePublicIpAddr net.IP = nil
	var maybePublicPorts map[string]*port_spec.PortSpec
	if status == container_status.ContainerStatus_Running && len(privatePorts) > 0 {
		maybePublicIpAddr, maybePublicPorts, err = condensePublicNetworkInfoFromHostMachineBindings(
			allHostMachinePortBindings,
			privatePorts,
			portIdsForDockerPortObjs,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred extracting public IP addr & ports from the host machine ports returned by the container engine")
		}
	}

	var privateIpAddr net.IP
	privateIpAddrStr, found := labels[label_key_consts.PrivateIPLabelKey.GetString()]
	if !found {
		 return nil, stacktrace.NewError("Expected to find user service private IP label key '%v' but none was found", label_key_consts.PrivateIPLabelKey.GetString())
	}

	newObject := service.NewService(
		user_service_registration.UserServiceRegistrationGUID(registrationGuid),
		service.ServiceGUID(guid),
		status,
		enclave.EnclaveID(enclaveId),
		privateIpAddr,
		privatePorts,
		maybePublicIpAddr,
		maybePublicPorts,
	)

	return newObject, nil
}
*/

/*
func getUserServicePrivatePortsFromContainerLabels(containerLabels map[string]string) (map[string]*port_spec.PortSpec, error) {
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		if err != nil {
			return nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
		}
	}

	return portSpecs, nil
}

 */

/*
// TODO Replace with the method that the API containers use for getting & retrieving port specs
func getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts map[string]*port_spec.PortSpec) (map[nat.Port]docker_manager.PortPublishSpec, map[nat.Port]string, error) {
	publishSpecs := map[nat.Port]docker_manager.PortPublishSpec{}
	portIdsForDockerPortObjs := map[nat.Port]string{}
	for portId, portSpec := range privatePorts {
		dockerPort, err := transformPortSpecToDockerPort(portSpec)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred transforming the '%+v' port spec to a Docker port", portSpec)
		}
		publishSpecs[dockerPort] = docker_manager.NewAutomaticPublishingSpec()

		if preexistingPortId, found := portIdsForDockerPortObjs[dockerPort]; found {
			return nil, nil, stacktrace.NewError(
				"Port '%v' declares Docker port spec '%v', but this port spec is already in use by port '%v'",
				portId,
				dockerPort,
				preexistingPortId,
			)
		}
		portIdsForDockerPortObjs[dockerPort] = portId

	}
	return publishSpecs, portIdsForDockerPortObjs, nil
}

 */

/*
// TODO Replace with the simpler method that the API container uses for getting public port specs using private port specs
// condensePublicNetworkInfoFromHostMachineBindings
// Condenses declared private port bindings and the host machine port bindings returned by the container engine lib into:
//  1) a single host machine IP address
//  2) a map of private port binding IDs -> public ports
// An error is thrown if there are multiple host machine IP addresses
func condensePublicNetworkInfoFromHostMachineBindings(
	hostMachinePortBindings map[nat.Port]*nat.PortBinding,
	privatePorts map[string]*port_spec.PortSpec,
	portIdsForDockerPortObjs map[nat.Port]string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPorts map[string]*port_spec.PortSpec,
	resultErr error,
) {
	if len(hostMachinePortBindings) == 0 {
		return nil, nil, stacktrace.NewError("Cannot condense public network info if no host machine port bindings are provided")
	}

	publicIpAddrStr := uninitializedPublicIpAddrStrValue
	publicPorts := map[string]*port_spec.PortSpec{}
	for dockerPortObj, hostPortBinding := range hostMachinePortBindings {
		portId, found := portIdsForDockerPortObjs[dockerPortObj]
		if !found {
			// If the container engine reports a host port binding that wasn't declared in the input used-ports object, ignore it
			// This could happen if a port is declared in the Dockerfile
			continue
		}

		privatePort, found := privatePorts[portId]
		if !found {
			return nil, nil, stacktrace.NewError(
				"The container engine returned a host machine port binding for Docker port spec '%v', but this port spec didn't correspond to any port ID; this is very likely a bug in Kurtosis",
				dockerPortObj,
			)
		}

		hostIpAddr := hostPortBinding.HostIP
		if publicIpAddrStr == uninitializedPublicIpAddrStrValue {
			publicIpAddrStr = hostIpAddr
		} else if publicIpAddrStr != hostIpAddr {
			return nil, nil, stacktrace.NewError(
				"A public IP address '%v' was already declared for the service, but Docker port object '%v' declares a different public IP address '%v'",
				publicIpAddrStr,
				dockerPortObj,
				hostIpAddr,
			)
		}

		hostPortStr := hostPortBinding.HostPort
		hostPortUint64, err := strconv.ParseUint(hostPortStr, dockerContainerPortNumUintBase, dockerContainerPortNumUintBits)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred parsing host machine port string '%v' into a uint with %v bits and base %v",
				hostPortStr,
				dockerContainerPortNumUintBits,
				dockerContainerPortNumUintBase,
			)
		}
		hostPortUint16 := uint16(hostPortUint64) // Safe to do because our ParseUint declares the expected number of bits
		portProtocol := privatePort.GetProtocol()

		portSpec, err := port_spec.NewPortSpec(hostPortUint16, portProtocol)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred creating a new public port spec object using number '%v' and protocol '%v'",
				hostPortUint16,
				portProtocol,
			)
		}

		publicPorts[portId] = portSpec
	}
	if publicIpAddrStr == uninitializedPublicIpAddrStrValue {
		return nil, nil, stacktrace.NewError("No public IP address string was retrieved from host port bindings: %+v", hostMachinePortBindings)
	}
	publicIpAddr := net.ParseIP(publicIpAddrStr)
	if publicIpAddr == nil {
		return nil, nil, stacktrace.NewError("Couldn't parse service's public IP address string '%v' to an IP object", publicIpAddrStr)
	}
	return publicIpAddr, publicPorts, nil
}
 */

/*
func doesHttpResponseMatchExpected(
	expectedResponseBody string,
	responseBody io.ReadCloser,
	url string) bool {

	defer responseBody.Close()
	if expectedResponseBody != "" {

		bodyBytes, err := ioutil.ReadAll(responseBody)
		if err != nil {
			logrus.Errorf("An error occurred reading the response body from endpoint '%v':\n%v", url, err)
			return false
		}

		bodyStr := string(bodyBytes)

		if bodyStr != expectedResponseBody {
			logrus.Errorf("Expected response body text '%v' from endpoint '%v' but got '%v' instead", expectedResponseBody, url, bodyStr)
			return false
		}
	}
	return true
}
 */

func extractServiceGUIDFromServiceObj(uncastedObj interface{}) (string, error) {
	castedObj, ok := uncastedObj.(*service.Service)
	if !ok {
		return "", stacktrace.NewError("An error occurred downcasting the user service object")
	}
	return string(castedObj.GetGUID()), nil
}
