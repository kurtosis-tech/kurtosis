package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/label_value_consts"
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

	// TODO UNCOMMENT THIS ONCE WE HAVE GRPC-PROXY WIRED UP!!
	/*
	if err := waitForEnginePortAvailability(ctx, backendCore.dockerManager, containerId, grpcProxyPortNum); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server's grpc proxy port to become available")
	}
	 */

	result, err := getEngineObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an engine object from container with ID '%v'", containerId)
	}

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
	successfulEngineIds map[string]bool,
	erroredEngineIds map[string]error,
	resultErr error,
) {
	matchingEnginesByContainerId, err := backendCore.getMatchingEnginesByContainerId(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	successIds := map[string]bool{}
	errorIds := map[string]error{}
	for containerId, engineObj := range matchingEnginesByContainerId {
		engineId := engineObj.GetID()
		if err := backendCore.dockerManager.StopContainer(ctx, containerId, engineStopTimeout); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred stopping engine '%v' with container ID '%v'", engineId, containerId)
			errorIds[engineId] = wrappedErr
		} else {
			successIds[engineId] = true
		}
	}
	return successIds, errorIds, nil
}

func (backendCore *DockerKurtosisBackendCore) DestroyEngines(
	ctx context.Context,
	filters *engine.GetEnginesFilters,
) (
	successfulEngineIds map[string]bool,
	erroredEngineIds map[string]error,
	resultErr error,
) {
	matchingEnginesByContainerId, err := backendCore.getMatchingEnginesByContainerId(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}

	successIds := map[string]bool{}
	errorIds := map[string]error{}
	for containerId, engineObj := range matchingEnginesByContainerId {
		engineId := engineObj.GetID()
		if err := backendCore.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred removing engine '%v' with container ID '%v'", engineId, containerId)
			errorIds[engineId] = wrappedErr
		} else {
			successIds[engineId] = true
		}
	}
	return successIds, errorIds, nil
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
	engineContainerSearchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():         label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.EngineContainerTypeLabelValue.GetString(),
		// NOTE: we do NOT use the engine ID label here, and instead do postfiltering, because Docker has no way to do disjunctive search!
	}
	allEngineContainers, err := backendCore.dockerManager.GetContainersByLabels(ctx, engineContainerSearchLabels, shouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching engine containers using labels: %+v", engineContainerSearchLabels)
	}

	allMatchingEngines := map[string]*engine.Engine{}
	for _, engineContainer := range allEngineContainers {
		containerId := engineContainer.GetId()
		engineObj, err := getEngineObjectFromContainerInfo(
			containerId,
			engineContainer.GetLabels(),
			engineContainer.GetStatus(),
			engineContainer.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with ID '%v' into an engine object", engineContainer.GetId())
		}

		// If the ID filter is specified, drop engines not matching it
		if filters.IDs != nil && len(filters.IDs) > 0 {
			if _, found := filters.IDs[engineObj.GetID()]; !found {
				continue
			}
		}

		// If status filter is specified, drop engines not matching it
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[engineObj.GetStatus()]; !found {
				continue
			}
		}

		allMatchingEngines[containerId] = engineObj
	}

	return allMatchingEngines, nil
}

