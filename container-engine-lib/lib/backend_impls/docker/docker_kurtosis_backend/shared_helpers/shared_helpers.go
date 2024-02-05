package shared_helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	shouldGetStoppedContainersWhenGettingServiceInfo = true

	// Engine container port number string parsing constants
	hostMachinePortNumStrParsingBase = 10
	hostMachinePortNumStrParsingBits = 16

	netstatSuccessExitCode = 0

	// permissions and constants useful for dumping containers
	createdFilePerms                     = 0644
	shouldFollowContainerLogsWhenDumping = false
	containerSpecJsonSerializationIndent = "  "
	containerSpecJsonSerializationPrefix = ""
	containerInspectResultFilename       = "spec.json"
	containerLogsFilename                = "output.log"
	createdDirPerms                      = 0755
	numContainersToDumpAtOnce            = 20
)

// !!!WARNING!!!
// This files contains functions that are shared by multiple DockerKurtosisBackend functions.
// Generally, we want to prevent long utils folders with functionality that is difficult to find, so be careful
// when adding functionality in this folder.
// Things to think about: Could this function be a private helper function that's scope is smaller than you think?
// E.g. only used by start user services functions thus could go in start_user_services.go

// Unfortunately, Docker doesn't have an enum for the protocols it supports, so we have to create this translation map
var portSpecProtosToDockerPortProtos = map[port_spec.TransportProtocol]string{
	port_spec.TransportProtocol_TCP:  "tcp",
	port_spec.TransportProtocol_SCTP: "sctp",
	port_spec.TransportProtocol_UDP:  "udp",
}

// NOTE: Normally we'd have a "canonical" resource here, where that resource is always guaranteed to exist. For Kurtosis services,
// we want this to be the container engine's representation of a user service registration. Unfortunately, Docker has no way
// of representing a user service registration, so we store them in an in-memory map on the DockerKurtosisBackend. Therefore, an
// entry in that map is actually the canonical representation, which means that any of these fields could be nil!
type UserServiceDockerResources struct {
	ServiceContainer *types.Container

	// Will never be nil but may be empty if no expander volumes exist
	ExpanderVolumeNames []string
}

func GetEnclaveNetworkByEnclaveUuid(ctx context.Context, enclaveUuid enclave.EnclaveUUID, dockerManager *docker_manager.DockerManager) (*types.Network, error) {
	networkSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():       label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.EnclaveUUIDDockerLabelKey.GetString(): string(enclaveUuid),
	}

	enclaveNetworksFound, err := dockerManager.GetNetworksByLabels(ctx, networkSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Docker networks by enclave ID '%v'", enclaveUuid)
	}
	numMatchingNetworks := len(enclaveNetworksFound)
	if numMatchingNetworks == 0 {
		return nil, stacktrace.NewError("No network was found for enclave with ID '%v'", enclaveUuid)
	}
	if numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Expected exactly one network matching enclave ID '%v', but got %v",
			enclaveUuid,
			numMatchingNetworks,
		)
	}
	return enclaveNetworksFound[0], nil
}

func GetDockerPortFromPortSpec(portSpec *port_spec.PortSpec) (nat.Port, error) {
	var (
		dockerPort nat.Port
		err        error
	)
	portNum := portSpec.GetNumber()

	// Convert port spec protocol -> Docker protocol string
	portSpecProto := portSpec.GetTransportProtocol()
	portDockerProto, found := portSpecProtosToDockerPortProtos[portSpecProto]
	if !found {
		return dockerPort, stacktrace.NewError(
			"No Docker protocol was defined for port spec proto '%v'; this is a bug in Kurtosis",
			portSpecProto.String(),
		)
	}

	portNumStr := fmt.Sprintf("%v", portNum)
	dockerPort, err = nat.NewPort(
		portDockerProto,
		portNumStr,
	)
	if err != nil {
		return dockerPort, stacktrace.Propagate(
			err,
			"An error occurred creating the Docker port object from port number '%v' and protocol '%v'",
			portNumStr,
			portDockerProto,
		)
	}
	return dockerPort, nil
}

