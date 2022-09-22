package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

const (
	shouldShowStoppedLogsCollectorContainers = true
)

func getLogsCollectorPrivatePorts(containerLabels map[string]string) (
	resultTcpPortSpec *port_spec.PortSpec,
	resultHttpPortSpec *port_spec.PortSpec,
	resultErr error,
) {

	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsDockerLabelKey.GetString())
	}

	portSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, nil,  stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
	}

	tcpPortSpec, foundTcpPort := portSpecs[logsCollectorTcpPortId]
	if !foundTcpPort {
		return nil, nil, stacktrace.NewError("No logs collector tcp port with ID '%v' found in the logs collector port specs", logsCollectorTcpPortId)
	}

	httpPortSpec, foundHttpPort := portSpecs[logsCollectorHttpPortId]
	if !foundHttpPort {
		return nil, nil, stacktrace.NewError("No logs collector http port with ID '%v' found in the logs collector port specs", logsCollectorHttpPortId)
	}

	return tcpPortSpec, httpPortSpec, nil
}


func getLogsCollectorObjectFromContainerInfo(
	ctx context.Context,
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	dockerManager *docker_manager.DockerManager,
) (*logs_collector.LogsCollector, error) {

	privateIpAddrStr, err := dockerManager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the private IP address of container '%v' in network '%v'", containerId, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn)
	}
	privateIpAddr := net.ParseIP(privateIpAddrStr)
	if privateIpAddr == nil {
		return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
	}

	privateTcpPortSpec, privateHttpPortSpec, err := getLogsCollectorPrivatePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for logs collector container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var logsCollectorStatus container_status.ContainerStatus
	if isContainerRunning {
		logsCollectorStatus = container_status.ContainerStatus_Running
	} else {
		logsCollectorStatus = container_status.ContainerStatus_Stopped
	}

	logsCollectorObj := logs_collector.NewLogsCollector(
		logsCollectorStatus,
		privateIpAddr,
		privateTcpPortSpec,
		privateHttpPortSpec,
	)

	return logsCollectorObj, nil
}

func getAllLogsCollectorContainers(ctx context.Context, dockerManager *docker_manager.DockerManager) ([]*types.Container, error) {
	matchingLogsCollectorContainers := []*types.Container{}

	logsCollectorContainerSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.LogsCollectorTypeDockerLabelValue.GetString(),
	}

	matchingLogsCollectorContainers, err := dockerManager.GetContainersByLabels(ctx, logsCollectorContainerSearchLabels, shouldShowStoppedLogsCollectorContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching logs collector containers using labels: %+v", logsCollectorContainerSearchLabels)
	}
	return matchingLogsCollectorContainers, nil
}
