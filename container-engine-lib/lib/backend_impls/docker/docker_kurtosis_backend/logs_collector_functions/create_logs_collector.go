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

func CreateLogsCollector(
	ctx context.Context,
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

	logsDatabaseHost := logsDatabase.GetPrivateIpAddr().String()
	logsDatabasePort := logsDatabase.GetPrivateHttpPort().GetNumber()

	containerId, containerLabels, removeLogsCollectorContainerFunc, err := logsCollectorContainer.CreateAndStart(
		ctx,
		logsDatabaseHost,
		logsDatabasePort,
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
			"An error occurred running the logs collector container with logs database host '%v', logs database port '%v', http port '%v', tcp port id '%v', and http port id '%v' in Docker network with ID '%v'",
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
		types.ContainerStatus_Running,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs collector object using container ID '%v', labels '%+v' and the status '%v'", containerId, containerLabels, types.ContainerStatus_Running)
	}

	shouldRemoveLogsCollectorContainerFunc = false
	return logsCollectorObj, nil
}