func GetPublicPortBindingFromPrivatePortSpec(privatePortSpec *port_spec.PortSpec, allHostMachinePortBindings map[nat.Port]*nat.PortBinding) (
	resultPublicIpAddr net.IP,
	resultPublicPortSpec *port_spec.PortSpec,
	resultErr error,
) {

	dockerPrivatePort, err := GetDockerPortFromPortSpec(privatePortSpec)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Docker private port object from private port spec '%+v', which is necessary for getting the corresponding host machine port bindings",
			privatePortSpec,
		)
	}

	hostMachinePortBinding, found := allHostMachinePortBindings[dockerPrivatePort]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No host machine port binding was specified for Docker port '%v' which corresponds to port spec with num '%v' and protocol '%v'",
			dockerPrivatePort,
			privatePortSpec.GetNumber(),
			privatePortSpec.GetTransportProtocol().String(),
		)
	}

	hostMachineIpAddrStr := hostMachinePortBinding.HostIP
	hostMachineIp := net.ParseIP(hostMachineIpAddrStr)
	if hostMachineIp == nil {
		return nil, nil, stacktrace.NewError(
			"Found host machine IP string '%v' for port spec with number '%v' and protocol '%v', but it wasn't a valid IP",
			hostMachineIpAddrStr,
			privatePortSpec.GetNumber(),
			privatePortSpec.GetTransportProtocol().String(),
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

	maybeApplicationProtocol := ""
	if privatePortSpec.GetMaybeApplicationProtocol() != nil {
		maybeApplicationProtocol = *privatePortSpec.GetMaybeApplicationProtocol()
	}
	publicPortSpec, err := port_spec.NewPortSpec(hostMachinePortNumUint16, privatePortSpec.GetTransportProtocol(), maybeApplicationProtocol, privatePortSpec.GetWait())
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err, "An error occurred creating public port spec with host machine port num '%v',  transport protocol '%v' and application protocol '%v'", hostMachinePortNumUint16, privatePortSpec.GetTransportProtocol().String(), maybeApplicationProtocol)
	}

	return hostMachineIp, publicPortSpec, nil
}

