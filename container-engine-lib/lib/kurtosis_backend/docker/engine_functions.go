package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	// The location where the engine data directory (on the Docker host machine) will be bind-mounted
	//  on the engine server
	engineDataDirpathOnEngineServerContainer = "/engine-data"

	// This needs to be bind-mounted into the engine & API containers so they can manipulate Docker
	dockerSocketFilepath = "/var/run/docker.sock"

	nameOfNetworkToStartEngineContainerIn = "bridge"

	// The engine server uses gRPC so MUST listen on TCP (no other protocols are supported), which also
	// means that its grpc-proxy must listen on TCP
	enginePortProtocol          = port_spec.PortProtocol_TCP

	// The protocol string we use in the netstat command used to ensure the engine container's grpc & grpc-proxy
	// ports are available
	netstatWaitForAvailabilityPortProtocol = "tcp"

	maxWaitForEngineAvailabilityRetries         = 10
	timeBetweenWaitForEngineAvailabilityRetries = 1 * time.Second

	engineAvailabilityWaitingExecCmdSuccessExitCode = 0

	engineStopTimeout = 10 * time.Second
)

// ====================================================================================================
//                                     Engine CRUD Methods
// ====================================================================================================

func (backendCore *DockerKurtosisBackendCore) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	engineDataDirpathOnHostMachine string,
	envVars map[string]string,
) (
	*engine.Engine,
	error,
) {
	matchingNetworks, err := backendCore.dockerManager.GetNetworksByName(ctx, nameOfNetworkToStartEngineContainerIn)
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

	containerStartTimeUnixSecs := time.Now().Unix()
	engineIdStr := fmt.Sprintf("%v", containerStartTimeUnixSecs)

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, enginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			enginePortProtocol.String(),
		)
	}
	privateGrpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyPortNum, enginePortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the engine's private grpc proxy port spec object using number '%v' and protocol '%v'",
			grpcProxyPortNum,
			enginePortProtocol.String(),
		)
	}

	engineAttrs, err := backendCore.objAttrsProvider.ForEngineServer(
		engineIdStr,
		kurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
		kurtosisInternalContainerGrpcProxyPortId,
		privateGrpcProxyPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the engine server container attributes using id '%v', grpc port num '%v', and " +
				"grpc proxy port num '%v'",
			engineIdStr,
			grpcPortNum,
			grpcProxyPortNum,
		)
	}

	privateGrpcDockerPort, err := transformPortSpecToDockerPort(privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}
	privateGrpcProxyDockerPort, err := transformPortSpecToDockerPort(privateGrpcProxyPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc proxy port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort: docker_manager.NewManualPublishingSpec(grpcPortNum),
		privateGrpcProxyDockerPort: docker_manager.NewManualPublishingSpec(grpcProxyPortNum),
	}

	bindMounts := map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		dockerSocketFilepath:           dockerSocketFilepath,
		engineDataDirpathOnHostMachine: engineDataDirpathOnEngineServerContainer,
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

	// Best-effort pull attempt
	if err = backendCore.dockerManager.PullImage(ctx, containerImageAndTag); err != nil {
		logrus.Warnf("Failed to pull the latest version of engine server image '%v'; you may be running an out-of-date version", containerImageAndTag)
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

	containerId, hostMachinePortBindings, err := backendCore.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}
	shouldKillEngineContainer := true
	defer func() {
		if shouldKillEngineContainer {
			// We kill the container, rather than destroyign it, to leave debugging information around
			if err := backendCore.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf("Launching the engine server didn't complete successfully so we tried to kill the container we started, but doing so exited with an error:\n%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill engine server with container ID '%v'!!!!!!", containerId)
			}
		}
	}()

	if err := waitForEnginePortAvailability(ctx, backendCore.dockerManager, containerId, grpcPortNum); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server's grpc port to become available")
	}

	if err := waitForEnginePortAvailability(ctx, backendCore.dockerManager, containerId, grpcProxyPortNum); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server's grpc proxy port to become available")
	}

	result, err := getEngineObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an engine object from container with ID '%v'", containerId)
	}

	/*
		publicGrpcIpAddr, publicGrpcPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, hostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the engine's public port binding info from the host machine port bindings")
		}
		publicGrpcProxyIpAddr, publicGrpcProxyPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privateGrpcProxyPortSpec, hostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the engine's public grpc proxy port binding info from the host machine port bindings")
		}

		publicGrpcIpAddrStr := publicGrpcIpAddr.String()
		publicGrpcProxyIpAddrStr := publicGrpcProxyIpAddr.String()
		if publicGrpcIpAddrStr != publicGrpcProxyIpAddrStr {
			return nil, stacktrace.NewError(
				"Expected public IP address '%v' for the grpc port to be the same as public IP address '%v' for " +
					"the grpc proxy port, but they were different",
				publicGrpcIpAddrStr,
				publicGrpcProxyIpAddrStr,
			)
		}
		publicIpAddr := publicGrpcIpAddr

		result := engine.NewEngine(
			engineIdStr,
			engine.EngineStatus_Running,
			publicIpAddr,
			publicGrpcPortSpec,
			publicGrpcProxyPortSpec,
		)

	*/

	shouldKillEngineContainer = false
	return result, nil
}

