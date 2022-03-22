package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
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

	enclaveNetwork, _, err := backendCore.getEnclaveNetwork(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Docker network by enclave ID '%v'", enclaveId)
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
	filters *service.ServiceFilters,
)(
	map[service.ServiceGUID]*service.Service,
	error,
){
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	map[service.ServiceGUID]io.ReadCloser,
	error,
){
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) RunUserServiceExecCommand (
	ctx context.Context,
	serviceGUID service.ServiceGUID,
	commandArgs []string,
)(
	exitCode int32,
	output string,
	resultErr error,
){
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) WaitForUserServiceHttpEndpointAvailability(
	ctx context.Context,
	serviceGUID service.ServiceGUID,
	httpMethod string,
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
	panic("Implement me")
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
	filters *service.ServiceFilters,
)(
	successfulUserServiceIds map[service.ServiceGUID]bool,
	erroredUserServiceIds map[service.ServiceGUID]error,
	resultErr error,
) {
	panic("Implement me")
}

func (backendCore *DockerKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	filters *service.ServiceFilters,
)(
	successfulUserServiceIds map[service.ServiceGUID]bool,
	erroredUserServiceIds map[service.ServiceGUID]error,
	resultErr error,
) {
	panic("Implement me")
}
