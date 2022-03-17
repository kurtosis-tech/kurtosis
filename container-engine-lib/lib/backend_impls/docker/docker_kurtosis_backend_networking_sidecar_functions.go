package docker

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
)

const (
	networkingSidecarImageName = "kurtosistech/iproute2"
	succesfulExecCmdExitCode = 0
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
	networkLabels := map[string]string{
		label_key_consts.IDLabelKey.GetString(): string(enclaveId),
	}
	matchingNetworks, err := backendCore.dockerManager.GetNetworksByLabels(ctx, networkLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting docker networks by labels '%+v'", networkLabels)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 {
		return nil, stacktrace.NewError("No network found for enclave with ID '%v'", enclaveId)
	}
	if numMatchingNetworks > 1 {
		return nil, stacktrace.NewError("Found '%v' enclave networks with ID '%v', which shouldn't happen", numMatchingNetworks, enclaveId)
	}
	enclaveNetwork := matchingNetworks[0]

	enclaveStatus, enclaveContainers, err := backendCore.getEnclaveStatusAndContainers(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
	}

	if enclaveStatus != enclave.EnclaveStatus_Running {
		return nil, stacktrace.NewError("Networking sidecar for user service with GUID '%v' can not be created inside enclave with ID '%v' because its current status is '%v' and it must be '%v' to accept new nodes", serviceGuid, enclaveId, enclaveStatus, enclave.EnclaveStatus_Running.String())
	}

	userServiceContainer := getUserServiceContainerFromContainerListByEnclaveIdAndUserServiceGUID(enclaveContainers, enclaveId, serviceGuid)

	enclaveObjAttrsProvider, err := backendCore.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForNetworkingSidecarContainer(serviceGuid)
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

	networkingSidecarGuid := networking_sidecar.NetworkingSidecarGUID(serviceGuid)

	networkingSidecar := networking_sidecar.NewNetworkingSidecar(networkingSidecarGuid, ipAddr, enclaveId)

	return networkingSidecar, nil

}

func (backend *DockerKurtosisBackend) GetNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[networking_sidecar.NetworkingSidecarGUID]*networking_sidecar.NetworkingSidecar,
	error,
) {
	enclaveId := filters.EnclaveId

	enclaveIDs := map[enclave.EnclaveID]bool{
		enclaveId: true,
	}

	networks, err := backend.getEnclaveNetworksByEnclaveIds(ctx, enclaveIDs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave networks by enclave IDs '%+v'", enclaveIDs)
	}
	numMatchingNetworks := len(networks)
	if numMatchingNetworks == 0 {
		return nil, stacktrace.Propagate(err, "No Enclave was founded with some of these IDs '%+v'", enclaveIDs)
	}
	if numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Found %v networks matching name '%v' when we expected just one - this is likely a bug in Kurtosis!",
			numMatchingNetworks,
			enclaveId,
		)
	}
	enclaveNetwork := networks[0]

	enclaveContainers, err := backend.getEnclaveContainers(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
	}

	networkingSidecarContainers:= getNetworkingSidecarContainersFromContainerListByGUIDs(enclaveContainers, filters.GUIDs)

	networkingSidecars := map[networking_sidecar.NetworkingSidecarGUID]*networking_sidecar.NetworkingSidecar{}
	for networkingSidecarGuid, networkingSidecarContainer := range networkingSidecarContainers {
		ip, found := networkingSidecarContainer.GetNetworksIPAddresses()[enclaveNetwork.GetId()]
		if !found {
			return nil, stacktrace.Propagate(err, "Networking sidecar container with container ID '%v' does not have and IP address defined in Docker Network with ID '%v'; it should never happen it's a bug in Kurtosis", networkingSidecarContainer.GetId(), enclaveNetwork.GetId())
		}

		networkingSidecar := networking_sidecar.NewNetworkingSidecar(networkingSidecarGuid, ip, enclaveId)

		networkingSidecars[networkingSidecarGuid] = networkingSidecar
	}

	return networkingSidecars, nil
}

