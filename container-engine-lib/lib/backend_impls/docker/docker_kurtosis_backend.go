package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_network_allocator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	// The Docker API's default is to return just containers whose status is "running"
	// However, we'd rather do our own filtering on what "running" means (because, e.g., "restarting"
	// should also be considered as running)
	shouldFetchAllContainersWhenRetrievingContainers = true

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

	netstatSuccessExitCode = 0

	uninitializedPublicIpAddrStrValue = ""

	dockerContainerPortNumUintBase = 10
	dockerContainerPortNumUintBits = 16
)

// This maps a Docker container's status to a binary "is the container considered running?" determiner
// Its completeness is enforced via unit test
var isContainerRunningDeterminer = map[types.ContainerStatus]bool{
	types.ContainerStatus_Paused:     false,
	types.ContainerStatus_Restarting: true,
	types.ContainerStatus_Running:    true,
	types.ContainerStatus_Removing:   false,
	types.ContainerStatus_Dead:       false,
	types.ContainerStatus_Created:    false,
	types.ContainerStatus_Exited:     false,
}

// Unfortunately, Docker doesn't have an enum for the protocols it supports, so we have to create this translation map
var portSpecProtosToDockerPortProtos = map[port_spec.PortProtocol]string{
	port_spec.PortProtocol_TCP:  "tcp",
	port_spec.PortProtocol_SCTP: "sctp",
	port_spec.PortProtocol_UDP:  "udp",
}

type DockerKurtosisBackend struct {
	dockerManager *docker_manager.DockerManager

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator

	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider
}

func NewDockerKurtosisBackend(dockerManager *docker_manager.DockerManager) *DockerKurtosisBackend {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator(dockerManager)
	return &DockerKurtosisBackend{
		dockerManager:    dockerManager,
		dockerNetworkAllocator: dockerNetworkAllocator,
		objAttrsProvider: object_attributes_provider.GetDockerObjectAttributesProvider(),
	}
}

func (backendCore *DockerKurtosisBackend) PullImage(image string) error {
	//TODO implement me
	panic("implement me")
}

// Engine methods in separate file

