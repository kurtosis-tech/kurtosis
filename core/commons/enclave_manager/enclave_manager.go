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
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/api_container_launcher_lib"
	"github.com/kurtosis-tech/kurtosis/api_container_availability_waiter/api_container_availability_waiter_consts"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/docker_network_allocator"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/enclave_context"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
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

	// Protocol that the API container listens on
	apiContainerListenProtocol = "tcp"

	// The port that the API container listens on
	apiContainerListenPort = 7443
)

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// Will be wrapped in the DockerManager that logs to the proper location
	dockerClient *client.Client

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator

	// TODO This shouldn't be passed in at constructor time, but should be auto-detected from the core API version!!!
	apiContainerImage string
}

func NewEnclaveManager(dockerClient *client.Client, apiContainerImage string) *EnclaveManager {
	dockerNetworkAllocator := docker_network_allocator.NewDockerNetworkAllocator()
	return &EnclaveManager{
		dockerClient:           dockerClient,
		dockerNetworkAllocator: dockerNetworkAllocator,
		apiContainerImage:      apiContainerImage,
	}
}

func (manager *EnclaveManager) CreateEnclave(
		setupCtx context.Context,
		log *logrus.Logger,
		apiContainerLogLevel logrus.Level,
		// TODO put in coreApiVersion as a param here!
		externalContainerIdsToMount map[string]bool,  // Preexisting containers that should be mounted inside the enclave network
		enclaveId string,
		isPartitioningEnabled bool,
		shouldPublishAllPorts bool) (*enclave_context.EnclaveContext, error) {
	dockerManager := docker_manager.NewDockerManager(log, manager.dockerClient)

	matchingNetworks, err := dockerManager.GetNetworkIdsByName(setupCtx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finding enclaves with name '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if len(matchingNetworks) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveId)
	}

	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)

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

	log.Debugf("Connecting external containers to the enclave network so that they can interact with the containers in the enclave...")
	externalContainerIpAddrs := []net.IP{}
	externalContainerIdsToDisconnectSet := map[string]bool{}
	defer func() {
		for containerId := range externalContainerIdsToDisconnectSet {
			if err := dockerManager.DisconnectContainerFromNetwork(teardownCtx, containerId, networkId); err != nil {
				log.Errorf("Creating the enclave didn't complete successfully, so we tried to disconnect container with ID '%v' from enclave network but an error was thrown:", containerId)
				fmt.Fprintln(log.Out, err)
				log.Errorf("ACTION REQUIRED: You'll need to manually disconnect container with ID '%v' from network with ID '%v'!!!!!!!", containerId, networkId)
			}
		}
	}()
	for containerId := range externalContainerIdsToMount {
		ipInsideEnclaveNetwork, err := freeIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting a free IP for mounting external container with ID '%v' inside the enclave", containerId)
		}
		externalContainerIpAddrs = append(externalContainerIpAddrs, ipInsideEnclaveNetwork)
		if err := dockerManager.ConnectContainerToNetwork(setupCtx, networkId, containerId, ipInsideEnclaveNetwork, ""); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred connecting container with ID '%v' to the enclave network", containerId)
		}
		externalContainerIdsToDisconnectSet[containerId] = true
	}
	log.Debugf("Successfully connected external containers to the enclave network so that they can interact with the containers in the enclave")

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
	alreadyTakenIps := append(
		[]net.IP{testsuiteContainerIpAddr, replContainerIpAddr},
		externalContainerIpAddrs...
	)

	// TODO This shouldn't be hardcoded!!! We should instead detect the launch API version from the core API version
	launchApiVersion := uint(0)
	apiContainerLauncher, err := lib.GetAPIContainerLauncherForLaunchAPIVersion(
		launchApiVersion,
		dockerManager,
		log,
		manager.apiContainerImage,
		apiContainerListenPort,
		apiContainerListenProtocol,
		apiContainerLogLevel,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the API container launcher for launch API version '%v'", launchApiVersion)
	}

	apiContainerId, err := apiContainerLauncher.Launch(
		setupCtx,
		apiContainerName,
		enclaveId,
		networkId,
		networkIpAndMask.String(),
		gatewayIp,
		apiContainerIpAddr,
		alreadyTakenIps,
		isPartitioningEnabled,
		externalContainerIdsToMount,
		shouldPublishAllPorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}
	// The API container is started successfully and it will disconnect/stop everything in its network when it shuts down,
	//  so it takes over the responsibility of disconnecting the external containers
	externalContainerIdsToDisconnectSet = map[string]bool{}
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
		replContainerIpAddr,
		testsuiteContainerIpAddr,
		enclaveObjNameProvider.ForTestRunningTestsuiteContainer(),
		dockerManager,
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
};

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
