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

	// check if aggregator already exists
	// if so don't start another one

	// get the network this aggregator should connect to
	logsAggregatorNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database network")
	}
	targetNetworkId := logsAggregatorNetwork.GetId()

	// start the aggregator container
	containerId, _, removeLogsAggregatorContainerFunc, err := logsAggregatorContainer.CreateAndStart(
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

	// get the resulting object from container info
	logsAggregator, err := getLogsAggregatorObjectFromContainerInfo(
		ctx,
		containerId,
		defaultContainerStatusForNewLogsAggregatorContainer,
		dockerManager)

	// return the object
	shouldRemoveLogsAggregatorContainer = false
	return logsAggregator, nil
}
