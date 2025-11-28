package logs_collector_functions

import (
	"context"
	"net"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"

	"github.com/docker/docker/api/types/volume"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shouldShowStoppedLogsCollectorContainers = true
)

func getLogsCollectorPrivatePorts(containerLabels map[string]string) (*port_spec.PortSpec, *port_spec.PortSpec, error) {

	serializedPortSpecs, found := containerLabels[docker_label_key.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", docker_label_key.PortSpecsDockerLabelKey.GetString())
	}

	portSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
	}

	tcpPortSpec, foundTcpPort := portSpecs[logsCollectorTcpPortId]
	if !foundTcpPort {
		return nil, nil, stacktrace.NewError("No logs collector TCP port with ID '%v' found in the logs collector port specs", logsCollectorTcpPortId)
	}

	httpPortSpec, foundHttpPort := portSpecs[logsCollectorHttpPortId]
	if !foundHttpPort {
		return nil, nil, stacktrace.NewError("No logs collector HTTP port with ID '%v' found in the logs collector port specs", logsCollectorHttpPortId)
	}

	return tcpPortSpec, httpPortSpec, nil
}

func getLogsCollectorObjectFromContainerInfo(
	ctx context.Context,
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	enclaveNetworkId *types.Network,
	dockerManager *docker_manager.DockerManager,
) (*logs_collector.LogsCollector, error) {

	var (
		logsCollectorStatus     container.ContainerStatus
		privateIpAddr           net.IP
		bridgeNetworkIpAddr     net.IP
		enclaveNetworkIpAddress string
		bridgeNetworkIpAddress  string
		err                     error
	)

	privateTcpPortSpec, privateHttpPortSpec, err := getLogsCollectorPrivatePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for logs collector container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}

	if isContainerRunning {
		logsCollectorStatus = container.ContainerStatus_Running

		enclaveNetworkIpAddress, err = dockerManager.GetContainerIP(ctx, enclaveNetworkId.GetName(), containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the ip address of container '%v' in network '%v'", containerId, enclaveNetworkId)
		}
		privateIpAddr = net.ParseIP(enclaveNetworkIpAddress)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse '%v' network ip address string '%v' to an IP", enclaveNetworkId, enclaveNetworkIpAddress)
		}

		bridgeNetworkName := dockerManager.GetBridgeNetworkName()
		bridgeNetworkIpAddress, err = dockerManager.GetContainerIP(ctx, bridgeNetworkName, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while getting the '%v' network ip address of container '%v' in network", bridgeNetworkName, containerId)
		}
		bridgeNetworkIpAddr = net.ParseIP(bridgeNetworkIpAddress)
		if bridgeNetworkIpAddr == nil {
			return nil, stacktrace.Propagate(err, "Couldn't parse '%v' network ip address string '%v' to ip", bridgeNetworkName, bridgeNetworkIpAddress)
		}
	} else {
		logsCollectorStatus = container.ContainerStatus_Stopped
	}

	logsCollectorObj := logs_collector.NewLogsCollector(
		logsCollectorStatus,
		privateIpAddr,
		bridgeNetworkIpAddr,
		privateTcpPortSpec,
		privateHttpPortSpec,
	)

	return logsCollectorObj, nil
}

func getLogsCollectorForTheGivenEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, dockerManager *docker_manager.DockerManager) ([]*types.Container, error) {
	logsCollectorContainerSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.ContainerTypeDockerLabelKey.GetString(): label_value_consts.LogsCollectorTypeDockerLabelValue.GetString(),
		docker_label_key.EnclaveUUIDDockerLabelKey.GetString():   string(enclaveUuid),
	}

	matchingLogsCollectorContainers, err := dockerManager.GetContainersByLabels(ctx, logsCollectorContainerSearchLabels, shouldShowStoppedLogsCollectorContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching logs collector containers using labels: %+v", logsCollectorContainerSearchLabels)
	}
	return matchingLogsCollectorContainers, nil
}

// If nothing is found returns nil
func getLogsCollectorObjectAndContainerId(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	enclaveNetworkId *types.Network,
	dockerManager *docker_manager.DockerManager,
) (*logs_collector.LogsCollector, string, error) {
	allLogsCollectorContainers, err := getLogsCollectorForTheGivenEnclave(ctx, enclaveUuid, dockerManager)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting all logs collector containers")
	}

	if len(allLogsCollectorContainers) == 0 {
		return nil, "", nil
	}
	if len(allLogsCollectorContainers) > 1 {
		return nil, "", stacktrace.NewError("Found more than one logs collector Docker container'; this is a bug in Kurtosis")
	}

	logsCollectorContainer := allLogsCollectorContainers[0]
	logsCollectorContainerID := logsCollectorContainer.GetId()
	hostMachinePortBindings := logsCollectorContainer.GetHostPortBindings()

	logsCollectorObject, err := getLogsCollectorObjectFromContainerInfo(
		ctx,
		logsCollectorContainerID,
		logsCollectorContainer.GetLabels(),
		logsCollectorContainer.GetStatus(),
		enclaveNetworkId,
		dockerManager,
	)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting logs collector object using container ID '%v', labels '%+v', status '%v' and host machine port bindings '%+v'", logsCollectorContainer.GetId(), logsCollectorContainer.GetLabels(), logsCollectorContainer.GetStatus(), hostMachinePortBindings)
	}

	return logsCollectorObject, logsCollectorContainerID, nil
}

// if nothing is found we return empty volume name
func getEnclaveLogsCollectorVolumeName(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaveUuid enclave.EnclaveUUID,
) (string, error) {

	var volumes []*volume.Volume

	searchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():       label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.VolumeTypeDockerLabelKey.GetString():  label_value_consts.LogsCollectorVolumeTypeDockerLabelValue.GetString(),
		docker_label_key.EnclaveUUIDDockerLabelKey.GetString(): string(enclaveUuid),
	}

	volumes, err := dockerManager.GetVolumesByLabels(ctx, searchLabels)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the volumes for enclave '%v' by labels '%+v'", enclaveUuid, searchLabels)
	}

	if len(volumes) == 0 {
		return "", nil
	}

	if len(volumes) > 1 {
		return "", stacktrace.NewError("Attempted to get logs collector volume name for enclave '%v' but got more than one matches", enclaveUuid)
	}

	return volumes[0].Name, nil
}