func TransformPortSpecToDockerPort(portSpec *port_spec.PortSpec) (nat.Port, error) {
	portSpecProto := portSpec.GetTransportProtocol()
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

// TODO Extract this to DockerKurtosisBackend and use it everywhere, for Engines and API containers?
func GetIpAndPortInfoFromContainer(
	containerName string,
	labels map[string]string,
	hostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (
	resultPrivateIp net.IP,
	resultPrivatePortSpecs map[string]*port_spec.PortSpec,
	resultPublicIp net.IP,
	resultPublicPortSpecs map[string]*port_spec.PortSpec,
	resultErr error,
) {
	privateIpAddrStr, found := labels[docker_label_key.PrivateIPDockerLabelKey.GetString()]
	if !found {
		return nil, nil, nil, nil, stacktrace.NewError("Expected to find label '%v' on container '%v' but label was missing", docker_label_key.PrivateIPDockerLabelKey.GetString(), containerName)
	}
	privateIp := net.ParseIP(privateIpAddrStr)
	if privateIp == nil {
		return nil, nil, nil, nil, stacktrace.NewError("Couldn't parse private IP string '%v' on container '%v' to an IP address", privateIpAddrStr, containerName)
	}

	serializedPortSpecs, found := labels[docker_label_key.PortSpecsDockerLabelKey.GetString()]
	if !found {
		return nil, nil, nil, nil, stacktrace.NewError(
			"Expected to find port specs label '%v' on container '%v' but none was found",
			containerName,
			docker_label_key.PortSpecsDockerLabelKey.GetString(),
		)
	}

	privatePortSpecs, err := docker_port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		if err != nil {
			return nil, nil, nil, nil, stacktrace.Propagate(err, "Couldn't deserialize port spec string '%v'", serializedPortSpecs)
		}
	}

	var containerPublicIp net.IP
	var publicPortSpecs map[string]*port_spec.PortSpec
	if len(hostMachinePortBindings) == 0 {
		return privateIp, privatePortSpecs, containerPublicIp, publicPortSpecs, nil
	}

	for portId, privatePortSpec := range privatePortSpecs {
		portPublicIp, publicPortSpec, err := GetPublicPortBindingFromPrivatePortSpec(privatePortSpec, hostMachinePortBindings)
		if err != nil {
			return nil, nil, nil, nil, stacktrace.Propagate(
				err,
				"An error occurred getting public port spec for private port '%v' with spec '%v/%v' on container '%v'",
				portId,
				privatePortSpec.GetNumber(),
				privatePortSpec.GetTransportProtocol().String(),
				containerName,
			)
		}

		if containerPublicIp == nil {
			containerPublicIp = portPublicIp
		} else {
			if !containerPublicIp.Equal(portPublicIp) {
				return nil, nil, nil, nil, stacktrace.NewError(
					"Private port '%v' on container '%v' yielded a public IP '%v', which doesn't agree with "+
						"previously-seen public IP '%v'",
					portId,
					containerName,
					portPublicIp.String(),
					containerPublicIp.String(),
				)
			}
		}

		if publicPortSpecs == nil {
			publicPortSpecs = map[string]*port_spec.PortSpec{}
		}
		publicPortSpecs[portId] = publicPortSpec
	}

	return privateIp, privatePortSpecs, containerPublicIp, publicPortSpecs, nil
}

// Gets the service objects & Docker resources for services matching the given filters
func GetMatchingUserServiceObjsAndDockerResourcesNoMutex(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	dockerManager *docker_manager.DockerManager,
) (
	map[service.ServiceUUID]*service.Service,
	map[service.ServiceUUID]*UserServiceDockerResources,
	error,
) {
	matchingDockerResources, err := getMatchingUserServiceDockerResources(ctx, enclaveId, filters.UUIDs, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting matching user service resources")
	}

	matchingServiceObjs, err := getUserServiceObjsFromDockerResources(enclaveId, matchingDockerResources)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis service objects from user service Docker resources")
	}

	resultServiceObjs := map[service.ServiceUUID]*service.Service{}
	resultDockerResources := map[service.ServiceUUID]*UserServiceDockerResources{}
	for uuid, serviceObj := range matchingServiceObjs {
		if filters.UUIDs != nil && len(filters.UUIDs) > 0 {
			if _, found := filters.UUIDs[serviceObj.GetRegistration().GetUUID()]; !found {
				continue
			}
		}

		if filters.Names != nil && len(filters.Names) > 0 {
			if _, found := filters.Names[serviceObj.GetRegistration().GetName()]; !found {
				continue
			}
		}

		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[serviceObj.GetContainer().GetStatus()]; !found {
				continue
			}
		}

		dockerResources, found := matchingDockerResources[uuid]
		if !found {
			// This should never happen; the Services map and the Docker resources maps should have the same UUIDs
			return nil, nil, stacktrace.Propagate(
				err,
				"Needed to return Docker resources for service with UUID '%v', but none was "+
					"found; this is a bug in Kurtosis",
				uuid,
			)
		}

		resultServiceObjs[uuid] = serviceObj
		resultDockerResources[uuid] = dockerResources
	}
	return resultServiceObjs, resultDockerResources, nil
}

