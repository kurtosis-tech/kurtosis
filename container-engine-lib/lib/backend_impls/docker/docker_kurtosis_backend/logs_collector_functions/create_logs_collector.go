package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_collector_functions/implementations/fluentbit"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
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
	logsDatabase *logs_database.LogsDatabase,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*logs_collector.LogsCollector,
	error,
) {

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

	if logsDatabase.GetMaybePrivateIpAddr() == nil {
		return nil, stacktrace.NewError("Expected the logs database has private IP address but this is nil")
	}

	logsDatabaseHost := logsDatabase.GetMaybePrivateIpAddr().String()
	logsDatabasePort := logsDatabase.GetPrivateHttpPort().GetNumber()

	containerId, containerLabels, hostMachinePortBindings, removeLogsCollectorContainerFunc, err := logsCollectorContainer.CreateAndStart(
		ctx,
		enclaveUuid,
		logsDatabaseHost,
		logsDatabasePort,
		logsCollectorTcpPortNumber,
		logsCollectorHttpPortNumber,
		logsCollectorTcpPortId,
		logsCollectorHttpPortId,
		enclaveNetwork.GetId(),
		objAttrsProvider,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running the logs collector container with container ID '%v' with logs database host '%v', logs database port '%v', HTTP port number '%v', TCP port id '%v', and HTTP port id '%v' in Docker network with ID '%v'",
			containerId,
			logsDatabaseHost,
			logsDatabasePort,
			logsCollectorHttpPortNumber,
			logsCollectorTcpPortId,
			logsCollectorHttpPortId,
			enclaveNetwork,
		)
	}
	shouldRemoveLogsCollectorContainerFunc := true
	defer func() {
		if shouldRemoveLogsCollectorContainerFunc {
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

	logsCollectorAvailabilityChecker := fluentbit.NewFluentbitAvailabilityChecker(logsCollectorObj.GetBridgeNetworkIpAddress(), logsCollectorObj.GetPrivateHttpPort().GetNumber())

	if err = logsCollectorAvailabilityChecker.WaitForAvailability(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while waiting for the log container to become available")
	}

	shouldDisconnectLogsCollectorFromEnclaveNetwork = false
	shouldRemoveLogsCollectorContainerFunc = false
	return logsCollectorObj, nil
}
