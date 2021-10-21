package enclave_manager

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core/api_container_availability_waiter/api_container_availability_waiter_consts"
	"github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher_lib"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/enclave_manager/docker_network_allocator"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)
// !!!!!!!!!!!!!!!!!!! WARNING WARNING WARNING WARNING WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// Be VERY careful modifying these! If you add a new label here, it's possible to leak Kurtosis resources:
//  1) the user creates an enclave using the old engine, and the network & volume get the old labels
//  2) the user upgrades their CLI, and restarts with the new engine
//  3) the new engine searches for enclaves/volumes using the new labels, and doesn't find the old network/volume
var enclaveNetworkLabels = map[string]string{
	enclave_object_labels.AppIDLabel: enclave_object_labels.AppIDValue,
}
var enclaveDataVolLabels = map[string]string{
	enclave_object_labels.AppIDLabel: enclave_object_labels.AppIDValue,
}
// !!!!!!!!!!!!!!!!!!! WARNING WARNING WARNING WARNING WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

const (
	// The API container is responsible for disconnecting/stopping everything in its network when stopped, so we need
	//  to give it some time to do so
	apiContainerStopTimeout = 3 * time.Minute

	// This is set in the API container Dockerfile
	availabilityWaiterBinaryFilepath = "/run/api-container-availability-waiter"

	shouldFetchStoppedContainersWhenGettingEnclaveStatus = true

	// We set this to true in case there are any race conditions with a container starting as we're trying to stop the enclave
	shouldKillAlreadyStoppedContainersWhenStoppingEnclave = true

	shouldFetchStoppedContainersWhenDestroyingEnclave = true

	// API container port number string parsing constants
	portNumStrParsingBase = 10
	portNumStrParsingBits = 32
)

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// We use Docker as our backing datastore, but it has tons of race conditions so we use this mutex to ensure
	//  enclave modifications are atomic
	mutex *sync.Mutex
	
	dockerManager *docker_manager.DockerManager

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator
}

