package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/wait_for_availability_http_methods"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const(
	// The path on the user service container where the enclave data dir will be bind-mounted
	serviceEnclaveDataDirMountpoint = "/kurtosis-enclave-data"
)

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
)(
	newUserService *service.Service,
	resultErr error,
){

	enclaveObjAttrsProvider, err := backendCore.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForUserServiceContainer(id, guid, privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the user service container attributes for user service with GUID '%v'", guid)
	}
	containerName := containerAttrs.GetName()

	labelStrs := map[string]string{}
	for labelKey, labelValue := range containerAttrs.GetLabels(){
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	enclaveNetwork, err := backendCore.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	usedPorts, portIdsForDockerPortObjs, err := getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts)
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

	_, hostPortBindingsByPortObj, err  := backendCore.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the user service container for user service with GUID '%v'", guid)
	}

	var maybePublicIpAddr net.IP = nil
	publicPorts := map[string]*port_spec.PortSpec{}
	if len(privatePorts) > 0 {
		maybePublicIpAddr, publicPorts, err = condensePublicNetworkInfoFromHostMachineBindings(
			hostPortBindingsByPortObj,
			privatePorts,
			portIdsForDockerPortObjs,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred extracting public IP addr & ports from the host machine ports returned by the container engine")
		}
	}

	userService := service.NewService(id, guid, enclaveId, maybePublicIpAddr, publicPorts)

	return userService, nil
}

func (backendCore *DockerKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
)(
	map[service.ServiceGUID]*service.Service,
	map[service.ServiceGUID]error,
	error,
){

	userServiceContainers, err := backendCore.getUserServiceContainersByEnclaveIDAndUserServiceGUIDs(ctx, enclaveId, filters.GUIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user-service-containers by enclave ID '%v' and user service GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	successfulUserServices := map[service.ServiceGUID]*service.Service{}
	erroredUserServices := map[service.ServiceGUID]error{}
	for guid, container := range userServiceContainers {
		id, err := getServiceIdFromContainer(container)
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting service ID from container with ID '%v'", container.GetId())
			erroredUserServices[guid] = serviceError
			continue
		}

		privatePorts, err := getPrivatePortsFromContainerLabels(container.GetLabels())
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting port specs from container labels '%+v'", container.GetLabels())
			erroredUserServices[guid] = serviceError
			continue
		}

		_, portIdsForDockerPortObjs, err := getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts)
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting used port from private port spec '%+v'", privatePorts)
			erroredUserServices[guid] = serviceError
			continue
		}

		var maybePublicIpAddr net.IP = nil
		publicPorts := map[string]*port_spec.PortSpec{}
		if len(privatePorts) > 0 {
			maybePublicIpAddr, publicPorts, err = condensePublicNetworkInfoFromHostMachineBindings(
				container.GetHostPortBindings(),
				privatePorts,
				portIdsForDockerPortObjs,
			)
			if err != nil {
				serviceError := stacktrace.Propagate(err, "An error occurred extracting public IP addr & ports from the host machine ports returned by the container engine")
				erroredUserServices[guid] = serviceError
				continue
			}
		}

		service := service.NewService(id, guid, enclaveId, maybePublicIpAddr, publicPorts)
		successfulUserServices[guid] = service
	}
	return successfulUserServices, erroredUserServices, nil
}

func (backendCore *DockerKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
)(
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
){
	userServiceContainers, err := backendCore.getUserServiceContainersByEnclaveIDAndUserServiceGUIDs(ctx, enclaveId, filters.GUIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user-service-containers by enclave ID '%v' and user service GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	successfulUserServicesLogs := map[service.ServiceGUID]io.ReadCloser{}
	erroredUserServices := map[service.ServiceGUID]error{}

	//TODO use concurrency to improve perf
	for userServiceGuid, container := range userServiceContainers {
		readCloserLogs, err := backendCore.dockerManager.GetContainerLogs(ctx, container.GetId(), shouldFollowLogs)
		if err != nil {
			serviceError := stacktrace.Propagate(err, "An error occurred getting logs for user service with GUID '%v' and container ID '%v'", userServiceGuid, container.GetId())
			erroredUserServices[userServiceGuid] = serviceError
			continue
		}
		successfulUserServicesLogs[userServiceGuid] = readCloserLogs
	}

	return successfulUserServicesLogs, erroredUserServices, nil
}

func (backendCore *DockerKurtosisBackend) RunUserServiceExecCommand (
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGUID service.ServiceGUID,
	command []string,
)(
	resultExitCode int32,
	resultOutput string,
	resultErr error,
){

	userServiceContainer, err := backendCore.getUserServiceContainerByEnclaveIDAndUserServiceGUID(ctx, enclaveId, serviceGUID)
	if err != nil {
		return 0, "", stacktrace.Propagate(err, "An error occurred getting user service container by enclave id '%v' and user service GUID '%v'", enclaveId, serviceGUID)
	}

	execOutputBuf := &bytes.Buffer{}
	exitCode, err := backendCore.dockerManager.RunExecCommand(ctx, userServiceContainer.GetId(), command, execOutputBuf)
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
)(
	resultErr error,
) {

	if requestBody != "" && httpMethod != wait_for_availability_http_methods.WaitForAvailabilityHttpMethod_POST {
		return stacktrace.NewError("Is not possible to execute the http request with body '%v' using the http '%v' method, it is only possible to use http POST method with request body", requestBody, httpMethod)
	}

	enclaveNetwork, err := backendCore.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	userServiceContainer, err := backendCore.getUserServiceContainerByEnclaveIDAndUserServiceGUID(ctx, enclaveId, serviceGUID)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service container by enclave id '%v' and user service GUID '%v'", enclaveId, serviceGUID)
	}

	privateIpAddr, found := userServiceContainer.GetNetworksIPAddresses()[enclaveNetwork.GetId()]
	if !found {
		return stacktrace.Propagate(err, "User service container with container ID '%v' does not have and IP address defined in Docker Network with ID '%v'; it should never happen it's a bug in Kurtosis", userServiceContainer.GetId(), enclaveNetwork.GetId())
	}

	url := fmt.Sprintf("http://%v:%v/%v", privateIpAddr, port, path)

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
	serviceGUID service.ServiceGUID,
)(
	resultErr error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) StopUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
)(
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	userServiceContainers, err := backendCore.getUserServiceContainersByEnclaveIDAndUserServiceGUIDs(ctx, enclaveId, filters.GUIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user-service-containers by enclave ID '%v' and user service GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	for userServiceGuid, userServiceContainer := range userServiceContainers {
		if err := backendCore.killContainerAndWaitForExit(ctx, userServiceContainer); err != nil {
			erroredUserServiceGuids[userServiceGuid] = err
			continue
		}
		successfulUserServiceGuids[userServiceGuid] = true
	}

	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

func (backendCore *DockerKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *service.ServiceFilters,
)(
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	userServiceContainers, err := backendCore.getUserServiceContainersByEnclaveIDAndUserServiceGUIDs(ctx, enclaveId, filters.GUIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user-service-containers by enclave ID '%v' and user service GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting networking-sidecar-containers by enclave ID '%v' and networking sidecar GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	for userServiceGuid, userServiceContainer := range userServiceContainers {
		if err := backendCore.removeContainer(ctx, userServiceContainer); err != nil {
			erroredUserServiceGuids[userServiceGuid] = err
			continue
		}
		successfulUserServiceGuids[userServiceGuid] = true
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