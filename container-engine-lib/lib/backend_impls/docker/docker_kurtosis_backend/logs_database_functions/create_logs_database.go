package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultContainerStatusForNewLogsDatabaseContainer = types.ContainerStatus_Running
)

func CreateLogsDatabase(
	ctx context.Context,
	logsDatabaseContainer LogsDatabaseContainer,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*logs_database.LogsDatabase,
	error,
){

	preExistingLogsDatabaseContainers, err := getAllLogsDatabaseContainers(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting all logs database containers")
	}
	if len(preExistingLogsDatabaseContainers) > 0 {
		return nil, stacktrace.NewError("Found existing logs database container(s); cannot start a new one")
	}

	logsDatabaseNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database network")
	}
	targetNetworkId := logsDatabaseNetwork.GetId()

	containerId, containerLabels, removeLogsDatabaseContainerFunc, err := logsDatabaseContainer.CreateAndStart(
		ctx,
		logsDatabaseHttpPortId,
		targetNetworkId,
		objAttrsProvider,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the logs database container with http port id '%v' in Docker network with ID '%v'",
			logsDatabaseHttpPortId,
			targetNetworkId,
		)
	}
	shouldRemoveLogsDatabaseContainer := true
	defer func() {
		if shouldRemoveLogsDatabaseContainer {
			removeLogsDatabaseContainerFunc()
		}
	}()

	logsDatabaseObject, err := getLogsDatabaseObjectFromContainerInfo(
		ctx,
		containerId,
		containerLabels,
		defaultContainerStatusForNewLogsDatabaseContainer,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err, "An error occurred getting logs database object using container ID '%v', labels '%+v' and the status '%v'", containerId, containerLabels, defaultContainerStatusForNewLogsDatabaseContainer)
	}

	shouldRemoveLogsDatabaseContainer = false
	return logsDatabaseObject, nil
}
