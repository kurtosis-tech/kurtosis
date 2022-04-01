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
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/shell"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/wait_for_availability_http_methods"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
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

func (backendCore *DockerKurtosisBackend) CreateUserService(
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

	enclaveObjAttrsProvider, err := backendCore.objAttrsProvider.ForEnclave(enclaveId)
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

	enclaveNetwork, err := backendCore.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

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
	if err = backendCore.dockerManager.PullImage(ctx, containerImageName); err != nil {
		logrus.Warnf("Failed to pull the latest version of user service container image '%v'; you may be running an out-of-date version", containerImageName)
	}

	containerId, hostMachinePortBindings, err := backendCore.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the user service container for user service with GUID '%v'", guid)
	}

	userService, err := getUserServiceObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service object from container info")
	}

	return userService, nil
}

func (backendCore *DockerKurtosisBackend) GetUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.Service,
	error,
) {

	userServices, err := backendCore.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	successfulUserServices := map[service.ServiceGUID]*service.Service{}
	for _, userService := range userServices {
		successfulUserServices[userService.GetGUID()] = userService
	}
	return successfulUserServices,  nil
}

func (backendCore *DockerKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
) {
	userServices, err := backendCore.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	successfulUserServicesLogs := map[service.ServiceGUID]io.ReadCloser{}
	erroredUserServices := map[service.ServiceGUID]error{}

	//TODO use concurrency to improve perf
	for containerId, userService := range userServices {
		readCloserLogs, err := backendCore.dockerManager.GetContainerLogs(ctx, containerId, shouldFollowLogs)
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting logs for user service with GUID '%v' and container ID '%v'", userService.GetGUID(), containerId)
			erroredUserServices[userService.GetGUID()] = serviceError
			continue
		}
		successfulUserServicesLogs[userService.GetGUID()] = readCloserLogs
	}

	return successfulUserServicesLogs, erroredUserServices, nil
}

func (backendCore *DockerKurtosisBackend) RunUserServiceExecCommand(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
	command []string,
) (
	resultExitCode int32,
	resultOutput string,
	resultErr error,
) {

	userServiceContainerId, _, err := backendCore.getContainerIDAndUserServiceObjectByEnclaveIDAndUserServiceGUID(ctx, enclaveId, serviceGUID)
	if err != nil {
		return 0, "", stacktrace.Propagate(err, "An error occurred getting container ID and user service object for enclave ID '%v' and user service GUID '%v'", enclaveId, serviceGUID)
	}

	execOutputBuf := &bytes.Buffer{}
	exitCode, err := backendCore.dockerManager.RunExecCommand(ctx, userServiceContainerId, command, execOutputBuf)
	if err != nil {
		return 0, "", stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service with GUID '%v'",
			command,
			serviceGUID)
	}

	return exitCode, execOutputBuf.String(), nil
}

func (backendCore *DockerKurtosisBackend) WaitForUserServiceHttpEndpointAvailability(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
	httpMethod wait_for_availability_http_methods.WaitForAvailabilityHttpMethod,
	port uint32,
	path string,
	requestBody string,
	bodyText string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
) (
	resultErr error,
) {

	if requestBody != "" && httpMethod != wait_for_availability_http_methods.WaitForAvailabilityHttpMethod_POST {
		return stacktrace.NewError("Is not possible to execute the http request with body '%v' using the http '%v' method, it is only possible to use http POST method with request body", requestBody, httpMethod)
	}

	_, userService, err := backendCore.getContainerIDAndUserServiceObjectByEnclaveIDAndUserServiceGUID(ctx, enclaveId, serviceGUID)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting container ID and user service object for enclave ID '%v' and user service GUID '%v'", enclaveId, serviceGUID)
	}

	url := fmt.Sprintf("http://%v:%v/%v", userService.GetPrivateIp(), port, path)

	time.Sleep(time.Duration(initialDelayMilliseconds) * time.Millisecond)

	httpMethodStr := httpMethod.String()
	var resp *http.Response
	for i := uint32(0); i < retries; i++ {
		resp, err = makeHttpRequest(httpMethodStr, url, requestBody)
		if err == nil {
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

	if bodyText != "" {
		body := resp.Body
		defer body.Close()

		bodyBytes, err := ioutil.ReadAll(body)

		if err != nil {
			return stacktrace.Propagate(err,
				"An error occurred reading the response body from endpoint '%v'", url)
		}

		bodyStr := string(bodyBytes)

		if bodyStr != bodyText {
			return stacktrace.NewError("Expected response body text '%v' from endpoint '%v' but got '%v' instead", bodyText, url, bodyStr)
		}
	}
	return nil
}

func (backendCore *DockerKurtosisBackend) GetShellOnUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
) (
	*shell.Shell,
	error,
) {

	containerId, _, err := backendCore.getContainerIDAndUserServiceObjectByEnclaveIDAndUserServiceGUID(ctx, enclaveId, serviceGUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting container ID and user service object for enclave ID '%v' and user service GUID '%v'", enclaveId, serviceGUID)
	}

	hijackedResponse, err := backendCore.dockerManager.ContainerExecCreate(ctx, containerId, commandToRunWhenCreatingUserServiceShell)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing container exec create on user service with GUID '%v'", serviceGUID)
	}

	newShell := shell.NewShell(hijackedResponse.Conn, hijackedResponse.Reader)

	return newShell, nil
}

