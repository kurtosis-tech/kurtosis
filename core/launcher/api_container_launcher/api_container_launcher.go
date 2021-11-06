/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
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
)

const (
	dockerSocket = "/var/run/docker.sock"

	// We ALWAYS publish service ports now
	shouldPublishServicePorts = true

	listenProtocol = "tcp"

	// The location where the enclave data directory (on the Docker host machine) will be bind-mounted
	//  on the API container
	enclaveDataDirpathOnAPIContainer = "/kurtosis-enclave-data"
)

type ApiContainerLauncher struct {
	dockerManager *docker_manager.DockerManager

	objAttrsProvider schema.ObjectAttributesProvider
}

func NewApiContainerLauncher(dockerManager *docker_manager.DockerManager, objAttrsProvider schema.ObjectAttributesProvider) *ApiContainerLauncher {
	return &ApiContainerLauncher{dockerManager: dockerManager, objAttrsProvider: objAttrsProvider}
}

func (launcher ApiContainerLauncher) Launch(
	ctx context.Context,
	containerImage string,
	logLevel logrus.Level,
	enclaveId string,
	networkId string,
	subnetMask string,
	listenPort uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	otherTakenIpAddrsInEnclave []net.IP,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnHostMachine string,
) (string, *nat.PortBinding, error){
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
	for _, takenIp := range otherTakenIpAddrsInEnclave {
		takenIpAddrStrSet[takenIp.String()] = true
	}
	argsObj, err := args.NewAPIContainerArgs(
		containerName,
		logLevel.String(),
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


	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
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
	shouldDeleteContainer := true
	defer func() {
		if shouldDeleteContainer {
			if killErr := launcher.dockerManager.KillContainer(context.Background(), containerId); killErr != nil {
				logrus.Errorf("The function to create the API container didn't finish successful so we tried to kill the container we created, but the killing threw an error:")
				logrus.Error(killErr)
			}
		}
	}()

	hostPortBinding, found := hostPortBindings[kurtosisApiPort]
	if !found {
		return "", nil, stacktrace.NewError("No host port binding was found for API container port '%v' - this is very strange!", kurtosisApiPort)
	}

	shouldDeleteContainer = false
	return containerId, hostPortBinding, nil
}
