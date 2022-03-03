package docker

import (
	"errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	port_spec2 "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"strconv"
	"strings"
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
)

// This maps a Docker container's status to a binary "is the container considered running?" determiner
// Its completeness is enforced via unit test
var isContainerRunningDeterminer = map[types.ContainerStatus]bool{
	types.ContainerStatus_Paused:     false,
	types.ContainerStatus_Restarting: true,
	types.ContainerStatus_Running:    true,
	types.ContainerStatus_Removing:   true,
	types.ContainerStatus_Dead:       false,
	types.ContainerStatus_Created:    false,
	types.ContainerStatus_Exited:     false,
}

// Unfortunately, Docker doesn't have an enum for the protocols it supports, so we have to create this translation map
var portSpecProtosToDockerPortProtos = map[port_spec2.PortProtocol]string{
	port_spec2.PortProtocol_TCP:  "tcp",
	port_spec2.PortProtocol_SCTP: "sctp",
	port_spec2.PortProtocol_UDP:  "udp",
}

type DockerKurtosisBackend struct {
	dockerManager *docker_manager.DockerManager

	objAttrsProvider object_attributes_provider.DockerObjectAttributesProvider
}

func NewDockerKurtosisBackend(
	dockerManager *docker_manager.DockerManager,
) *DockerKurtosisBackend {
	return &DockerKurtosisBackend{
		dockerManager:    dockerManager,
		objAttrsProvider: object_attributes_provider.GetDockerObjectAttributesProvider(),
	}
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
func transformPortSpecToDockerPort(portSpec *port_spec2.PortSpec) (nat.Port, error) {
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


func getPublicPortBindingFromPrivatePortSpec(privatePortSpec *port_spec2.PortSpec, allHostMachinePortBindings map[nat.Port]*nat.PortBinding) (
	resultPublicIpAddr net.IP,
	resultPublicPortSpec *port_spec2.PortSpec,
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
	publicPortSpec, err := port_spec2.NewPortSpec(hostMachinePortNumUint16, portSpecProto)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating public port spec with host machine port num '%v' and protocol '%v'", hostMachinePortNumUint16, portSpecProto.String())
	}

	return hostMachineIp, publicPortSpec, nil
}
