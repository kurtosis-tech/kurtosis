package loki

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	maxWaitForLokiServiceAvailabilityRetries         = 10
	timeBetweenWaitForLokiServiceAvailabilityRetries = 1 * time.Second
	shouldFollowLogsWhenTheContainerWillBeRemoved    = false
)

type lokiLogsDatabaseContainer struct{}

func NewLokiLogDatabaseContainer() *lokiLogsDatabaseContainer {
	return &lokiLogsDatabaseContainer{}
}

func (lokiContainer *lokiLogsDatabaseContainer) CreateAndStart(
	ctx context.Context,
	httpPortId string,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (
	resultContainerId string,
	resultContainerLabels map[string]string,
	resultRemoveLogsDatabaseContainerFunc func(),
	resultErr error,
) {

	lokiContainerConfigProviderObj := createLokiContainerConfigProviderForKurtosis()

	privateHttpPortSpec, err := lokiContainerConfigProviderObj.GetPrivateHttpPortSpec()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs database container's private port spec")
	}

	logsDatabaseAttrs, err := objAttrsProvider.ForLogsDatabase(
		httpPortId,
		privateHttpPortSpec,
	)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(
			err,
			"An error occurred getting the logs database container attributes using the HTTP port spec '%+v'",
			privateHttpPortSpec,
		)
	}
	logsDbVolumeAttrs, err := objAttrsProvider.ForLogsDatabaseVolume()
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the logs database volume attributes")
	}

	containerLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsDatabaseAttrs.GetLabels() {
		containerLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	containerName := logsDatabaseAttrs.GetName().GetString()
	volumeName := logsDbVolumeAttrs.GetName().GetString()
	volumeLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsDbVolumeAttrs.GetLabels() {
		volumeLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}
	//This method will create the volume if it doesn't exist, or it will get it if it exists
	//From Docker docs: If you specify a volume name already in use on the current driver, Docker assumes you want to re-use the existing volume and does not return an error.
	//https://docs.docker.com/engine/reference/commandline/volume_create/
	if err := dockerManager.CreateVolume(ctx, volumeName, volumeLabelStrs); err != nil {
		return "", nil, nil,
		stacktrace.Propagate(
			err,
			"An error occurred creating logs database volume with name '%v' and labels '%+v'",
			volumeName,
			volumeLabelStrs,
		)
	}
	//We do not defer undo volume creation because the volume could already exist from previous executions
	//for this reason the logs database volume creation has to be idempotent, we ALWAYS want to create it if it doesn't exist, no matter what

	createAndStartArgs, err := lokiContainerConfigProviderObj.GetContainerArgs(containerName, containerLabelStrs, volumeName, targetNetworkId)
	if err != nil {
		return "", nil, nil,
			stacktrace.Propagate(
				err,
				"An error occurred getting the logs database container args with container name '%v', labels '%+v', volume name '%v' and network ID '%v",
				containerName,
				containerLabelStrs,
				volumeName,
				targetNetworkId,
			)
	}

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the logs database container with these args '%+v'", createAndStartArgs)
	}
	removeContainerFunc := func() {
		removeCtx := context.Background()
		containerLogsReadCloser, err := dockerManager.GetContainerLogs(removeCtx, containerId, shouldFollowLogsWhenTheContainerWillBeRemoved)
		if err != nil {
			logrus.Errorf(
				"Launching the logs databaes container with ID '%v' didn't complete successfully so we "+
					"tried to get the container's logs, but doing so throw the following error:\n%v",
				containerId,
				err)
		}
		defer containerLogsReadCloser.Close()

		if containerLogsReadCloser != nil {
			containerReadCloserBuffer := new(bytes.Buffer)
			if  _, err :=containerReadCloserBuffer.ReadFrom(containerLogsReadCloser); err != nil {
				logrus.Errorf(
					"Launching the logs database container with ID '%v' didn't complete successfully so we "+
						"tried to read the container's logs, but doing so throw the following error:\n%v",
					containerId,
					err)
			} else {
				containerLogsStr := containerReadCloserBuffer.String()
				containerLogsHeader := "\n--------------------- LOKI CONTAINER LOGS -----------------------\n"
				containerLogsFooter := "\n------------------- END LOKI CONTAINER LOGS --------------------"
				logrus.Infof("Could not start the logs database container; logs are below:%v%v%v", containerLogsHeader, containerLogsStr, containerLogsFooter)
			}
		}

		if err := dockerManager.RemoveContainer(removeCtx, containerId); err != nil {
			logrus.Errorf(
				"Launching the logs database server with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs database server with Docker container ID '%v'!!!!!!", containerId)
		}
	}
	shouldRemoveLogsDbContainer := true
	defer func() {
		if shouldRemoveLogsDbContainer {
			removeContainerFunc()
		}
	}()

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		ctx,
		dockerManager,
		containerId,
		privateHttpPortSpec,
		maxWaitForLokiServiceAvailabilityRetries,
		timeBetweenWaitForLokiServiceAvailabilityRetries,
	); err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the log database's HTTP port to become available")
	}

	shouldRemoveLogsDbContainer = false
	return containerId, containerLabelStrs, removeContainerFunc, nil
}
