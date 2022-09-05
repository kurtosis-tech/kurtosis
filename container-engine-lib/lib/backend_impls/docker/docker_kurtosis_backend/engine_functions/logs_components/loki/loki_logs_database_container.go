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
)

type lokiLogsDatabaseContainer struct {}

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
	resultKillLogsDatabaseContainerFunc func(),
	resultErr error,
) {

	lokiContainerConfigProvider := createLokiContainerConfigProviderForKurtosis()

	privateHttpPortSpec, err := lokiContainerConfigProvider.GetPrivateHttpPortSpec()
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred getting the logs database container's private port spec")
	}

	logsDatabaseAttrs, err := objAttrsProvider.ForLogsDatabase(
		engineGuid,
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
	killContainerFunc := func() {
		if err := dockerManager.KillContainer(context.Background(), containerId); err != nil {
			logrus.Errorf(
				"Launching the logs database server with GUID '%v' and container ID '%v' didn't complete successfully so we "+
					"tried to kill the container we started, but doing so exited with an error:\n%v",
				engineGuid,
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually stop the logs database server with GUID '%v' and Docker container ID '%v'!!!!!!", engineGuid, containerId)
		}
	}
	shouldKillLogsDbContainer := true
	defer func() {
		if shouldKillLogsDbContainer {
			killContainerFunc()
		}
	}()
	//We do not delete the volume because the volume could already exist from previous engine executions
	//for this reason the logs database volume creation has to be idempotent, we ALWAYS want to create it if it doesn't exist, no matter what

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

	shouldKillLogsDbContainer = false
	return logsDatabaseIP, privateHttpPortSpec.GetNumber(), killContainerFunc, nil
}

