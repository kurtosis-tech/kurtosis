package docker_kurtosis_backend

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_operation_parallelizer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	networkingSidecarImageName = "kurtosistech/iproute2"
	succesfulExecCmdExitCode   = 0
)

// TODO: MIGRATE THIS FOLDER TO USE STRUCTURE OF USER_SERVICE_FUNCTIONS MODULE

// We sleep forever because all the commands this container will run will be executed
//
//	via Docker exec
var sidecarContainerCommand = []string{
	"sleep", "infinity",
}

func (backend *DockerKurtosisBackend) CreateNetworkingSidecar(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
) (
	*networking_sidecar.NetworkingSidecar,
	error,
) {
	// Get the Docker network ID where we'll start the new sidecar container
	enclaveNetwork, err := backend.getEnclaveNetworkByEnclaveUuid(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	freeIpAddrProvider, found := backend.enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"Received a request to create networking sidecar for service with UUID '%v' in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			serviceUuid,
			enclaveId,
		)
	}

	_, dockerResources, err := shared_helpers.GetSingleUserServiceObjAndResourcesNoMutex(ctx, enclaveId, serviceUuid, backend.dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting network sidecar's user service '%v'", serviceUuid)
	}
	container := dockerResources.ServiceContainer

	enclaveObjAttrsProvider, err := backend.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForNetworkingSidecarContainer(serviceUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the networking sidecar container attributes for user service with UUID '%v'", serviceUuid)
	}
	containerName := containerAttrs.GetName()
	containerDockerLabels := containerAttrs.GetLabels()

	containerLabels := map[string]string{}
	for dockerLabelKey, dockerLabelValue := range containerDockerLabels {
		containerLabels[dockerLabelKey.GetString()] = dockerLabelValue.GetString()
	}

	ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a free IP address")
	}
	shouldReleaseIp := true
	defer func() {
		if shouldReleaseIp {
			if err = freeIpAddrProvider.ReleaseIpAddr(ipAddr); err != nil {
				logrus.Errorf("Error releasing IP address '%v'", ipAddr)
			}
		}
	}()

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
		docker_manager.NewContainerNetworkMode(container.GetId()),
	).WithCmdArgs(
		sidecarContainerCommand,
	).WithLabels(
		containerLabels,
	).Build()

	// Best-effort pull attempt
	if err = backend.dockerManager.PullImage(ctx, networkingSidecarImageName); err != nil {
		logrus.Warnf("Failed to pull the latest version of networking sidecar container image '%v'; you may be running an out-of-date version", networkingSidecarImageName)
	}

	containerId, _, err := backend.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the networking sidecar container")
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			// NOTE: We use the background context here so that the kill will still go off even if the reason for
			// the failure was the original context being cancelled
			if err := backend.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf(
					"Launching networking sidecar container '%v' with container ID '%v' didn't complete successfully so we "+
						"tried to kill the container we started, but doing so exited with an error:\n%v",
					containerName.GetString(),
					containerId,
					err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop networking sidecar container with ID '%v'!!!!!!", containerId)
			}
		}
	}()

	networkingSidecar, err := getNetworkingSidecarObjectFromContainerInfo(containerLabels, types.ContainerStatus_Running)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting networking sidecar object from container info with labels '%+v' and container status '%v'", containerLabels, types.ContainerStatus_Running)
	}

	shouldKillContainer = false
	shouldReleaseIp = false
	return networkingSidecar, nil

}

func (backend *DockerKurtosisBackend) GetNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceUUID]*networking_sidecar.NetworkingSidecar,
	error,
) {

	networkingSidecars, err := backend.getMatchingNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting networking sidecars matching filters '%+v'", filters)
	}

	successfulNetworkingSidecars := map[service.ServiceUUID]*networking_sidecar.NetworkingSidecar{}
	for _, networkingSidecar := range networkingSidecars {
		successfulNetworkingSidecars[networkingSidecar.GetServiceUUID()] = networkingSidecar
	}

	return successfulNetworkingSidecars, nil
}

