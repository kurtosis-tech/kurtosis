package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/wait_for_availability_http_methods"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	// The path on the user service container where the enclave data dir will be bind-mounted
	serviceEnclaveDataDirMountpoint = "/kurtosis-enclave-data"
)

// We'll try to use the nicer-to-use shells first before we drop down to the lower shells
var commandToRunWhenCreatingUserServiceShell = []string{
	"sh",
	"-c",
	"if command -v 'bash' > /dev/null; then echo \"Found bash on container; creating bash shell...\"; bash; else echo \"No bash found on container; dropping down to sh shell...\"; sh; fi",
}

func (backend *DockerKurtosisBackend) CreateUserService(
	ctx context.Context,
	id service.ServiceID,
	guid service.ServiceGUID,
	containerImageName string,
	enclaveId enclave.EnclaveID,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	privatePorts map[string]*port_spec.PortSpec,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	enclaveDataDirpathOnHostMachine string,
	filesArtifactMountDirpaths map[string]string,
) (
	newUserService *service.Service,
	resultErr error,
) {

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForUserServiceContainer(id, guid, ipAddr, privatePorts)
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

	// TODO Replace with the (simpler) way that's currently done when creating API container/engine container
	usedPorts, _, err := getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting used port from private port spec '%+v'", privatePorts)
	}

	bindMounts := map[string]string{
		enclaveDataDirpathOnHostMachine: serviceEnclaveDataDirMountpoint,
	}

	createAndStartArgsBuilder := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImageName,
		containerName.GetString(),
		enclaveNetwork.GetId(),
	).WithStaticIP(
		ipAddr,
	).WithUsedPorts(
		usedPorts,
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithLabels(
		labelStrs,
	).WithAlias(
		string(id),
	)
	if entrypointArgs != nil {
		createAndStartArgsBuilder.WithEntrypointArgs(entrypointArgs)
	}
	if cmdArgs != nil {
		createAndStartArgsBuilder.WithCmdArgs(cmdArgs)
	}
	if filesArtifactMountDirpaths != nil {
		createAndStartArgsBuilder.WithVolumeMounts(filesArtifactMountDirpaths)
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
					"Launching user service container '%v' with container ID '%v' didn't complete successfully so we " +
						"tried to kill the container we started, but doing so exited with an error:\n%v",
					containerName.GetString(),
					containerId,
					err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop user service container with ID '%v'!!!!!!", containerId)
			}
		}
	}()

	userService, err := getUserServiceObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service object from container info, using container ID '%v' and labels '%+v'", containerId, labelStrs)
	}

	shouldKillContainer = false
	return userService, nil
}

func (backend *DockerKurtosisBackend) GetUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.Service,
	error,
) {

	userServices, err := backend.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	successfulUserServices := map[service.ServiceGUID]*service.Service{}
	for _, userService := range userServices {
		successfulUserServices[userService.GetGUID()] = userService
	}
	return successfulUserServices,  nil
}

func (backend *DockerKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
) {
	userServices, err := backend.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	successfulUserServicesLogs := map[service.ServiceGUID]io.ReadCloser{}
	erroredUserServices := map[service.ServiceGUID]error{}

	//TODO use concurrency to improve perf
	for containerId, userService := range userServices {
		readCloserLogs, err := backend.dockerManager.GetContainerLogs(ctx, containerId, shouldFollowLogs)
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting logs for user service with GUID '%v' and container ID '%v'", userService.GetGUID(), containerId)
			erroredUserServices[userService.GetGUID()] = serviceError
			continue
		}
		successfulUserServicesLogs[userService.GetGUID()] = readCloserLogs
	}

	return successfulUserServicesLogs, erroredUserServices, nil
}

