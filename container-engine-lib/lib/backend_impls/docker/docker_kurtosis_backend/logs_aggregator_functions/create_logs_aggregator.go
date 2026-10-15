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
	logsAggregatorHttpPortNumber uint16,
	sinks logs_aggregator.Sinks,
	shouldEnablePersistentVolumeLogsCollection bool,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*logs_aggregator.LogsAggregator,
	func(),
	error,
) {
	logsAggregatorNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator network.")
	}
	targetNetworkId := logsAggregatorNetwork.GetId()

	existingLogsAggregatorContainer, found, err := getLogsAggregatorContainer(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator container.")
	}
	if found {
		reconfigureErr := logsAggregatorContainer.Reconfigure(
			ctx,
			existingLogsAggregatorContainer,
			defaultLogsListeningPortNum,
			sinks,
			logsAggregatorHttpPortNumber,
			targetNetworkId,
			shouldEnablePersistentVolumeLogsCollection,
			objAttrsProvider,
			dockerManager,
		)
		if reconfigureErr == nil {
			logsAggregatorObj, err := reuseLogsAggregator(ctx, existingLogsAggregatorContainer.GetId(), logsAggregatorContainer, targetNetworkId, dockerManager)
			if err != nil {
				return nil, nil, stacktrace.Propagate(err, "An error occurred reusing the existing logs aggregator.")
			}
			removeCtx := context.Background()
			containerId := existingLogsAggregatorContainer.GetId()
			removeLogsAggregatorContainerFunc := func() {
				if err := dockerManager.RemoveContainer(removeCtx, containerId); err != nil {
					logrus.Errorf(
						"Something failed while trying to remove the logs aggregator container with ID '%v'. Error was:\n%v",
						containerId,
						err)
					logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator server with Docker container ID '%v'!!!!!!", containerId)
				}
			}
			return logsAggregatorObj, removeLogsAggregatorContainerFunc, nil
		}
		logrus.Warnf("Recreating the logs aggregator because the existing one could not be reconfigured: %v", reconfigureErr)
		if err := DestroyLogsAggregator(ctx, dockerManager); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred removing the existing logs aggregator.")
		}
	}

	containerId, containerLabels, removeLogsAggregatorContainerFunc, err := logsAggregatorContainer.CreateAndStart(
		ctx,
		defaultLogsListeningPortNum,
		sinks,
		logsAggregatorHttpPortNumber,
		logsAggregatorHttpPortId,
		targetNetworkId,
		shouldEnablePersistentVolumeLogsCollection,
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
		containerLabels,
		defaultContainerStatusForNewLogsAggregatorContainer,
		dockerManager,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator object using container ID '%v', labels '%+v', status '%v'.", containerId, containerLabels, defaultContainerStatusForNewLogsAggregatorContainer)
	}

	logrus.Debugf("Checking for logs aggregator availability...")

	logsAggregatorAvailabilityEndpoint := logsAggregatorContainer.GetHttpHealthCheckEndpoint()
	if err = waitForLogsAggregatorAvailability(ctx, logsAggregator.GetMaybePrivateIpAddr(), logsAggregatorAvailabilityEndpoint, logsAggregator.GetPrivateHttpPort().GetNumber(), targetNetworkId, dockerManager); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while waiting for the log aggregator to become available")
	}

	logrus.Debugf("Logs aggregator successfully created and available")

	removeLogsAggregatorFunc := func() {
		removeLogsAggregatorContainerFunc()
	}

	shouldRemoveLogsAggregatorContainer = false
	return logsAggregator, removeLogsAggregatorFunc, nil
}

func reuseLogsAggregator(
	ctx context.Context,
	containerId string,
	logsAggregatorContainer LogsAggregatorContainer,
	targetNetworkId string,
	dockerManager *docker_manager.DockerManager,
) (*logs_aggregator.LogsAggregator, error) {
	logsAggregatorObj, _, err := getLogsAggregatorObjectAndContainerId(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator with container ID '%v'.", containerId)
	}
	if logsAggregatorObj == nil {
		return nil, stacktrace.NewError("The logs aggregator with container ID '%v' disappeared while reconfiguring it", containerId)
	}
	if err := waitForLogsAggregatorAvailability(ctx, logsAggregatorObj.GetMaybePrivateIpAddr(), logsAggregatorContainer.GetHttpHealthCheckEndpoint(), logsAggregatorObj.GetPrivateHttpPort().GetNumber(), targetNetworkId, dockerManager); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the reconfigured logs aggregator to become available")
	}
	return logsAggregatorObj, nil
}
