package vector

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	maxCleanRetries                    = 2
	timeBetweenCleanRetries            = 200 * time.Millisecond
	cleaningSuccessStatusCode          = 0
	stopLogsAggregatorContainerTimeout = 10 * time.Second
)

type vectorLogsAggregatorContainer struct{}

func NewVectorLogsAggregatorContainer() *vectorLogsAggregatorContainer {
	return &vectorLogsAggregatorContainer{}
}

func (vector *vectorLogsAggregatorContainer) CreateAndStart(
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

	logsAggregatorConfigVolumeAttrs, err := objAttrsProvider.ForLogsAggregatorConfigVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator config volume attributes")
	}

	logsAggregatorDataVolumeAttrs, err := objAttrsProvider.ForLogsAggregatorDataVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs aggregator data volume attributes")
	}

	configVolumeName, err := createVolume(ctx, logsAggregatorConfigVolumeAttrs, dockerManager)
	if err != nil {
		return "", nil, nil, err
	}

	dataVolumeName, err := createVolume(ctx, logsAggregatorDataVolumeAttrs, dockerManager)
	if err != nil {
		return "", nil, nil, err
	}

	// Engine handles creating the volume, but we need to mount the aggregator can send logs to logs storage
	logsStorageAttrs, err := objAttrsProvider.ForLogsStorageVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs storage volume attributes.")
	}
	logsStorageVolNameStr := logsStorageAttrs.GetName().GetString()

	//We do not defer undo volume creation because the volume could already exist from previous executions
	//for this reason the logs collector volume creation has to be idempotent, we ALWAYS want to create it if it doesn't exist, no matter what

	if err := vectorConfigurationCreatorObj.CreateConfiguration(ctx, targetNetworkId, configVolumeName, dockerManager); err != nil {
		return "", nil, nil, stacktrace.Propagate(
			err,
			"An error occurred running the logs aggregator configuration creator in network ID '%v' and with volume name '%+v'",
			targetNetworkId,
			configVolumeName,
		)
	}

	containerArgs, err := vectorContainerConfigProviderObj.GetContainerArgs(containerName, containerLabelStrs, targetNetworkId, configVolumeName, dataVolumeName, logsStorageVolNameStr)
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

func (vector *vectorLogsAggregatorContainer) Clean(
	ctx context.Context,
	logsAggregator *types.Container,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) error {
	if err := dockerManager.StopContainer(ctx, logsAggregator.GetId(), stopLogsAggregatorContainerTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the logs aggregator container with ID %s", logsAggregator.GetId())
	}

	logsAggregatorDataDirVolumeAttrs, err := objAttrsProvider.ForLogsAggregatorDataVolume()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs aggregator data volume attributes")
	}

	if err := emptyVolume(ctx, logsAggregatorDataDirVolumeAttrs.GetName().GetString(), targetNetworkId, dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred emptying the logs aggregator data volume")
	}

	if err := dockerManager.StartContainer(ctx, logsAggregator.GetId()); err != nil {
		return stacktrace.Propagate(err, "An error occurred restarting the logs aggregator container with ID %s", logsAggregator.GetId())
	}

	return nil
}

func emptyVolume(ctx context.Context, volumeName string, targetNetworkId string, dockerManager *docker_manager.DockerManager) error {
	entrypointArgs := []string{
		shBinaryFilepath,
		shCmdFlag,
		fmt.Sprintf("sleep %v", sleepSeconds),
	}

	dirToRemoveAndEmpty := "/dir-to-remove"

	volumeMounts := map[string]string{
		volumeName: dirToRemoveAndEmpty,
	}

	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred generating a UUID for the cleaning container name")
	}

	containerName := fmt.Sprintf("%s-%s", "remove-dir-container", uuid)

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		"busybox",
		containerName,
		targetNetworkId,
	).WithEntrypointArgs(
		entrypointArgs,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the logs aggregator cleaning container with these args '%+v'", createAndStartArgs)
	}
	//The killing step has to be executed always in the success and also in the failed case
	defer func() {
		if err = dockerManager.RemoveContainer(context.Background(), containerId); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator cleaning container with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the container with ID '%v'!!!!!!", containerId)
		}
	}()

	commandStr := fmt.Sprintf(
		"rm -rf %v/*",
		dirToRemoveAndEmpty,
	)

	execCmd := []string{
		shBinaryFilepath,
		shCmdFlag,
		commandStr,
	}

	for i := uint(0); i < maxCleanRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunUserServiceExecCommands(ctx, containerId, "", execCmd, outputBuffer)
		if err == nil {
			if exitCode == cleaningSuccessStatusCode {
				logrus.Debugf("The logs aggregator data volume was successfully cleaned")
				return nil
			}

			logrus.Debugf("Logs aggregator cleaning container exited with non-zero status code %d; errors are below:\n%s", exitCode, outputBuffer.String())
		} else {
			logrus.Debugf(
				"Logs aggregator cleaning command '%v' experienced a Docker error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxCleanRetries-1 {
			time.Sleep(timeBetweenCleanRetries)
		}
	}

	return stacktrace.NewError(
		"The logs aggregator cleaning container didn't return success (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxCleanRetries,
		timeBetweenCleanRetries,
	)
}

func createVolume(ctx context.Context, provider object_attributes_provider.DockerObjectAttributes, dockerManager *docker_manager.DockerManager) (string, error) {
	volumeName := provider.GetName().GetString()
	volumeLabelStrs := map[string]string{}
	for labelKey, labelValue := range provider.GetLabels() {
		volumeLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	//This method will create the volume if it doesn't exist, or it will get it if it exists
	//From Docker docs: If you specify a volume name already in use on the current driver, Docker assumes you want to re-use the existing volume and does not return an error.
	//https://docs.docker.com/engine/reference/commandline/volume_create/
	if err := dockerManager.CreateVolume(ctx, volumeName, volumeLabelStrs); err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred creating logs aggregator volume with name '%v' and labels '%+v'",
			volumeName,
			volumeLabelStrs,
		)
	}

	return volumeName, nil
}
