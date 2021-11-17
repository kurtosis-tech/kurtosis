package enclave_manager

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/server/engine/enclave_manager/docker_network_allocator"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)
const (
	shouldFetchStoppedContainersWhenGettingEnclaveStatus = true

	// We set this to true in case there are any race conditions with a container starting as we're trying to stop the enclave
	shouldKillAlreadyStoppedContainersWhenStoppingEnclave = true

	shouldFetchStoppedContainersWhenDestroyingEnclave = true

	// API container port number string parsing constants
	portNumStrParsingBase = 10
	portNumStrParsingBits = 32

	allEnclavesDirname = "enclaves"

	apiContainerListenPortNumInsideNetwork = uint16(7443)

	// These are the old labels that the API container used to use before 2021-11-15 for declaring its port num protocol
	// We can get rid of this after 2022-05-15, when we're confident no users will be running API containers with the old label
	oldApiContainerPortNumLabel = "com.kurtosistech.api-container-port-number"
	oldApiContainerPortProtocolLabel = "com.kurtosistech.api-container-port-protocol"

	// NOTE: It's very important that all directories created inside the engine data directory are created with 0777
	//  permissions, because:
	//  a) the engine data directory is bind-mounted on the Docker host machine
	//  b) the engine container, and pretty much every Docker container, runs as 'root'
	//  c) the Docker host machine will not be running as root
	//  d) For the host machine to be able to read & write files inside the engine data directory, it needs to be able
	//      to access the directories inside the engine data directory
	// The longterm fix to this is probably to:
	//  1) make the engine server data a Docker volume
	//  2) have the engine server expose the engine data directory to the Docker host machine via some filesystem-sharing
	//      server, like NFS or CIFS
	// This way, we preserve the host machine's ability to write to services as if they were local on the filesystem, while
	//  actually having the data live inside a Docker volume. This also sets the stage for Kurtosis-as-a-Service (where bind-mount
	//  the engine data dirpath would be impossible).
	engineDataSubdirectoryPerms = 0777
)

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// We use Docker as our backing datastore, but it has tons of race conditions so we use this mutex to ensure
	//  enclave modifications are atomic
	mutex *sync.Mutex

	dockerManager *docker_manager.DockerManager

	objAttrsProvider schema.ObjectAttributesProvider

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator

	// We need this so that when the engine container starts enclaves & containers, it can tell Docker where the enclave
	//  data directory is on the host machine os Docker can bind-mount it in
	engineDataDirpathOnHostMachine string

	engineDataDirpathOnEngineContainer string
}

