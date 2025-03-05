package logs_aggregator_functions

import (
	"context"
	"net"

	"github.com/docker/docker/api/types/volume"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shouldShowStoppedLogsAggregatorContainers = true
)

func getLogsAggregatorPrivatePorts(containerLabels map[string]string) (*port_spec.PortSpec, error) {
	serializedPortSpecs, found := containerLabels[docker_label_key.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", docker_label_key.PortSpecsDockerLabelKey.GetString())
	}

	portSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
	}

	httpPortSpec, foundHttpPort := portSpecs[logsAggregatorHttpPortId]
	if !foundHttpPort {
		return nil, stacktrace.NewError("No logs aggregator HTTP port with ID '%v' found in the logs aggregator port specs", logsAggregatorHttpPortId)
	}

	return httpPortSpec, nil
}

// Returns nil [LogsAggregator] object if no container is found
func getLogsAggregatorObjectAndContainerId(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (*logs_aggregator.LogsAggregator, string, error) {
	logsAggregatorContainer, found, err := getLogsAggregatorContainer(ctx, dockerManager)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting all logs aggregator containers")
	}
	if !found {
		return nil, "", nil
	}

	logsAggregatorContainerID := logsAggregatorContainer.GetId()

	logsAggregatorObject, err := getLogsAggregatorObjectFromContainerInfo(
		ctx,
		logsAggregatorContainerID,
		logsAggregatorContainer.GetLabels(),
		logsAggregatorContainer.GetStatus(),
		dockerManager,
	)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting the logs Aggregator object using container ID '%v', labels '%+v' and the status '%v'", logsAggregatorContainer.GetId(), logsAggregatorContainer.GetLabels(), logsAggregatorContainer.GetStatus())
	}

	return logsAggregatorObject, logsAggregatorContainerID, nil
}

// Returns nil [Container] object and false if no logs aggregator container is found
func getLogsAggregatorContainer(ctx context.Context, dockerManager *docker_manager.DockerManager) (*types.Container, bool, error) {
	logsAggregatorContainerSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.ContainerTypeDockerLabelKey.GetString(): label_value_consts.LogsAggregatorTypeDockerLabelValue.GetString(),
	}

	matchingLogsAggregatorContainers, err := dockerManager.GetContainersByLabels(ctx, logsAggregatorContainerSearchLabels, shouldShowStoppedLogsAggregatorContainers)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "An error occurred fetching the logs aggregator container using labels: %+v", logsAggregatorContainerSearchLabels)
	}

	if len(matchingLogsAggregatorContainers) == 0 {
		return nil, false, nil
	}
	if len(matchingLogsAggregatorContainers) > 1 {
		return nil, false, stacktrace.NewError("Found more than one logs aggregator Docker container'; this is a bug in Kurtosis")
	}
	return matchingLogsAggregatorContainers[0], true, nil
}

func getLogsAggregatorObjectFromContainerInfo(
	ctx context.Context,
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	dockerManager *docker_manager.DockerManager,
) (*logs_aggregator.LogsAggregator, error) {
	var privateIpAddr net.IP

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for logs aggregator container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}

	privateHttpPortSpec, err := getLogsAggregatorPrivatePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	var logsAggregatorStatus container.ContainerStatus
	if isContainerRunning {
		logsAggregatorStatus = container.ContainerStatus_Running

		privateIpAddrStr, err := dockerManager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the private IP address of container '%v' in network '%v'", containerId, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn)
		}
		privateIpAddr = net.ParseIP(privateIpAddrStr)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
		}
	} else {
		logsAggregatorStatus = container.ContainerStatus_Stopped
	}

	logsAggregatorObj := logs_aggregator.NewLogsAggregator(
		logsAggregatorStatus,
		privateIpAddr,
		defaultLogsListeningPortNum,
		privateHttpPortSpec,
	)

	return logsAggregatorObj, nil
}

// if nothing is found we return empty volume name
func getLogsAggregatorVolumeName(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (string, error) {

	var volumes []*volume.Volume

	searchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():      label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.VolumeTypeDockerLabelKey.GetString(): label_value_consts.LogsAggregatorVolumeTypeDockerLabelValue.GetString(),
	}

	volumes, err := dockerManager.GetVolumesByLabels(ctx, searchLabels)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the volumes for logs aggregator by labels '%+v'", searchLabels)
	}

	if len(volumes) == 0 {
		return "", nil
	}

	if len(volumes) > 1 {
		return "", stacktrace.NewError("Attempted to get logs collector volume name for logs aggregator but got more than one matches")
	}

	return volumes[0].Name, nil
}
