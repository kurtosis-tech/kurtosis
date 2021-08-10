/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_manager

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/docker_network_allocator"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis/initializer/api_container_launcher"
	"github.com/labstack/gommon/log"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// Will be wrapped in the DockerManager that logs to the proper location
	dockerClient *client.Client

	dockerNetworkAllocator *docker_network_allocator.DockerNetworkAllocator

	apiContainerLauncher *api_container_launcher.ApiContainerLauncher
}

func (manager *EnclaveManager) CreateEnclave(ctx context.Context, enclaveId string, log *logrus.Logger) (*EnclaveContext, error) {
	dockerManager := docker_manager.NewDockerManager(log, manager.dockerClient)

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
