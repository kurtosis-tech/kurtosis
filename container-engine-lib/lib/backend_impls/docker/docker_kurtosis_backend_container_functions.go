package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func (backendCore *DockerKurtosisBackend) killContainerAndWaitForExit(
	ctx context.Context,
	container *types.Container,
) error {
	containerId := container.GetId()
	containerName := container.GetName()
	if err := backendCore.dockerManager.KillContainer(ctx, containerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred killing container '%v' with ID '%v'",
			containerName,
			containerId,
		)
	}
	if _, err := backendCore.dockerManager.WaitForExit(ctx, containerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred waiting for container '%v' with ID '%v' to exit after killing it",
			container.GetName(),
			containerId,
		)
	}

	return nil
}

func (backendCore *DockerKurtosisBackend) killContainers(
	ctx context.Context,
	containers []*types.Container,
)(
	successfulContainers map[string]bool,
	erroredContainers map[string]error,
){

	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if err := backendCore.dockerManager.KillContainer(ctx, containerId); err != nil {
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

func (backendCore *DockerKurtosisBackend) waitForExitContainers(
	ctx context.Context,
	containers []*types.Container,
)(
	successfulContainers map[string]bool,
	erroredContainers map[string]error,
){
	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if _, err := backendCore.dockerManager.WaitForExit(ctx, containerId); err != nil {
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

func (backendCore *DockerKurtosisBackend) removeContainer(
	ctx context.Context,
	container *types.Container) error {

	containerId := container.GetId()
	if err := backendCore.dockerManager.RemoveContainer(ctx, containerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred removing container '%v' with ID '%v'",
			container.GetName(),
			containerId,
		)
	}
	return nil
}

func (backendCore *DockerKurtosisBackend) removeContainers(
	ctx context.Context,
	containers []*types.Container,
)(
	successfulContainers map[string]bool,
	erroredContainers map[string]error,
){
	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if err := backendCore.dockerManager.RemoveContainer(ctx, containerId); err != nil {
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

func (backendCore *DockerKurtosisBackend) getNetworkingSidecarContainersByEnclaveIdAndNetworkingSidecarGUIDs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	networkingSidecarGUIDs map[networking_sidecar.NetworkingSidecarGUID]bool,
) (map[networking_sidecar.NetworkingSidecarGUID]*types.Container, error) {

	enclaveContainers, err := backendCore.getEnclaveContainers(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
	}

	networkingSidecarContainers := map[networking_sidecar.NetworkingSidecarGUID]*types.Container{}
	for _, container := range enclaveContainers {
		if isNetworkingSidecarContainer(container) {
			for networkingSidecarGuid := range networkingSidecarGUIDs {
				userServiceGuid := service.ServiceGUID(networkingSidecarGuid)
				if hasUserServiceGuidLabel(container, userServiceGuid){
					networkingSidecarContainers[networkingSidecarGuid] = container
				}
			}
		}
	}
	return networkingSidecarContainers, nil
}

func (backendCore *DockerKurtosisBackend) getUserServiceContainerByEnclaveIDAndUserServiceGUID(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceGuid service.ServiceGUID,
)(
	*types.Container,
	error,
) {
	userServiceGuids := map[service.ServiceGUID]bool{
		userServiceGuid: true,
	}

	userServiceContainers, err := backendCore.getUserServiceContainersByEnclaveIDAndUserServiceGUIDs(ctx, enclaveId, userServiceGuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user-service-containers by enclave ID '%v' user service GUID '%v'", enclaveId, userServiceGuid)
	}
	numOfUserServiceContainers := len(userServiceContainers)
	if numOfUserServiceContainers == 0 {
		return nil, stacktrace.NewError("No user service with GUID '%v' in enclave with ID '%v' was found to wait for availability", userServiceGuid, enclaveId)
	}
	if numOfUserServiceContainers > 1 {
		return nil, stacktrace.NewError("Expected to find only one user service with GUID '%v' in enclave with ID '%v', but '%v' was found", userServiceGuid, enclaveId, numOfUserServiceContainers)
	}

	userServiceContainer := userServiceContainers[userServiceGuid]

	return userServiceContainer, nil
}

func (backendCore *DockerKurtosisBackend) getUserServiceContainersByEnclaveIDAndUserServiceGUIDs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	userServiceGuids map[service.ServiceGUID]bool,
) (map[service.ServiceGUID]*types.Container, error) {


	enclaveContainers, err := backendCore.getEnclaveContainers(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave status and containers for enclave with ID '%v'", enclaveId)
	}

	userServiceContainers := map[service.ServiceGUID]*types.Container{}
	for _, container := range enclaveContainers {
		if isUserServiceContainer(container) {
			for userServiceGuid := range userServiceGuids {
				if hasUserServiceGuidLabel(container, userServiceGuid){
					userServiceContainers[userServiceGuid] = container
				}
			}
		}
	}
	return userServiceContainers, nil
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

func getServiceIdFromContainer(container *types.Container) (service.ServiceID, error) {
	if !isUserServiceContainer(container) {
		return "", stacktrace.NewError("Can not possible to get service ID from container with ID '%v' because it's not a user service container", container.GetId())
	}
	labels := container.GetLabels()
	serviceIdLabelValue, found := labels[label_key_consts.IDLabelKey.GetString()]
	if !found {
		return "",  stacktrace.NewError("Expected to find container's label with key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
	}
	serviceId := service.ServiceID(serviceIdLabelValue)

	return serviceId, nil
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

func getPrivatePortsFromContainerLabels(containerLabels map[string]string) (map[string]*port_spec.PortSpec, error) {
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return  nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		// TODO AFTER 2022-05-02 SWITCH THIS TO A PLAIN ERROR WHEN WE'RE SURE NOBODY WILL BE USING THE OLD PORT SPEC STRING!
		oldPortSpecs, err := deserialize_pre_2022_03_02_PortSpecs(serializedPortSpecs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v' even when trying the old method", serializedPortSpecs)
		}
		portSpecs = oldPortSpecs
	}

	return portSpecs, nil
}