func (backend *DockerKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceCommands map[service.ServiceGUID][]string,
)(
	map[service.ServiceGUID]*exec_result.ExecResult,
	map[service.ServiceGUID]error,
	error,
){
	succesfulUserServiceExecResults := map[service.ServiceGUID]*exec_result.ExecResult{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	userServiceGuids := map[service.ServiceGUID]bool{}
	for userServiceGuid := range userServiceCommands {
		userServiceGuids[userServiceGuid] = true
	}

	filters := &service.ServiceFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
		GUIDs: userServiceGuids,
	}

	userServices, err := backend.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	if len(userServiceCommands) != len(userServices) {
		return nil, nil, stacktrace.NewError("The amount of user services found '%v' are not equal to the amount of user service to run exec commands '%v'", len(userServices), len(userServiceCommands))
	}
	for _, userService := range userServices{
		if _,found := userServiceCommands[userService.GetGUID()]; !found {
			return nil,
				nil,
				stacktrace.NewError(
					"User service with GUID '%v' was found when getting matching " +
						"user services with filters '%+v' but it was not declared in the user " +
						"service exec commands list '%+v'",
					userService.GetGUID(),
					filters,
					userServiceCommands,
				)
		}
	}

	// TODO Parallelize to increase perf
	for containerId, userService := range userServices {
		userServiceUnwrappedCommand:= userServiceCommands[userService.GetGUID()]

		userServiceShWrappedCmd := wrapShCommand(userServiceUnwrappedCommand)

		execOutputBuf := &bytes.Buffer{}
		exitCode, err := backend.dockerManager.RunExecCommand(
			ctx,
			containerId,
			userServiceShWrappedCmd,
			execOutputBuf)
		if err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred executing command '%+v' on user service with GUID '%v' and container ID '%v'",
				userServiceShWrappedCmd,
				userService.GetGUID(),
				containerId,
			)
			erroredUserServiceGuids[userService.GetGUID()] = wrappedErr
			continue
		}
		newExecResult := exec_result.NewExecResult(exitCode, execOutputBuf.String())
		succesfulUserServiceExecResults[userService.GetGUID()] = newExecResult
	}

	return succesfulUserServiceExecResults, erroredUserServiceGuids, nil
}

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

func (backend *DockerKurtosisBackend) GetConnectionWithUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
) (
	net.Conn,
	error,
) {

	containerId, _, err := backend.getSingleUserService(ctx, enclaveId, serviceGUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting container ID and user service object for enclave ID '%v' and user service GUID '%v'", enclaveId, serviceGUID)
	}

	hijackedResponse, err := backend.dockerManager.CreateContainerExec(ctx, containerId, commandToRunWhenCreatingUserServiceShell)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing container exec create on user service with GUID '%v'", serviceGUID)
	}

	newConnection := hijackedResponse.Conn

	return newConnection, nil
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

	containerIdsToKill := map[string]bool{}
	for containerId := range matchingUserServicesByContainerId {
		containerIdsToKill[containerId] = true
	}

	successfulContainerIds, erroredContainerIds := backend.killContainers(ctx, containerIdsToKill)

	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	for containerId := range successfulContainerIds {
		serviceObj, found := matchingUserServicesByContainerId[containerId]
		if !found {
			return nil, nil, stacktrace.NewError("Successfully killed container with ID '%v' that wasn't requested; this is a bug in Kurtosis!", containerId)
		}
		successfulUserServiceGuids[serviceObj.GetGUID()] = true
	}

	erroredUserServiceGuids := map[service.ServiceGUID]error{}
	for containerId := range erroredContainerIds {
		serviceObj, found := matchingUserServicesByContainerId[containerId]
		if !found {
			return nil, nil, stacktrace.NewError("An error occurred killing container with ID '%v' that wasn't requested; this is a bug in Kurtosis!", containerId)
		}
		wrappedErr := stacktrace.Propagate(err, "An error occurred killing service with GUID '%v' and container ID '%v'", serviceObj.GetGUID(), containerId)
		erroredUserServiceGuids[serviceObj.GetGUID()] = wrappedErr
	}

	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

func (backend *DockerKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	userServices, err := backend.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	for containerId, userService := range userServices {
		if err := backend.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred removing user service container with ID '%v'",
				containerId,
			)
			erroredUserServiceGuids[userService.GetGUID()] = wrappedErr
			continue
		}
		successfulUserServiceGuids[userService.GetGUID()] = true
	}
	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
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

