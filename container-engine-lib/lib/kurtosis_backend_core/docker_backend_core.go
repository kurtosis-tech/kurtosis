package kurtosis_backend_core

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	container_status_calculator "github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend_core/helpers"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	dockerSocketFilepath = "/var/run/docker.sock"

	networkToStartEngineContainerIn = "bridge"

	// The engine server uses gRPC so MUST listen on TCP (no other protocols are supported)
	// This is the Docker constant indicating a TCP port
	engineContainerDockerPortProtocol = "tcp"

	// The protocol string we use in the netstat command used to ensure the engine container is available
	netstatWaitForAvailabilityPortProtocol = "tcp"

	maxWaitForAvailabilityRetries         = 10
	timeBetweenWaitForAvailabilityRetries = 1 * time.Second

	availabilityWaitingExecCmdSuccessExitCode = 0

	shouldGetStoppedContainersWhenCheckingForExistingEngines = false

	engineStopTimeout = 30 * time.Second

	// Obviously yes
	shouldFetchStoppedContainersWhenDestroyingStoppedContainers = true

	// --------------------------- Old port parsing constants ------------------------------------
	// These are the old labels that the API container used to use before 2021-11-15 for declaring its port num protocol
	// We can get rid of this after 2022-05-15, when we're confident no users will be running API containers with the old label
	pre2021_11_15_portNum   = uint16(9710)
	pre2021_11_15_portProto = schema.PortProtocol_TCP

	// These are the old labels that the API container used to use before 2021-12-02 for declaring its port num protocol
	// We can get rid of this after 2022-06-02, when we're confident no users will be running API containers with the old label
	pre2021_12_02_portNumLabel    = "com.kurtosistech.port-number"
	pre2021_12_02_portNumBase     = 10
	pre2021_12_02_portNumUintBits = 16
	pre2021_12_02_portProtocol    = schema.PortProtocol_TCP
	// --------------------------- Old port parsing constants ------------------------------------

	// Engine container port number string parsing constants
	hostMachinePortNumStrParsingBase = 10
	hostMachinePortNumStrParsingBits = 16
)

// Unfortunately, Docker doesn't have constants for the protocols it supports declared
var objAttrsSchemaPortProtosToDockerPortProtos = map[schema.PortProtocol]string{
	schema.PortProtocol_TCP:  "tcp",
	schema.PortProtocol_SCTP: "sctp",
	schema.PortProtcol_UDP:   "udp",
}

type KurtosisDockerBackendCore struct {
	// The logger that all log messages will be written to
	log *logrus.Logger // NOTE: This log should be used for all log statements - the system-wide logger should NOT be used!

	dockerManager *docker_manager.DockerManager

	objAttrsProvider schema.ObjectAttributesProvider
}

func NewDockerBackendCore(log *logrus.Logger, dockerManager *docker_manager.DockerManager, objAttrsProvider schema.ObjectAttributesProvider) *KurtosisDockerBackendCore {
	return &KurtosisDockerBackendCore{
		log:              log,
		dockerManager:    dockerManager,
		objAttrsProvider: objAttrsProvider,
	}
}

