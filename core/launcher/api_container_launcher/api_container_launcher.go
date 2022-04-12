/* * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/martian/log"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!
	DefaultVersion = "1.41.0"
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
var cmdArgs []string = nil
var volumeMounts map[string]string = nil

type ApiContainerLauncher struct {
	enclaveContainerLauncher *enclave_container_launcher.EnclaveContainerLauncher
	dockerManager            *docker_manager.DockerManager
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
	grpcListenPort uint16,
	grpcProxyListenPort uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnHostMachine string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (
	resultContainerId string,
	resultPublicIpAddr net.IP,
	resultGrpcPublicPort *enclave_container_launcher.EnclaveContainerPort,
	resultGrpcProxyPublicPort *enclave_container_launcher.EnclaveContainerPort,
	resultErr error,
) {
	containerId, publicIpAddr, publicGrpcPort, publicGrpcProxyPort, err := launcher.LaunchWithCustomVersion(
		ctx,
		DefaultVersion,
		logLevel,
		enclaveId,
		networkId,
		subnetMask,
		grpcListenPort,
		grpcProxyListenPort,
		gatewayIpAddr,
		apiContainerIpAddr,
		isPartitioningEnabled,
		enclaveDataDirpathOnHostMachine,
		metricsUserID,
		didUserAcceptSendingMetrics,
	)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred launching the API container with default version tag '%v'", DefaultVersion)
	}
	return containerId, publicIpAddr, publicGrpcPort, publicGrpcProxyPort, nil
}

func (launcher ApiContainerLauncher) LaunchWithCustomVersion(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	enclaveId string,
	networkId string,
	subnetMask string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnHostMachine string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (
	resultContainerId string,
	resultPublicIpAddr net.IP,
	resultGrpcPublicPort *enclave_container_launcher.EnclaveContainerPort,
	resultGrpcProxyPublicPort *enclave_container_launcher.EnclaveContainerPort,
	resultErr error,
) {
	objAttrsSupplier := func(enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider) (schema.ObjectAttributes, error) {
		apiContainerAttrs, err := enclaveObjAttrsProvider.ForApiContainer(
			apiContainerIpAddr,
			grpcPortNum,
			grpcProxyPortNum,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting the API container object attributes using port num '%v' and proxy port '%v'",
				grpcPortNum,
				grpcProxyPortNum,
			)
		}
		return apiContainerAttrs, nil
	}

	takenIpAddrStrSet := map[string]bool{
		gatewayIpAddr.String():      true,
		apiContainerIpAddr.String(): true,
	}
	argsObj, err := args.NewAPIContainerArgs(
		imageVersionTag,
		logLevel.String(),
		grpcPortNum,
		grpcProxyPortNum,
		enclaveId,
		networkId,
		subnetMask,
		apiContainerIpAddr.String(),
		takenIpAddrStrSet,
		isPartitioningEnabled,
		enclaveDataDirpathOnAPIContainer,
		enclaveDataDirpathOnHostMachine,
		metricsUserID,
		didUserAcceptSendingMetrics,
	)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating the API container args")
	}

	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		containerImage,
		imageVersionTag,
	)

	grpcPort, err := enclave_container_launcher.NewEnclaveContainerPort(grpcPortNum, enclave_container_launcher.EnclaveContainerPortProtocol_TCP)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred constructing the enclave container port object representing the API container's gRPC port '%v'", grpcPortNum)
	}

	grpcProxyPort, err := enclave_container_launcher.NewEnclaveContainerPort(grpcProxyPortNum, enclave_container_launcher.EnclaveContainerPortProtocol_TCP)
	if err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred constructing the enclave container port object representing the API container's gRPC port with portNum '%v' and grpcPortNum '%v'", grpcPortNum, grpcProxyPortNum)
	}

	privatePorts := map[string]*enclave_container_launcher.EnclaveContainerPort{
		schema.KurtosisInternalContainerGRPCPortID:      grpcPort,
		schema.KurtosisInternalContainerGRPCProxyPortID: grpcProxyPort,
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
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred launching the API container")
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

	if err := waitForAvailability(ctx, launcher.dockerManager, containerId, grpcPortNum); err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the API container's grpc port to become available")
	}

	if err := waitForAvailability(ctx, launcher.dockerManager, containerId, grpcProxyPortNum); err != nil {
		return "", nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the API container's grpc-proxy port to become available")
	}

	publicGrpcPort, found := publicPorts[schema.KurtosisInternalContainerGRPCPortID]
	if !found {
		return "", nil, nil, nil, stacktrace.NewError("No public port was found for '%v' - this is very strange!", schema.KurtosisInternalContainerGRPCPortID)
	}

	publicGrpcProxyPort, found := publicPorts[schema.KurtosisInternalContainerGRPCProxyPortID]
	if !found {
		return "", nil, nil, nil, stacktrace.NewError("No public port was found for '%v' - this is very strange!", schema.KurtosisInternalContainerGRPCProxyPortID)
	}

	shouldKillContainer = false
	return containerId, publicIpAddr, publicGrpcPort, publicGrpcProxyPort, nil
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
			if exitCode == availabilityWaitingExecCmdSuccessExitCode {
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