func NewEnclaveManager(dockerManager *docker_manager.DockerManager) *EnclaveManager {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator(dockerManager)
	return &EnclaveManager{
		mutex:                  &sync.Mutex{},
		dockerManager:          dockerManager,
		dockerNetworkAllocator: dockerNetworkAllocator,
	}
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//  is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) CreateEnclave(
	setupCtx context.Context,
	// TODO This shouldn't be passed as an argument, but should be auto-detected from the core API version!!!
	apiContainerImage string,
	apiContainerLogLevel logrus.Level,
	// TODO put in coreApiVersion as a param here!
	enclaveId string,
	isPartitioningEnabled bool,
	shouldPublishAllPorts bool,
) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {

	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	_, found, err := manager.getEnclaveNetwork(setupCtx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for networks with name '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if found {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveId)
	}

	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)
	enclaveObjLabelsProvider := object_labels_providers.NewEnclaveObjectLabelsProvider(enclaveId)

	teardownCtx := context.Background()  // Separate context for tearing stuff down in case the input context is cancelled

	logrus.Debugf("Creating Docker network for enclave '%v'...", enclaveId)
	networkId, networkIpAndMask, gatewayIp, freeIpAddrTracker, err := manager.dockerNetworkAllocator.CreateNewNetwork(
		setupCtx,
		enclaveId,
		enclaveNetworkLabels,
	)
	if err != nil {
		// TODO If the user Ctrl-C's while the CreateNetwork call is ongoing then the CreateNetwork will error saying
		//  that the Context was cancelled as expected, but *the Docker engine will still create the network*!!! We'll
		//  need to parse the log message for the string "context canceled" and, if found, do another search for
		//  networks with our network name and delete them
		return nil, stacktrace.Propagate(err, "An error occurred allocating a new network for enclave '%v'", enclaveId)
	}
	shouldDeleteNetwork := true
	defer func() {
		if shouldDeleteNetwork {
			if err := manager.dockerManager.RemoveNetwork(teardownCtx, networkId); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to delete network '%v' that we created but an error was thrown:", networkId)
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove network with ID '%v'!!!!!!!", networkId)
			}
		}
	}()
	logrus.Debugf("Docker network '%v' created successfully with ID '%v' and subnet CIDR '%v'", enclaveId, networkId, networkIpAndMask.String())

	// TODO use hostnames rather than IPs, which makes things nicer and which we'll need for Docker swarm support
	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	apiContainerIpAddr, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}

	if err := manager.dockerManager.CreateVolume(setupCtx, enclaveId, enclaveDataVolLabels); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave volume '%v'", enclaveId)
	}
	// NOTE: We could defer a deletion of this volume unless the function completes successfully - right now, Kurtosis
	//  doesn't do any volume deletion

	// TODO We want to get rid of this; see the detailed TODO on EnclaveContext
	testsuiteContainerIpAddr, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't reserve an IP address for a possible testsuite container")
	}

	replContainerIpAddr, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't reserve an IP address for a possible REPL container")
	}

	apiContainerName := enclaveObjNameProvider.ForApiContainer()

	alreadyTakenIps := []net.IP{testsuiteContainerIpAddr, replContainerIpAddr}
	apiContainerLabels := enclaveObjLabelsProvider.ForApiContainer(apiContainerIpAddr)

	//Pulling latest image version
	if err = manager.dockerManager.PullImage(setupCtx, apiContainerImage); err != nil {
		logrus.Warnf("Failed to pull the latest version of image '%v'; you may be running an out-of-date version", apiContainerImage)
	}

	// TODO This shouldn't be hardcoded!!! We should instead detect the launch API version from the core API version
	launchApiVersion := uint(0)
	apiContainerLauncher, err := api_container_launcher_lib.GetAPIContainerLauncherForLaunchAPIVersion(
		launchApiVersion,
		manager.dockerManager,
		logrus.StandardLogger(),
		apiContainerImage,
		kurtosis_core_rpc_api_consts.ListenPort,
		kurtosis_core_rpc_api_consts.ListenProtocol,
		apiContainerLogLevel,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the API container launcher for launch API version '%v'", launchApiVersion)
	}

	apiContainerId, apiContainerHostPortBinding, err := apiContainerLauncher.Launch(
		setupCtx,
		apiContainerName,
		apiContainerLabels,
		enclaveId,
		networkId,
		networkIpAndMask.String(),
		gatewayIp,
		apiContainerIpAddr,
		alreadyTakenIps,
		isPartitioningEnabled,
		shouldPublishAllPorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}
	shouldStopApiContainer := true
	defer func() {
		if shouldStopApiContainer {
			if err := manager.dockerManager.StopContainer(teardownCtx, apiContainerId, apiContainerStopTimeout); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to stop the API container but an error was thrown:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop API container with ID '%v'", apiContainerId)
			}
		}
	}()

	if err := waitForApiContainerAvailability(setupCtx, manager.dockerManager, apiContainerId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container to become available")
	}

	hostMachinePortNumStr := apiContainerHostPortBinding.HostPort
	hostMachinePortUint32, err := parsePortNumStrToUint32(hostMachinePortNumStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the API container host machine port string '%v' to uint32", hostMachinePortNumStr)
	}

	result := &kurtosis_engine_rpc_api_bindings.EnclaveInfo{
		EnclaveId:          enclaveId,
		NetworkId:          networkId,
		NetworkCidr:        networkIpAndMask.String(),
		ContainersStatus:   kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING,
		ApiContainerStatus: kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING,
		ApiContainerInfo:   &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo{
			ContainerId:       apiContainerId,
			IpInsideEnclave:   apiContainerIpAddr.String(),
			PortInsideEnclave: kurtosis_core_rpc_api_consts.ListenPort,
			IpOnHostMachine:   apiContainerHostPortBinding.HostIP,
			PortOnHostMachine: hostMachinePortUint32,
		},
	}

	// Everything started successfully, so the responsibility of deleting the network is now transferred to the caller
	shouldDeleteNetwork = false
	shouldStopApiContainer = false
	return result, nil
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//  is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) GetEnclaves(
	ctx context.Context,
) (map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	networks, err := manager.dockerManager.GetNetworksByLabels(ctx, enclaveNetworkLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis networks")
	}
	result := map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	for _, network := range networks {
		enclaveId := network.GetName()
		// Container retrieval requires an extra call to the Docker engine per enclave, so therefore could be expensive
		//  if you have a LOT of enclaves. Maybe we want to make the getting of enclave container information be a separate
		//  engine server endpoint??
		containersStatus, apiContainerStatus, apiContainerInfo, err := getEnclaveContainerInformation(ctx, manager.dockerManager, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting information about the containers in enclave '%v'", enclaveId)
		}

		enclaveInfo := &kurtosis_engine_rpc_api_bindings.EnclaveInfo{
			EnclaveId:          enclaveId,
			NetworkId:          network.GetId(),
			NetworkCidr:        network.GetIpAndMask().String(),
			ContainersStatus:   containersStatus,
			ApiContainerStatus: apiContainerStatus,
			ApiContainerInfo:   apiContainerInfo,
		}
		result[enclaveId] = enclaveInfo
	}
	return result, nil
}