func (backend *DockerKurtosisBackend) RunNetworkingSidecarsExecCommand(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	networkingSidecarsCommands map[networking_sidecar.NetworkingSidecarGUID][]string,
)(
	successfulSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]bool,
	erroredSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]error,
	resultErr error,
){
	enclaveContainers, err := backend.getEnclaveContainers(ctx, enclaveId)
	if err != nil {
		return nil, nil,  stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
	}

	networkingSidecarGuids := map[networking_sidecar.NetworkingSidecarGUID]bool{}
	for networkingSidecarGuid := range networkingSidecarsCommands {
		networkingSidecarGuids[networkingSidecarGuid] = true
	}

	networkingSidecarContainers := getNetworkingSidecarContainersFromContainerListByGUIDs(enclaveContainers, networkingSidecarGuids)

	for networkingSidecarGuid, networkingSidecarContainer := range networkingSidecarContainers {
		networkingSidecarUnwrappedCommand := networkingSidecarsCommands[networkingSidecarGuid]

		networkingSidecarShWrappedCmd := wrapNetworkingSidecarContainerShCommand(networkingSidecarUnwrappedCommand)

		execOutputBuf := &bytes.Buffer{}

		exitCode, err := backend.dockerManager.RunExecCommand(
			ctx,
			networkingSidecarContainer.GetId(),
			networkingSidecarShWrappedCmd,
			execOutputBuf)
		if err != nil || exitCode != succesfulExecCmdExitCode {
			logrus.Errorf("------------------ Exec command output for command '%v' --------------------", networkingSidecarShWrappedCmd)
			if _, outputErr := io.Copy(logrus.StandardLogger().Out, execOutputBuf); outputErr != nil {
				logrus.Errorf("An error occurred printing the exec logs: %v", outputErr)
			}
			logrus.Errorf("------------------ END Exec command output for command '%v' --------------------", networkingSidecarShWrappedCmd)
			var resultErr error
			if err != nil {
				resultErr = stacktrace.Propagate(err, "An error occurred running exec command '%v'", networkingSidecarShWrappedCmd)
			}
			if exitCode != succesfulExecCmdExitCode {
				resultErr = stacktrace.NewError(
					"Expected exit code '%v' when running exec command '%v', but got exit code '%v' instead",
					succesfulExecCmdExitCode,
					networkingSidecarShWrappedCmd,
					exitCode)
			}
			erroredSidecarGuids[networkingSidecarGuid] = resultErr
			continue
		}
		successfulSidecarGuids[networkingSidecarGuid] = true
	}

	return successfulSidecarGuids, erroredSidecarGuids, nil
}

func (backend *DockerKurtosisBackend) StopNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	successfulSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]bool,
	erroredSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]error,
	resultErr error,
) {

	enclaveContainers, err := backend.getEnclaveContainers(ctx, filters.EnclaveId)
	if err != nil {
		return nil, nil,  stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", filters.EnclaveId)
	}

	networkingSidecarContainers:= getNetworkingSidecarContainersFromContainerListByGUIDs(enclaveContainers, filters.GUIDs)

	for networkingSidecarGuid, networkingSidecarContainer := range networkingSidecarContainers {
		if err := backend.killContainerAndWaitForExit(ctx, networkingSidecarContainer); err != nil {
			erroredSidecarGuids[networkingSidecarGuid] = err
			continue
		}
		successfulSidecarGuids[networkingSidecarGuid] = true
	}

	return successfulSidecarGuids, erroredSidecarGuids, nil
}

func (backend *DockerKurtosisBackend) DestroyNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	successfulSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]bool,
	erroredSidecarGuids map[networking_sidecar.NetworkingSidecarGUID]error,
	resultErr error,
) {
	enclaveContainers, err := backend.getEnclaveContainers(ctx, filters.EnclaveId)
	if err != nil {
		return nil, nil,  stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", filters.EnclaveId)
	}

	networkingSidecarContainers:= getNetworkingSidecarContainersFromContainerListByGUIDs(enclaveContainers, filters.GUIDs)

	for networkingSidecarGuid, networkingSidecarContainer := range networkingSidecarContainers {
		if err := backend.removeContainer(ctx, networkingSidecarContainer); err != nil {
			erroredSidecarGuids[networkingSidecarGuid] = err
			continue
		}
		successfulSidecarGuids[networkingSidecarGuid] = true
	}
	return successfulSidecarGuids, erroredSidecarGuids, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
// Embeds the given command in a call to whichever shell is native to the image, so that a command with things
//  like '&&' will get executed as expected
func wrapNetworkingSidecarContainerShCommand(unwrappedCmd []string) []string {
	return []string{
		"sh",
		"-c",
		strings.Join(unwrappedCmd, " "),
	}
}