func (backend *DockerKurtosisBackend) RunNetworkingSidecarExecCommands(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	networkingSidecarsCommands map[service.ServiceUUID][]string,
) (
	map[service.ServiceUUID]*exec_result.ExecResult,
	map[service.ServiceUUID]error,
	error,
) {
	successfulNetworkingSidecarExecResults := map[service.ServiceUUID]*exec_result.ExecResult{}
	erroredUserServiceUuids := map[service.ServiceUUID]error{}

	userServiceUuids := map[service.ServiceUUID]bool{}
	for userServiceUuid := range networkingSidecarsCommands {
		userServiceUuids[userServiceUuid] = true
	}

	filters := &networking_sidecar.NetworkingSidecarFilters{
		EnclaveUUIDs: map[enclave.EnclaveUUID]bool{
			enclaveUuid: true,
		},
		UserServiceUUIDs: userServiceUuids,
		Statuses:         nil,
	}

	networkingSidecars, err := backend.getMatchingNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting networking sidecars matching filters '%+v'", filters)
	}

	if len(networkingSidecarsCommands) != len(networkingSidecars) {
		return nil, nil, stacktrace.NewError("The amount of networking sidecars found '%v' are not equal to the amount of networking sidecars to run exec commands '%v'", len(networkingSidecars), len(networkingSidecarsCommands))
	}
	for _, networkingSidecar := range networkingSidecars {
		if _, found := networkingSidecarsCommands[networkingSidecar.GetServiceUUID()]; !found {
			return nil,
				nil,
				stacktrace.NewError(
					"Networking sidecar with user service UUID '%v' was found when getting matching "+
						"networking sidecars with filters '%+v' but it was not declared in the networking "+
						"sidecar exec commands list '%+v'",
					networkingSidecar.GetServiceUUID(),
					filters,
					networkingSidecarsCommands,
				)
		}
	}

	// TODO Parallelize to increase perf
	for containerId, networkingSidecar := range networkingSidecars {
		networkingSidecarCommand := networkingSidecarsCommands[networkingSidecar.GetServiceUUID()]

		execOutputBuf := &bytes.Buffer{}
		exitCode, err := backend.dockerManager.RunExecCommand(
			ctx,
			containerId,
			networkingSidecarCommand,
			execOutputBuf)
		if err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred executing command '%+v' on networking sidecar with user service UUID '%v'",
				networkingSidecarCommand,
				containerId,
			)
			erroredUserServiceUuids[networkingSidecar.GetServiceUUID()] = wrappedErr
			continue
		}
		newExecResult := exec_result.NewExecResult(exitCode, execOutputBuf.String())
		successfulNetworkingSidecarExecResults[networkingSidecar.GetServiceUUID()] = newExecResult
	}

	return successfulNetworkingSidecarExecResults, erroredUserServiceUuids, nil
}

func (backend *DockerKurtosisBackend) StopNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	matchingNetworkingSidecarsByContainerId, err := backend.getMatchingNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting networking sidecars matching filters '%+v'", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedObjectsByContainerId := map[string]interface{}{}
	for containerId, object := range matchingNetworkingSidecarsByContainerId {
		matchingUncastedObjectsByContainerId[containerId] = interface{}(object)
	}

	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.KillContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred killing networking sidecar container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulServiceUuidStrs, erroredServiceUuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedObjectsByContainerId,
		backend.dockerManager,
		extractServiceUUIDFromNetworkSidecarObj,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred killing networking sidecar containers matching filters '%+v'", filters)
	}

	successfulServiceUuids := map[service.ServiceUUID]bool{}
	for serviceUuidStr := range successfulServiceUuidStrs {
		successfulServiceUuids[service.ServiceUUID(serviceUuidStr)] = true
	}
	erroredUuids := map[service.ServiceUUID]error{}
	for serviceUuidStr, removalErr := range erroredServiceUuidStrs {
		erroredUuids[service.ServiceUUID(serviceUuidStr)] = removalErr
	}

	return successfulServiceUuids, erroredUuids, nil
}