func (backendCore *DockerKurtosisBackendCore) GetEngines(ctx context.Context, filters *engine.GetEnginesFilters) (map[string]*engine.Engine, error) {
	matchingEnginesByContainerId, err := backendCore.getMatchingEnginesByContainerId(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	matchingEnginesByEngineId := map[string]*engine.Engine{}
	for _, engineObj := range matchingEnginesByContainerId {
		matchingEnginesByEngineId[engineObj.GetID()] = engineObj
	}

	return matchingEnginesByEngineId, nil
}

func (backendCore *DockerKurtosisBackendCore) StopEngines(
	ctx context.Context,
	filters *engine.GetEnginesFilters,
) (
	map[string]error,
	error,
) {
	matchingEnginesByContainerId, err := backendCore.getMatchingEnginesByContainerId(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	engineStopErrorsByEngineId := map[string]error{}
	for containerId, engineObj := range matchingEnginesByContainerId {
		engineId := engineObj.GetID()
		if err := backendCore.dockerManager.StopContainer(ctx, containerId, engineStopTimeout); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred stopping engine '%v' with container ID '%v'", engineId, containerId)
			engineStopErrorsByEngineId[engineId] = wrappedErr
		} else {
			engineStopErrorsByEngineId[engineId] = nil
		}
	}
	return engineStopErrorsByEngineId, nil
}

func (backendCore *DockerKurtosisBackendCore) DestroyEngines(
	ctx context.Context,
	filters *engine.GetEnginesFilters,
) (
	map[string]error,
	error,
) {
	matchingEnginesByContainerId, err := backendCore.getMatchingEnginesByContainerId(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	engineDestroyErrorsByEngineId := map[string]error{}
	for containerId, engineObj := range matchingEnginesByContainerId {
		engineId := engineObj.GetID()
		if err := backendCore.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred removing engine '%v' with container ID '%v'", engineId, containerId)
			engineDestroyErrorsByEngineId[engineId] = wrappedErr
		} else {
			engineDestroyErrorsByEngineId[engineId] = nil
		}
	}
	return engineDestroyErrorsByEngineId, nil
}


// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func waitForEnginePortAvailability(ctx context.Context, dockerManager *docker_manager.DockerManager, containerId string, listenPortNum uint16) error {
	commandStr := fmt.Sprintf(
		"[ -n \"$(netstat -anp %v | grep LISTEN | grep %v)\" ]",
		netstatWaitForAvailabilityPortProtocol,
		listenPortNum,
	)
	execCmd := []string{
		"sh",
		"-c",
		commandStr,
	}
	for i := 0; i < maxWaitForEngineAvailabilityRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunExecCommand(ctx, containerId, execCmd, outputBuffer)
		if err == nil {
			if exitCode == engineAvailabilityWaitingExecCmdSuccessExitCode {
				return nil
			}
			logrus.Debugf(
				"Engine server availability-waiting command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				engineAvailabilityWaitingExecCmdSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"Engine server availability-waiting command '%v' experienced a Docker error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxWaitForEngineAvailabilityRetries {
			time.Sleep(timeBetweenWaitForEngineAvailabilityRetries)
		}
	}

	return stacktrace.NewError(
		"The engine server didn't become available (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxWaitForEngineAvailabilityRetries,
		timeBetweenWaitForEngineAvailabilityRetries,
	)
}

// Gets engines matching the search filters, indexed by their container ID
func (backendCore *DockerKurtosisBackendCore) getMatchingEnginesByContainerId(ctx context.Context, filters *engine.GetEnginesFilters) (map[string]*engine.Engine, error) {
	searchLabels := map[string]string{
		// TODO extract this into somewhere better so we're ALWAYS getting containers using the Kurtosis label and nothing else????
		object_attributes_provider.AppIDLabelKey.GetString(): object_attributes_provider.AppIDLabelValue.GetString(),
	}
	containersMatchingLabels, err := backendCore.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching engine containers using labels: %+v", searchLabels)
	}

	allMatchingEngines := map[string]*engine.Engine{}
	for _, matchingContainer := range containersMatchingLabels {
		containerId := matchingContainer.GetId()
		engineObj, err := getEngineObjectFromContainerInfo(
			containerId,
			matchingContainer.GetLabels(),
			matchingContainer.GetStatus(),
			matchingContainer.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with ID '%v' into an engine object", matchingContainer.GetId())
		}
		allMatchingEngines[containerId] = engineObj
	}

	allEnginesMatchingStatus := map[string]*engine.Engine{}
	for containerId, engineObj := range allMatchingEngines {
		engineStatus := engineObj.GetStatus()
		if _, found := filters.Statuses[engineStatus]; found {
			allEnginesMatchingStatus[containerId] = engineObj
		}
	}
	return allEnginesMatchingStatus, nil
}

