package engine_functions

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_components"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_components/fluentbit"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_components/loki"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	nameOfNetworkToStartEngineContainersIn = "bridge"

	maxWaitForEngineAvailabilityRetries         = 10
	timeBetweenWaitForEngineAvailabilityRetries = 1 * time.Second

	logsDatabaseHttpPortId = "http"

	logsCollectorTcpPortId  = "tcp"
	logsCollectorHttpPortId = "http"
)

func CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	envVars map[string]string,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
) (
	*engine.Engine,
	error,
) {
	matchingNetworks, err := dockerManager.GetNetworksByName(ctx, nameOfNetworkToStartEngineContainersIn)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting networks matching the network we want to start the engine in, '%v'",
			nameOfNetworkToStartEngineContainersIn,
		)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 && numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Expected exactly one network matching the name of the network that we want to start the engine in, '%v', but got %v",
			nameOfNetworkToStartEngineContainersIn,
			numMatchingNetworks,
		)
	}
	targetNetwork := matchingNetworks[0]
	targetNetworkId := targetNetwork.GetId()

	engineGuidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID string for the engine")
	}
	engineGuid := engine.EngineGUID(engineGuidStr)

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, consts.EnginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			consts.EnginePortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, consts.EnginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc proxy port spec object using number '%v' and protocol '%v'",
			grpcProxyPortNum,
			consts.EnginePortProtocol.String(),
		)
	}

	engineAttrs, err := objAttrsProvider.ForEngineServer(
		engineGuid,
		consts.KurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
		consts.KurtosisInternalContainerGrpcProxyPortId,
		privateGrpcProxyPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the engine server container attributes using GUID '%v', grpc port num '%v', and "+
				"grpc proxy port num '%v'",
			engineGuid,
			grpcPortNum,
			grpcProxyPortNum,
		)
	}

	privateGrpcDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}
	privateGrpcProxyDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateGrpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc proxy port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort:      docker_manager.NewManualPublishingSpec(grpcPortNum),
		privateGrpcProxyDockerPort: docker_manager.NewManualPublishingSpec(grpcProxyPortNum),
	}

	bindMounts := map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		consts.DockerSocketFilepath: consts.DockerSocketFilepath,
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
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		labelStrs,
	).Build()

	// Best-effort pull attempt
	if err = dockerManager.PullImage(ctx, containerImageAndTag); err != nil {
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

	// TODO UNCOMMENT THIS ONCE WE HAVE GRPC-PROXY WIRED UP!!
	/*
		if err := waitForPortAvailabilityUsingNetstat(ctx, backend.dockerManager, containerId, grpcProxyPortNum); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server's grpc proxy port to become available")
		}
	*/

	result, err := getEngineObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an engine object from container with GUID '%v'", containerId)
	}

	killCentralizedLogsComponentsContainersAndVolumesFunc, err := createCentralizedLogsComponents(
		ctx,
		engineGuid,
		targetNetworkId,
		objAttrsProvider,
		dockerManager,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the centralized logs components for the engine with GUID '%v' and network ID '%v'", engineGuid, targetNetworkId)
	}
	shouldKillCentralizedLogsComponentsContainers := true
	defer func() {
		if shouldKillCentralizedLogsComponentsContainers {
			killCentralizedLogsComponentsContainersAndVolumesFunc()
		}
	}()

	shouldKillEngineContainer = false
	shouldKillCentralizedLogsComponentsContainers = false
	return result, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
//TODO we can run it in parallel after the network creation, and we can wait before returning the EngineInfo object
func createCentralizedLogsComponents(
	ctx context.Context,
	engineGuid engine.EngineGUID,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
) (func(), error) {

	logsDatabase := loki.CreateLokiConfiguredForKurtosis()

	logsDatabaseHost, logsDatabasePort, killLogsDatabaseContainerAndVolumeFunc, err := createLogsDatabaseContainer(
		ctx,
		engineGuid,
		targetNetworkId,
		objAttrsProvider,
		dockerManager,
		logsDatabase,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the logs database container for engine with GUID '%v' in Docker network with ID '%v'",
			engineGuid,
			targetNetworkId,
		)
	}
	shouldKillLogsDatabaseContainerAndVolume := true
	defer func() {
		if shouldKillLogsDatabaseContainerAndVolume {
			killLogsDatabaseContainerAndVolumeFunc()
		}
	}()

	logsCollector := fluentbit.CreateFluentbitConfiguredForKurtosis(logsDatabaseHost, logsDatabasePort)

	killLogsCollectorContainerAndVolumeFunc, err := createLogsCollectorContainer(
		ctx,
		engineGuid,
		targetNetworkId,
		objAttrsProvider,
		dockerManager,
		logsCollector,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the logs collector container for engine with GUID '%v' in Docker network with ID '%v'",
			engineGuid,
			targetNetworkId,
		)
	}

	killCentralizedLogsComponentsContainersAndVolumesFunc := func() {
		killLogsDatabaseContainerAndVolumeFunc()
		killLogsCollectorContainerAndVolumeFunc()
	}

	shouldKillLogsDatabaseContainerAndVolume = false
	return killCentralizedLogsComponentsContainersAndVolumesFunc, nil
}