func (backend *DockerKurtosisBackend) getSingleUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceGuid service.ServiceGUID,
) (
	containerId string,
	userService *service.Service,
	error error,
) {

	filters := &service.ServiceFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
		GUIDs: map[service.ServiceGUID]bool{
			userServiceGuid: true,
		},
	}

	userServices, err := backend.getMatchingUserServices(ctx, filters)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred getting user services using filters '%v'", filters)
	}
	numOfUserServices := len(userServices)
	if numOfUserServices == 0 {
		return "", nil, stacktrace.NewError("No user service with GUID '%v' in enclave with ID '%v' was found", userServiceGuid, enclaveId)
	}
	if numOfUserServices > 1 {
		return "", nil, stacktrace.NewError("Expected to find only one user service with GUID '%v' in enclave with ID '%v', but '%v' was found", userServiceGuid, enclaveId, numOfUserServices)
	}

	var resultUserServiceContainerId string
	var resultUserService *service.Service

	for containerId, userService := range userServices {
		resultUserServiceContainerId = containerId
		resultUserService = userService
	}

	return resultUserServiceContainerId, resultUserService, nil
}

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

	id, found := labels[label_key_consts.IDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find user service ID label key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
	}

	guid, found := labels[label_key_consts.GUIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find user service GUID label key '%v' but none was found", label_key_consts.GUIDLabelKey.GetString())
	}

	privatePorts, err := getUserServicePrivatePortsFromContainerLabels(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting port specs from container '%v' with labels '%+v'", containerId, labels)
	}

	// TODO Replace with the (simpler) way that's currently done when creating API container/engine container
	_, portIdsForDockerPortObjs, err := getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting used ports from private ports spec '%+v'", privatePorts)
	}

	var maybePublicIpAddr net.IP = nil
	publicPorts := map[string]*port_spec.PortSpec{}
	if len(privatePorts) > 0 {
		maybePublicIpAddr, publicPorts, err = condensePublicNetworkInfoFromHostMachineBindings(
			allHostMachinePortBindings,
			privatePorts,
			portIdsForDockerPortObjs,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred extracting public IP addr & ports from the host machine ports returned by the container engine")
		}
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

	var privateIpAddr net.IP
	privateIpAddrStr, found := labels[label_key_consts.PrivateIPLabelKey.GetString()]
	// UNCOMMENT THIS AFTER 2022-06-30 WHEN NOBODY HAS USER SERVICES WITHOUT THE PRIVATE IP ADDRESS LABEL
	/*
		if !found {
			return nil, stacktrace.NewError("Expected to find user service private IP label key '%v' but none was found", label_key_consts.PrivateIPLabelKey.GetString())
		}
	*/
	if found {
		candidatePrivateIpAddr := net.ParseIP(privateIpAddrStr)
		if candidatePrivateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
		}
		privateIpAddr = candidatePrivateIpAddr
	}

	newObject := service.NewService(
		service.ServiceID(id),
		service.ServiceGUID(guid),
		status,
		enclave.EnclaveID(enclaveId),
		maybePublicIpAddr,
		publicPorts,
		privateIpAddr,
	)

	return newObject, nil
}

func getUserServicePrivatePortsFromContainerLabels(containerLabels map[string]string) (map[string]*port_spec.PortSpec, error) {
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return  nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		if err != nil {
			return nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
		}
	}

	return portSpecs, nil
}

// TODO Replace with the method that the API containers use for getting & retrieving port specs
func getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts map[string]*port_spec.PortSpec) (map[nat.Port]docker_manager.PortPublishSpec, map[nat.Port]string, error) {
	publishSpecs := map[nat.Port]docker_manager.PortPublishSpec{}
	portIdsForDockerPortObjs := map[nat.Port]string{}
	for portId, portSpec := range privatePorts {
		dockerPort, err := transformPortSpecToDockerPort(portSpec)
		if err != nil {
			return nil, nil,  stacktrace.Propagate(err, "An error occurred transforming the '%+v' port spec to a Docker port", portSpec)
		}
		publishSpecs[dockerPort] =  docker_manager.NewAutomaticPublishingSpec()

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
			return nil,  nil, stacktrace.NewError(
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