func getEngineObjectFromContainerInfo(
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (*engine.Engine, error) {
	engineGuid, found := labels[object_attributes_provider.GUIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError(
			"Expected a '%v' label on engine container with ID '%v', but none was found",
			object_attributes_provider.GUIDLabelKey.GetString(),
			containerId,
		)
	}

	privateGrpcPortSpec, privateGrpcProxyPortSpec, err := getPrivateEnginePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the engine container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := isContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for engine container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var engineStatus engine.EngineStatus
	if isContainerRunning {
		engineStatus = engine.EngineStatus_Running
	} else {
		engineStatus = engine.EngineStatus_Stopped
	}

	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	var publicGrpcProxyPortSpec *port_spec.PortSpec
	if engineStatus == engine.EngineStatus_Running {
		publicGrpcPortIpAddr, candidatePublicGrpcPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The engine is running, but an error occurred getting the public port spec for the engine's grpc private port spec")
		}
		publicGrpcPortSpec = candidatePublicGrpcPortSpec

		publicGrpcProxyPortIpAddr, candidatePublicGrpcProxyPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privateGrpcProxyPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The engine is running, but an error occurred getting the public port spec for the engine's grpc private port spec")
		}
		publicGrpcProxyPortSpec = candidatePublicGrpcProxyPortSpec

		if publicGrpcPortIpAddr.String() != publicGrpcProxyPortIpAddr.String() {
			return nil, stacktrace.NewError(
				"Expected the engine's grpc port public IP address '%v' and grpc-proxy port public IP address '%v' to be the same, but they were different",
				publicGrpcPortIpAddr.String(),
				publicGrpcProxyPortIpAddr.String(),
			)
		}
		publicIpAddr = publicGrpcPortIpAddr
	}

	result := engine.NewEngine(
		engineGuid,
		engineStatus,
		publicIpAddr,
		publicGrpcPortSpec,
		publicGrpcProxyPortSpec,
	)

	return result, nil
}

func getPrivateEnginePorts(containerLabels map[string]string) (
	resultGrpcPortSpec *port_spec.PortSpec,
	resultGrpcProxyPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	serializedPortSpecs, found := containerLabels[object_attributes_provider.PortSpecsLabelKey.GetString()]
	if !found {
		return nil, nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", object_attributes_provider.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred deserializing engine server port spec string '%v'", serializedPortSpecs)
	}
	grpcPortSpec, foundGrpcPort := portSpecs[kurtosisInternalContainerGrpcPortId]
	if !foundGrpcPort {
		return nil, nil, stacktrace.NewError("No engine grpc port with ID '%v' found in the engine server port specs", kurtosisInternalContainerGrpcPortId)
	}
	grpcProxyPortSpec, foundGrpcProxyPort := portSpecs[kurtosisInternalContainerGrpcProxyPortId]
	if !foundGrpcProxyPort {
		return nil, nil, stacktrace.NewError("No engine grpc-proxy port with ID '%v' found in the engine server port specs", kurtosisInternalContainerGrpcProxyPortId)
	}
	return grpcPortSpec, grpcProxyPortSpec, nil
}
