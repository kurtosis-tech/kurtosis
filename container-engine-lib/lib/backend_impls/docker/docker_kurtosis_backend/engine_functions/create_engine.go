package engine_functions

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_database"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/logs_database/loki"
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
	nameOfNetworkToStartEngineContainerIn = "bridge"

	maxWaitForEngineAvailabilityRetries         = 10
	timeBetweenWaitForEngineAvailabilityRetries = 1 * time.Second

	logsDatabaseHttpPortId = "http"

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
	matchingNetworks, err := dockerManager.GetNetworksByName(ctx, nameOfNetworkToStartEngineContainerIn)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting networks matching the network we want to start the engine in, '%v'",
			nameOfNetworkToStartEngineContainerIn,
		)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 && numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Expected exactly one network matching the name of the network that we want to start the engine in, '%v', but got %v",
			nameOfNetworkToStartEngineContainerIn,
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

	lokiContainerConfigProvider := loki.CreateLokiContainerConfigProviderForKurtosis()

	killCentralizedLogsComponentsContainersFunc, err := createCentralizedLogsComponents(
		ctx,
		engineGuid,
		targetNetworkId,
		objAttrsProvider,
		dockerManager,
		lokiContainerConfigProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the centralized logs components for the engine with GUID '%v' and network ID '%v'", engineGuid, targetNetworkId)
	}
	shouldKillCentralizedLogsComponentsContainers := true
	defer func() {
		if shouldKillCentralizedLogsComponentsContainers {
			killCentralizedLogsComponentsContainersFunc()
		}
	}()

	shouldKillEngineContainer = false
	shouldKillCentralizedLogsComponentsContainers = false
	return result, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
//TODO we can run it in parallel after the network creation, and we can wait before returnin the EngineInfo object
func createCentralizedLogsComponents(
	ctx context.Context,
	engineGuid engine.EngineGUID,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
	logsDatabaseContainerConfigProvider logs_database.LogsDatabaseContainerConfigProvider,
) (func(), error) {
	killLogsDatabaseContainerFunc, err := createLogsDatabaseContainer(
		ctx,
		engineGuid,
		targetNetworkId,
		objAttrsProvider,
		dockerManager,
		logsDatabaseContainerConfigProvider,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs database container")
	}
	shouldKillLogsDatabaseContainer := true
	defer func() {
		if shouldKillLogsDatabaseContainer {
			killLogsDatabaseContainerFunc()
		}
	}()


	//TODO add createLogsCollectorContainer and handle killing function

	killCentralizedLogsComponentsContainersFunc := func(){
		killLogsDatabaseContainerFunc()
		//TODO call killLogsCollectorContainerFunc()
	}

	shouldKillLogsDatabaseContainer = false
	return killCentralizedLogsComponentsContainersFunc, nil
}

func createLogsDatabaseContainer(
	ctx context.Context,
	engineGuid engine.EngineGUID,
	targetNetworkId string,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	dockerManager *docker_manager.DockerManager,
	logsDatabaseContainerConfigProvider logs_database.LogsDatabaseContainerConfigProvider,
) (func(), error) {
	privateHttpPortSpec, err := logsDatabaseContainerConfigProvider.GetPrivateHttpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database container's private port spec")
	}

	logsDatabaseAttrs, err := objAttrsProvider.ForLogsDatabase(
		engineGuid,
		logsDatabaseHttpPortId,
		privateHttpPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the logs database container attributes using GUID '%v' and the HTTP port spec '%+v'",
			engineGuid,
			privateHttpPortSpec,
		)
	}
	logsDbVolumeAttrs, err := objAttrsProvider.ForLogsDatabaseVolume(engineGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database volume for engine with GUID %v", engineGuid)
	}

	labelStrs := map[string]string{}
	for labelKey, labelValue := range logsDatabaseAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	containerName := logsDatabaseAttrs.GetName().GetString()
	volumeName := logsDbVolumeAttrs.GetName().GetString()

	createAndStartArgs, err := logsDatabaseContainerConfigProvider.GetContainerArgs(containerName, labelStrs, volumeName, targetNetworkId)
	if err != nil {
		return nil,
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
		return nil, stacktrace.Propagate(err, "An error occurred starting the logs database container")
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

	if err := shared_helpers.WaitForPortAvailabilityUsingNetstat(
		ctx,
		dockerManager,
		containerId,
		privateHttpPortSpec,
		maxWaitForEngineAvailabilityRetries,
		timeBetweenWaitForEngineAvailabilityRetries,
	); err != nil {
		return killContainerFunc, stacktrace.Propagate(err, "An error occurred waiting for the log database's HTTP port to become available")
	}

	return killContainerFunc, nil
}