func NewEnclaveManager(
	dockerManager *docker_manager.DockerManager,
	objectAttributesProvider schema.ObjectAttributesProvider,
	engineDataDirpathOnHostMachine string,
	engineDataDirpathOnEngineContainer string,
) *EnclaveManager {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator(dockerManager)
	return &EnclaveManager{
		mutex:                              &sync.Mutex{},
		dockerManager:                      dockerManager,
		objAttrsProvider:                   objectAttributesProvider,
		dockerNetworkAllocator:             dockerNetworkAllocator,
		engineDataDirpathOnHostMachine:     engineDataDirpathOnHostMachine,
		engineDataDirpathOnEngineContainer: engineDataDirpathOnEngineContainer,
	}
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//  is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) CreateEnclave(
	setupCtx context.Context,
	// If blank, will use the default
	apiContainerImageVersionTag string,
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

	enclaveObjAttrsProvider := manager.objAttrsProvider.ForEnclave(enclaveId)
	enclaveNetworkAttrs := enclaveObjAttrsProvider.ForEnclaveNetwork()
	enclaveNetworkName := enclaveNetworkAttrs.GetName()
	enclaveNetworkLabels := enclaveNetworkAttrs.GetLabels()

	logrus.Debugf("Creating Docker network for enclave '%v'...", enclaveId)
	networkId, networkIpAndMask, gatewayIp, freeIpAddrTracker, err := manager.dockerNetworkAllocator.CreateNewNetwork(
		setupCtx,
		enclaveNetworkName,
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

	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	apiContainerIpAddr, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}

	apiContainerLauncher := api_container_launcher.NewApiContainerLauncher(
		manager.dockerManager,
		manager.objAttrsProvider,
	)
	var apiContainerId string
	var apiContainerHostPortBinding *nat.PortBinding
	var launchApiContainerErr error
	if apiContainerImageVersionTag == "" {
		apiContainerId, apiContainerHostPortBinding, launchApiContainerErr = apiContainerLauncher.LaunchWithDefaultVersion(
			setupCtx,
			apiContainerLogLevel,
			enclaveId,
			networkId,
			networkIpAndMask.String(),
			apiContainerListenPortNumInsideNetwork,
			gatewayIp,
			apiContainerIpAddr,
			isPartitioningEnabled,
			enclaveDataDirpathOnHostMachine,
		)
	} else {
		apiContainerId, apiContainerHostPortBinding, launchApiContainerErr = apiContainerLauncher.LaunchWithCustomVersion(
			setupCtx,
			apiContainerImageVersionTag,
			apiContainerLogLevel,
			enclaveId,
			networkId,
			networkIpAndMask.String(),
			apiContainerListenPortNumInsideNetwork,
			gatewayIp,
			apiContainerIpAddr,
			isPartitioningEnabled,
			enclaveDataDirpathOnHostMachine,
		)
	}
	if launchApiContainerErr != nil {
		return nil, stacktrace.Propagate(launchApiContainerErr, "An error occurred launching the API container")
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
		ApiContainerInfo: &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo{
			ContainerId:       apiContainerId,
			IpInsideEnclave:   apiContainerIpAddr.String(),
			PortInsideEnclave: uint32(apiContainerListenPortNumInsideNetwork),
		},
		ApiContainerHostMachineInfo: &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
			IpOnHostMachine:   apiContainerHostPortBinding.HostIP,
			PortOnHostMachine: hostMachinePortUint32,
		},
		EnclaveDataDirpathOnHostMachine: enclaveDataDirpathOnHostMachine,
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
	// TODO this is janky; we need a better way to find enclave networks!!! Ideally, we shouldn't actually know the labbel keys or values here
	kurtosisNetworkLabels := map[string]string{
		forever_constants.AppIDLabel: forever_constants.AppIDValue,
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

		enclaveDataDirpathOnHostMachine, _ := manager.getEnclaveDataDirpath(enclaveId)
		enclaveInfo := &kurtosis_engine_rpc_api_bindings.EnclaveInfo{
			EnclaveId:                       enclaveId,
			NetworkId:                       network.GetId(),
			NetworkCidr:                     network.GetIpAndMask().String(),
			ContainersStatus:                containersStatus,
			ApiContainerStatus:              apiContainerStatus,
			ApiContainerInfo:                apiContainerInfo,
			ApiContainerHostMachineInfo:     apiContainerHostMachineInfo,
			EnclaveDataDirpathOnHostMachine: enclaveDataDirpathOnHostMachine,
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
		schema.EnclaveIDContainerLabel: enclaveId,
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
		containerTypeLabelValue, found := containerLabels[forever_constants.ContainerTypeLabel]
		if found && containerTypeLabelValue == schema.ContainerTypeAPIContainer {
			if resultApiContainerInfo != nil {
				return 0, 0, nil, nil, stacktrace.NewError("Found a second API container inside the network; this should never happen!")
			}

			if isContainerRunning {
				resultApiContainerStatus = kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING
			} else {
				resultApiContainerStatus = kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_STOPPED
			}

			apiContainerIpInsideNetwork, found := containerLabels[schema.APIContainerIPLabel]
			if !found {
				return 0, 0, nil, nil, stacktrace.NewError(
					"No label '%v' was found on the API container indicating its IP inside the network",
					schema.APIContainerIPLabel,
				)
			}

			apiContainerPortNumStr, found := containerLabels[schema.PortNumLabel]
			if !found {
				// We can get rid of this after 2022-05-15, when we're confident no users will be running API containers with the old label
				maybeApiContainerPortNumStr, foundOldLabel := containerLabels[oldApiContainerPortNumLabel]
				if !foundOldLabel {
					return 0, 0, nil, nil, stacktrace.NewError(
						"Neither the current label '%v' nor old label '%v' were found on the API container, which is necessary for getting its host machine port bindings",
						schema.PortNumLabel,
						oldApiContainerPortNumLabel,
					)
				}
				apiContainerPortNumStr = maybeApiContainerPortNumStr
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
				apiContainerPortProtocol, found := containerLabels[schema.PortProtocolLabel]
				if !found {
					// We can get rid of this after 2022-05-15, when we're confident no users will be running API containers with the old label
					maybeApiContainerPortProtocol, foundOldLabel := containerLabels[oldApiContainerPortProtocolLabel]
					if !foundOldLabel {
						return 0, 0, nil, nil, stacktrace.NewError(
							"Neither the current label '%v' nor the old label '%v' was found on the API container, which is necessary for getting its host machine port bindings",
							schema.PortProtocolLabel,
							oldApiContainerPortProtocolLabel,
						)
					}
					apiContainerPortProtocol = maybeApiContainerPortProtocol
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
		schema.EnclaveIDContainerLabel: enclaveId,
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
		schema.EnclaveIDContainerLabel: enclaveId,
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
	if _, statErr := os.Stat(absoluteDirpath); os.IsNotExist(statErr) {
		if mkdirErr := os.Mkdir(absoluteDirpath, engineDataSubdirectoryPerms); mkdirErr != nil {
			return stacktrace.Propagate(
				mkdirErr,
				"Directory '%v' in the engine data volume didn't exist, and an error occurred trying to create it",
				absoluteDirpath)
		}
	}
	// This is necessary because the os.Mkdir might not create the directory with the perms that we want due to the umask
	// Chmod is not affected by the umask, so this will guarantee we get a directory with the perms that we want
	// NOTE: This has the added benefit of, if this directory already exists (due to the user running an old engine from
	//  before 2021-11-16), we'll correct the perms
	if err := os.Chmod(absoluteDirpath, engineDataSubdirectoryPerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred setting the permissions on directory '%v' to '%v'",
			absoluteDirpath,
			engineDataSubdirectoryPerms,
		)
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
