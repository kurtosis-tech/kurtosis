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
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
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

	shouldFetchStoppedContainers = false
)

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	dockerManager *docker_manager.DockerManager

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator
}

func NewEnclaveManager(dockerManager *docker_manager.DockerManager) *EnclaveManager {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator(dockerManager)
	return &EnclaveManager{
		dockerManager:           dockerManager,
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
	shouldPublishAllPorts bool) (string, *net.IPNet, string, *net.IP, *nat.PortBinding, error) {

	matchingNetworks, err := manager.dockerManager.GetNetworksByName(setupCtx, enclaveId)
	if err != nil {
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "An error occurred finding enclaves with name '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if len(matchingNetworks) > 0 {
		return "", nil, "", nil, nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveId)
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
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "An error occurred allocating a new network for enclave '%v'", enclaveId)
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
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}

	if err := manager.dockerManager.CreateVolume(setupCtx, enclaveId); err != nil {
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "An error occurred creating enclave volume '%v'", enclaveId)
	}
	// NOTE: We could defer a deletion of this volume unless the function completes successfully - right now, Kurtosis
	//  doesn't do any volume deletion

	// TODO We want to get rid of this; see the detailed TODO on EnclaveContext
	testsuiteContainerIpAddr, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "Couldn't reserve an IP address for a possible testsuite container")
	}

	replContainerIpAddr, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "Couldn't reserve an IP address for a possible REPL container")
	}

	apiContainerName := enclaveObjNameProvider.ForApiContainer()

	alreadyTakenIps := []net.IP{testsuiteContainerIpAddr, replContainerIpAddr}
	apiContainerLabels := enclaveObjLabelsProvider.ForApiContainer(apiContainerIpAddr, apiContainerListenPort)

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
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the API container launcher for launch API version '%v'", launchApiVersion)
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
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "An error occurred launching the API container")
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
		return "", nil, "", nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the API container to become available")
	}

	// Everything started successfully, so the responsibility of deleting the network is now transferred to the caller
	shouldDeleteNetwork = false
	shouldStopApiContainer = false
	return networkId, networkIpAndMask, apiContainerId, &apiContainerIpAddr, apiContainerHostPortBinding, nil
}

func (manager *EnclaveManager) DestroyEnclave(ctx context.Context, enclaveId string) error {

	networks, err := manager.dockerManager.GetNetworksByName(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the network  matching the '%v' network name", enclaveId)
	}
	if len(networks) == 0 || len(networks) > 1 {
		return stacktrace.NewError("%v Docker network were returned for the '%v' network - this is very strange!", len(networks), enclaveId)
	}

	networkId := networks[0].GetId()

	apiContainer, err := manager.getAPIContainer(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting API container for enclave ID '%v'", enclaveId)
	}
	apiContainerId := apiContainer.GetId()

	if err := manager.dockerManager.StopContainer(ctx, apiContainerId, apiContainerStopTimeout); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred stopping the API container with ID '%v' for enclave '%v'",
			apiContainerId,
			enclaveId,
		)
	}

	// The API container's shutdown logic disconnects/stops all other containers, so we're good to remove the network now
	if err := manager.dockerManager.RemoveNetwork(ctx, networkId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred deleting the network with ID '%v' for enclave '%v'",
			networkId,
			enclaveId,
		)
	}

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (manager *EnclaveManager) getAPIContainer(ctx context.Context, enclaveId string) (*types.Container, error) {
	labels := getLabelsForAPIContainer(enclaveId)
	containers, err := manager.dockerManager.GetContainersByLabels(ctx, labels, shouldFetchStoppedContainers)
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
