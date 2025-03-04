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
	logsListeningPortNumber uint16,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (string, map[string]string, func(), error) {
	vectorContainerConfigProviderObj := createVectorContainerConfigProvider(logsListeningPortNumber)

	logsAggregatorAttrs, err := objAttrsProvider.ForLogsAggregator()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator container attributes.")
	}
	containerName := logsAggregatorAttrs.GetName().GetString()
	containerLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsAggregatorAttrs.GetLabels() {
		containerLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	// Engine handles creating the volume, but we need to mount the aggregator can send logs to logs storage
	logsStorageAttrs, err := objAttrsProvider.ForLogsStorageVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs storage volume attributes.")
	}
	logsStorageVolNameStr := logsStorageAttrs.GetName().GetString()

	createAndStartArgs, err := vectorContainerConfigProviderObj.GetContainerArgs(containerName, containerLabelStrs, targetNetworkId, logsStorageVolNameStr)
	if err != nil {
		return "", nil, nil, err
	}

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

	shouldRemoveLogsAggregatorContainer = false
	return containerId, containerLabelStrs, removeContainerFunc, nil
}

func (vector *vectorLogsAggregatorContainer) GetLogsBaseDirPath() string {
	return logsStorageDirpath
}
