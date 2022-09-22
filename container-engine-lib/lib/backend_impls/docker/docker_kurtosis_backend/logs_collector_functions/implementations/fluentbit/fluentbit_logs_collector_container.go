package fluentbit

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	shouldFollowLogsWhenTheContainerWillBeRemoved = false
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
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (
	resultContainerId  string,
	resultContainerLabels map[string]string,
	resultRemoveLogsCollectorContainerFunc func(),
	resultErr error,
) {

	logsCollectorConfigurationCreator := createFluentbitConfigurationCreatorForKurtosis(logsDatabaseHost, logsDatabasePort, httpPortNumber)
	logsCollectorContainerConfigProvider := createFluentbitContainerConfigProviderForKurtosis(logsDatabaseHost, logsDatabasePort, httpPortNumber)

	privateTcpPortSpec, err := logsCollectorContainerConfigProvider.GetPrivateTcpPortSpec()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs collector private TCP port spec")
	}

	privateHttpPortSpec, err := logsCollectorContainerConfigProvider.GetPrivateHttpPortSpec()
	if err != nil {
		return "",  nil, nil,stacktrace.Propagate(err, "An error occurred getting the logs collector private HTTP port spec")
	}

	logsCollectorAttrs, err := objAttrsProvider.ForLogsCollector(
		logsCollectorTcpPortId,
		privateTcpPortSpec,
		logsCollectorHttpPortId,
		privateHttpPortSpec,
	)
	if err != nil {
		return "",  nil, nil,stacktrace.Propagate(
			err,
			"An error occurred getting the logs collector container attributes with TCP port spec '%+v' and HTTP port spec '%+v'",
			privateTcpPortSpec,
			privateHttpPortSpec,
		)
	}

	containerName := logsCollectorAttrs.GetName().GetString()
	containerLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsCollectorAttrs.GetLabels() {
		containerLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	logsCollectorVolumeAttrs, err := objAttrsProvider.ForLogsCollectorVolume()
	if err != nil {
		return "",  nil, nil,stacktrace.Propagate(err, "An error occurred getting the logs collector volume attributes")
	}

	volumeName := logsCollectorVolumeAttrs.GetName().GetString()
	volumeLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsCollectorVolumeAttrs.GetLabels() {
		volumeLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	//This method will create the volume if it doesn't exist, or it will get it if it exists
	//From Docker docs: If you specify a volume name already in use on the current driver, Docker assumes you want to re-use the existing volume and does not return an error.
	//https://docs.docker.com/engine/reference/commandline/volume_create/
	if err := dockerManager.CreateVolume(ctx, volumeName, volumeLabelStrs); err != nil {
		return "",  nil, nil,stacktrace.Propagate(
			err,
			"An error occurred creating logs collector volume with name '%v' and labels '%+v'",
			volumeName,
			volumeLabelStrs,
		)
	}
	//We do not defer undo volume creation because the volume could already exist from previous executions
	//for this reason the logs collector volume creation has to be idempotent, we ALWAYS want to create it if it doesn't exist, no matter what

	if err := logsCollectorConfigurationCreator.CreateConfiguration(context.Background(), targetNetworkId, volumeName, dockerManager); err != nil {
		return "",  nil, nil,stacktrace.Propagate(
			err,
			"An error occurred running the logs collector configuration creator in network ID '%v' and with volume name '%+v'",
			targetNetworkId,
			volumeName,
		)
	}

	createAndStartArgs, err := logsCollectorContainerConfigProvider.GetContainerArgs(containerName, containerLabelStrs, volumeName, targetNetworkId)
	if err != nil {
		return "",  nil, nil,
			stacktrace.Propagate(
				err,
				"An error occurred getting the logs-collector-container-args with container name '%v', labels '%+v', and network ID '%v",
				containerName,
				containerLabelStrs,
				targetNetworkId,
			)
	}

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "",  nil, nil,stacktrace.Propagate(err, "An error occurred starting the logs collector container with these args '%+v'", createAndStartArgs)
	}
	removeContainerFunc := func() {
		removeCtx := context.Background()
		//First try to print the logs before removing it
		containerLogsReadCloser, err := dockerManager.GetContainerLogs(removeCtx, containerId, shouldFollowLogsWhenTheContainerWillBeRemoved)
		if err != nil {
			logrus.Errorf(
				"Launching the logs collector container with ID '%v' didn't complete successfully so we "+
				"tried to get the container's logs, but doing so throw the following error:\n%v",
				containerId,
				err)
		}
		defer containerLogsReadCloser.Close()

		if containerLogsReadCloser != nil {
			containerReadCloserBuffer := new(bytes.Buffer)
			if  _, err := containerReadCloserBuffer.ReadFrom(containerLogsReadCloser); err != nil {
				logrus.Errorf(
					"Launching the logs collector container with ID '%v' didn't complete successfully so we "+
						"tried to read the container's logs, but doing so throw the following error:\n%v",
					containerId,
					err)
			} else {
				containerLogsStr := containerReadCloserBuffer.String()
				containerLogsHeader := "\n--------------------- FLUENTBIT CONTAINER LOGS -----------------------\n"
				containerLogsFooter := "\n------------------- END FLUENTBIT CONTAINER LOGS --------------------"
				logrus.Infof("Could not start the logs collector container; logs are below:%v%v%v", containerLogsHeader, containerLogsStr, containerLogsFooter)
			}
		}

		if err := dockerManager.RemoveContainer(removeCtx, containerId); err != nil {
			logrus.Errorf(
				"Launching the logs collector container with ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector server with Docker container ID '%v'!!!!!!",  containerId)
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
		return "",  nil, nil,stacktrace.Propagate(err, "An error occurred waiting for the log collector's to become available")
	}

	shouldRemoveLogsCollectorContainer = false
	return containerId, containerLabelStrs, removeContainerFunc, nil
}
