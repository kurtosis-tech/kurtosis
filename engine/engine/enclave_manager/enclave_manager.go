package enclave_manager

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-core/api_container_availability_waiter/api_container_availability_waiter_consts"
	"github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher_lib"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/enclave_manager/docker_network_allocator"
	enclave_manager_types "github.com/kurtosis-tech/kurtosis-engine-server/engine/enclave_manager/types"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	// TODO This should come from the Kurt Client constant!!!
	// The port that the API container listens on
	apiContainerListenPort = 7443

	// TODO This should come from the Kurt Client constant!!!
	// Protocol that the API container listens on
	apiContainerListenProtocol = "tcp"

	// The API container is responsible for disconnecting/stopping everything in its network when stopped, so we need
	//  to give it some time to do so
	apiContainerStopTimeout = 3 * time.Minute

	// This is set in the API container Dockerfile
	availabilityWaiterBinaryFilepath = "/run/api-container-availability-waiter"

	shouldFetchStoppedContainersWhenGettingAPIContainer = false

	// We set this to true in case there are any race conditions with a container starting as we're trying to stop the enclave
	shouldKillAlreadyStoppedContainersWhenStoppingEnclave = true

	shouldFetchStoppedContainersWhenDestroyingEnclave = true
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

func (manager *EnclaveManager) CreateEnclave(
	setupCtx context.Context,
	// TODO This shouldn't be passed as an argument, but should be auto-detected from the core API version!!!
	apiContainerImage string,
	apiContainerLogLevel logrus.Level,
	// TODO put in coreApiVersion as a param here!
	enclaveId string,
	isPartitioningEnabled bool,
	shouldPublishAllPorts bool) (*enclave_manager_types.Enclave, error) {

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

	if err := manager.dockerManager.CreateVolume(setupCtx, enclaveId); err != nil {
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
	apiContainerLabels := enclaveObjLabelsProvider.ForApiContainer(apiContainerIpAddr, apiContainerListenPort)

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
		apiContainerListenPort,
		apiContainerListenProtocol,
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

	// Everything started successfully, so the responsibility of deleting the network is now transferred to the caller
	shouldDeleteNetwork = false
	shouldStopApiContainer = false

	enclave := enclave_manager_types.NewEnclave(
		networkId,
		networkIpAndMask,
		apiContainerId,
		&apiContainerIpAddr,
		apiContainerHostPortBinding)

	return enclave, nil
}

func (manager *EnclaveManager) GetEnclave(ctx context.Context, enclaveId string) (*enclave_manager_types.Enclave, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	network, found, err := manager.getEnclaveNetwork(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting network for enclave ID '%v'", enclaveId)
	}
	if !found {
		return nil, stacktrace.NewError("No enclave with ID '%v' exists", enclaveId)
	}
	networkId := network.GetId()

	apiContainer, err := manager.getAPIContainer(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API container for enclave ID '%v'", enclaveId)
	}
	apiContainerId := apiContainer.GetId()

	apiContainerLabels := apiContainer.GetLabels()
	apiContainerIPAddressString, found := apiContainerLabels[enclave_object_labels.APIContainerIPLabel]
	if !found {
		return nil, stacktrace.NewError(
			"No '%v' container label was found on API container with ID '%v' with labels '%+v'",
			enclave_object_labels.APIContainerIPLabel,
			apiContainer.GetId(),
			apiContainerLabels,
		)
	}
	apiContainerIpAddr := net.ParseIP(apiContainerIPAddressString)
	if apiContainerIpAddr == nil {
		return nil, stacktrace.NewError("Couldn't parse API container IP address string '%v' to an IP address", apiContainerIPAddressString)
	}

	apiContainerListenPortString, found := apiContainerLabels[enclave_object_labels.APIContainerPortLabel]
	apiContainerNatPort, err := nat.NewPort(apiContainerListenProtocol, apiContainerListenPortString)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new API container port with protocol '%v' and port number '%v'", apiContainerListenProtocol, apiContainerListenPortString)
	}

	allApiContainerHostPortBindings := apiContainer.GetHostPortBindings()
	apiContainerHostPortBinding, found := allApiContainerHostPortBindings[apiContainerNatPort]
	if !found {
		return nil, stacktrace.NewError("No API container host port binding found for port '%v' among host port bindings '%+v'", apiContainerNatPort, allApiContainerHostPortBindings)
	}

	enclave := enclave_manager_types.NewEnclave(
		networkId,
		network.GetIpAndMask(),
		apiContainerId,
		&apiContainerIpAddr,
		apiContainerHostPortBinding)

	return enclave, nil
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

	// First, delete all containers in the network (which is necessary to delete the network)
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

func (manager *EnclaveManager) getAPIContainer(ctx context.Context, enclaveId string) (*types.Container, error) {
	labels := getLabelsForAPIContainer(enclaveId)
	containers, err := manager.dockerManager.GetContainersByLabels(ctx, labels, shouldFetchStoppedContainersWhenGettingAPIContainer)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API container by labels: '%+v'", labels)
	}
	if len(containers) == 0 || len(containers) > 1 {
		return nil, stacktrace.NewError("%v Docker container were returned for labels '%+v' and should be only one API container running on each enclave - this is very strange!", len(containers), labels)
	}

	apiContainer := containers[0]

	return apiContainer, nil
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

func getLabelsForAPIContainer(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeAPIContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	return labels
}
