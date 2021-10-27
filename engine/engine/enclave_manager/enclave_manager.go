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
	"github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/enclave_manager/docker_network_allocator"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)
const (
	// This is set in the API container Dockerfile
	availabilityWaiterBinaryFilepath = "/run/api-container-availability-waiter"

	shouldFetchStoppedContainersWhenGettingEnclaveStatus = true

	// We set this to true in case there are any race conditions with a container starting as we're trying to stop the enclave
	shouldKillAlreadyStoppedContainersWhenStoppingEnclave = true

	shouldFetchStoppedContainersWhenDestroyingEnclave = true

	// API container port number string parsing constants
	portNumStrParsingBase = 10
	portNumStrParsingBits = 32

	allEnclavesDirname = "enclaves"
)

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// We use Docker as our backing datastore, but it has tons of race conditions so we use this mutex to ensure
	//  enclave modifications are atomic
	mutex *sync.Mutex
	
	dockerManager *docker_manager.DockerManager

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator

	// We need this so that when the engine container starts enclaves & containers, it can tell Docker where the enclave
	//  data directory is on the host machine os Docker can bind-mount it in
	engineDataDirpathOnHostMachine string

	engineDataDirpathOnEngineContainer string
}

func NewEnclaveManager(
	dockerManager *docker_manager.DockerManager,
	engineDataDirpathOnHostMachine string,
	engineDataDirpathOnEngineContainer string,
) *EnclaveManager {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator(dockerManager)
	return &EnclaveManager{
		mutex:                  &sync.Mutex{},
		dockerManager:          dockerManager,
		dockerNetworkAllocator: dockerNetworkAllocator,
		engineDataDirpathOnHostMachine: engineDataDirpathOnHostMachine,
		engineDataDirpathOnEngineContainer: engineDataDirpathOnEngineContainer,
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

	teardownCtx := context.Background()  // Separate context for tearing stuff down in case the input context is cancelled

	_, found, err := manager.getEnclaveNetwork(setupCtx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for networks with name '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if found {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveId)
	}

	_, allEnclavesDirpathOnEngineContainer := manager.getAllEnclavesDirpaths()
	if err := ensureDirpathExists(allEnclavesDirpathOnEngineContainer); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring enclaves directory '%v' exists", allEnclavesDirpathOnEngineContainer)
	}

	enclaveDataDirpathOnHostMachine, enclaveDataDirpathOnEngineContainer := manager.getEnclaveDataDirpath(enclaveId)
	if _, err := os.Stat(enclaveDataDirpathOnEngineContainer); err == nil {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave data directory already exists at '%v'", enclaveId, enclaveDataDirpathOnEngineContainer)
	}
	if err := ensureDirpathExists(enclaveDataDirpathOnEngineContainer); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring enclave data directory '%v' exists", enclaveDataDirpathOnEngineContainer)
	}
	shouldDeleteEnclaveDataDir := true
	defer func() {
		if shouldDeleteEnclaveDataDir {
			if err := os.RemoveAll(enclaveDataDirpathOnEngineContainer); err != nil {
				// 'clean' will remove this dangling directories, so this is a warn rather than an error
				logrus.Warnf("Enclave creation didn't complete successfully so we tried to remove the enclave data directory '%v', but removing threw the following error; this directory will stay around until a clean is run", enclaveDataDirpathOnEngineContainer)
				fmt.Fprintln(logrus.StandardLogger().Out, err)
			}
		}
	}()

	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)
	enclaveObjLabelsProvider := object_labels_providers.NewEnclaveObjectLabelsProvider(enclaveId)

	logrus.Debugf("Creating Docker network for enclave '%v'...", enclaveId)
	networkId, networkIpAndMask, gatewayIp, freeIpAddrTracker, err := manager.dockerNetworkAllocator.CreateNewNetwork(
		setupCtx,
		enclaveId,
		enclaveObjLabelsProvider.ForEnclaveNetwork(),
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

	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	apiContainerIpAddr, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}

	apiContainerName := enclaveObjNameProvider.ForApiContainer()

	//Pulling latest image version
	if err = manager.dockerManager.PullImage(setupCtx, apiContainerImage); err != nil {
		logrus.Warnf("Failed to pull the latest version of image '%v'; you may be running an out-of-date version", apiContainerImage)
	}

	apiContainerLauncher := api_container_launcher.NewApiContainerLauncher(
		manager.dockerManager,
	)
	apiContainerId, apiContainerHostPortBinding, err := apiContainerLauncher.Launch(
		setupCtx,
		apiContainerImage,
		apiContainerName,
		apiContainerLogLevel,
		enclaveId,
		networkId,
		networkIpAndMask.String(),
		gatewayIp,
		apiContainerIpAddr,
		[]net.IP{}, // TODO remove this, as we don't need it anymore now that we have the RegisterExternalContainer endpoint
		isPartitioningEnabled,
		enclaveDataDirpathOnHostMachine,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}
	shouldStopApiContainer := true
	defer func() {
		if shouldStopApiContainer {
			if err := manager.dockerManager.KillContainer(teardownCtx, apiContainerId); err != nil {
				logrus.Errorf("Creating the enclave didn't complete successfully, so we tried to kill the API container but an error was thrown:")
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
		},
		ApiContainerHostMachineInfo: &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
			IpOnHostMachine:   apiContainerHostPortBinding.HostIP,
			PortOnHostMachine: hostMachinePortUint32,
		},
	}

	// Everything started successfully, so the responsibility of deleting the enclave is now transferred to the caller
	shouldDeleteEnclaveDataDir = false
	shouldDeleteNetwork = false
	shouldStopApiContainer = false
	return result, nil
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//  is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) GetEnclaves(
	ctx context.Context,
) (map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	// TODO this is janky; we need a better way to find enclave networks!!! Ideally, we shouldn't actually know the label keys or values here
	kurtosisNetworkLabels := map[string]string{
		enclave_object_labels.AppIDLabel: enclave_object_labels.AppIDValue,
	}

	networks, err := manager.dockerManager.GetNetworksByLabels(ctx, kurtosisNetworkLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis networks")
	}
	result := map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	for _, network := range networks {
		enclaveId := network.GetName()
		// Container retrieval requires an extra call to the Docker engine per enclave, so therefore could be expensive
		//  if you have a LOT of enclaves. Maybe we want to make the getting of enclave container information be a separate
		//  engine server endpoint??
		containersStatus, apiContainerStatus, apiContainerInfo, apiContainerHostMachineInfo, err := getEnclaveContainerInformation(ctx, manager.dockerManager, enclaveId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting information about the containers in enclave '%v'", enclaveId)
		}

		enclaveInfo := &kurtosis_engine_rpc_api_bindings.EnclaveInfo{
			EnclaveId:                   enclaveId,
			NetworkId:                   network.GetId(),
			NetworkCidr:                 network.GetIpAndMask().String(),
			ContainersStatus:            containersStatus,
			ApiContainerStatus:          apiContainerStatus,
			ApiContainerInfo:            apiContainerInfo,
			ApiContainerHostMachineInfo: apiContainerHostMachineInfo,
		}
		result[enclaveId] = enclaveInfo
	}
	return result, nil
}

