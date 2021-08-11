/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_manager

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/docker_network_allocator"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis/initializer/api_container_launcher"
	"github.com/labstack/gommon/log"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// Will be wrapped in the DockerManager that logs to the proper location
	dockerClient *client.Client

	// TODO Hide this all inside this class, rather than taking it in as a constructor param????
	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator

	apiContainerLauncher *api_container_launcher.ApiContainerLauncher
}

// TODO Constructor

// NOTE: thisContainerId is the ID of the container in which this code is executing, so that it can be mounted inside
//  the new enclave such that it can communicate with the API container
func (manager *EnclaveManager) CreateEnclave(ctx context.Context, thisContainerId string, enclaveId string, log *logrus.Logger, isPartitioningEnabled bool) (*EnclaveContext, error) {
	dockerManager := docker_manager.NewDockerManager(log, manager.dockerClient)

	matchingNetworks, err := dockerManager.GetNetworkIdsByName(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finding enclaves with name '%v', which is necessary to ensure that our enclave doesn't exist yet", enclaveId)
	}
	if len(matchingNetworks) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveId)
	}

	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)

	log.Debugf("Creating Docker network for enclave '%v'...", enclaveId)
	networkId, networkIpAndMask, gatewayIp, freeIpAddrTracker, err := manager.dockerNetworkAllocator.CreateNewNetwork(
		ctx,
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
	log.Debugf("Docker network '%v' created successfully with ID '%v' and subnet CIDR '%v'", enclaveId, networkId, networkIpAndMask.String())

	// TODO Mount other containers (i.e. initializer) inside the network

	// TODO use hostnames rather than IPs, which makes things nicer and which we'll need for Docker swarm support
	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	kurtosisApiIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}

	if err := dockerManager.CreateVolume(ctx, enclaveId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave volume '%v'", enclaveId)
	}

	apiContainerName := enclaveObjNameProvider.ForApiContainer()
	apiContainerId, err := manager.apiContainerLauncher.Launch(
		ctx,
		log,
		dockerManager,
		apiContainerName,
		enclaveId,
		networkId,
		networkIpAndMask.String(),
		gatewayIp,
		kurtosisApiIp,
		[]net.IP{},  // TODO Add the other containers that we mount in here
		isPartitioningEnabled,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}

	grpc.Dial()
	apiClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient()
	networkCtx := networks.NewNetworkContext()



}

/*
Helper function for making a best-effort attempt at removing a network and the containers inside after the task is
	done running against the enclave
	exited (either normally or with error)
*/
func removeNetworkDeferredFunc(
	testTeardownContext context.Context,
	log *logrus.Logger,
	dockerManager *docker_manager.DockerManager,
	networkId string,
	initializerContainerId string) {
	log.Debugf("Attempting to remove Docker network with ID %v...", networkId)
	containerIds, err := dockerManager.GetContainerIdsConnectedToNetwork(testTeardownContext, networkId)
	if err != nil {
		errorDesc := fmt.Sprintf("An error occurred getting the containers connected to network '%v' so we can stop them:", networkId)
		logErrorAndRecommendManualIntervention(log, errorDesc, err, networkId)
		return
	}

	for _, containerId := range containerIds {
		if containerId == initializerContainerId {
			// We don't want to kill the initializer since it could be coordinating other tests, but we need it gone
			//  from the network before we can delete the network
			if err := dockerManager.DisconnectContainerFromNetwork(testTeardownContext, initializerContainerId, networkId); err != nil {
				errorDesc := fmt.Sprintf("An error occurred disconnecting the initializer container from the network, which prevents the network from being deleted:")
				logErrorAndRecommendManualIntervention(log, errorDesc, err, networkId)
				return
			}
		} else {
			if err := dockerManager.KillContainer(testTeardownContext, containerId); err != nil {
				errorDesc := fmt.Sprintf("An error occurred killing container '%v', which prevents the network from being deleted:", containerId)
				logErrorAndRecommendManualIntervention(log, errorDesc, err, networkId)
				return
			}
		}
	}

	// Give a tiny bit of time for the container-kills to complete before removing the network
	time.Sleep(networkTeardownWaitTime)

	if err := dockerManager.RemoveNetwork(testTeardownContext, networkId); err != nil {
		errorDesc := fmt.Sprintf("An error occurred removing Docker network with ID %v:", networkId)
		logErrorAndRecommendManualIntervention(log, errorDesc, err, networkId)
		return
	}
	log.Debugf("Successfully removed Docker network with ID %v", networkId)
}

func logErrorAndRecommendManualIntervention(log *logrus.Logger, humanReadableErrorDesc string, err error, networkId string) {
	log.Errorf(humanReadableErrorDesc)
	log.Error(err.Error())
	log.Errorf("ACTION REQUIRED: You'll need to delete any remaining containers and the network with ID '%v' manually!!!", networkId)
}