func (manager *EnclaveManager) StopEnclave(ctx context.Context, enclaveId string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	_, found, err := manager.getEnclaveNetwork(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking for the existence of a network for enclave '%v'", enclaveId)
	}
	if !found {
		return stacktrace.Propagate(err, "No enclave with ID '%v' exists", enclaveId)
	}

	enclaveContainerSearchLabels := map[string]string{
		enclave_object_labels.EnclaveIDContainerLabel: enclaveId,
	}
	allEnclaveContainers, err := manager.dockerManager.GetContainersByLabels(ctx, enclaveContainerSearchLabels, shouldKillAlreadyStoppedContainersWhenStoppingEnclave)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting containers for enclave '%v'", enclaveId)
	}

	// TODO Parallelize for perf
	containerKillErrorStrs := []string{}
	for _, enclaveContainer := range allEnclaveContainers {
		containerId := enclaveContainer.GetId()
		containerName := enclaveContainer.GetName()
		if err := manager.dockerManager.KillContainer(ctx, containerId); err != nil {
			wrappedContainerKillErr := stacktrace.Propagate(
				err,
				"An error occurred killing container '%v' with ID '%v'",
				containerName,
				containerId,
			)
			containerKillErrorStrs = append(
				containerKillErrorStrs,
				wrappedContainerKillErr.Error(),
			)
		}
	}

	if len(containerKillErrorStrs) > 0 {
		errorStr := strings.Join(containerKillErrorStrs, "\n\n")
		return stacktrace.NewError(
			"One or more errors occurred killing the containers in enclave '%v':\n%v",
			enclaveId,
			errorStr,
		)
	}

	// If all the kills went off successfully, wait for all the containers we just killed to definitively exit
	//  before we return
	containerWaitErrorStrs := []string{}
	for _, enclaveContainer := range allEnclaveContainers {
		containerName := enclaveContainer.GetName()
		containerId := enclaveContainer.GetId()
		if _, err := manager.dockerManager.WaitForExit(ctx, containerId); err != nil {
			wrappedContainerWaitErr := stacktrace.Propagate(
				err,
				"An error occurred waiting for container '%v' with ID '%v' to exit after killing",
				containerName,
				containerId,
			)
			containerWaitErrorStrs = append(
				containerWaitErrorStrs,
				wrappedContainerWaitErr.Error(),
			)
		}
	}

	if len(containerWaitErrorStrs) > 0 {
		errorStr := strings.Join(containerWaitErrorStrs, "\n\n")
		return stacktrace.NewError(
			"One or more errors occurred waiting for containers in enclave '%v' to exit after killing, meaning we can't guarantee the enclave is completely stopped:\n%v",
			enclaveId,
			errorStr,
		)
	}

	return nil
}

