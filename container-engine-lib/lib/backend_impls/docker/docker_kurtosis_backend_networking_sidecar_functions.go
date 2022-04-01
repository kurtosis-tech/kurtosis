package docker

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
)

const (
	networkingSidecarImageName = "kurtosistech/iproute2"
	succesfulExecCmdExitCode = 0

	shouldFetchStoppedContainersWhenGettingNetworkingSidecarContainers = true
)

// We sleep forever because all the commands this container will run will be executed
//  via Docker exec
var sidecarContainerCommand = []string{
	"sleep", "infinity",
}

func (backendCore *DockerKurtosisBackend) CreateNetworkingSidecar(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
)(
	*networking_sidecar.NetworkingSidecar,
	error,
){
	// Get the Docker network ID where we'll start the new sidecar container
	enclaveNetwork, err := backendCore.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	enclaveStatus, enclaveContainers, err := backendCore.getEnclaveStatusAndContainers(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
	}

	if enclaveStatus != enclave.EnclaveStatus_Running {
		return nil, stacktrace.NewError("Networking sidecar for user service with GUID '%v' can not be created inside enclave with ID '%v' because its current status is '%v' and it must be '%v' to accept new nodes", serviceGuid, enclaveId, enclaveStatus, enclave.EnclaveStatus_Running.String())
	}

	userServiceContainer, err := getUserServiceContainerFromContainerListByEnclaveIdAndUserServiceGUID(enclaveContainers, enclaveId, serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service container with GUID '%v' from container list '%+v' which are part of enclave with ID '%v'", serviceGuid, enclaveContainers, enclaveId)
	}

	enclaveObjAttrsProvider, err := backendCore.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForNetworkingSidecarContainer(serviceGuid, ipAddr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the networking sidecar container attributes for user service with GUID '%v'", serviceGuid)
	}
	containerName := containerAttrs.GetName()
	containerDockerLabels := containerAttrs.GetLabels()

	containerLabels := map[string]string{}
	for dockerLabelKey, dockerLabelValue := range containerDockerLabels {
		containerLabels[dockerLabelKey.GetString()] = dockerLabelValue.GetString()
	}
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		networkingSidecarImageName,
		containerName.GetString(),
		enclaveNetwork.GetId(),
	).WithAlias(
		containerName.GetString(),
	).WithStaticIP(
		ipAddr,
	).WithAddedCapabilities(map[docker_manager.ContainerCapability]bool{
		docker_manager.NetAdmin: true,
	}).WithNetworkMode(
		docker_manager.NewContainerNetworkMode(userServiceContainer.GetId()),
	).WithCmdArgs(
		sidecarContainerCommand,
	).WithLabels(
		containerLabels,
	).Build()

	// Best-effort pull attempt
	if err = backendCore.dockerManager.PullImage(ctx, networkingSidecarImageName); err != nil {
		logrus.Warnf("Failed to pull the latest version of networking sidecar container image '%v'; you may be running an out-of-date version", networkingSidecarImageName)
	}

	_, _, err = backendCore.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the networking sidecar container")
	}

	networkingSidecar := networking_sidecar.NewNetworkingSidecar(serviceGuid, ipAddr, enclaveId)

	return networkingSidecar, nil

}

func (backendCore *DockerKurtosisBackend) GetNetworkingSidecars(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]*networking_sidecar.NetworkingSidecar,
	error,
) {

	enclaveNetwork, err := backendCore.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	networkingSidecarContainers, err := backendCore.getNetworkingSidecarContainersByEnclaveIdAndUserServiceGUIDs(ctx, enclaveId, filters.GUIDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting networking-sidecar-containers by enclave ID '%v' and user service GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	networkingSidecars := map[service.ServiceGUID]*networking_sidecar.NetworkingSidecar{}
	for userServiceGuid, networkingSidecarContainer := range networkingSidecarContainers {
		//TODO remove this part, use the private ip label to get the IP address
		/*privateIpAddr, found := networkingSidecarContainer.GetNetworkIPAddresses()[enclaveNetwork.GetId()]
		if !found {
			return nil, stacktrace.Propagate(err, "Networking sidecar container with container ID '%v' does not have and IP address defined in Docker Network with ID '%v'; it should never happen it's a bug in Kurtosis", networkingSidecarContainer.GetId(), enclaveNetwork.GetId())
		}*/

		networkingSidecar := networking_sidecar.NewNetworkingSidecar(userServiceGuid, privateIpAddr, enclaveId)

		networkingSidecars[userServiceGuid] = networkingSidecar
	}

	return networkingSidecars, nil
}

func (backendCore *DockerKurtosisBackend) RunNetworkingSidecarExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	networkingSidecarsCommands map[service.ServiceGUID][]string,
)(
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
){
	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	userServiceGuids := map[service.ServiceGUID]bool{}
	for userServiceGuid := range networkingSidecarsCommands {
		userServiceGuids[userServiceGuid] = true
	}

	networkingSidecarContainers, err := backendCore.getNetworkingSidecarContainersByEnclaveIdAndUserServiceGUIDs(ctx, enclaveId, userServiceGuids)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting networking-sidecar-containers by enclave ID '%v' and user service GUIDs '%+v'", enclaveId, userServiceGuids)
	}

	// TODO Parallelize to increase perf
	for userServiceGuid, networkingSidecarContainer := range networkingSidecarContainers {
		networkingSidecarUnwrappedCommand := networkingSidecarsCommands[userServiceGuid]

		networkingSidecarShWrappedCmd := wrapNetworkingSidecarContainerShCommand(networkingSidecarUnwrappedCommand)

		execOutputBuf := &bytes.Buffer{}
		exitCode, err := backendCore.dockerManager.RunExecCommand(
			ctx,
			networkingSidecarContainer.GetId(),
			networkingSidecarShWrappedCmd,
			execOutputBuf)
		if err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred executing command '%+v' on networking sidecar with user service GUID '%v'",
				networkingSidecarShWrappedCmd,
				userServiceGuid,
			)
			erroredUserServiceGuids[userServiceGuid] = wrappedErr
			continue
		}
		if exitCode != succesfulExecCmdExitCode {
			exitCodeErr := stacktrace.NewError(
				"Expected exit code '%v' when running exec command '%+v' on networking sidecar with user service GUID '%v', but got exit code '%v' instead with the following output:\n%v",
				succesfulExecCmdExitCode,
				networkingSidecarShWrappedCmd,
				userServiceGuid,
				exitCode,
				execOutputBuf.String(),
			)
			erroredUserServiceGuids[userServiceGuid] = exitCodeErr
			continue
		}
		successfulUserServiceGuids[userServiceGuid] = true
	}

	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