func (kdb *KurtosisDockerBackendCore) CreateEngine(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	listenPortNum uint16,
	engineDataDirpathOnHostMachine string,
	containerImage string,
	engineServerArgs *args.EngineServerArgs,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultErr error,
) {
	matchingNetworks, err := kdb.dockerManager.GetNetworksByName(ctx, networkToStartEngineContainerIn)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred getting networks matching the network we want to start the engine in, '%v'",
			networkToStartEngineContainerIn,
		)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 && numMatchingNetworks > 1 {
		return nil, 0, stacktrace.NewError(
			"Expected exactly one network matching the name of the network that we want to start the engine in, '%v', but got %v",
			networkToStartEngineContainerIn,
			numMatchingNetworks,
		)
	}
	targetNetwork := matchingNetworks[0]
	targetNetworkId := targetNetwork.GetId()

	engineAttrs, err := kdb.objAttrsProvider.ForEngineServer(listenPortNum)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred getting the engine server container attributes using port num '%v'", listenPortNum)
	}

	enginePortObj, err := nat.NewPort(
		engineContainerDockerPortProtocol,
		fmt.Sprintf("%v", listenPortNum),
	)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred creating a port object with port num '%v' and protocol '%v' to represent the engine's port",
			listenPortNum,
			engineContainerDockerPortProtocol,
		)
	}

	envVars, err := args.GetEnvFromArgs(engineServerArgs)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		enginePortObj: docker_manager.NewManualPublishingSpec(listenPortNum),
	}

	bindMounts := map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		dockerSocketFilepath:           dockerSocketFilepath,
		engineDataDirpathOnHostMachine: EngineDataDirpathOnEngineServerContainer,
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		containerImage,
		imageVersionTag,
	)

	// Best-effort pull attempt
	if err = kdb.dockerManager.PullImage(ctx, containerImageAndTag); err != nil {
		logrus.Warnf("Failed to pull the latest version of engine server image '%v'; you may be running an out-of-date version", containerImageAndTag)
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImageAndTag,
		engineAttrs.GetName(),
		targetNetworkId,
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		engineAttrs.GetLabels(),
	).Build()

	containerId, hostMachinePortBindings, err := kdb.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}
	shouldKillEngineContainer := true
	defer func() {
		if shouldKillEngineContainer {
			if err := kdb.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf("Launching the engine server didn't complete successfully so we tried to kill the container we started, but doing so exited with an error:\n%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill engine server with container ID '%v'!!!!!!", containerId)
			}
		}
	}()

	if err := waitForAvailability(ctx, kdb.dockerManager, containerId, listenPortNum); err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred waiting for the engine server to become available")
	}

	hostMachineEnginePortBinding, found := hostMachinePortBindings[enginePortObj]
	if !found {
		return nil, 0, stacktrace.NewError("The Kurtosis engine server started successfully, but no host machine port binding was found")
	}

	publicIpAddrStr := hostMachineEnginePortBinding.HostIP
	publicIpAddr := net.ParseIP(publicIpAddrStr)
	if publicIpAddr == nil {
		return nil, 0, stacktrace.NewError("The engine server's port was reported to be bound on host machine interface IP '%v', but this is not a valid IP string", publicIpAddrStr)
	}

	publicPortNumStr := hostMachineEnginePortBinding.HostPort
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, publicPortNumParsingBase, publicPortNumParsingUintBits)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			publicPortNumParsingBase,
			publicPortNumParsingUintBits,
		)
	}
	publicPortNumUint16 := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command

	shouldKillEngineContainer = false
	return publicIpAddr, publicPortNumUint16, nil
}

func (kdb *KurtosisDockerBackendCore) StopEngine(ctx context.Context) error {
	matchingEngineContainers, err := kdb.dockerManager.GetContainersByLabels(
		ctx,
		engineLabels,
		shouldGetStoppedContainersWhenCheckingForExistingEngines,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers, which we need to check for existing engines")
	}

	numMatchingEngineContainers := len(matchingEngineContainers)
	if numMatchingEngineContainers == 0 {
		return nil
	}
	if numMatchingEngineContainers > 1 {
		logrus.Warnf(
			"Found %v Kurtosis engine containers, which is strange because there should never be more than 1 engine container; all will be stopped",
			numMatchingEngineContainers,
		)
	}

	engineStopErrorStrs := []string{}
	for _, engineContainer := range matchingEngineContainers {
		containerName := engineContainer.GetName()
		containerId := engineContainer.GetId()
		if err := kdb.dockerManager.StopContainer(ctx, containerId, engineStopTimeout); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred stopping engine container '%v' with ID '%v'",
				containerName,
				containerId,
			)
			engineStopErrorStrs = append(engineStopErrorStrs, wrappedErr.Error())
		}
	}

	if len(engineStopErrorStrs) > 0 {
		return stacktrace.NewError(
			"One or more errors occurred stopping the engine(s):\n%v",
			strings.Join(
				engineStopErrorStrs,
				"\n\n",
			),
		)
	}
	return nil
}

func (kdb *KurtosisDockerBackendCore) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	successfullyDestroyedContainerNames, containerDestructionErrors, err := kdb.cleanContainers(ctx, engineLabels, shouldCleanRunningEngineContainers)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred cleaning stopped Kurtosis engine containers")
	}
	return successfullyDestroyedContainerNames, containerDestructionErrors, nil
}

