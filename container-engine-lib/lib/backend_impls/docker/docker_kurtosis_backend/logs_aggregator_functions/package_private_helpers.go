package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

func getLogsAggregatorObjectFromContainerInfo(
	ctx context.Context,
	containerId string,
	containerStatus types.ContainerStatus,
	dockerManager *docker_manager.DockerManager,
) (*logs_aggregator.LogsAggregator, error) {

	privateIpAddr := net.IP{}

	isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for logs aggregator container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var logsAggregatorStatus container_status.ContainerStatus
	if isContainerRunning {
		logsAggregatorStatus = container_status.ContainerStatus_Running

		privateIpAddrStr, err := dockerManager.GetContainerIP(ctx, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn, containerId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the private IP address of container '%v' in network '%v'", containerId, consts.NameOfNetworkToStartEngineAndLogServiceContainersIn)
		}
		privateIpAddr = net.ParseIP(privateIpAddrStr)
		if privateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
		}
	} else {
		logsAggregatorStatus = container_status.ContainerStatus_Stopped
	}

	logsAggregatorObj := logs_aggregator.NewLogsAggregator(
		logsAggregatorStatus,
		privateIpAddr,
	)

	return logsAggregatorObj, nil
}
