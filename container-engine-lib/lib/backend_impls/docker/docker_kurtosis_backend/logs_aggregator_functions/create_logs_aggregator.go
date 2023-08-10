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
	logsAggregatorContainer LogsAggregatorContainer,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*logs_aggregator.LogsAggregator,
	error,
) {
	preExistingLogsAggregatorContainer, err := getLogsAggregatorContainer(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator container")
	}
	if preExistingLogsAggregatorContainer != nil {
		return nil, stacktrace.NewError("Found existing logs aggregator; cannot start a new one.")
	}

	logsAggregatorNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator network.")
	}
	targetNetworkId := logsAggregatorNetwork.GetId()

	containerId, containerLabels, removeLogsAggregatorContainerFunc, err := logsAggregatorContainer.CreateAndStart(
		ctx,
		defaultLogsListeningPortNum,
		targetNetworkId,
		objAttrsProvider,
		dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the logs aggregator container in Docker network with ID '%v'",
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
