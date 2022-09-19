package fluentbit

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type fluentbitLogsCollectorContainer struct {}

func NewFluentbitLogsCollectorContainer() *fluentbitLogsCollectorContainer {
	return &fluentbitLogsCollectorContainer{}
}

func (fluentbitContainer *fluentbitLogsCollectorContainer) CreateAndStart(
	ctx context.Context,
	logsDatabaseHost string,
	logsDatabasePort uint16,
	httpPortNumber uint16,
	logsCollectorTcpPortId string,
	logsCollectorHttpPortId string,
	engineGuid engine.EngineGUID,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (func(), error) {

	logsCollectorConfigurationCreator := createFluentbitConfigurationCreatorForKurtosis(logsDatabaseHost, logsDatabasePort, httpPortNumber)
	logsCollectorContainerConfigProvider := createFluentbitContainerConfigProviderForKurtosis(logsDatabaseHost, logsDatabasePort, httpPortNumber)

	privateTcpPortSpec, err := logsCollectorContainerConfigProvider.GetPrivateTcpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector private TCP port spec")
	}

	privateHttpPortSpec, err := logsCollectorContainerConfigProvider.GetPrivateHttpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector private HTTP port spec")
	}

	logsCollectorAttrs, err := objAttrsProvider.ForLogsCollector(
		logsCollectorTcpPortId,
		privateTcpPortSpec,
		logsCollectorHttpPortId,
		privateHttpPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the logs collector container attributes using GUID '%v' with TCP port spec '%+v' and HTTP port spec '%+v'",
			engineGuid,
			privateTcpPortSpec,
			privateHttpPortSpec,
		)
	}

	containerName := logsCollectorAttrs.GetName().GetString()
	labelStrs := map[string]string{}
	for labelKey, labelValue := range logsCollectorAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	logsCollectorVolumeAttrs, err := objAttrsProvider.ForLogsCollectorVolume()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector volume attributes for engine with GUID %v", engineGuid)
	}

	volumeName := logsCollectorVolumeAttrs.GetName().GetString()
	volumeLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsCollectorVolumeAttrs.GetLabels() {
		volumeLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}
	if err := dockerManager.CreateVolume(ctx, volumeName, volumeLabelStrs); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating logs collector volume with name '%v' and labels '%+v'",
			volumeName,
			volumeLabelStrs,
		)
	}
	removeVolumeFunc := func() {
		if err := dockerManager.RemoveVolume(ctx, volumeName); err != nil {
			logrus.Errorf(
				"Launching the logs collector server for the engine with GUID '%v' didn't complete successfully so we "+
					"tried to remove the associated volume '%v' we started, but doing so exited with an error:\n%v",
				engineGuid,
				volumeName,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector volume '%v'!!!!!!", volumeName)
		}
	}
	shouldRemoveLogsCollectorVolume := true
	defer func() {
		if shouldRemoveLogsCollectorVolume {
			removeVolumeFunc()
		}
	}()

	if err := logsCollectorConfigurationCreator.CreateConfiguration(context.Background(), targetNetworkId, volumeName, dockerManager); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running the logs collector configuration creator in network ID '%v' and with volume name '%+v'",
			targetNetworkId,
			volumeName,
		)
	}

	createAndStartArgs, err := logsCollectorContainerConfigProvider.GetContainerArgs(containerName, labelStrs, volumeName, targetNetworkId)
	if err != nil {
		return nil,
			stacktrace.Propagate(
				err,
				"An error occurred getting the logs-collector-container-args with container name '%v', labels '%+v', and network ID '%v",
				containerName,
				labelStrs,
				targetNetworkId,
			)
	}

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the logs collector container with these args '%+v'", createAndStartArgs)
	}
	removeContainerFunc := func() {
		removeCtx := context.Background()
		if err := dockerManager.RemoveContainer(removeCtx, containerId); err != nil {
			logrus.Errorf(
				"Launching the logs collector server for engine with GUID '%v' and container ID '%v' didn't complete successfully so we "+
					"tried to kill the container we started, but doing so exited with an error:\n%v",
				engineGuid,
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually stop the logs collector server for engine with GUID '%v' and Docker container ID '%v'!!!!!!", engineGuid, containerId)
		}
	}
	shouldRemoveLogsCollectorContainer := true
	defer func() {
		if shouldRemoveLogsCollectorContainer {
			removeContainerFunc()
		}
	}()

	logsCollectorAvailabilityChecker := newFluentbitAvailabilityChecker(privateHttpPortSpec.GetNumber())

	if err := logsCollectorAvailabilityChecker.WaitForAvailability(); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for logs collector container '%v' to become available",
			containerId,
		)
	}

	removeContainerAndVolumeFunc := func() {
		removeContainerFunc()
		removeVolumeFunc()
	}

	shouldRemoveLogsCollectorContainer = false
	shouldRemoveLogsCollectorVolume = false
	return removeContainerAndVolumeFunc, nil
}