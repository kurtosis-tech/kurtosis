package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

const (
	shouldShowStoppedLogsDatabaseContainers = true
)

func getLogsDatabasePrivatePorts(containerLabels map[string]string) (
	resultHttpPortSpec *port_spec.PortSpec,
	resultErr error,
) {

	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsDockerLabelKey.GetString())
	}

	portSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
	}

	httpPortSpec, foundHttpPort := portSpecs[logsDatabaseHttpPortId]
	if !foundHttpPort {
		return nil, stacktrace.NewError("No logs database http port with ID '%v' found in the logs database port specs", logsDatabaseHttpPortId)
	}

	return httpPortSpec,  nil
}

func getLogsDatabaseObjectFromContainerInfo(
	ctx context.Context,
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	dockerManager *docker_manager.DockerManager,
) (*logs_database.LogsDatabase, error) {

	privateIpAddrStr, err := dockerManager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the private IP address of container '%v' in network '%v'", containerId, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn)
	}
	privateIpAddr := net.ParseIP(privateIpAddrStr)
	if privateIpAddr == nil {
		return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
	}

	privateHttpPortSpec, err := getLogsDatabasePrivatePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for logs database container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var logsDatabaseStatus container_status.ContainerStatus
	if isContainerRunning {
		logsDatabaseStatus = container_status.ContainerStatus_Running
	} else {
		logsDatabaseStatus = container_status.ContainerStatus_Stopped
	}

	logsDatabaseObj := logs_database.NewLogsDatabase(
		logsDatabaseStatus,
		privateIpAddr,
		privateHttpPortSpec,
	)

	return logsDatabaseObj, nil
}

func getAllLogsDatabaseContainers(ctx context.Context, dockerManager *docker_manager.DockerManager) ([]*types.Container, error) {
	matchingLogsDatabaseContainers := []*types.Container{}

	logsDatabaseContainerSearchLabels := map[string]string{
		label_key_consts.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(): label_value_consts.LogsDatabaseTypeDockerLabelValue.GetString(),
	}

	matchingLogsDatabaseContainers, err := dockerManager.GetContainersByLabels(ctx, logsDatabaseContainerSearchLabels, shouldShowStoppedLogsDatabaseContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching logs database containers using labels: %+v", logsDatabaseContainerSearchLabels)
	}
	return matchingLogsDatabaseContainers, nil
}
