package loki

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
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
	logsDatabaseHttpPortId string,
	engineGuid engine.EngineGUID,
	targetNetworkId string,
	targetNetworkName string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (
	resultLogsDatabasePrivateHost string,
	resultLogsDatabasePrivatePort uint16,
	resultRemoveLogsDatabaseContainerFunc func(),
	resultErr error,
) {

	lokiContainerConfigProvider := createLokiContainerConfigProviderForKurtosis()

	privateHttpPortSpec, err := lokiContainerConfigProvider.GetPrivateHttpPortSpec()
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred getting the logs database container's private port spec")
	}

	logsDatabaseAttrs, err := objAttrsProvider.ForLogsDatabase(
		logsDatabaseHttpPortId,
		privateHttpPortSpec,
	)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(
			err,
			"An error occurred getting the logs database container attributes using GUID '%v' and the HTTP port spec '%+v'",
			engineGuid,
			privateHttpPortSpec,
		)
	}
	logsDbVolumeAttrs, err := objAttrsProvider.ForLogsDatabaseVolume()
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred getting the logs database volume attributes")
	}

	labelStrs := map[string]string{}
	for labelKey, labelValue := range logsDatabaseAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
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
		return "", 0, nil, stacktrace.Propagate(
			err,
			"An error occurred creating logs database volume with name '%v' and labels '%+v'",
			volumeName,
			volumeLabelStrs,
		)
	}
	//We do not defer undo volume creation because the volume could already exist from previous executions
	//for this reason the logs database volume creation has to be idempotent, we ALWAYS want to create it if it doesn't exist, no matter what

	createAndStartArgs, err := lokiContainerConfigProvider.GetContainerArgs(containerName, labelStrs, volumeName, targetNetworkId)
	if err != nil {
		return "", 0, nil,
			stacktrace.Propagate(
				err,
				"An error occurred getting the logs database container args with container name '%v', labels '%+v', volume name '%v' and network ID '%v",
				containerName,
				labelStrs,
				volumeName,
				targetNetworkId,
			)
	}

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred starting the logs database container with these args '%+v'", createAndStartArgs)
	}
	removeContainerFunc := func() {
		removeCtx := context.Background()
		if err := dockerManager.RemoveContainer(removeCtx, containerId); err != nil {
			logrus.Errorf(
				"Launching the logs database server with GUID '%v' and container ID '%v' didn't complete successfully so we "+
					"tried to kill the container we started, but doing so exited with an error:\n%v",
				engineGuid,
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually stop the logs database server with GUID '%v' and Docker container ID '%v'!!!!!!", engineGuid, containerId)
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
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred waiting for the log database's HTTP port to become available")
	}

	logsDatabaseIP, err := dockerManager.GetContainerIP(ctx, targetNetworkName, containerId)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred getting the IP address of container '%v' in network '%v'", containerId, targetNetworkName)
	}

	shouldRemoveLogsDbContainer = false
	return logsDatabaseIP, privateHttpPortSpec.GetNumber(), removeContainerFunc, nil
}