/*
func (backendCore *DockerKurtosisBackend) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	successfullyDestroyedContainerNames, containerDestructionErrors, err := backendCore.cleanContainers(ctx, engineLabels, shouldCleanRunningEngineContainers)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred cleaning stopped Kurtosis engine containers")
	}
	return successfullyDestroyedContainerNames, containerDestructionErrors, nil
}


func (backendCore *DockerKurtosisBackend) GetEnginePublicIPAndPort(
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
			hostMachineIpAddrStr,
			portNum,
			portSpecProto.String(),
		)
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

func waitForPortAvailabilityUsingNetstat(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	containerId string,
	portSpec *port_spec.PortSpec,
	maxRetries uint,
	timeBetweenRetries time.Duration,
) error {
	commandStr := fmt.Sprintf(
		"[ -n \"$(netstat -anp %v | grep LISTEN | grep %v)\" ]",
		strings.ToLower(portSpec.GetProtocol().String()),
		portSpec.GetNumber(),
	)
	execCmd := []string{
		"sh",
		"-c",
		commandStr,
	}
	for i := uint(0); i < maxRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunExecCommand(ctx, containerId, execCmd, outputBuffer)
		if err == nil {
			if exitCode == netstatSuccessExitCode {
				return nil
			}
			logrus.Debugf(
				"Netstat availability-waiting command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				netstatSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"Netstat availability-waiting command '%v' experienced a Docker error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxRetries {
			time.Sleep(timeBetweenRetries)
		}
	}

	return stacktrace.NewError(
		"The port didn't become available (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxRetries,
		timeBetweenRetries,
	)
}

func getUsedPortsFromPrivatePortSpecMapAndPortIdsForDockerPortObjs(privatePorts map[string]*port_spec.PortSpec) (map[nat.Port]docker_manager.PortPublishSpec, map[nat.Port]string, error) {
	publishSpecs := map[nat.Port]docker_manager.PortPublishSpec{}
	portIdsForDockerPortObjs := map[nat.Port]string{}
	for portId, portSpec := range privatePorts {
		dockerPort, err := transformPortSpecToDockerPort(portSpec)
		if err != nil {
			return nil, nil,  stacktrace.Propagate(err, "An error occurred transforming the '%+v' port spec to a Docker port", portSpec)
		}
		publishSpecs[dockerPort] =  docker_manager.NewManualPublishingSpec(portSpec.GetNumber())

		if preexistingPortId, found := portIdsForDockerPortObjs[dockerPort]; found {
			return nil, nil, stacktrace.NewError(
				"Port '%v' declares Docker port spec '%v', but this port spec is already in use by port '%v'",
				portId,
				dockerPort,
				preexistingPortId,
			)
		}
		portIdsForDockerPortObjs[dockerPort] = portId

	}
	return publishSpecs, portIdsForDockerPortObjs, nil
}

// condensePublicNetworkInfoFromHostMachineBindings
// Condenses declared private port bindings and the host machine port bindings returned by the container engine lib into:
//  1) a single host machine IP address
//  2) a map of private port binding IDs -> public ports
// An error is thrown if there are multiple host machine IP addresses
func condensePublicNetworkInfoFromHostMachineBindings(
	hostMachinePortBindings map[nat.Port]*nat.PortBinding,
	privatePorts map[string]*port_spec.PortSpec,
	portIdsForDockerPortObjs map[nat.Port]string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPorts map[string]*port_spec.PortSpec,
	resultErr error,
) {
	if len(hostMachinePortBindings) == 0 {
		return nil, nil, stacktrace.NewError("Cannot condense public network info if no host machine port bindings are provided")
	}

	publicIpAddrStr := uninitializedPublicIpAddrStrValue
	publicPorts := map[string]*port_spec.PortSpec{}
	for dockerPortObj, hostPortBinding := range hostMachinePortBindings {
		portId, found := portIdsForDockerPortObjs[dockerPortObj]
		if !found {
			// If the container engine reports a host port binding that wasn't declared in the input used-ports object, ignore it
			// This could happen if a port is declared in the Dockerfile
			continue
		}

		privatePort, found := privatePorts[portId]
		if !found {
			return nil,  nil, stacktrace.NewError(
				"The container engine returned a host machine port binding for Docker port spec '%v', but this port spec didn't correspond to any port ID; this is very likely a bug in Kurtosis",
				dockerPortObj,
			)
		}

		hostIpAddr := hostPortBinding.HostIP
		if publicIpAddrStr == uninitializedPublicIpAddrStrValue {
			publicIpAddrStr = hostIpAddr
		} else if publicIpAddrStr != hostIpAddr {
			return nil, nil, stacktrace.NewError(
				"A public IP address '%v' was already declared for the service, but Docker port object '%v' declares a different public IP address '%v'",
				publicIpAddrStr,
				dockerPortObj,
				hostIpAddr,
			)
		}

		hostPortStr := hostPortBinding.HostPort
		hostPortUint64, err := strconv.ParseUint(hostPortStr, dockerContainerPortNumUintBase, dockerContainerPortNumUintBits)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred parsing host machine port string '%v' into a uint with %v bits and base %v",
				hostPortStr,
				dockerContainerPortNumUintBits,
				dockerContainerPortNumUintBase,
			)
		}
		hostPortUint16 := uint16(hostPortUint64) // Safe to do because our ParseUint declares the expected number of bits
		portProtocol := privatePort.GetProtocol()

		portSpec, err := port_spec.NewPortSpec(hostPortUint16, portProtocol)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred creating a new public port spec object using number '%v' and protocol '%v'",
				hostPortUint16,
				portProtocol,
			)
		}

		publicPorts[portId] = portSpec
	}
	if publicIpAddrStr == uninitializedPublicIpAddrStrValue {
		return nil, nil, stacktrace.NewError("No public IP address string was retrieved from host port bindings: %+v", hostMachinePortBindings)
	}
	publicIpAddr := net.ParseIP(publicIpAddrStr)
	if publicIpAddr == nil {
		return nil, nil, stacktrace.NewError("Couldn't parse service's public IP address string '%v' to an IP object", publicIpAddrStr)
	}
	return publicIpAddr, publicPorts, nil
}

func getEnclaveIdFromNetwork(network *types.Network) (enclave.EnclaveID, error) {
	labels := network.GetLabels()
	enclaveIdLabelValue, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return "", stacktrace.NewError("Expected to find network's label with key '%v' but none was found", label_key_consts.EnclaveIDLabelKey.GetString())
	}
	enclaveId := enclave.EnclaveID(enclaveIdLabelValue)
	return enclaveId, nil
}

func hasEnclaveIdLabel(
	container *types.Container,
	enclaveId enclave.EnclaveID) bool {

	labels := container.GetLabels()
	enclaveIdLabelValue, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		//TODO Do all containers should have enclave ID label key??? we should return and error here if this answer is yes??
		logrus.Debugf("Container with ID '%v' haven't label '%v'", container.GetId(), label_key_consts.EnclaveIDLabelKey.GetString())
		return false
	}
	if enclaveIdLabelValue == string(enclaveId) {
		return true
	}
	return false
}

func hasGuidLabel(container *types.Container, guid string) bool {
	labels := container.GetLabels()
	guidLabelValue, found := labels[label_key_consts.GUIDLabelKey.GetString()]
	if !found {
		return false
	}
	if guidLabelValue == guid {
		return true
	}
	return false
}

func (backendCore *DockerKurtosisBackend) killContainerAndWaitForExit(
	ctx context.Context,
	containerId string,
) error {

	if err := backendCore.dockerManager.KillContainer(ctx, containerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred killing container with ID '%v'",
			containerId,
		)
	}
	if _, err := backendCore.dockerManager.WaitForExit(ctx, containerId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred waiting for container with ID '%v' to exit after killing it",
			containerId,
		)
	}

	return nil
}

func (backendCore *DockerKurtosisBackend) killContainers(
	ctx context.Context,
	containers []*types.Container,
)(
	successfulContainers map[string]bool,
	erroredContainers map[string]error,
){

	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if err := backendCore.dockerManager.KillContainer(ctx, containerId); err != nil {
			containerError :=  stacktrace.Propagate(
				err,
				"An error occurred killing container '%v' with ID '%v'",
				container.GetName(),
				containerId,
			)
			erroredContainers[container.GetId()] = containerError
			continue
		}
		successfulContainers[containerId] = true
	}

	return successfulContainers, erroredContainers
}

func (backendCore *DockerKurtosisBackend) waitForContainerExits(
	ctx context.Context,
	containers []*types.Container,
)(
	successfulContainers map[string]bool,
	erroredContainers map[string]error,
){
	// TODO Parallelize for perf
	for _, container := range containers {
		containerId := container.GetId()
		if _, err := backendCore.dockerManager.WaitForExit(ctx, containerId); err != nil {
			containerError := stacktrace.Propagate(
				err,
				"An error occurred waiting for container '%v' with ID '%v' to exit",
				container.GetName(),
				containerId,
			)
			erroredContainers[container.GetId()] = containerError
			continue
		}
		successfulContainers[containerId] = true
	}

	return successfulContainers, erroredContainers
}

func (backendCore *DockerKurtosisBackend) removeContainers(
	ctx context.Context,
	containerIds []string,
)(
	map[string]bool,
	map[string]error,
){
	successfulContainers := map[string]bool{}
	erroredContainers := map[string]error{}
	// TODO Parallelize for perf
	for _, containerId := range containerIds {
		if err := backendCore.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			containerError := stacktrace.Propagate(
				err,
				"An error occurred removing container with ID '%v'",
				containerId,
			)
			erroredContainers[containerId] = containerError
			continue
		}
		successfulContainers[containerId] = true
	}

	return successfulContainers, erroredContainers
}


func getPrivatePortsFromContainerLabels(containerLabels map[string]string) (map[string]*port_spec.PortSpec, error) {
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return  nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		// TODO AFTER 2022-05-02 SWITCH THIS TO A PLAIN ERROR WHEN WE'RE SURE NOBODY WILL BE USING THE OLD PORT SPEC STRING!
		oldPortSpecs, err := deserialize_pre_2022_03_02_PortSpecs(serializedPortSpecs)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v' even when trying the old method", serializedPortSpecs)
		}
		portSpecs = oldPortSpecs
	}

	return portSpecs, nil
}