func (kkb *KurtosisDockerBackendCore) GetEngineStatus(
	ctx context.Context,
) (engineStatus string, ipAddr net.IP, portNum uint16, err error) {
	runningEngineContainers, err := kkb.dockerManager.GetContainersByLabels(ctx, engineLabels, shouldGetStoppedContainersWhenCheckingForExistingEngines)
	if err != nil {
		return "", nil, 0, stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers")
	}

	numRunningEngineContainers := len(runningEngineContainers)
	if numRunningEngineContainers > 1 {
		return "", nil, 0, stacktrace.NewError("Cannot report engine status because we found %v running Kurtosis engine containers; this is very strange as there should never be more than one", numRunningEngineContainers)
	}
	if numRunningEngineContainers == 0 {
		return EngineStatus_Stopped, nil, 0, nil
	}
	engineContainer := runningEngineContainers[0]

	currentlyRunningEngineContainerLabels := engineContainer.GetLabels()
	objAttrPrivatePort, err := getPrivateEnginePort(currentlyRunningEngineContainerLabels)
	if err != nil {
		return "", nil, 0, stacktrace.Propagate(err, "An error occurred getting the engine container's object attributes schema private port")
	}

	privatePortObjAttrProto := objAttrPrivatePort.GetProtocol()
	privatePortDockerProto, found := objAttrsSchemaPortProtosToDockerPortProtos[privatePortObjAttrProto]
	if !found {
		return "", nil, 0, stacktrace.NewError(
			"No Docker protocol was defined for obj attr proto '%v'; this is a bug in Kurtosis",
			privatePortObjAttrProto,
		)
	}

	privatePortNumStr := fmt.Sprintf("%v", objAttrPrivatePort.GetNumber())
	dockerPrivatePort, err := nat.NewPort(
		privatePortDockerProto,
		privatePortNumStr,
	)
	if err != nil {
		return "", nil, 0, stacktrace.Propagate(
			err,
			"An error occurred creating the engine container Docker private port object from port number '%v' and protocol '%v', which is necessary for getting its host machine port bindings",
			privatePortNumStr,
			privatePortDockerProto,
		)
	}

	hostMachineEnginePortBinding, found := engineContainer.GetHostPortBindings()[dockerPrivatePort]
	if !found {
		return "", nil, 0, stacktrace.NewError("Found a Kurtosis engine server container, but it didn't have a host machine port binding - this is likely a Kurtosis bug")
	}

	hostMachineIpAddrStr := hostMachineEnginePortBinding.HostIP
	hostMachineIp := net.ParseIP(hostMachineIpAddrStr)
	if hostMachineIp == nil {
		return "", nil, 0, stacktrace.NewError("We got host machine IP '%v' for accessing the engine container, but it wasn't a valid IP", hostMachineIpAddrStr)
	}

	hostMachinePortNumStr := hostMachineEnginePortBinding.HostPort
	hostMachinePortNumUint64, err := strconv.ParseUint(hostMachinePortNumStr, hostMachinePortNumStrParsingBase, hostMachinePortNumStrParsingBits)
	if err != nil {
		return "", nil, 0, stacktrace.Propagate(
			err,
			"An error occurred parsing engine container host machine port num string '%v' using base '%v' and num bits '%v'",
			hostMachinePortNumStr,
			hostMachinePortNumStrParsingBase,
			hostMachinePortNumStrParsingBits,
		)
	}
	hostMachinePortNumUint16 := uint16(hostMachinePortNumUint64) // Okay to do due to specifying the number of bits above

	return "", hostMachineIp, hostMachinePortNumUint16, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
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
	for i := 0; i < maxWaitForAvailabilityRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunExecCommand(ctx, containerId, execCmd, outputBuffer)
		if err == nil {
			if exitCode == availabilityWaitingExecCmdSuccessExitCode {
				return nil
			}
			logrus.Debugf(
				"Engine server availability-waiting command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				availabilityWaitingExecCmdSuccessExitCode,
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
		if i < maxWaitForAvailabilityRetries {
			time.Sleep(timeBetweenWaitForAvailabilityRetries)
		}
	}

	return stacktrace.NewError(
		"The engine server didn't become available (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxWaitForAvailabilityRetries,
		timeBetweenWaitForAvailabilityRetries,
	)
}

func (kdb *KurtosisDockerBackendCore) cleanContainers(ctx context.Context, searchLabels map[string]string, shouldKillRunningContainers bool) ([]string, []error, error) {
	matchingContainers, err := kdb.dockerManager.GetContainersByLabels(
		ctx,
		searchLabels,
		shouldFetchStoppedContainersWhenDestroyingStoppedContainers,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting containers matching labels '%+v'", searchLabels)
	}

	containersToDestroy := []*types.Container{}
	for _, container := range matchingContainers {
		containerName := container.GetName()
		containerStatus := container.GetStatus()
		if shouldKillRunningContainers {
			containersToDestroy = append(containersToDestroy, container)
			continue
		}

		isRunning, err := container_status_calculator.IsContainerRunning(containerStatus)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred determining if container '%v' with status '%v' is running", containerName, containerStatus)
		}
		if !isRunning {
			containersToDestroy = append(containersToDestroy, container)
		}
	}

	successfullyDestroyedContainerNames := []string{}
	removeContainerErrors := []error{}
	for _, container := range containersToDestroy {
		containerId := container.GetId()
		containerName := container.GetName()
		if err := kdb.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred removing stopped container '%v'", containerName)
			removeContainerErrors = append(removeContainerErrors, wrappedErr)
			continue
		}
		successfullyDestroyedContainerNames = append(successfullyDestroyedContainerNames, containerName)
	}

	return successfullyDestroyedContainerNames, removeContainerErrors, nil

}

