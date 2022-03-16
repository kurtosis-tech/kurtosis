package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func (backend *DockerKurtosisBackend) killContainerAndWaitForExit(
	ctx context.Context,
	container *types.Container,
) error {
	containerId := container.GetId()
	containerName := container.GetName()
	if err := backend.dockerManager.KillContainer(ctx, containerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred killing container '%v' with ID '%v'",
			containerName,
			containerId,
		)
	}
	if _, err := backend.dockerManager.WaitForExit(ctx, containerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred waiting for container '%v' with ID '%v' to exit after killing it",
			container.GetName(),
			containerId,
		)
	}

	return nil
}

func (backend *DockerKurtosisBackend) killContainers(
	ctx context.Context,
	containers []*types.Container,
)(
	successfulContainers map[string]bool,
	erroredContainers map[string]error,
){

	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if err := backend.dockerManager.KillContainer(ctx, containerId); err != nil {
			containerError :=  stacktrace.Propagate(
				err,
				"An error occurred killing container '%v' with ID '%v'",
				container.GetName(),
				containerId,
			)
			erroredContainers[container.GetId()] = containerError
			continue
		}
		successfulContainers[containerId] = true
	}

	return successfulContainers, erroredContainers
}

func (backend *DockerKurtosisBackend) waitForExitContainers(
	ctx context.Context,
	containers []*types.Container,
)(
	successfulContainers map[string]bool,
	erroredContainers map[string]error,
){
	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if _, err := backend.dockerManager.WaitForExit(ctx, containerId); err != nil {
			containerError := stacktrace.Propagate(
				err,
				"An error occurred waiting for container '%v' with ID '%v' to exit",
				container.GetName(),
				containerId,
			)
			erroredContainers[container.GetId()] = containerError
			continue
		}
		successfulContainers[containerId] = true
	}

	return successfulContainers, erroredContainers
}

func (backend *DockerKurtosisBackend) removeContainer(
	ctx context.Context,
	container *types.Container) error {

	containerId := container.GetId()
	if err := backend.dockerManager.RemoveContainer(ctx, containerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred removing container '%v' with ID '%v'",
			container.GetName(),
			containerId,
		)
	}
	return nil
}

func (backend *DockerKurtosisBackend) removeContainers(
	ctx context.Context,
	containers []*types.Container,
)(
	successfulContainers map[string]bool,
	erroredContainers map[string]error,
){
	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if err := backend.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			containerError := stacktrace.Propagate(
				err,
				"An error occurred removing container '%v' with ID '%v'",
				container.GetName(),
				containerId,
			)
			erroredContainers[container.GetId()] = containerError
			continue
		}
		successfulContainers[containerId] = true
	}

	return successfulContainers, erroredContainers
}

func getNetworkingSidecarContainersFromContainerListByGUIDs(
	containers []*types.Container,
	guids map[networking_sidecar.NetworkingSidecarGUID]bool,
) map[networking_sidecar.NetworkingSidecarGUID]*types.Container {

	networkingSidecarContainers := map[networking_sidecar.NetworkingSidecarGUID]*types.Container{}
	for _, container := range containers {
		if isNetworkingSidecarContainer(container) {
			for networkingSidecarGuid := range guids {
				userServiceGuid := service.ServiceGUID(networkingSidecarGuid)
				if hasUserServiceGuidLabel(container, userServiceGuid){
					networkingSidecarContainers[networkingSidecarGuid] = container
				}
			}
		}
	}
	return networkingSidecarContainers
}

func getUserServiceContainerFromContainerListByEnclaveIdAndUserServiceGUID(
	containers []*types.Container,
	enclaveId enclave.EnclaveID,
	userServiceGUID service.ServiceGUID) *types.Container {

	for _, container := range containers {
		if isUserServiceContainer(container) && hasEnclaveIdLabelAndUserServiceGuidLabel(container, enclaveId, userServiceGUID) {
			return container
		}
	}
	return nil
}

func isUserServiceContainer(container *types.Container) bool {
	labels := container.GetLabels()
	containerTypeValue, found := labels[label_key_consts.ContainerTypeLabelKey.GetString()]
	if !found {
		//TODO Do all containers should have container type label key??? we should return and error here if this answer is yes??
		logrus.Debugf("Container with ID '%v' haven't label '%v'", container.GetId(), label_key_consts.ContainerTypeLabelKey.GetString())
		return false
	}
	if containerTypeValue == label_value_consts.UserServiceContainerTypeLabelValue.GetString() {
		return true
	}
	return false
}

func isNetworkingSidecarContainer(container *types.Container) bool {
	labels := container.GetLabels()
	containerTypeValue, found := labels[label_key_consts.ContainerTypeLabelKey.GetString()]
	if !found {
		//TODO Do all containers should have container type label key??? we should return and error here if this answer is yes??
		logrus.Debugf("Container with ID '%v' haven't label '%v'", container.GetId(), label_key_consts.ContainerTypeLabelKey.GetString())
		return false
	}
	if containerTypeValue == label_value_consts.NetworkingSidecarContainerTypeLabelValue.GetString() {
		return true
	}
	return false
}

func hasEnclaveIdLabelAndUserServiceGuidLabel(
		container *types.Container,
		enclaveId enclave.EnclaveID,
		userServiceGuid service.ServiceGUID) bool {

	labels := container.GetLabels()
	enclaveIdLabelValue, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		//TODO Do all containers should have enclave ID label key??? we should return and error here if this answer is yes??
		logrus.Debugf("Container with ID '%v' haven't label '%v'", container.GetId(), label_key_consts.EnclaveIDLabelKey.GetString())
		return false
	}
	if enclaveIdLabelValue == string(enclaveId) && hasUserServiceGuidLabel(container, userServiceGuid) {
		return true
	}
	return false
}

func hasUserServiceGuidLabel(container *types.Container, userServiceGuid service.ServiceGUID) bool {
	labels := container.GetLabels()
	userServiceGuidLabelValueStr, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return false
	}
	userServiceGuidLabelValue := service.ServiceGUID(userServiceGuidLabelValueStr)
	if userServiceGuidLabelValue == userServiceGuid {
		return true
	}
	return false
}

