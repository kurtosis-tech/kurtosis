package vector

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
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
	sinks logs_aggregator.Sinks,
	httpPortNumber uint16,
	logsAggregatorHttpPortId string,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (string, map[string]string, func(), error) {
	vectorConfigurationCreatorObj := createVectorConfigurationCreatorForKurtosis(logsListeningPortNumber, httpPortNumber, sinks)
	vectorContainerConfigProviderObj := createVectorContainerConfigProvider(httpPortNumber)

	// Start vector

	privateHttpPortSpec, err := vectorContainerConfigProviderObj.GetPrivateHttpPortSpec()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs collector private HTTP port spec")
	}

	logsAggregatorAttrs, err := objAttrsProvider.ForLogsAggregator(logsAggregatorHttpPortId, privateHttpPortSpec)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator container attributes.")
	}
	containerName := logsAggregatorAttrs.GetName().GetString()
	containerLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsAggregatorAttrs.GetLabels() {
		containerLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	logsAggregatorVolumeAttrs, err := objAttrsProvider.ForLogsAggregatorVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator volume attributes")
	}

	volumeName := logsAggregatorVolumeAttrs.GetName().GetString()
	volumeLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsAggregatorVolumeAttrs.GetLabels() {
		volumeLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	//This method will create the volume if it doesn't exist, or it will get it if it exists
	//From Docker docs: If you specify a volume name already in use on the current driver, Docker assumes you want to re-use the existing volume and does not return an error.
	//https://docs.docker.com/engine/reference/commandline/volume_create/
	if err := dockerManager.CreateVolume(ctx, volumeName, volumeLabelStrs); err != nil {
		return "", nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating logs aggregator volume with name '%v' and labels '%+v'",
			volumeName,
			volumeLabelStrs,
		)
	}

	// Engine handles creating the volume, but we need to mount the aggregator can send logs to logs storage
	logsStorageAttrs, err := objAttrsProvider.ForLogsStorageVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs storage volume attributes.")
	}
	logsStorageVolNameStr := logsStorageAttrs.GetName().GetString()

	//We do not defer undo volume creation because the volume could already exist from previous executions
	//for this reason the logs collector volume creation has to be idempotent, we ALWAYS want to create it if it doesn't exist, no matter what

	if err := vectorConfigurationCreatorObj.CreateConfiguration(ctx, targetNetworkId, volumeName, dockerManager); err != nil {
		return "", nil, nil, stacktrace.Propagate(
			err,
			"An error occurred running the logs aggregator configuration creator in network ID '%v' and with volume name '%+v'",
			targetNetworkId,
			volumeName,
		)
	}

	containerArgs, err := vectorContainerConfigProviderObj.GetContainerArgs(containerName, containerLabelStrs, targetNetworkId, volumeName, logsStorageVolNameStr)
	if err != nil {
		return "", nil, nil, err
	}

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, containerArgs)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the logs aggregator container with these args '%+v'", containerArgs)
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

func (vector *vectorLogsAggregatorContainer) GetHttpHealthCheckEndpoint() string {
	return healthCheckEndpointPath
}
