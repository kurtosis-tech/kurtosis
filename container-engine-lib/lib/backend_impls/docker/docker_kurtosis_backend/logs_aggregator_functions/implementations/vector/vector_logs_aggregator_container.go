package vector

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type vectorLogsAggregatorContainer struct{}

func NewVectorLogsAggregatorContainer() *vectorLogsAggregatorContainer {
	return &vectorLogsAggregatorContainer{}
}

func (vectorContainer *vectorLogsAggregatorContainer) CreateAndStart(
	ctx context.Context,
	portNumber uint16,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (
	resultContainerId string,
	resultContainerLabels map[string]string,
	resultRemoveLogsAggregatorContainerFunc func(),
	resultErr error,
) {
	vectorContainerConfigProviderObj := createVectorContainerConfigProvider(portNumber)

	// TODO create the object attrs properly
	volumeName := "log-aggregator-volume"
	volumeLabelStrs := map[string]string{}
	containerName := "log-aggregator"
	containerLabelStrs := map[string]string{}

	// create volume
	if err := dockerManager.CreateVolume(ctx, volumeName, volumeLabelStrs); err != nil {
		return "", nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating logs collector volume with name '%v' and labels '%+v'",
			volumeName,
			volumeLabelStrs,
		)
	}

	// create container args
	createAndStartArgs, err := vectorContainerConfigProviderObj.GetContainerArgs(containerName, containerLabelStrs, volumeName, targetNetworkId)
	if err != nil {
		return "", nil, nil, err
	}

	// create container
	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the logs aggregator container with these args '%+v'", createAndStartArgs)
	}
	removeContainerFunc := func() {
		removeCtx := context.Background()

		if err := dockerManager.RemoveContainer(removeCtx, containerId); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator server with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator server with Docker container ID '%v'!!!!!!", containerId)
		}
	}
	shouldRemoveLogsAggregatorContainer := true
	defer func() {
		if shouldRemoveLogsAggregatorContainer {
			removeContainerFunc()
		}
	}()

	// TODO: wait for availability

	shouldRemoveLogsAggregatorContainer = false
	return containerId, containerLabelStrs, removeContainerFunc, nil
}