// Destroys an enclave, deleting all objects associated with it in the container engine (containers, volumes, networks, etc.)
func (manager *EnclaveManager) DestroyEnclave(ctx context.Context, enclaveId string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	enclaveNetwork, found, err := manager.getEnclaveNetwork(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking for a network for enclave '%v'", enclaveId)
	}
	if !found {
		return stacktrace.NewError("Cannot destroy enclave '%v' because no enclave with that ID exists", enclaveId)
	}

	if err := manager.StopEnclave(ctx, enclaveId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred stopping enclave with ID '%v', which is a prerequisite for destroying the enclave",
			enclaveId,
		)
	}

	// First, delete all enclave containers
	enclaveContainersSearchLabels := map[string]string{
		enclave_object_labels.EnclaveIDContainerLabel: enclaveId,
	}
	allEnclaveContainers, err := manager.dockerManager.GetContainersByLabels(
		ctx,
		enclaveContainersSearchLabels,
		shouldFetchStoppedContainersWhenDestroyingEnclave,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the containers in enclave '%v'", enclaveId)
	}
	removeContainerErrorStrs := []string{}
	for _, container := range allEnclaveContainers {
		containerName := container.GetName()
		containerId := container.GetId()
		if err := manager.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred removing container '%v' with ID '%v'",
				containerName,
				containerId,
			)
			removeContainerErrorStrs = append(
				removeContainerErrorStrs,
				wrappedErr.Error(),
			)
		}
	}
	if len(removeContainerErrorStrs) > 0 {
		return stacktrace.NewError(
			"An error occurred removing one or more containers in enclave '%v':\n%v",
			enclaveId,
			strings.Join(
				removeContainerErrorStrs,
				"\n\n",
			),
		)
	}

	// Next, remove the volume (if it exists)
	matchingVolumeNames, err := manager.dockerManager.GetVolumesByName(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking for volumes for enclave '%v'", enclaveId)
	}
	numMatchingVolumeNames := len(matchingVolumeNames)
	if numMatchingVolumeNames > 1 {
		return stacktrace.NewError(
			"Couldn't remove enclave volumes because we found %v volumes matching enclave '%v' when we expect just one; this is a bug in Kurtosis!",
			numMatchingVolumeNames,
			enclaveId,
		)
	}
	if numMatchingVolumeNames > 0 {
		enclaveVolumeName := matchingVolumeNames[0]
		if enclaveVolumeName != enclaveId {
			return stacktrace.NewError(
				"Couldn't remove volume for enclave ID '%v' because volume name '%v' doesn't match enclave ID; this is a Kurtosis bug",
				enclaveId,
				enclaveVolumeName,
			)
		}
		if err := manager.dockerManager.RemoveVolume(ctx, enclaveId); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing volume '%v' for enclave '%v'", enclaveVolumeName, enclaveId)
		}
	}

	// Finally, remove the network
	if err := manager.dockerManager.RemoveNetwork(ctx, enclaveNetwork.GetId()); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing the network for enclave '%v'", enclaveId)
	}

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
// There is a 1:1 mapping between Docker network and enclave - no network, no enclave, and vice versa
// We therefore use this function to check for the existence of an enclave, as well as get network info about existing enclaves
func (manager *EnclaveManager) getEnclaveNetwork(ctx context.Context, enclaveId string) (*types.Network, bool, error) {
	matchingNetworks, err := manager.dockerManager.GetNetworksByName(ctx, enclaveId)
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "An error occurred getting networks matching name '%v'", enclaveId)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks > 1 {
		return nil, false, stacktrace.NewError(
			"Found %v networks matching name '%v' when we expected just one - this is likely a bug in Kurtosis!",
			numMatchingNetworks,
			enclaveId,
		)
	}
	if numMatchingNetworks == 0 {
		return nil, false, nil
	}
	network := matchingNetworks[0]
	return network, true, nil
}

func waitForApiContainerAvailability(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	apiContainerId string) error {
	cmdOutputBuffer := &bytes.Buffer{}
	waitForAvailabilityExitCode, err := dockerManager.RunExecCommand(
		ctx,
		apiContainerId,
		[]string{availabilityWaiterBinaryFilepath},
		cmdOutputBuffer,
	)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred executing binary '%v' to wait for the API container to become available",
			availabilityWaiterBinaryFilepath,
		)
	}
	if waitForAvailabilityExitCode != api_container_availability_waiter_consts.SuccessExitCode {
		return stacktrace.NewError(
			"Expected API container availability waiter binary '%v' to return " +
				"success code %v, but got '%v' instead with the following log output:\n%v",
			availabilityWaiterBinaryFilepath,
			api_container_availability_waiter_consts.SuccessExitCode,
			waitForAvailabilityExitCode,
			cmdOutputBuffer.String(),
		)
	}
	return nil
}