func (backendCore *DockerKurtosisBackend) StopNetworkingSidecars(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {

	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	networkingSidecarContainers, err := backendCore.getNetworkingSidecarContainersByEnclaveIdAndUserServiceGUIDs(ctx, enclaveId, filters.GUIDs)

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting networking-sidecar-containers by enclave ID '%v' and user service GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	for userServiceGuid, networkingSidecarContainer := range networkingSidecarContainers {
		if err := backendCore.killContainerAndWaitForExit(ctx, networkingSidecarContainer); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred killing networking sidecar container with user service GUID '%v' and container ID '%v'", userServiceGuid, networkingSidecarContainer.GetId())
			erroredUserServiceGuids[userServiceGuid] = wrappedErr
			continue
		}
		successfulUserServiceGuids[userServiceGuid] = true
	}

	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

func (backendCore *DockerKurtosisBackend) DestroyNetworkingSidecars(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids := map[service.ServiceGUID]bool{}
	erroredUserServiceGuids := map[service.ServiceGUID]error{}

	networkingSidecarContainers, err := backendCore.getNetworkingSidecarContainersByEnclaveIdAndUserServiceGUIDs(ctx, enclaveId, filters.GUIDs)

	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting networking-sidecar-containers by enclave ID '%v' and user service GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	for userServiceGuid, networkingSidecarContainer := range networkingSidecarContainers {
		containersToRemove := []*types.Container{networkingSidecarContainer}
		if _, erroredContainers := backendCore.removeContainers(ctx, containersToRemove); len(erroredContainers) > 0 {
			containerError, found := erroredContainers[networkingSidecarContainer.GetId()]
			var wrappedErr error
			if !found {
				wrappedErr = stacktrace.NewError("Expected to find an error for container with ID '%v' in error list '%+v' but it was not found; it should never happens, it's a bug in Kurtosis", networkingSidecarContainer.GetId(), erroredContainers)
			} else {
				wrappedErr = stacktrace.Propagate(containerError, "An error occurred removing networking sidecar container with user service GUID '%v' and container ID '%v'", userServiceGuid, networkingSidecarContainer.GetId())
			}
			erroredUserServiceGuids[userServiceGuid] = wrappedErr
			continue
		}
		successfulUserServiceGuids[userServiceGuid] = true
	}
	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (backendCore *DockerKurtosisBackend) getNetworkingSidecarContainersByEnclaveIdAndUserServiceGUIDs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceGUIDs map[service.ServiceGUID]bool,
) (map[service.ServiceGUID]*types.Container, error) {

	searchLabels := map[string]string{
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.NetworkingSidecarContainerTypeLabelValue.GetString(),
	}
	foundContainers, err := backendCore.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchStoppedContainersWhenGettingNetworkingSidecarContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting containers using labels '%+v'", searchLabels)
	}

	networkingSidecarContainers := map[service.ServiceGUID]*types.Container{}
	for _, container := range foundContainers {
		for userServiceGuid := range userServiceGUIDs {
			//TODO we could improve this doing only one container iteration? or is this ok this way because is not to expensive?
			if hasEnclaveIdLabel(container, enclaveId) && hasGuidLabel(container, string(userServiceGuid)){
				networkingSidecarContainers[userServiceGuid] = container
			}
		}
	}
	return networkingSidecarContainers, nil
}


// Embeds the given command in a call to sh shell, so that a command with things
//  like '&&' will get executed as expected
func wrapNetworkingSidecarContainerShCommand(unwrappedCmd []string) []string {
	return []string{
		"sh",
		"-c",
		strings.Join(unwrappedCmd, " "),
	}
}