func createLogsDatabaseContainer(
	ctx context.Context,
	engineGuid engine.EngineGUID,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
	logsDatabase logs_components.LogsDatabase,
) (
	resultLogsDatabasePrivateHost string,
	resultLogsDatabasePrivatePort uint16,
	resultKillLogsDatabaseContainerFunc func(),
	resultErr error,
) {
	privateHttpPortSpec, err := logsDatabase.GetPrivateHttpPortSpec()
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred getting the logs database private port spec")
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
	logsDbVolumeAttrs, err := objAttrsProvider.ForLogsDatabaseVolume(engineGuid)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred getting the logs database volume attributes for engine with GUID %v", engineGuid)
	}

	labelStrs := map[string]string{}
	for labelKey, labelValue := range logsDatabaseAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	containerName := logsDatabaseAttrs.GetName().GetString()
	volumeName := logsDbVolumeAttrs.GetName().GetString()

	createAndStartArgs, err := logsDatabase.GetContainerArgs(containerName, labelStrs, volumeName, targetNetworkId)
	if err != nil {
		return "", 0, nil,
			stacktrace.Propagate(
				err,
				"An error occurred getting the logs-database-container-args with container name '%v', labels '%+v', volume name '%v' and network ID '%v",
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
		if err := dockerManager.RemoveVolume(ctx, volumeName); err != nil {
			logrus.Errorf(
				"Launching the logs database server with GUID '%v' and container ID '%v' didn't complete successfully so we "+
					"tried to remove the associated volume we started, but doing so exited with an error:\n%v",
				engineGuid,
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs database volume '%v'!!!!!!", volumeName)
		}
	}
	shouldKillLogsDbContainerAndVolume := true
	defer func() {
		if shouldKillLogsDbContainerAndVolume {
			killContainerFunc()
		}
	}()

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		ctx,
		dockerManager,
		containerId,
		privateHttpPortSpec,
		maxWaitForEngineAvailabilityRetries,
		timeBetweenWaitForEngineAvailabilityRetries,
	); err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred waiting for the log database's HTTP port to become available")
	}

	logsDatabaseIP, err := dockerManager.GetContainerIP(ctx, nameOfNetworkToStartEngineContainersIn, containerId)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred ")
	}

	shouldKillLogsDbContainerAndVolume = false
	return logsDatabaseIP, privateHttpPortSpec.GetNumber(), killContainerFunc, nil
}

func createLogsCollectorContainer(
	ctx context.Context,
	engineGuid engine.EngineGUID,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
	logsCollector logs_components.LogsCollector,
) (func(), error) {

	privateTcpPortSpec, err := logsCollector.GetPrivateTcpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector private TCP port spec")
	}

	privateHttpPortSpec, err := logsCollector.GetPrivateHttpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector private HTTP port spec")
	}

	logsCollectorAttrs, err := objAttrsProvider.ForLogsCollector(
		engineGuid,
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

	logsCollectorVolumeAttrs, err := objAttrsProvider.ForLogsCollectorVolume(engineGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector volume attributes for engine with GUID %v", engineGuid)
	}

	volumeName := logsCollectorVolumeAttrs.GetName().GetString()

	createAndStartArgs, err := logsCollector.GetContainerArgs(containerName, labelStrs, volumeName, targetNetworkId, dockerManager)
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
		return nil, stacktrace.Propagate(err, "An error occurred starting the logs controller container with these args '%+v'", createAndStartArgs)
	}
	killContainerFunc := func() {
		if err := dockerManager.KillContainer(context.Background(), containerId); err != nil {
			logrus.Errorf(
				"Launching the logs controller server for engine with GUID '%v' and container ID '%v' didn't complete successfully so we "+
					"tried to kill the container we started, but doing so exited with an error:\n%v",
				engineGuid,
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually stop the logs controller server for engine with GUID '%v' and Docker container ID '%v'!!!!!!", engineGuid, containerId)
		}
		if err := dockerManager.RemoveVolume(ctx, volumeName); err != nil {
			logrus.Errorf(
				"Launching the logs controller server for engine with GUID '%v' and container ID '%v' didn't complete successfully so we "+
					"tried to remove the associated volume we started, but doing so exited with an error:\n%v",
				engineGuid,
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the logs collector volume '%v'!!!!!!", volumeName)
		}
	}
	shouldKillLogsCollectorContainer := true
	defer func() {
		if shouldKillLogsCollectorContainer {
			killContainerFunc()
		}
	}()


	if err := logsCollector.WaitForAvailability(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the log collector's to become available")
	}

	shouldKillLogsCollectorContainer = false
	return killContainerFunc, nil
}