func getEnclaveContainerInformation(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaveId string,
) (kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus, kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus, *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo, error) {
	containers, err := getEnclaveContainers(ctx, dockerManager, enclaveId)
	if err != nil {
		return 0, 0, nil, stacktrace.Propagate(err, "An error occurred getting the containers for enclave '%v'", enclaveId)
	}
	if len(containers) == 0 {
		return kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_EMPTY,
			kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_NONEXISTENT,
			nil,
			nil
	}

	resultContainersStatus := kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED
	resultApiContainerStatus := kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_NONEXISTENT
	var resultApiContainerInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo = nil
	for _, container := range containers {
		containerStatus := container.GetStatus()
		isContainerRunning := containerStatus == types.Running || containerStatus == types.Restarting
		if isContainerRunning {
			resultContainersStatus = kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING
		}

		// Parse API container info, if it exists
		containerLabels := container.GetLabels()
		containerTypeLabelValue, found := containerLabels[enclave_object_labels.ContainerTypeLabel]
		if found && containerTypeLabelValue == enclave_object_labels.ContainerTypeAPIContainer {
			if resultApiContainerInfo != nil {
				return 0, 0, nil, stacktrace.NewError("Found a second API container inside the network; this should never happen!")
			}

			if isContainerRunning {
				resultApiContainerStatus = kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING
			} else {
				resultApiContainerStatus = kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_STOPPED
			}

			apiContainerIpInsideNetwork, found := containerLabels[enclave_object_labels.APIContainerIPLabel]
			if !found {
				return 0, 0, nil, stacktrace.NewError(
					"No label '%v' was found on the API container indicating its IP inside the network",
					enclave_object_labels.APIContainerIPLabel,
				)
			}

			apiContainerPortNumStr, found := containerLabels[enclave_object_labels.APIContainerPortNumLabel]
			if !found {
				return 0, 0, nil, stacktrace.NewError(
					"No label '%v' was found on the API container, which is necessary for getting its host machine port bindings",
					enclave_object_labels.APIContainerPortNumLabel,
				)
			}

			apiContainerPortProtocol, found := containerLabels[enclave_object_labels.APIContainerPortProtocolLabel]
			if !found {
				return 0, 0, nil, stacktrace.NewError(
					"No label '%v' was found on the API container, which is necessary for getting its host machine port bindings",
					enclave_object_labels.APIContainerPortProtocolLabel,
				)
			}

			apiContainerPortObj, err := nat.NewPort(apiContainerPortProtocol, apiContainerPortNumStr)
			if err != nil {
				return 0, 0, nil, stacktrace.Propagate(
					err,
					"An error occurred creating the API container port object from port number '%v' and protocol '%v', which is necessary for getting its host machine port bindings",
					apiContainerPortNumStr,
					apiContainerPortProtocol,
				)
			}

			containerHostMachinePortBindings := container.GetHostPortBindings()
			apiContainerPortHostMachineBinding, found := containerHostMachinePortBindings[apiContainerPortObj]
			if !found {
				return 0, 0, nil, stacktrace.NewError(
					"No host machine port binding was found for API container port '%v'; this is a bug in Kurtosis!",
					apiContainerPortObj,
				)
			}

			apiContainerInternalPortNumUint32, err := parsePortNumStrToUint32(apiContainerPortNumStr)
			if err != nil {
				return 0, 0, nil, stacktrace.Propagate(err, "An error occurred converting the API container internal port string '%v' to uint32", apiContainerPortNumStr)
			}

			apiContainerHostMachinePortNumStr := apiContainerPortHostMachineBinding.HostPort
			apiContainerHostMachinePortNumUint32, err := parsePortNumStrToUint32(apiContainerHostMachinePortNumStr)
			if err != nil {
				return 0, 0, nil, stacktrace.Propagate(err, "An error occurred converting the API container host machine port string '%v' to uint32", apiContainerHostMachinePortNumStr)
			}

			resultApiContainerInfo = &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo{
				ContainerId:       container.GetId(),
				IpInsideEnclave:   apiContainerIpInsideNetwork,
				PortInsideEnclave: apiContainerInternalPortNumUint32,
				IpOnHostMachine:   apiContainerPortHostMachineBinding.HostIP,
				PortOnHostMachine: apiContainerHostMachinePortNumUint32,
			}
		}
	}

	return resultContainersStatus, resultApiContainerStatus, resultApiContainerInfo, nil
}

func getEnclaveContainers(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	enclaveId string,
) ([]*types.Container, error) {
	searchLabels := map[string]string{
		enclave_object_labels.EnclaveIDContainerLabel: enclaveId,
	}
	containers, err := dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchStoppedContainersWhenGettingEnclaveStatus)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the containers for enclave '%v' by labels '%+v'", enclaveId, searchLabels)
	}
	return containers, nil
}

func parsePortNumStrToUint32(input string) (uint32, error) {
	portNumUint64, err := strconv.ParseUint(input, portNumStrParsingBase, portNumStrParsingBits)
	if err != nil {
		return 0, stacktrace.Propagate(
			err,
			"An error occurred parsing port number string '%v' to an integer of base %v with %v bits",
			input,
			portNumStrParsingBase,
			portNumStrParsingBits,
		)
	}
	return uint32(portNumUint64), nil
}