func getEngineObjectFromContainerInfo(
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (*engine.Engine, error) {
	engineGuid, found := labels[label_key_consts.GUIDLabelKey.GetString()]
	if !found {
		// TODO Delete this after 2022-05-02 when we're confident there won't be any engines running
		//  without the engine ID label!
		engineGuid = containerId

		// TODO Uncomment this error after 2022-05-02 when we're confident there won't be any engines
		//  running without the engine ID label!
		/*
		return nil, stacktrace.NewError(
			"Expected a '%v' label on engine container with ID '%v', but none was found",
			object_attributes_provider.GUIDLabelKey.GetString(),
			containerId,
		)
		 */
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

		// TODO REMOVE THIS CONDITIONAL AFTER 2022-05-03 WHEN WE'RE CONFIDENT NOBODY IS USING ENGINES WITHOUT A PROXY
		if privateGrpcProxyPortSpec != nil {
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
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return nil, nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		// TODO AFTER 2022-05-02 SWITCH THIS TO A PLAIN ERROR WHEN WE'RE SURE NOBODY WILL BE USING THE OLD PORT SPEC STRING!
		oldPortSpecs, err := deserialize_pre_2022_03_02_PortSpecs(serializedPortSpecs)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v' even when trying the old method", serializedPortSpecs)
		}
		portSpecs = oldPortSpecs
	}

	grpcPortSpec, foundGrpcPort := portSpecs[kurtosisInternalContainerGrpcPortId]
	if !foundGrpcPort {
		return nil, nil, stacktrace.NewError("No engine grpc port with ID '%v' found in the engine server port specs", kurtosisInternalContainerGrpcPortId)
	}

	grpcProxyPortSpec, foundGrpcProxyPort := portSpecs[kurtosisInternalContainerGrpcProxyPortId]
	if !foundGrpcProxyPort {
		// TODO AFTER 2022-05-02 SWITCH THIS TO AN ERROR WHEN WE'RE SURE NOBODY WILL HAVE AN ENGINE WITHOUT THE PROXY
		grpcProxyPortSpec = nil
		// return nil, nil, stacktrace.NewError("No engine grpc-proxy port with ID '%v' found in the engine server port specs", kurtosisInternalContainerGrpcProxyPortId)
	}

	return grpcPortSpec, grpcProxyPortSpec, nil
}

// TODO DELETE THIS AFTER 2022-05-02, WHEN WE'RE CONFIDENT NO ENGINES WILL BE USING THE OLD PORT SPEC!
func deserialize_pre_2022_03_02_PortSpecs(specsStr string) (map[string]*port_spec.PortSpec, error) {
	const (
		portIdAndInfoSeparator      = "."
		portNumAndProtocolSeparator = "-"
		portSpecsSeparator          = "_"

		expectedNumPortIdAndSpecFragments      = 2
		expectedNumPortNumAndProtocolFragments = 2
		portUintBase                           = 10
		portUintBits                           = 16
	)

	result := map[string]*port_spec.PortSpec{}
	portIdAndSpecStrs := strings.Split(specsStr, portSpecsSeparator)
	for _, portIdAndSpecStr := range portIdAndSpecStrs {
		portIdAndSpecFragments := strings.Split(portIdAndSpecStr, portIdAndInfoSeparator)
		numPortIdAndSpecFragments := len(portIdAndSpecFragments)
		if numPortIdAndSpecFragments != expectedNumPortIdAndSpecFragments {
			return nil, stacktrace.NewError(
				"Expected splitting port ID & spec string '%v' to yield %v fragments but got %v",
				portIdAndSpecStr,
				expectedNumPortIdAndSpecFragments,
				numPortIdAndSpecFragments,
			)
		}
		portId := portIdAndSpecFragments[0]
		portSpecStr := portIdAndSpecFragments[1]

		portNumAndProtocolFragments := strings.Split(portSpecStr, portNumAndProtocolSeparator)
		numPortNumAndProtocolFragments := len(portNumAndProtocolFragments)
		if numPortNumAndProtocolFragments != expectedNumPortNumAndProtocolFragments {
			return nil, stacktrace.NewError(
				"Expected splitting port num & protocol string '%v' to yield %v fragments but got %v",
				portSpecStr,
				expectedNumPortIdAndSpecFragments,
				numPortIdAndSpecFragments,
			)
		}
		portNumStr := portNumAndProtocolFragments[0]
		portProtocolStr := portNumAndProtocolFragments[1]

		portNumUint64, err := strconv.ParseUint(portNumStr, portUintBase, portUintBits)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred parsing port num string '%v' to uint with base %v and %v bits",
				portNumStr,
				portUintBase,
				portUintBits,
			)
		}
		portNumUint16 := uint16(portNumUint64)
		portProtocol, err := port_spec.PortProtocolString(portProtocolStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting port protocol string '%v' to a port protocol enum", portProtocolStr)
		}

		portSpec, err := port_spec.NewPortSpec(portNumUint16, portProtocol)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating port spec object from ID & spec string '%v'",
				portIdAndSpecStr,
			)
		}

		result[portId] = portSpec
	}
	return result, nil
}
