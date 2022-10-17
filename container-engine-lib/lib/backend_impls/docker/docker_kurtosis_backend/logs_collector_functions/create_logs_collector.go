package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultContainerStatusForNewLogsCollectorContainer = types.ContainerStatus_Running
)

func CreateLogsCollector(
	ctx context.Context,
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

	preExistingLogsCollectorContainers, err := getAllLogsCollectorContainers(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting all logs collector containers")
	}
	if len(preExistingLogsCollectorContainers) > 0 {
		return nil, stacktrace.NewError("Found existing logs collector container(s); cannot start a new one")
	}

	logsCollectorNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector network")
	}
	targetNetworkId := logsCollectorNetwork.GetId()

	if logsDatabase.GetMaybePrivateIpAddr() == nil {
		return nil, stacktrace.NewError("Expected the logs database has private IP address but this is nil")
	}

	logsDatabaseHost := logsDatabase.GetMaybePrivateIpAddr().String()
	logsDatabasePort := logsDatabase.GetPrivateHttpPort().GetNumber()

	containerId, containerLabels, hostMachinePortBindings, removeLogsCollectorContainerFunc, err := logsCollectorContainer.CreateAndStart(
		ctx,
		logsDatabaseHost,
		logsDatabasePort,
		logsCollectorTcpPortNumber,
		logsCollectorHttpPortNumber,
		logsCollectorTcpPortId,
		logsCollectorHttpPortId,
		targetNetworkId,
		objAttrsProvider,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running the logs collector container with logs database host '%v', logs database port '%v', HTTP port number '%v', TCP port id '%v', and HTTP port id '%v' in Docker network with ID '%v'",
			logsDatabaseHost,
			logsDatabasePort,
			logsCollectorHttpPortNumber,
			logsCollectorTcpPortId,
			logsCollectorHttpPortId,
			targetNetworkId,
		)
	}
	shouldRemoveLogsCollectorContainerFunc := true
	defer func() {
		if shouldRemoveLogsCollectorContainerFunc {
			removeLogsCollectorContainerFunc()
		}
	}()

	logsCollectorObj, err := getLogsCollectorObjectFromContainerInfo(
		ctx,
		containerId,
		containerLabels,
		defaultContainerStatusForNewLogsCollectorContainer,
		hostMachinePortBindings,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs collector object using container ID '%v', labels '%+v', status '%v' and host machine port bindings '%+v'", containerId, containerLabels, defaultContainerStatusForNewLogsCollectorContainer, hostMachinePortBindings)
	}

	shouldRemoveLogsCollectorContainerFunc = false
	return logsCollectorObj, nil
}
