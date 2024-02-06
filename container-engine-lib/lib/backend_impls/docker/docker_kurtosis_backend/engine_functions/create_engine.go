package engine_functions

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_aggregator_functions/implementations/vector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/reverse_proxy_functions"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/reverse_proxy_functions/implementations/traefik"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	//TODO: pass this parameter
	enclaveManagerUIPort                        = 9711
	enclaveManagerAPIPort                       = 8081
	engineDebugServerPort                       = 50102 // in ClI this is 50101 and 50103 for the APIC
	maxWaitForEngineAvailabilityRetries         = 10
	timeBetweenWaitForEngineAvailabilityRetries = 1 * time.Second
	logsStorageDirpath                          = "/var/log/kurtosis/"
	removeLogsWaitHours                         = 6 * time.Hour
)

func CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	envVars map[string]string,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider,
	shouldStartInDebugMode bool,
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
	shouldRemoveLogsAggregator := true
	defer func() {
		if shouldRemoveLogsAggregator {
			removeLogsAggregatorFunc()
		}
	}()
	logrus.Infof("Centralized logs components started.")

	reverseProxyContainer := traefik.NewTraefikReverseProxyContainer()
	_, removeReverseProxyFunc, err := reverse_proxy_functions.CreateReverseProxy(
		ctx,
		engineGuid,
		reverseProxyContainer,
		dockerManager,
		objAttrsProvider)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"An error occurred attempting to create reverse proxy for engine with GUID '%v' in Docker network with network id '%v'.", engineGuidStr, targetNetworkId)
	}
	shouldRemoveReverseProxy := true
	defer func() {
		if shouldRemoveReverseProxy {
			removeReverseProxyFunc()
		}
	}()
	if err = reverse_proxy_functions.ConnectReverseProxyToEnclaveNetworks(ctx, dockerManager); err != nil {
		return nil, stacktrace.Propagate(err, "An error occured connecting the reverse proxy to the enclave networks")
	}
	logrus.Infof("Reverse proxy started.")

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

	restAPIPortSpec, err := port_spec.NewPortSpec(engine.RESTAPIPortAddr, consts.EngineTransportProtocol, consts.HttpApplicationProtocol, defaultWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the REST API server's http port spec object using number '%v' and protocol '%v'",
			engine.RESTAPIPortAddr,
			consts.EngineTransportProtocol.String(),
		)
	}

	engineAttrs, err := objAttrsProvider.ForEngineServer(
		engineGuid,
		consts.KurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
		consts.KurtosisInternalContainerRESTAPIPortId,
		restAPIPortSpec,
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

	restAPIDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(restAPIPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the Enclave Manager API port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort:       docker_manager.NewManualPublishingSpec(grpcPortNum),
		enclaveManagerUIDockerPort:  docker_manager.NewManualPublishingSpec(uint16(enclaveManagerUIPort)),
		enclaveManagerAPIDockerPort: docker_manager.NewManualPublishingSpec(uint16(enclaveManagerAPIPort)),
		restAPIDockerPort:           docker_manager.NewManualPublishingSpec(engine.RESTAPIPortAddr),
	}

	// Configure the debug port only if it's required
	if shouldStartInDebugMode {
		debugServerPortSpec, err := port_spec.NewPortSpec(
			uint16(engineDebugServerPort),
			consts.EngineTransportProtocol,
			consts.HttpApplicationProtocol,
			defaultWait,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating the Engine's debug server port spec object using number '%v' and protocol '%v'",
				engineDebugServerPort,
				consts.EngineTransportProtocol.String(),
			)
		}

		debugServerDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(debugServerPortSpec)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred transforming the debug server port spec to a Docker port")
		}

		usedPorts[debugServerDockerPort] = docker_manager.NewManualPublishingSpec(uint16(engineDebugServerPort))
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

	createAndStartArgsBuilder := docker_manager.NewCreateAndStartContainerArgsBuilder(
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
	)

	if shouldStartInDebugMode {
		// Adding systrace capabilities when starting the debug server in the engine's container
		capabilities := map[docker_manager.ContainerCapability]bool{
			docker_manager.SysPtrace: true,
		}
		createAndStartArgsBuilder.WithAddedCapabilities(capabilities)

		// Setting security for debugging the engine's container
		securityOpts := map[docker_manager.ContainerSecurityOpt]bool{
			docker_manager.AppArmorUnconfined: true,
		}
		createAndStartArgsBuilder.WithSecurityOpts(securityOpts)
	}

	createAndStartArgs := createAndStartArgsBuilder.Build()

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

	shouldRemoveLogsAggregator = false
	shouldRemoveReverseProxy = false
	shouldKillEngineContainer = false
	return result, nil
}
