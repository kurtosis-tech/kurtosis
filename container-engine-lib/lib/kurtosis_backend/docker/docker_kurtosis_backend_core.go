package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	dockerSocketFilepath = "/var/run/docker.sock"

	// The Docker API's default is to return just containers whose status is "running"
	// However, we'd rather do our own filtering on what "running" means (because, e.g., "restarting"
	// should also be considered as running)
	shouldFetchAllContainersWhenRetrievingContainers = true

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

	shouldGetStoppedContainersWhenCheckingForExistingEngines = false

	engineStopTimeout = 10 * time.Second

	// The ID of the GRPC port for Kurtosis-internal containers (e.g. API container, engine, modules, etc.) which will
	//  be stored in the port spec label
	kurtosisInternalContainerGrpcPortId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	// need a proxy  that will translate grpc-web requests before they hit the main GRPC server
	kurtosisInternalContainerGrpcProxyPortId = "grpcProxy"

	// Engine container port number string parsing constants
	hostMachinePortNumStrParsingBase = 10
	hostMachinePortNumStrParsingBits = 16
)

// This maps a Docker container's status to a binary "is the container considered running?" determiner
// Its completeness is enforced via unit test
var isContainerRunningDeterminer = map[types.ContainerStatus]bool{
	types.ContainerStatus_Paused: false,
	types.ContainerStatus_Restarting: true,
	types.ContainerStatus_Running: true,
	types.ContainerStatus_Removing: true,
	types.ContainerStatus_Dead: false,
	types.ContainerStatus_Created: false,
	types.ContainerStatus_Exited: false,
}

// Unfortunately, Docker doesn't have an enum for the protocols it supports, so we have to create this translation map
var portSpecProtosToDockerPortProtos = map[port_spec.PortProtocol]string{
	port_spec.PortProtocol_TCP:  "tcp",
	port_spec.PortProtocol_SCTP: "sctp",
	port_spec.PortProtocol_UDP:   "udp",
}

type DockerKurtosisBackendCore struct {
	// The logger that all log messages will be written to
	log *logrus.Logger // NOTE: This log should be used for all log statements - the system-wide logger should NOT be used!

	dockerManager *docker_manager.DockerManager

	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider
}

func NewDockerKurtosisBackendCore(log *logrus.Logger, dockerManager *docker_manager.DockerManager) *DockerKurtosisBackendCore {
	return &DockerKurtosisBackendCore{
		log:              log,
		dockerManager:    dockerManager,
		objAttrsProvider: object_attributes_provider.GetDockerObjectAttributesProvider(),
	}
}

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
		engineDataDirpathOnHostMachine: kurtosis_backend.EngineDataDirpathOnEngineServerContainer,
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

	if err := waitForAvailability(ctx, backendCore.dockerManager, containerId, grpcPortNum); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server's grpc port to become available")
	}

	if err := waitForAvailability(ctx, backendCore.dockerManager, containerId, grpcProxyPortNum); err != nil {
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

/*
func (backendCore *DockerKurtosisBackendCore) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	successfullyDestroyedContainerNames, containerDestructionErrors, err := backendCore.cleanContainers(ctx, engineLabels, shouldCleanRunningEngineContainers)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred cleaning stopped Kurtosis engine containers")
	}
	return successfullyDestroyedContainerNames, containerDestructionErrors, nil
}


func (backendCore *DockerKurtosisBackendCore) GetEnginePublicIPAndPort(
	ctx context.Context,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultIsEngineStopped bool,
	resultErr error,
) {
	runningEngineContainers, err := backendCore.dockerManager.GetContainersByLabels(ctx, engineLabels, shouldGetStoppedContainersWhenCheckingForExistingEngines)
	if err != nil {
		return nil, 0, false, stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers")
	}

	numRunningEngineContainers := len(runningEngineContainers)
	if numRunningEngineContainers > 1 {
		return nil, 0, false, stacktrace.NewError("Cannot report engine status because we found %v running Kurtosis engine containers; this is very strange as there should never be more than one", numRunningEngineContainers)
	}
	if numRunningEngineContainers == 0 {
		return nil, 0, true, nil
	}
	engineContainer := runningEngineContainers[0]

	currentlyRunningEngineContainerLabels := engineContainer.GetLabels()
}
*/

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func transformPortSpecToDockerPort(portSpec *port_spec.PortSpec) (nat.Port, error) {
	portSpecProto := portSpec.GetProtocol()
	dockerProto, found := portSpecProtosToDockerPortProtos[portSpecProto]
	if !found {
		// This should never happen because we enforce completeness via a unit test
		return "", stacktrace.NewError("Expected a Docker port protocol for port spec protocol '%v' but none was found; this is a bug in Kurtosis!", portSpecProto.String())
	}

	portSpecNum := portSpec.GetNumber()
	dockerPort, err := nat.NewPort(
		dockerProto,
		fmt.Sprintf("%v", portSpecNum),
	)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred creating a Docker port object using port num '%v' and Docker protocol '%v'",
			portSpecNum,
			dockerProto,
		)
	}
	return dockerPort, nil
}

