package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultContainerStatusForNewLogsAggregatorContainer = types.ContainerStatus_Running
)

func CreateLogsAggregator(
	ctx context.Context,
	portNumber uint16,
	logsAggregatorContainer LogsAggregatorContainer,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*logs_aggregator.LogsAggregator,
	error,
) {
	preExistingLogsAggregatorContainers, err := getAllLogsAggregatorContainers(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting all logs aggregator containers")
	}
	if len(preExistingLogsAggregatorContainers) > 0 {
		return nil, stacktrace.NewError("Found existing logs database aggregator(s); cannot start a new one")
	}

	logsAggregatorNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator network")
	}
	targetNetworkId := logsAggregatorNetwork.GetId()

	containerId, containerLabels, removeLogsAggregatorContainerFunc, err := logsAggregatorContainer.CreateAndStart(
		ctx,
		portNumber,
		targetNetworkId,
		objAttrsProvider,
		dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the logs aggregator container with port number '%v' in Docker network with ID '%v'",
			portNumber,
			targetNetworkId,
		)
	}
	shouldRemoveLogsAggregatorContainer := true
	defer func() {
		if shouldRemoveLogsAggregatorContainer {
			removeLogsAggregatorContainerFunc()
		}
	}()

	logsAggregator, err := getLogsAggregatorObjectFromContainerInfo(
		ctx,
		containerId,
		defaultContainerStatusForNewLogsAggregatorContainer,
		dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator object using container ID '%v', labels '%+v', status '%v'.", containerId, containerLabels, defaultContainerStatusForNewLogsAggregatorContainer)
	}

	shouldRemoveLogsAggregatorContainer = false
	return logsAggregator, nil
}
