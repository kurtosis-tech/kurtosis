package engine_functions

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions/implementations/vector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_remover"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	//TODO: pass this parameter
	enclaveManagerUIPort                        = 9711
	enclaveManagerAPIPort                       = 8081
	maxWaitForEngineAvailabilityRetries         = 10
	timeBetweenWaitForEngineAvailabilityRetries = 1 * time.Second
	logsStorageDirpath                          = "/var/log/kurtosis/"
)

var (
	logRemovalTicker = time.NewTicker(6 * time.Hour)
)

func CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	envVars map[string]string,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*engine.Engine,
	error,
) {
	serverArgs, err := args.GetArgsFromEnvVars(envVars)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't retrieve engine server args from the env vars")
	}

	engineGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID string for the engine")
	}
	engineGuid := engine.EngineGUID(engineGuidStr)

	defaultWait, err := port_spec.CreateWaitWithDefaultValues()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a wait with default values")
	}

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, consts.EngineTransportProtocol, consts.HttpApplicationProtocol, defaultWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			consts.EngineTransportProtocol.String(),
		)
	}

	engineNetwork, err := shared_helpers.GetEngineAndLogsComponentsNetwork(ctx, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine network")
	}
	targetNetworkId := engineNetwork.GetId()

	logrus.Infof("Starting the centralized logs components...")
	logsStorageAttrs, err := objAttrsProvider.ForLogsStorageVolume()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving logs storage object attributes.")
	}
	logsStorageVolNameStr := logsStorageAttrs.GetName().GetString()
	volumeLabelStrs := map[string]string{}
	for labelKey, labelValue := range logsStorageAttrs.GetLabels() {
		volumeLabelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	// Creation of volume should be idempotent because the volume with persisted logs in it could already exist
	// Thus, we don't defer an undo volume if this operation fails
	if err = dockerManager.CreateVolume(ctx, logsStorageVolNameStr, volumeLabelStrs); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating logs storage.")
	}

	logsAggregatorContainer := vector.NewVectorLogsAggregatorContainer() // Declaring implementation
	_, removeLogsAggregatorFunc, err := logs_aggregator_functions.CreateLogsAggregator(
		ctx,
		logsAggregatorContainer,
		dockerManager,
		objAttrsProvider)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"An error occurred attempting to create logging components for engine with GUID '%v' in Docker network with network id '%v'.", engineGuidStr, targetNetworkId)
	}

	// schedule log removal for log retention
	osFs := volume_filesystem.NewOsVolumeFilesystem()
	realTime := logs_clock.NewRealClock()
	logRemover := log_remover.NewLogRemover(osFs, realTime)
	go func() {
		for {
			select {

			// attempt to remove logs every six hours
			case <-logRemovalTicker.C:
				logRemover.Run()
			}
		}
	}()

	shouldRemoveCentralizedLogComponents := true
	defer func() {
		if shouldRemoveCentralizedLogComponents {
			removeLogsAggregatorFunc()
		}
	}()
	logrus.Infof("Centralized logs components started.")

	enclaveManagerUIPortSpec, err := port_spec.NewPortSpec(uint16(enclaveManagerUIPort), consts.EngineTransportProtocol, consts.HttpApplicationProtocol, defaultWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Enclave Manager UI's http port spec object using number '%v' and protocol '%v'",
			enclaveManagerUIPort,
			consts.EngineTransportProtocol.String(),
		)
	}

	enclaveManagerApiPortSpec, err := port_spec.NewPortSpec(
		uint16(enclaveManagerAPIPort),
		consts.EngineTransportProtocol,
		consts.HttpApplicationProtocol,
		defaultWait,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Enclave Manager API's http port spec object using number '%v' and protocol '%v'",
			enclaveManagerAPIPort,
			consts.EngineTransportProtocol.String(),
		)
	}

	engineAttrs, err := objAttrsProvider.ForEngineServer(
		engineGuid,
		consts.KurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the engine server container attributes using GUID '%v' and GRPC port num '%v'",
			engineGuid,
			grpcPortNum,
		)
	}

	privateGrpcDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}

	enclaveManagerUIDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(enclaveManagerUIPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the Enclave Manager UI port spec to a Docker port")
	}

	enclaveManagerAPIDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(enclaveManagerApiPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the Enclave Manager API port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort:       docker_manager.NewManualPublishingSpec(grpcPortNum),
		enclaveManagerUIDockerPort:  docker_manager.NewManualPublishingSpec(uint16(enclaveManagerUIPort)),
		enclaveManagerAPIDockerPort: docker_manager.NewManualPublishingSpec(uint16(enclaveManagerAPIPort)),
	}

	bindMounts := map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		consts.DockerSocketFilepath: consts.DockerSocketFilepath,
	}

	volumeMounts := map[string]string{
		logsStorageVolNameStr: logsStorageDirpath,
	}

	if serverArgs.OnBastionHost {
		// Mount the host engine config directory so the engine can access files like the remote backend config.
		bindMounts[consts.HostEngineConfigDirToMount] = consts.EngineConfigLocalDir
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		imageOrgAndRepo,
		imageVersionTag,
	)

	labelStrs := map[string]string{}
	for labelKey, labelValue := range engineAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImageAndTag,
		engineAttrs.GetName().GetString(),
		targetNetworkId,
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithVolumeMounts(
		volumeMounts,
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		labelStrs,
	).Build()

	// Best-effort pull attempt
	if err = dockerManager.FetchImage(ctx, containerImageAndTag); err != nil {
		logrus.Warnf("Failed to pull the latest version of engine server image '%v'; you may be running an out-of-date version", containerImageAndTag)
	}

	containerId, hostMachinePortBindings, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}
	shouldKillEngineContainer := true
	defer func() {
		if shouldKillEngineContainer {
			// NOTE: We use the background context here so that the kill will still go off even if the reason for
			// the failure was the original context being cancelled
			if err := dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf(
					"Launching the engine server with GUID '%v' and container ID '%v' didn't complete successfully so we "+
						"tried to kill the container we started, but doing so exited with an error:\n%v",
					engineGuid,
					containerId,
					err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop engine server with GUID '%v'!!!!!!", engineGuid)
			}
		}
	}()

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		ctx,
		dockerManager,
		containerId,
		privateGrpcPortSpec,
		maxWaitForEngineAvailabilityRetries,
		timeBetweenWaitForEngineAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server's grpc port to become available")
	}

	result, err := getEngineObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an engine object from container with GUID '%v'", containerId)
	}

	shouldRemoveCentralizedLogComponents = false
	shouldKillEngineContainer = false
	return result, nil
}
