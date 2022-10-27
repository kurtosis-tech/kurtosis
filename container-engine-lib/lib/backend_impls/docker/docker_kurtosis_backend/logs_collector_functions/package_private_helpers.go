package logs_collector_functions

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
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

	//This port number was used when whe run the logs collector without publishing the TCP port
	//now we need to check for this, in order to know if the logs collector does not bind
	oldLogsCollectorTCPPortNumber = 24224
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
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
	dockerManager *docker_manager.DockerManager,
) (*logs_collector.LogsCollector, error) {

	var (
		logsCollectorStatus container_status.ContainerStatus
		privateIpAddr net.IP
		publicIpAddr net.IP
		publicTcpPortSpec *port_spec.PortSpec
		publicHttpPortSpec *port_spec.PortSpec
		privateIpAddrStr string
		err error
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
		logsCollectorStatus = container_status.ContainerStatus_Running

		privateIpAddrStr, err = dockerManager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the private IP address of container '%v' in network '%v'", containerId, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn)
		}
		privateIpAddr = net.ParseIP(privateIpAddrStr)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
		}

		// TODO Remove this after 2022-12-27, when we're confident nobody will have the old port
		//checking for old port spec because old logs collectors container (older than engine v.0.50.3) do not publish the TCP port
		isPrivateTCPPortFromOldLogsCollectorContainers, err := isLogsCollectorPortSpecFromOlbLogsCollectorContainers(privateTcpPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred cheking if logs collector private TCP port spec '%+v' corresponds to and old version", privateTcpPortSpec)
		}

		if !isPrivateTCPPortFromOldLogsCollectorContainers {
			_, publicTcpPortSpec, err = shared_helpers.GetPublicPortBindingFromPrivatePortSpec(privateTcpPortSpec, allHostMachinePortBindings)
			if err != nil {
				return nil,
					stacktrace.Propagate(err, "The logs collector is running, but an error occurred getting the public port spec for the TCP private port spec '%+v'", privateTcpPortSpec)
			}
		}

		publicIpAddr, publicHttpPortSpec, err = shared_helpers.GetPublicPortBindingFromPrivatePortSpec(privateHttpPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The logs collector is running, but an error occurred getting the public port spec for the HTTP private port spec '%+v'", privateHttpPortSpec)
		}

	} else {
		logsCollectorStatus = container_status.ContainerStatus_Stopped
	}

	logsCollectorObj := logs_collector.NewLogsCollector(
		logsCollectorStatus,
		privateIpAddr,
		privateTcpPortSpec,
		privateHttpPortSpec,
		publicIpAddr,
		publicTcpPortSpec,
		publicHttpPortSpec,
	)

	return logsCollectorObj, nil
}

func isLogsCollectorPortSpecFromOlbLogsCollectorContainers(
	privatePortSpec *port_spec.PortSpec,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (bool, error) {
	// Convert port spec protocol -> Docker protocol string
	dockerPort, err := shared_helpers.GetDockerPortFromPortSpec(privatePortSpec)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred checking if this private TCP port spec '%+v' corresponds to an old version of the logs collector container", privatePortSpec)
	}

	_, found := allHostMachinePortBindings[dockerPort]
	if !found && privatePortSpec.GetNumber() == oldLogsCollectorTCPPortNumber {
		return true, nil
	}
	return false, nil
}

func getAllLogsCollectorContainers(ctx context.Context, dockerManager *docker_manager.DockerManager) ([]*types.Container, error) {
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

//If nothing is found returns nil
func getLogsCollectorObjectAndContainerId(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (
	resultMaybeLogsCollector *logs_collector.LogsCollector,
	resultMaybeContainerId string,
	resultErr error,
) {
	allLogsCollectorContainers, err := getAllLogsCollectorContainers(ctx, dockerManager)
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
		hostMachinePortBindings,
		dockerManager,
	)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred getting logs collector object using container ID '%v', labels '%+v', status '%v' and host machine port bindings '%+v'", logsCollectorContainer.GetId(), logsCollectorContainer.GetLabels(), logsCollectorContainer.GetStatus(), hostMachinePortBindings)
	}

	return logsCollectorObject, logsCollectorContainerID, nil
}