func (manager *EnclaveManager) StopEnclave(ctx context.Context, enclaveId string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if err := manager.stopEnclaveWithoutMutex(ctx, enclaveId); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveId)
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

	if err := manager.stopEnclaveWithoutMutex(ctx, enclaveId); err != nil {
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

	// Next, remove the enclave data dir (if it exists)
	_, enclaveDataDirpathOnEngineContainer := manager.getEnclaveDataDirpath(enclaveId)
	if _, statErr := os.Stat(enclaveDataDirpathOnEngineContainer); !os.IsNotExist(statErr) {
		if removeErr := os.RemoveAll(enclaveDataDirpathOnEngineContainer); removeErr != nil {
			return stacktrace.Propagate(removeErr, "An error occurred removing enclave data dir '%v' on engine container", enclaveDataDirpathOnEngineContainer)
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
	logrus.Debugf("Found %v networks matching name '%v': %+v", numMatchingNetworks, enclaveId, matchingNetworks)
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
) (
	kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus,
	kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus,
	*kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo,
	*kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo,
	error,
) {
	containers, err := getEnclaveContainers(ctx, dockerManager, enclaveId)
	if err != nil {
		return 0, 0, nil, nil, stacktrace.Propagate(err, "An error occurred getting the containers for enclave '%v'", enclaveId)
	}
	if len(containers) == 0 {
		return kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_EMPTY,
			kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_NONEXISTENT,
			nil,
			nil,
			nil
	}

	resultContainersStatus := kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED
	resultApiContainerStatus := kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_NONEXISTENT
	var resultApiContainerInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo = nil
	var resultApiContainerHostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo = nil
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
				return 0, 0, nil, nil, stacktrace.NewError("Found a second API container inside the network; this should never happen!")
			}

			if isContainerRunning {
				resultApiContainerStatus = kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING
			} else {
				resultApiContainerStatus = kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_STOPPED
			}

			apiContainerIpInsideNetwork, found := containerLabels[enclave_object_labels.APIContainerIPLabel]
			if !found {
				return 0, 0, nil, nil, stacktrace.NewError(
					"No label '%v' was found on the API container indicating its IP inside the network",
					enclave_object_labels.APIContainerIPLabel,
				)
			}

			apiContainerPortNumStr, found := containerLabels[enclave_object_labels.APIContainerPortNumLabel]
			if !found {
				return 0, 0, nil, nil, stacktrace.NewError(
					"No label '%v' was found on the API container, which is necessary for getting its host machine port bindings",
					enclave_object_labels.APIContainerPortNumLabel,
				)
			}

			apiContainerInternalPortNumUint32, err := parsePortNumStrToUint32(apiContainerPortNumStr)
			if err != nil {
				return 0, 0, nil, nil, stacktrace.Propagate(err, "An error occurred converting the API container internal port string '%v' to uint32", apiContainerPortNumStr)
			}

			resultApiContainerInfo = &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo{
				ContainerId:       container.GetId(),
				IpInsideEnclave:   apiContainerIpInsideNetwork,
				PortInsideEnclave: apiContainerInternalPortNumUint32,
			}

			// We only get host machine info if the container is running
			if isContainerRunning {
				apiContainerPortProtocol, found := containerLabels[enclave_object_labels.APIContainerPortProtocolLabel]
				if !found {
					return 0, 0, nil, nil, stacktrace.NewError(
						"No label '%v' was found on the API container, which is necessary for getting its host machine port bindings",
						enclave_object_labels.APIContainerPortProtocolLabel,
					)
				}

				apiContainerPortObj, err := nat.NewPort(apiContainerPortProtocol, apiContainerPortNumStr)
				if err != nil {
					return 0, 0, nil, nil, stacktrace.Propagate(
						err,
						"An error occurred creating the API container port object from port number '%v' and protocol '%v', which is necessary for getting its host machine port bindings",
						apiContainerPortNumStr,
						apiContainerPortProtocol,
					)
				}

				containerHostMachinePortBindings := container.GetHostPortBindings()
				apiContainerPortHostMachineBinding, found := containerHostMachinePortBindings[apiContainerPortObj]
				if !found {
					return 0, 0, nil, nil, stacktrace.NewError(
						"No host machine port binding was found for API container port '%v'; this is a bug in Kurtosis!",
						apiContainerPortObj,
					)
				}

				apiContainerHostMachinePortNumStr := apiContainerPortHostMachineBinding.HostPort
				apiContainerHostMachinePortNumUint32, err := parsePortNumStrToUint32(apiContainerHostMachinePortNumStr)
				if err != nil {
					return 0, 0, nil, nil, stacktrace.Propagate(err, "An error occurred converting the API container host machine port string '%v' to uint32", apiContainerHostMachinePortNumStr)
				}

				resultApiContainerHostMachineInfo = &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
					IpOnHostMachine:   apiContainerPortHostMachineBinding.HostIP,
					PortOnHostMachine: apiContainerHostMachinePortNumUint32,
				}
			}
		}
	}

	return resultContainersStatus, resultApiContainerStatus, resultApiContainerInfo, resultApiContainerHostMachineInfo, nil
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