// TODO MOVE THIS TO WHOMEVER CALLS KURTOSISBACKEND
// This is a helper function that will take multiple errors, each identified by an ID, and format them together
// If no errors are returned, this function returns nil
func buildCombinedError(errorsById map[string]error, titleStr string) error {
	allErrorStrs := []string{}
	for errorId, stopErr := range errorsById {
		errorFormatStr := ">>>>>>>>>>>>> %v %v <<<<<<<<<<<<<\n" +
			"%v\n" +
			">>>>>>>>>>>>> END %v %v <<<<<<<<<<<<<"
		errorStr := fmt.Sprintf(
			errorFormatStr,
			strings.ToUpper(titleStr),
			errorId,
			stopErr.Error(),
			strings.ToUpper(titleStr),
			errorId,
		)
		allErrorStrs = append(allErrorStrs, errorStr)
	}

	if len(allErrorStrs) > 0 {
		// NOTE: This is one of the VERY rare cases where we don't want to use stacktrace.Propagate, because
		// attaching stack information for this method (which simply combines errors) just isn't useful. The
		// expected behaviour is that the caller of this function will use stacktrace.Propagate
		return errors.New(strings.Join(
			allErrorStrs,
			"\n\n",
		))
	}

	return nil
}


func waitForAvailability(ctx context.Context, dockerManager *docker_manager.DockerManager, containerId string, listenPortNum uint16) error {
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

func getPublicPortBindingFromPrivatePortSpec(privatePortSpec *port_spec.PortSpec, allHostMachinePortBindings map[nat.Port]*nat.PortBinding) (
	resultPublicIpAddr net.IP,
	resultPublicPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	portNum := privatePortSpec.GetNumber()

	// Convert port spec protocol -> Docker protocol string
	portSpecProto := privatePortSpec.GetProtocol()
	privatePortDockerProto, found := portSpecProtosToDockerPortProtos[portSpecProto]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No Docker protocol was defined for port spec proto '%v'; this is a bug in Kurtosis",
			portSpecProto.String(),
		)
	}

	privatePortNumStr := fmt.Sprintf("%v", portNum)
	dockerPrivatePort, err := nat.NewPort(
		privatePortDockerProto,
		privatePortNumStr,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Docker private port object from port number '%v' and protocol '%v', which is necessary for getting the corresponding host machine port bindings",
			privatePortNumStr,
			privatePortDockerProto,
		)
	}

	hostMachinePortBinding, found := allHostMachinePortBindings[dockerPrivatePort]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No host machine port binding was specified for Docker port '%v' which corresponds to port spec with num '%v' and protocol '%v'",
			dockerPrivatePort,
			portNum,
			portSpecProto.String(),
		)
	}

	hostMachineIpAddrStr := hostMachinePortBinding.HostIP
	hostMachineIp := net.ParseIP(hostMachineIpAddrStr)
	if hostMachineIp == nil {
		return nil, nil, stacktrace.NewError(
			"Found host machine IP string '%v' for port spec with number '%v' and protocol '%v', but it wasn't a valid IP",

			hostMachineIpAddrStr)
	}

	hostMachinePortNumStr := hostMachinePortBinding.HostPort
	hostMachinePortNumUint64, err := strconv.ParseUint(hostMachinePortNumStr, hostMachinePortNumStrParsingBase, hostMachinePortNumStrParsingBits)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred parsing engine container host machine port num string '%v' using base '%v' and num bits '%v'",
			hostMachinePortNumStr,
			hostMachinePortNumStrParsingBase,
			hostMachinePortNumStrParsingBits,
		)
	}
	hostMachinePortNumUint16 := uint16(hostMachinePortNumUint64) // Okay to do due to specifying the number of bits above
	publicPortSpec, err := port_spec.NewPortSpec(hostMachinePortNumUint16, portSpecProto)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating public port spec with host machine port num '%v' and protocol '%v'", hostMachinePortNumUint16, portSpecProto.String())
	}

	return hostMachineIp, publicPortSpec, nil
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