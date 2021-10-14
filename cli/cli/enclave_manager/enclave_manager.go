/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_manager

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager/docker_network_allocator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager/enclave_context"
	"github.com/kurtosis-tech/kurtosis-core/api_container_availability_waiter/api_container_availability_waiter_consts"
	"github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher_lib"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	// The API container is responsible for disconnecting/stopping everything in its network when stopped, so we need
	//  to give it some time to do so
	apiContainerStopTimeout = 3 * time.Minute

	// This is set in the API container Dockerfile
	availabilityWaiterBinaryFilepath = "/run/api-container-availability-waiter"

	// TODO This should come from the Kurt Client constant!!!
	// Protocol that the API container listens on
	apiContainerListenProtocol = "tcp"

	// TODO This should come from the Kurt Client constant!!!
	// The port that the API container listens on
	apiContainerListenPort = 7443

	shouldConsiderStoppedAPIContainersWhenGettingEnclaves = true
)

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// TODO This can now be a DockerManager
	// Will be wrapped in the DockerManager that logs to the proper location
	dockerClient *client.Client

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator

}

func NewEnclaveManager(dockerClient *client.Client) *EnclaveManager {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator()
	return &EnclaveManager{
		dockerClient:           dockerClient,
		dockerNetworkAllocator: dockerNetworkAllocator,
	}
}

// TODO Because this is no longer running inside the test_executor_parallelizer in Kurt Core, we no longer need to pass
//  in a log, and therefore no longer need to store a DockerManager inside the EnclaveContext
func (manager *EnclaveManager) CreateEnclave(
		setupCtx context.Context,
		log *logrus.Logger,
	    // TODO This shouldn't be passed as an argument, but should be auto-detected from the core API version!!!
		apiContainerImage string,
		apiContainerLogLevel logrus.Level,
		// TODO put in coreApiVersion as a param here!
		enclaveId string,
		isPartitioningEnabled bool,
		shouldPublishAllPorts bool) (*enclave_context.EnclaveContext, error) {
	dockerManager := docker_manager.NewDockerManager(log, manager.dockerClient)

	matchingNetworks, err := dockerManager.GetNetworksByName(setupCtx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finding enclaves with name '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if len(matchingNetworks) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveId)
	}

	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)
	enclaveObjLabelsProvider := object_labels_providers.NewEnclaveObjectLabelsProvider(enclaveId)

	teardownCtx := context.Background()  // Separate context for tearing stuff down in case the input context is cancelled

	log.Debugf("Creating Docker network for enclave '%v'...", enclaveId)
	networkId, networkIpAndMask, gatewayIp, freeIpAddrTracker, err := manager.dockerNetworkAllocator.CreateNewNetwork(
		setupCtx,
		dockerManager,
		log,
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
			if err := dockerManager.RemoveNetwork(teardownCtx, networkId); err != nil {
				log.Errorf("Creating the enclave didn't complete successfully, so we tried to delete network '%v' that we created but an error was thrown:", networkId)
				fmt.Fprintln(log.Out, err)
				log.Errorf("ACTION REQUIRED: You'll need to manually remove network with ID '%v'!!!!!!!", networkId)
			}
		}
	}()
	log.Debugf("Docker network '%v' created successfully with ID '%v' and subnet CIDR '%v'", enclaveId, networkId, networkIpAndMask.String())


	// TODO use hostnames rather than IPs, which makes things nicer and which we'll need for Docker swarm support
	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	apiContainerIpAddr, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}

	if err := dockerManager.CreateVolume(setupCtx, enclaveId); err != nil {
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

	// TODO This shouldn't be hardcoded!!! We should instead detect the launch API version from the core API version
	launchApiVersion := uint(0)
	apiContainerLauncher, err := api_container_launcher_lib.GetAPIContainerLauncherForLaunchAPIVersion(
		launchApiVersion,
		dockerManager,
		log,
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
			if err := dockerManager.StopContainer(teardownCtx, apiContainerId, apiContainerStopTimeout); err != nil {
				log.Errorf("Creating the enclave didn't complete successfully, so we tried to stop the API container but an error was thrown:")
				fmt.Fprintln(log.Out, err)
				log.Errorf("ACTION REQUIRED: You'll need to manually stop API container with ID '%v'", apiContainerId)
			}
		}
	}()

	if err := waitForApiContainerAvailability(setupCtx, dockerManager, apiContainerId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the API container to become available")
	}

	result := enclave_context.NewEnclaveContext(
		enclaveId,
		networkId,
		networkIpAndMask,
		apiContainerId,
		apiContainerIpAddr,
		apiContainerHostPortBinding,
		dockerManager,
		enclaveObjNameProvider,
		enclaveObjLabelsProvider,
	)

	// Everything started successfully, so the responsibility of deleting the network is now transferred to the caller
	shouldDeleteNetwork = false
	shouldStopApiContainer = false
	return result, nil
}