func getPrivateEnginePort(containerLabels map[string]string) (*schema.PortSpec, error) {
	serializedPortSpecs, found := containerLabels[schema.PortSpecsLabel]
	if found {
		portSpecs, err := schema.DeserializePortSpecs(serializedPortSpecs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing engine server port spec string '%v'", serializedPortSpecs)
		}
		portSpec, foundInternalPortId := portSpecs[schema.KurtosisInternalContainerGRPCPortID]
		if !foundInternalPortId {
			return nil, stacktrace.NewError("No Kurtosis-internal port ID '%v' found in the engine server port specs", schema.KurtosisInternalContainerGRPCPortID)
		}
		return portSpec, nil
	}

	// We can get rid of this after 2022-06-02, when we're confident no users will be running API containers with this label
	pre2021_12_02Port, err := getApiContainerPrivatePortUsingPre2021_12_02Label(containerLabels)
	if err == nil {
		return pre2021_12_02Port, nil
	} else {
		logrus.Debugf("An error occurred getting the engine container private port num using the pre-2021-12-02 label: %v", err)
	}

	// We can get rid of this after 2022-05-15, when we're confident no users will be running API containers with this label
	pre2021_11_15Port, err := schema.NewPortSpec(pre2021_11_15_portNum, pre2021_11_15_portProto)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't create engine private port spec using pre-2021-11-15 constants")
	}
	return pre2021_11_15Port, nil
}

func getApiContainerPrivatePortUsingPre2021_12_02Label(containerLabels map[string]string) (*schema.PortSpec, error) {
	// We can get rid of this after 2022-06-02, when we're confident no users will be running API containers with this label
	portNumStr, found := containerLabels[pre2021_12_02_portNumLabel]
	if !found {
		return nil, stacktrace.NewError("Couldn't get engine container private port using the pre-2021-12-02 label '%v' because it doesn't exist", pre2021_12_02_portNumLabel)
	}
	portNumUint64, err := strconv.ParseUint(portNumStr, pre2021_12_02_portNumBase, pre2021_12_02_portNumUintBits)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing pre-2021-12-02 private port num string '%v' to a uint16", portNumStr)
	}
	portNumUint16 := uint16(portNumUint64) // Safe to do because we pass in the number of bits to the ParseUint call above
	result, err := schema.NewPortSpec(portNumUint16, pre2021_12_02_portProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a new port spec using pre-2021-12-02 port num '%v' and protocol '%v'",
			portNumUint16,
			pre2021_12_02_portProtocol,
		)
	}
	return result, nil
}