func (backendCore *DockerKurtosisBackend) StopUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	userServices, err := backendCore.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	for containerId, userService := range userServices {
		if err := backendCore.killContainerAndWaitForExit(ctx, containerId); err != nil {
			erroredUserServiceGuids[userService.GetGUID()] = err
			continue
		}
		successfulUserServiceGuids[userService.GetGUID()] = true
	}

	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

func (backendCore *DockerKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	userServices, err := backendCore.getMatchingUserServices(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user services matching filters '%+v'", filters)
	}

	for containerId, userService := range userServices {
		if err := backendCore.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred removing container with ID '%v'",
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
		return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v'", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v'", resp.StatusCode)
	}
	return resp, nil
}

func (backendCore *DockerKurtosisBackend) getContainerIDAndUserServiceObjectByEnclaveIDAndUserServiceGUID(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceGuid service.ServiceGUID,
) (
	string,
	*service.Service,
	error,
) {

	filters := &service.ServiceFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
		GUIDs: map[service.ServiceGUID]bool{
			userServiceGuid: true,
		},
	}

	userServices, err := backendCore.getMatchingUserServices(ctx, filters)
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

func (backendCore *DockerKurtosisBackend) getMatchingUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
) (map[string]*service.Service, error) {

	searchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():         label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.UserServiceContainerTypeLabelValue.GetString(),
	}
	matchingContainers, err := backendCore.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchAllContainersWhenRetrievingContainers)
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

	privatePorts, err := getPrivatePortsFromContainerLabels(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting port specs from container '%v' with labels '%+v'", containerId, labels)
	}

	_, portIdsForDockerPortObjs, err := getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting used port from private port spec '%+v'", privatePorts)
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

func getUserServiceContainerFromContainerListByEnclaveIdAndUserServiceGUID(
	containers []*types.Container,
	enclaveId enclave.EnclaveID,
	userServiceGUID service.ServiceGUID) (*types.Container, error) {

	for _, container := range containers {
		if isUserServiceContainer(container) && hasEnclaveIdLabel(container, enclaveId) && hasGuidLabel(container, string(userServiceGUID)) {
			return container, nil
		}
	}
	return nil, stacktrace.NewError("No user service container with user service GUID '%v' was found in container list '%+v'", userServiceGUID, containers)
}

func isUserServiceContainer(container *types.Container) bool {
	labels := container.GetLabels()
	containerTypeValue, found := labels[label_key_consts.ContainerTypeLabelKey.GetString()]
	if !found {
		//TODO Do all containers should have container type label key??? we should return and error here if this answer is yes??
		logrus.Debugf("Container with ID '%v' does not have label '%v'", container.GetId(), label_key_consts.ContainerTypeLabelKey.GetString())
		return false
	}
	if containerTypeValue == label_value_consts.UserServiceContainerTypeLabelValue.GetString() {
		return true
	}
	return false
}
