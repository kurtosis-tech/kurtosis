package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultContainerStatusForNewLogsAggregatorContainer = types.ContainerStatus_Running
)

// Create logs aggregator idempotently, if existing logs aggregator is found, then it is returned
func CreateLogsAggregator(
	ctx context.Context,
	logsAggregatorContainer LogsAggregatorContainer,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*logs_aggregator.LogsAggregator,
	func(),
	error,
) {
	_, found, err := getLogsAggregatorContainer(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator container.")
	}
	if found {
		logrus.Warnf("Found existing logs aggregator; cannot start a new one.")
		logsAggregatorObj, _, err := getLogsAggregatorObjectAndContainerId(ctx, dockerManager)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting existing logs aggregator.")
		}
		return logsAggregatorObj, nil, nil
	}

	logsAggregatorNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator network.")
	}
	targetNetworkId := logsAggregatorNetwork.GetId()

	containerId, containerLabels, removeLogsAggregatorContainerFunc, err := logsAggregatorContainer.CreateAndStart(
		ctx,
		defaultLogsListeningPortNum,
		targetNetworkId,
		objAttrsProvider,
		dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
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
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator object using container ID '%v', labels '%+v', status '%v'.", containerId, containerLabels, defaultContainerStatusForNewLogsAggregatorContainer)
	}

	removeLogsAggregatorFunc := func() {
		removeLogsAggregatorContainerFunc()
	}

	shouldRemoveLogsAggregatorContainer = false
	return logsAggregator, removeLogsAggregatorFunc, nil
}
