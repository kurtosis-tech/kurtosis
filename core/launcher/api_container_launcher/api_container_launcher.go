/* * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/martian/log"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!
	DefaultVersion = "1.36.4"
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!

	portProtocolToMonitorWhenWaitingForAvailability = "tcp"

	maxWaitForAvailabilityRetries         = 10
	timeBetweenWaitForAvailabilityRetries = 1 * time.Second

	availabilityWaitingExecCmdSuccessExitCode = 0

	// The location where the enclave data directory (on the Docker host machine) will be bind-mounted
	//  on the API container
	enclaveDataDirpathOnAPIContainer = "/kurtosis-enclave-data"

	// TODO This should come from the same logic that builds the server image!!!!!
	containerImage = "kurtosistech/kurtosis-core_api"

	// We always want the user to be running the latest version of the specified API container image
	shouldPullImageBeforeLaunching = true

	// The API container must have the Docker socket bind-mounted so that it can start other containers
	shouldBindMountDockerSocket = true

	// API container doesn't have an alias
	containerAlias = ""
)
// All the following are set to the value of "don't use these"
var entrypointArgs []string = nil
var cmdArgs        []string = nil
var volumeMounts   map[string]string = nil

type ApiContainerLauncher struct {
	enclaveContainerLauncher *enclave_container_launcher.EnclaveContainerLauncher
	dockerManager *docker_manager.DockerManager
}

func NewApiContainerLauncher(enclaveContainerLauncher *enclave_container_launcher.EnclaveContainerLauncher, dockerManager *docker_manager.DockerManager) *ApiContainerLauncher {
	return &ApiContainerLauncher{enclaveContainerLauncher: enclaveContainerLauncher, dockerManager: dockerManager}
}

func (launcher ApiContainerLauncher) LaunchWithDefaultVersion(
	ctx context.Context,
	logLevel logrus.Level,
	enclaveId string,
	networkId string,
	subnetMask string,
	listenPort uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnHostMachine string,
) (
	resultContainerId string,
	resultPublicIpAddr net.IP,
	resultPublicPort *enclave_container_launcher.EnclaveContainerPort,
	resultErr error,
) {
	containerId, publicIpAddr, publicPort, err := launcher.LaunchWithCustomVersion(
		ctx,
		DefaultVersion,
		logLevel,
		enclaveId,
		networkId,
		subnetMask,
		listenPort,
		gatewayIpAddr,
		apiContainerIpAddr,
		isPartitioningEnabled,
		enclaveDataDirpathOnHostMachine,
	)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred launching the API container with default version tag '%v'", DefaultVersion)
	}
	return containerId, publicIpAddr, publicPort, nil
}

func (launcher ApiContainerLauncher) LaunchWithCustomVersion(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	enclaveId string,
	networkId string,
	subnetMask string,
	portNum uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnHostMachine string,
) (
	resultContainerId string,
	resultPublicIpAddr net.IP,
	resultPublicPort *enclave_container_launcher.EnclaveContainerPort,
	resultErr error,
) {
	objAttrsSupplier := func(enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider) (schema.ObjectAttributes, error) {
		apiContainerAttrs, err := enclaveObjAttrsProvider.ForApiContainer(
			apiContainerIpAddr,
			portNum,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting the API container object attributes using port num '%v'",
				portNum,
			)
		}
		return apiContainerAttrs, nil
	}

	takenIpAddrStrSet := map[string]bool{
		gatewayIpAddr.String(): true,
		apiContainerIpAddr.String(): true,
	}
	argsObj, err := args.NewAPIContainerArgs(
		logLevel.String(),
		portNum,
		enclaveId,
		networkId,
		subnetMask,
		apiContainerIpAddr.String(),
		takenIpAddrStrSet,
		isPartitioningEnabled,
		enclaveDataDirpathOnAPIContainer,
		enclaveDataDirpathOnHostMachine,
	)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred creating the API container args")
	}

	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		containerImage,
		imageVersionTag,
	)

	grpcPort, err := enclave_container_launcher.NewEnclaveContainerPort(portNum, enclave_container_launcher.EnclaveContainerPortProtocol_TCP)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred constructing the enclave container port object representing the API container's gRPC port")
	}
	privatePorts := map[string]*enclave_container_launcher.EnclaveContainerPort{
		schema.KurtosisInternalContainerGRPCPortID: grpcPort,
	}

	log.Debugf("Launching Kurtosis API container...")
	containerId, publicIpAddr, publicPorts, err := launcher.enclaveContainerLauncher.Launch(
		ctx,
		containerImageAndTag,
		shouldPullImageBeforeLaunching,
		apiContainerIpAddr,
		networkId,
		enclaveDataDirpathOnAPIContainer,
		privatePorts,
		objAttrsSupplier,
		envVars,
		shouldBindMountDockerSocket,
		containerAlias,
		entrypointArgs,
		cmdArgs,
		volumeMounts,
	)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}

	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			if killErr := launcher.dockerManager.KillContainer(context.Background(), containerId); killErr != nil {
				logrus.Errorf("The function to create the API container didn't finish successful so we tried to kill the container we created, but the killing threw an error:")
				logrus.Error(killErr)
			}
		}
	}()

	if err := waitForAvailability(ctx, launcher.dockerManager, containerId, portNum); err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the API container to become available")
	}

	publicPort, found := publicPorts[schema.KurtosisInternalContainerGRPCPortID]
	if !found {
		return "", nil, nil, stacktrace.NewError("No public port was found for '%v' - this is very strange!", schema.KurtosisInternalContainerGRPCPortID)
	}

	shouldKillContainer = false
	return containerId, publicIpAddr, publicPort, nil
}

func waitForAvailability(ctx context.Context, dockerManager *docker_manager.DockerManager, containerId string, listenPortNum uint16) error {
	commandStr := fmt.Sprintf(
		"[ -n \"$(netstat -anp %v | grep LISTEN | grep %v)\" ]",
		portProtocolToMonitorWhenWaitingForAvailability,
		listenPortNum,
	)
	execCmd := []string{
		"sh",
		"-c",
		commandStr,
	}
	for i := 0; i < maxWaitForAvailabilityRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunExecCommand(ctx, containerId, execCmd, outputBuffer)
		if err == nil {
			if (exitCode == availabilityWaitingExecCmdSuccessExitCode) {
				return nil
			}
			logrus.Debugf(
				"API container availability-waiting command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				availabilityWaitingExecCmdSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"API container availability-waiting command '%v' experienced a Docker error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxWaitForAvailabilityRetries {
			time.Sleep(timeBetweenWaitForAvailabilityRetries)
		}
	}

	return stacktrace.NewError(
		"The API container didn't become available (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxWaitForAvailabilityRetries,
		timeBetweenWaitForAvailabilityRetries,
	)
}