// Both StopEnclave and DestroyEnclave need to be able to stop enclaves, but both have a mutex guard. Because Go mutexes
//  aren't reentrant, DestroyEnclave can't just call StopEnclave so we use this helper function
func (manager *EnclaveManager) stopEnclaveWithoutMutex(ctx context.Context, enclaveId string) error {
	_, found, err := manager.getEnclaveNetwork(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking for the existence of a network for enclave '%v'", enclaveId)
	}
	if !found {
		return stacktrace.NewError("No enclave with ID '%v' exists", enclaveId)
	}

	enclaveContainerSearchLabels := map[string]string{
		enclave_object_labels.EnclaveIDContainerLabel: enclaveId,
	}
	allEnclaveContainers, err := manager.dockerManager.GetContainersByLabels(ctx, enclaveContainerSearchLabels, shouldKillAlreadyStoppedContainersWhenStoppingEnclave)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting containers for enclave '%v'", enclaveId)
	}
	logrus.Debugf("Containers in enclave '%v' that will be killed: %+v", enclaveId, allEnclaveContainers)

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

// TODO This is copied from Kurt Core; merge these (likely into an EngineDataVolume object)!!!
func ensureDirpathExists(absoluteDirpath string) error {
	if _, err := os.Stat(absoluteDirpath); os.IsNotExist(err) {
		if err := os.Mkdir(absoluteDirpath, 0777); err != nil {
			return stacktrace.Propagate(
				err,
				"Directory '%v' in the engine data volume didn't exist, and an error occurred trying to create it",
				absoluteDirpath)
		}
	}
	return nil
}

func (manager *EnclaveManager) getAllEnclavesDirpaths() (onHostMachine string, onEngineContainer string) {
	onHostMachine = path.Join(
		manager.engineDataDirpathOnHostMachine,
		allEnclavesDirname,
	)
	onEngineContainer = path.Join(
		manager.engineDataDirpathOnEngineContainer,
		allEnclavesDirname,
	)
	return
}

func (manager *EnclaveManager) getEnclaveDataDirpath(enclaveId string) (onHostMachine string, onEngineContainer string) {
	allEnclavesOnHostMachine, allEnclavesOnEngineContainer := manager.getAllEnclavesDirpaths()
	onHostMachine = path.Join(
		allEnclavesOnHostMachine,
		enclaveId,
	)
	onEngineContainer = path.Join(
		allEnclavesOnEngineContainer,
		enclaveId,
	)
	return
}
