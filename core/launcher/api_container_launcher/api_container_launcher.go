/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/google/martian/log"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!
	defaultImageVersionTag = "1.32.0"
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!

	dockerSocket = "/var/run/docker.sock"

	// We ALWAYS publish service ports now
	shouldPublishServicePorts = true

	listenProtocol = "tcp"

	maxWaitForAvailabilityRetries         = 10
	timeBetweenWaitForAvailabilityRetries = 1 * time.Second

	availabilityWaitingExecCmdSuccessExitCode = 0

	// The location where the enclave data directory (on the Docker host machine) will be bind-mounted
	//  on the API container
	enclaveDataDirpathOnAPIContainer = "/kurtosis-enclave-data"

	// TODO This should come from the same logic that builds the server image!!!!!
	containerImage = "kurtosistech/kurtosis-core_api"
)

type ApiContainerLauncher struct {
	dockerManager *docker_manager.DockerManager

	objAttrsProvider schema.ObjectAttributesProvider
}

func NewApiContainerLauncher(dockerManager *docker_manager.DockerManager, objAttrsProvider schema.ObjectAttributesProvider) *ApiContainerLauncher {
	return &ApiContainerLauncher{dockerManager: dockerManager, objAttrsProvider: objAttrsProvider}
}

func (launcher *ApiContainerLauncher) GetDefaultVersion() string {
	return defaultImageVersionTag
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
) (string, *nat.PortBinding, error) {
	containerId, hostMachinePortBinding, err := launcher.LaunchWithCustomVersion(
		ctx,
		defaultImageVersionTag,
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
		return "", nil, stacktrace.Propagate(err, "An error occurred launching the API container with default version tag '%v'", defaultImageVersionTag)
	}
	return containerId, hostMachinePortBinding, nil
}

func (launcher ApiContainerLauncher) LaunchWithCustomVersion(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	enclaveId string,
	networkId string,
	subnetMask string,
	listenPort uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnHostMachine string,
) (string, *nat.PortBinding, error) {
	enclaveObjAttrsProvider := launcher.objAttrsProvider.ForEnclave(enclaveId)
	apiContainerAttrs := enclaveObjAttrsProvider.ForApiContainer(
		apiContainerIpAddr,
		listenPort,
		listenProtocol,
	)
	containerName := apiContainerAttrs.GetName()
	containerLabels := apiContainerAttrs.GetLabels()

	takenIpAddrStrSet := map[string]bool{
		gatewayIpAddr.String(): true,
		apiContainerIpAddr.String(): true,
	}
	argsObj, err := args.NewAPIContainerArgs(
		containerName,
		logLevel.String(),
		listenPort,
		listenProtocol,
		enclaveId,
		networkId,
		subnetMask,
		apiContainerIpAddr.String(),
		takenIpAddrStrSet,
		isPartitioningEnabled,
		shouldPublishServicePorts,
		enclaveDataDirpathOnAPIContainer,
		enclaveDataDirpathOnHostMachine,
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating the API container args")
	}

	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	log.Debugf("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf(
		"%v/%v",
		listenPort,
		listenProtocol,
	))

	// We always publish the API container's ports so that we can call its external container registration functions from the CLI
	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		kurtosisApiPort: docker_manager.NewAutomaticPublishingSpec(),
	}


	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		containerImage,
		imageVersionTag,
	)

	// Best-effort pull attempt
	if err = launcher.dockerManager.PullImage(ctx, containerImageAndTag); err != nil {
		logrus.Warnf("Failed to pull the latest version of API container image '%v'; you may be running an out-of-date version", containerImageAndTag)
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImageAndTag,
		containerName,
		networkId,
	).WithStaticIP(
		apiContainerIpAddr,
	).WithUsedPorts(
		usedPorts,
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(map[string]string{
		dockerSocket: dockerSocket,
		enclaveDataDirpathOnHostMachine: enclaveDataDirpathOnAPIContainer,
	}).WithLabels(
		containerLabels,
	).Build()

	containerId, hostPortBindings, err := launcher.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred starting the API container")
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

	if err := waitForAvailability(ctx, launcher.dockerManager, containerId, listenPort); err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred waiting for the API container to become available")
	}

	hostPortBinding, found := hostPortBindings[kurtosisApiPort]
	if !found {
		return "", nil, stacktrace.NewError("No host port binding was found for API container port '%v' - this is very strange!", kurtosisApiPort)
	}

	shouldKillContainer = false
	return containerId, hostPortBinding, nil
}

func waitForAvailability(ctx context.Context, dockerManager *docker_manager.DockerManager, containerId string, listenPortNum uint16) error {
	commandStr := fmt.Sprintf(
		"[ -n \"$(netstat -anp %v | grep LISTEN | grep %v)\" ]",
		listenProtocol,
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
