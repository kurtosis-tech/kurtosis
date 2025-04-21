package logs_collector_functions

import (
	"context"
	"net"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/availability_checker"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultContainerStatusForNewLogsCollectorContainer = types.ContainerStatus_Running
	emptyAliasForLogsCollector                         = ""
)

var (
	autoAssignIpAddressToLogsCollector net.IP = nil
)

func CreateLogsCollectorForEnclave(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	logsCollectorTcpPortNumber uint16,
	logsCollectorHttpPortNumber uint16,
	logsCollectorContainer LogsCollectorContainer,
	logsAggregator *logs_aggregator.LogsAggregator,
	logsCollectorFilters []logs_collector.Filter,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*logs_collector.LogsCollector,
	error,
) {
	logrus.Debugf("Creating logs collector for enclave '%v'", enclaveUuid)
	preExistingLogsCollectorContainers, err := getLogsCollectorForTheGivenEnclave(ctx, enclaveUuid, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs collector containers for given enclave '%v'", enclaveUuid)
	}
	if len(preExistingLogsCollectorContainers) > 0 {
		return nil, stacktrace.NewError("Found existing logs collector container(s); cannot start a new one")
	}

	enclaveNetwork, err := shared_helpers.GetEnclaveNetworkByEnclaveUuid(ctx, enclaveUuid, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while retrieving the network id for the enclave '%v'", enclaveUuid)
	}

	if logsAggregator.GetMaybePrivateIpAddr() == nil {
		return nil, stacktrace.NewError("Expected the logs aggregator has private IP address but this is nil")
	}

	logsAggregatorHost := logsAggregator.GetMaybePrivateIpAddr().String()
	logsAggregatorPortNum := logsAggregator.GetListeningPortNum()

	containerId, containerLabels, hostMachinePortBindings, removeLogsCollectorContainerFunc, err := logsCollectorContainer.CreateAndStart(
		ctx,
		enclaveUuid,
		logsAggregatorHost,
		logsAggregatorPortNum,
		logsCollectorTcpPortNumber,
		logsCollectorHttpPortNumber,
		logsCollectorTcpPortId,
		logsCollectorHttpPortId,
		logsCollectorFilters,
		enclaveNetwork.GetId(),
		objAttrsProvider,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running the logs collector container with container ID '%v' with logs aggregator host '%v', logs aggregator port '%v', HTTP port number '%v', TCP port id '%v', and HTTP port id '%v' in Docker network with ID '%v'",
			containerId,
			logsAggregatorHost,
			logsAggregatorPortNum,
			logsCollectorHttpPortNumber,
			logsCollectorTcpPortId,
			logsCollectorHttpPortId,
			enclaveNetwork,
		)
	}
	shouldRemoveLogsCollectorContainer := true
	defer func() {
		if shouldRemoveLogsCollectorContainer {
			removeLogsCollectorContainerFunc()
		}
	}()

	if err = dockerManager.ConnectContainerToNetwork(ctx, enclaveNetwork.GetId(), containerId, autoAssignIpAddressToLogsCollector, emptyAliasForLogsCollector); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while connecting container '%v' to the enclave network '%v'", containerId, enclaveNetwork.GetId())
	}
	shouldDisconnectLogsCollectorFromEnclaveNetwork := true
	defer func() {
		if shouldDisconnectLogsCollectorFromEnclaveNetwork {
			err = dockerManager.DisconnectContainerFromNetwork(ctx, containerId, enclaveNetwork.GetId())
			if err != nil {
				logrus.Errorf("Tried disconnecting failing logs collector container with ID '%v' from the enclave network '%v' but failed with err:\n'%v'", containerId, enclaveNetwork.GetId(), err)
			}
		}
	}()

	logsCollectorObj, err := getLogsCollectorObjectFromContainerInfo(
		ctx,
		containerId,
		containerLabels,
		defaultContainerStatusForNewLogsCollectorContainer,
		enclaveNetwork,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs collector object using container ID '%v', labels '%+v', status '%v' and host machine port bindings '%+v'", containerId, containerLabels, defaultContainerStatusForNewLogsCollectorContainer, hostMachinePortBindings)
	}

	logrus.Debugf("Checking for logs collector availability in enclave '%v'...", enclaveUuid)

	logsCollectorAvailabilityEndpoint := logsCollectorContainer.GetHttpHealthCheckEndpoint()
	if err = availability_checker.WaitForAvailability(logsCollectorObj.GetBridgeNetworkIpAddress(), logsCollectorObj.GetPrivateHttpPort().GetNumber(), logsCollectorAvailabilityEndpoint); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while waiting for the log container to become available")
	}
	logrus.Debugf("...logs collector is available in enclave '%v'", enclaveUuid)
	logrus.Debugf("Logs collector successfully created with container ID '%v' for enclave '%v'", containerId, enclaveUuid)

	shouldDisconnectLogsCollectorFromEnclaveNetwork = false
	shouldRemoveLogsCollectorContainer = false
	return logsCollectorObj, nil
}
