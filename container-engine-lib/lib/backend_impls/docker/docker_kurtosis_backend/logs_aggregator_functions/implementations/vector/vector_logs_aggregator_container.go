package vector

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	validationFailedExitCode = 78
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
	vectorContainerConfigProviderObj := createVectorContainerConfigProvider(logsListeningPortNumber, httpPortNumber, sinks)

	logsAggregatorInitAttrs, err := objAttrsProvider.ForLogsAggregatorInit()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator init container attributes.")
	}

	initContainerName := logsAggregatorInitAttrs.GetName().GetString()
	initContainerLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsAggregatorInitAttrs.GetLabels() {
		initContainerLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	initArgs, err := vectorContainerConfigProviderObj.GetInitContainerArgs(initContainerName, initContainerLabelStrs, targetNetworkId)
	if err != nil {
		return "", nil, nil, err
	}

	initContainerId, _, err := dockerManager.CreateAndStartContainer(ctx, initArgs)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the logs aggregator container with these args '%+v'", initArgs)
	}

	defer func() {
		removeCtx := context.Background()
		if err := dockerManager.RemoveContainer(removeCtx, initContainerId); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator init with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				initContainerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs aggregator init container with Docker container ID '%v'!!!!!!", initContainerId)
		}
	}()

	exitCode, err := dockerManager.WaitForExit(ctx, initContainerId)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred validating logs aggregator configurations")
	}

	// Vector returns a specific exit code if the validation of configurations failed
	// https://vector.dev/docs/administration/validating/#how-validation-works
	if exitCode == validationFailedExitCode {
		errorStr := dockerManager.GetFormattedFailedContainerLogsOrErrorString(ctx, initContainerId)
		return "", nil, nil, stacktrace.NewError("The logs aggregator component failed validation; logs are below:%v", errorStr)
	}

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

	// Engine handles creating the volume, but we need to mount the aggregator can send logs to logs storage
	logsStorageAttrs, err := objAttrsProvider.ForLogsStorageVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs storage volume attributes.")
	}
	logsStorageVolNameStr := logsStorageAttrs.GetName().GetString()

	containerArgs, err := vectorContainerConfigProviderObj.GetContainerArgs(containerName, containerLabelStrs, targetNetworkId, logsStorageVolNameStr)
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

func (vectorContainer *vectorLogsAggregatorContainer) GetHttpHealthCheckEndpoint() string {
	return healthCheckEndpointPath
}