func (manager *EnclaveManager) DestroyEnclave(ctx context.Context, log *logrus.Logger, enclaveCtx *enclave_context.EnclaveContext) error {
	enclaveId := enclaveCtx.GetEnclaveID()
	dockerManager := enclaveCtx.GetDockerManager()
	networkId := enclaveCtx.GetNetworkID()

	apiContainerId := enclaveCtx.GetAPIContainerID()
	if err := dockerManager.StopContainer(ctx, apiContainerId, apiContainerStopTimeout); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred stopping the API container with ID '%v' for enclave '%v'",
			apiContainerId,
			enclaveId,
		)
	}

	// The API container's shutdown logic disconnects/stops all other containers, so we're good to remove the network now
	if err := dockerManager.RemoveNetwork(ctx, networkId); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred deleting the network with ID '%v' for enclave '%v'",
			networkId,
			enclaveId,
		)
	}

	return nil
}

func (manager *EnclaveManager) GetEnclave(ctx context.Context, enclaveId string, log *logrus.Logger) (*enclave_context.EnclaveContext, error) {
	dockerManager := docker_manager.NewDockerManager(log, manager.dockerClient)

	matchingNetworks, err := dockerManager.GetNetworksByName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting networks matching name '%v'", enclaveId)
	}
	if len(matchingNetworks) == 0 {
		return nil, stacktrace.NewError("Docker network with name '%v' does not exist.", enclaveId)
	}
	if len(matchingNetworks) > 1 {
		return  nil, stacktrace.NewError("Found several networks matching name '%v' - this is very strange!", enclaveId)
	}
	network := matchingNetworks[0]

	labels := getLabelsForAPIContainer(enclaveId)

	containers, err := dockerManager.GetContainersByLabels(ctx, labels, shouldConsiderStoppedAPIContainersWhenGettingEnclaves)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers by labels: '%+v'", labels)
	}
	if len(containers) == 0 {
		return nil, stacktrace.NewError("No API container was found for enclave ID '%v' matching labels '%+v'; this is a bug in Kurtosis itself", enclaveId, labels)
	}
	if len(containers) > 1 {
		return nil, stacktrace.NewError("Found more than one API container for enclave ID '%v'; this is a bug in Kurtosis itself", enclaveId)
	}

	apiContainer := containers[0]

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
	apiContainerIPAddress := net.ParseIP(apiContainerIPAddressString)
	if apiContainerIPAddress == nil {
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
	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)
	enclaveObjLabelsProvider := object_labels_providers.NewEnclaveObjectLabelsProvider(enclaveId)

	enclaveContext := enclave_context.NewEnclaveContext(
		enclaveId,
		network.GetId(),
		network.GetIpAndMask(),
		apiContainer.GetId(),
		apiContainerIPAddress,
		apiContainerHostPortBinding,
		dockerManager,
		enclaveObjNameProvider,
		enclaveObjLabelsProvider,
	)

	return enclaveContext, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
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