func (backend *DockerKurtosisBackend) DestroyNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceUUID]bool,
	map[service.ServiceUUID]error,
	error,
) {
	networkingSidecars, err := backend.getMatchingNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting networking sidecars matching filters '%+v'", filters)
	}

	// TODO PLEAAASE GO GENERICS... but we can't use 1.18 yet because it'll break all Kurtosis clients :(
	matchingUncastedObjectsByContainerId := map[string]interface{}{}
	for containerId, object := range networkingSidecars {
		matchingUncastedObjectsByContainerId[containerId] = interface{}(object)
	}

	var dockerOperation docker_operation_parallelizer.DockerOperation = func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		dockerObjectId string,
	) error {
		if err := dockerManager.RemoveContainer(ctx, dockerObjectId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing networking sidecar container with ID '%v'", dockerObjectId)
		}
		return nil
	}

	successfulServiceUuidStrs, erroredServiceUuidStrs, err := docker_operation_parallelizer.RunDockerOperationInParallelForKurtosisObjects(
		ctx,
		matchingUncastedObjectsByContainerId,
		backend.dockerManager,
		extractServiceUUIDFromNetworkSidecarObj,
		dockerOperation,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred removing networking sidecar containers matching filters '%+v'", filters)
	}

	successfulServiceUuids := map[service.ServiceUUID]bool{}
	for serviceUuidStr := range successfulServiceUuidStrs {
		successfulServiceUuids[service.ServiceUUID(serviceUuidStr)] = true
	}
	erroredUuids := map[service.ServiceUUID]error{}
	for serviceUuidStr, removalErr := range erroredServiceUuidStrs {
		erroredUuids[service.ServiceUUID(serviceUuidStr)] = removalErr
	}

	return successfulServiceUuids, erroredUuids, nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func (backend *DockerKurtosisBackend) getMatchingNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (map[string]*networking_sidecar.NetworkingSidecar, error) {

	searchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.NetworkingSidecarContainerTypeDockerLabelValue.GetString(),
	}
	matchingContainers, err := backend.dockerManager.GetContainersByLabels(ctx, searchLabels, consts.ShouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching containers using labels: %+v", searchLabels)
	}

	matchingObjects := map[string]*networking_sidecar.NetworkingSidecar{}
	for _, container := range matchingContainers {
		containerId := container.GetId()
		object, err := getNetworkingSidecarObjectFromContainerInfo(
			container.GetLabels(),
			container.GetStatus(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with ID '%v' into a networking sidecar object", container.GetId())
		}

		if filters.EnclaveUUIDs != nil && len(filters.EnclaveUUIDs) > 0 {
			if _, found := filters.EnclaveUUIDs[object.GetEnclaveUUID()]; !found {
				continue
			}
		}

		if filters.UserServiceUUIDs != nil && len(filters.UserServiceUUIDs) > 0 {
			if _, found := filters.UserServiceUUIDs[object.GetServiceUUID()]; !found {
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

func getNetworkingSidecarObjectFromContainerInfo(
	labels map[string]string,
	containerStatus types.ContainerStatus,
) (*networking_sidecar.NetworkingSidecar, error) {

	enclaveId, found := labels[label_key_consts.EnclaveUUIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the networking sidecar's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveUUIDDockerLabelKey.GetString())
	}

	uuid, found := labels[label_key_consts.GUIDDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find GUID label key '%v' but none was found", label_key_consts.GUIDDockerLabelKey.GetString())
	}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for networking sidecar container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var status container_status.ContainerStatus
	if isContainerRunning {
		status = container_status.ContainerStatus_Running
	} else {
		status = container_status.ContainerStatus_Stopped
	}

	newObject := networking_sidecar.NewNetworkingSidecar(
		service.ServiceUUID(uuid),
		enclave.EnclaveUUID(enclaveId),
		status,
	)

	return newObject, nil
}

func extractServiceUUIDFromNetworkSidecarObj(uncastedObj interface{}) (string, error) {
	castedObj, ok := uncastedObj.(*networking_sidecar.NetworkingSidecar)
	if !ok {
		return "", stacktrace.NewError("An error occurred downcasting the networking sidecar object")
	}
	return string(castedObj.GetServiceUUID()), nil
}