func WaitForPortAvailabilityUsingNetstat(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	containerId string,
	portSpec *port_spec.PortSpec,
	maxRetries uint,
	timeBetweenRetries time.Duration,
) error {
	commandStr := fmt.Sprintf(
		"[ -n \"$(netstat -anp %v | grep LISTEN | grep %v)\" ]",
		strings.ToLower(portSpec.GetTransportProtocol().String()),
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

func GetEngineAndLogsComponentsNetwork(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
) (*types.Network, error) {
	return dockerManager.GetDefaultNetwork(ctx)
}

func DumpContainers(ctx context.Context, dockerManager *docker_manager.DockerManager, containers []*types.Container, outputDirpath string) error {
	// Create output directory
	if _, err := os.Stat(outputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", outputDirpath)
	}
	if err := os.Mkdir(outputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory at '%v'", outputDirpath)
	}

	workerPool := workerpool.New(numContainersToDumpAtOnce)
	resultErrsChan := make(chan error, len(containers))
	for _, container := range containers {
		containerName := container.GetName()
		containerId := container.GetId()
		logrus.Debugf("Submitting job to dump info about container with name '%v' and ID '%v'", containerName, containerId)

		/*
			!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			It's VERY important that the actual `func()` job function get created inside a helper function!!
			This is because variables declared inside for-loops are created BY REFERENCE rather than by-value, which
				means that if we inline the `func() {....}` creation here then all the job functions would get a REFERENCE to
				any variables they'd use.
			This means that by the time the job functions were run in the worker pool (long after the for-loop finished)
				then all the job functions would be using a reference from the last iteration of the for-loop.

			For more info, see the "Variables declared in for loops are passed by reference" section of:
				https://www.calhoun.io/gotchas-and-common-mistakes-with-closures-in-go/ for more details
			!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		*/
		jobToSubmit := createDumpContainerJob(
			ctx,
			dockerManager,
			outputDirpath,
			resultErrsChan,
			containerName,
			containerId,
		)

		workerPool.Submit(jobToSubmit)
	}
	workerPool.StopWait()
	close(resultErrsChan)

	allResultErrStrs := []string{}
	for resultErr := range resultErrsChan {
		allResultErrStrs = append(allResultErrStrs, resultErr.Error())
	}

	if len(allResultErrStrs) > 0 {
		allIndexedResultErrStrs := []string{}
		for idx, resultErrStr := range allResultErrStrs {
			indexedResultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR %v <<<<<<<<<<<<<<<<<\n%v", idx, resultErrStr)
			allIndexedResultErrStrs = append(allIndexedResultErrStrs, indexedResultErrStr)
		}

		// NOTE: We don't use stacktrace here because the actual stacktraces we care about are the ones from the threads!
		return fmt.Errorf("The following errors occurred when trying to dump information :\n%v",
			strings.Join(allIndexedResultErrStrs, "\n\n"))
	}

	return nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func getMatchingUserServiceDockerResources(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	maybeUuidsToMatch map[service.ServiceUUID]bool,
	dockerManager *docker_manager.DockerManager,
) (map[service.ServiceUUID]*UserServiceDockerResources, error) {
	result := map[service.ServiceUUID]*UserServiceDockerResources{}

	// Grab services, INDEPENDENT OF volumes
	userServiceContainerSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():         label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.EnclaveUUIDDockerLabelKey.GetString():   string(enclaveId),
		docker_label_key.ContainerTypeDockerLabelKey.GetString(): label_value_consts.UserServiceContainerTypeDockerLabelValue.GetString(),
	}
	userServiceContainers, err := dockerManager.GetContainersByLabels(ctx, userServiceContainerSearchLabels, shouldGetStoppedContainersWhenGettingServiceInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service containers in enclave '%v' by labels: %+v", enclaveId, userServiceContainerSearchLabels)
	}

	for _, container := range userServiceContainers {
		serviceUuidStr, found := container.GetLabels()[docker_label_key.GUIDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Found user service container '%v' that didn't have expected GUID label '%v'", container.GetId(), docker_label_key.GUIDDockerLabelKey.GetString())
		}
		serviceUuid := service.ServiceUUID(serviceUuidStr)

		if len(maybeUuidsToMatch) > 0 {
			if _, found := maybeUuidsToMatch[serviceUuid]; !found {
				continue
			}
		}

		resourceObj, found := result[serviceUuid]
		if !found {
			resourceObj = &UserServiceDockerResources{
				ServiceContainer:    nil,
				ExpanderVolumeNames: nil,
			}
		}
		resourceObj.ServiceContainer = container
		result[serviceUuid] = resourceObj
	}

	// Grab volumes, INDEPENDENT OF whether there are any containers
	filesArtifactExpansionVolumeSearchLabels := map[string]string{
		docker_label_key.AppIDDockerLabelKey.GetString():       label_value_consts.AppIDDockerLabelValue.GetString(),
		docker_label_key.EnclaveUUIDDockerLabelKey.GetString(): string(enclaveId),
		docker_label_key.VolumeTypeDockerLabelKey.GetString():  label_value_consts.FilesArtifactExpansionVolumeTypeDockerLabelValue.GetString(),
	}
	matchingFilesArtifactExpansionVolumes, err := dockerManager.GetVolumesByLabels(ctx, filesArtifactExpansionVolumeSearchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact expansion volumes in enclave '%v' by labels: %+v", enclaveId, filesArtifactExpansionVolumeSearchLabels)
	}

	for _, volume := range matchingFilesArtifactExpansionVolumes {
		serviceUuidStr, found := volume.Labels[docker_label_key.UserServiceGUIDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Found files artifact expansion volume '%v' that didn't have expected service GUID label '%v'", volume.Name, docker_label_key.UserServiceGUIDDockerLabelKey.GetString())
		}
		serviceUuid := service.ServiceUUID(serviceUuidStr)

		if len(maybeUuidsToMatch) > 0 {
			if _, found := maybeUuidsToMatch[serviceUuid]; !found {
				continue
			}
		}

		resourceObj, found := result[serviceUuid]
		if !found {
			resourceObj = &UserServiceDockerResources{
				ServiceContainer:    nil,
				ExpanderVolumeNames: nil,
			}
		}
		resourceObj.ExpanderVolumeNames = append(resourceObj.ExpanderVolumeNames, volume.Name)
		result[serviceUuid] = resourceObj
	}

	return result, nil
}

func getUserServiceObjsFromDockerResources(
	enclaveId enclave.EnclaveUUID,
	allDockerResources map[service.ServiceUUID]*UserServiceDockerResources,
) (map[service.ServiceUUID]*service.Service, error) {
	result := map[service.ServiceUUID]*service.Service{}

	// If we have an entry in the map, it means there's at least one Docker resource
	for serviceUuid, resources := range allDockerResources {
		serviceContainer := resources.ServiceContainer

		// If we don't have a container, we don't have the service ID label which means we can't actually construct a Service object
		// The only case where this would happen is if, during deletion, we delete the container but an error occurred deleting the volumes
		if serviceContainer == nil {
			return nil, stacktrace.NewError(
				"Service '%v' has Docker resources but not a container; this indicates that there the service's "+
					"container was deleted but errors occurred deleting the rest of the resources",
				serviceUuid,
			)
		}
		containerName := serviceContainer.GetName()
		containerLabels := serviceContainer.GetLabels()

		serviceIdStr, found := containerLabels[docker_label_key.IDDockerLabelKey.GetString()]
		if !found {
			return nil, stacktrace.NewError("Expected to find label '%v' on container '%v' but label was missing", docker_label_key.IDDockerLabelKey.GetString(), containerName)
		}
		serviceName := service.ServiceName(serviceIdStr)

		privateIp, privatePorts, maybePublicIp, maybePublicPorts, err := GetIpAndPortInfoFromContainer(
			containerName,
			containerLabels,
			serviceContainer.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting IP & port info from container '%v'", serviceContainer.GetName())
		}

		registration := service.NewServiceRegistration(
			serviceName,
			serviceUuid,
			enclaveId,
			privateIp,
			string(serviceName), // in Docker, hostname = serviceName because we're setting the "alias" of the container to serviceName
		)

		containerStatus := serviceContainer.GetStatus()
		isContainerRunning, found := consts.IsContainerRunningDeterminer[containerStatus]
		if !found {
			return nil, stacktrace.NewError("No is-running determination found for status '%v' for container '%v'", containerStatus.String(), containerName)
		}
		serviceContainerStatus := container.ContainerStatus_Stopped
		if isContainerRunning {
			serviceContainerStatus = container.ContainerStatus_Running
		}

		result[serviceUuid] = service.NewService(
			registration,
			privatePorts,
			maybePublicIp,
			maybePublicPorts,
			container.NewContainer(
				serviceContainerStatus,
				serviceContainer.GetImageName(),
				serviceContainer.GetEntrypointArgs(),
				serviceContainer.GetCmdArgs(),
				serviceContainer.GetEnvVars(),
			),
		)
	}
	return result, nil
}

func createDumpContainerJob(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	outputDirpath string,
	resultErrsChan chan error,
	containerName string,
	containerId string,
) func() {
	return func() {
		if err := dumpContainerInfo(ctx, dockerManager, outputDirpath, containerName, containerId); err != nil {
			resultErrsChan <- stacktrace.Propagate(
				err,
				"An error occurred dumping container info for container with name '%v' and ID '%v'",
				containerName,
				containerId,
			)
		}
	}
}

func dumpContainerInfo(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	outputDirpath string,
	containerName string,
	containerId string,
) error {
	// Make output directory
	containerOutputDirpath := path.Join(outputDirpath, containerName)
	if err := os.Mkdir(containerOutputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating directory '%v' to hold the output of container with name '%v' and ID '%v'",
			containerOutputDirpath,
			containerName,
			containerId,
		)
	}

	// Write container inspect results to file
	inspectResult, err := dockerManager.InspectContainer(
		ctx,
		containerId,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred inspecting container with ID '%v'", containerId)
	}
	jsonSerializedInspectResultBytes, err := json.MarshalIndent(inspectResult, containerSpecJsonSerializationPrefix, containerSpecJsonSerializationIndent)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing the results of inspecting container with ID '%v' to JSON", containerId)
	}
	specOutputFilepath := path.Join(containerOutputDirpath, containerInspectResultFilename)
	if err := os.WriteFile(specOutputFilepath, jsonSerializedInspectResultBytes, createdFilePerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred writing the inspect output of container with name '%v' and ID '%v' to file '%v'",
			containerName,
			containerId,
			specOutputFilepath,
		)
	}

	// Write container logs to file
	containerLogsReadCloser, err := dockerManager.GetContainerLogs(ctx, containerId, shouldFollowContainerLogsWhenDumping)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs for container with ID '%v'", containerId)
	}
	defer containerLogsReadCloser.Close()

	logsOutputFilepath := path.Join(containerOutputDirpath, containerLogsFilename)
	logsOutputFp, err := os.Create(logsOutputFilepath)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating file '%v' to hold the logs of container with name '%v' and ID '%v'",
			logsOutputFilepath,
			containerName,
			containerId,
		)
	}
	defer logsOutputFp.Close()

	// TODO Push this down into DockerManager as this is copied in multiple places!!! This check-if-the-container-is-TTY-and-use-io.Copy-if-so-and-stdcopy-if-not
	//  is copied straight from the Docker CLI, but it REALLY sucks that a Kurtosis dev magically needs to know that that's what
	//  they have to do if they want to read container logs
	// If we don't have this, reading the logs from a TTY container breaks
	if inspectResult.Config.Tty {
		if _, err := io.Copy(logsOutputFp, containerLogsReadCloser); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred copying the TTY container logs stream to file '%v' for container with name '%v' and ID '%v'",
				logsOutputFilepath,
				containerName,
				containerId,
			)
		}
	} else {
		if _, err := stdcopy.StdCopy(logsOutputFp, logsOutputFp, containerLogsReadCloser); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred copying the non-TTY container logs stream to file '%v' for container with name '%v' and ID '%v'",
				logsOutputFilepath,
				containerName,
				containerId,
			)
		}
	}

	return nil
}
