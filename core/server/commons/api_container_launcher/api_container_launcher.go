/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/google/martian/log"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/object_labels_providers"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	dockerSocket = "/var/run/docker.sock"

	// All API containers accept exactly one environment variable, which contains the serialized params that
	// dictate how the API container ought to behave
	serializedArgsEnvVar = "SERIALIZED_ARGS"

	// We ALWAYS publish service ports now
	shouldPublishServicePorts = true
)

type ApiContainerLauncher struct {
	dockerManager *docker_manager.DockerManager
}

func NewApiContainerLauncher(dockerManager *docker_manager.DockerManager) *ApiContainerLauncher {
	return &ApiContainerLauncher{dockerManager: dockerManager}
}

func (launcher ApiContainerLauncher) Launch(
	ctx context.Context,
	containerImage string,
	containerName string,
	logLevel logrus.Level,
	enclaveId string,
	networkId string,
	subnetMask string,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	otherTakenIpAddrsInEnclave []net.IP,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnHostMachine string,
) (string, *nat.PortBinding, error){
	envVars, err := launcher.genApiContainerEnvVars(
		containerName,
		logLevel,
		enclaveId,
		networkId,
		subnetMask,
		gatewayIpAddr,
		apiContainerIpAddr,
		otherTakenIpAddrsInEnclave,
		isPartitioningEnabled,
		enclaveDataDirpathOnHostMachine,
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	log.Debugf("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf(
		"%v/%v",
		kurtosis_core_rpc_api_consts.ListenPort,
		kurtosis_core_rpc_api_consts.ListenProtocol,
	))

	// We always publish the API container's ports so that we can call its external container registration functions from the CLI
	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		kurtosisApiPort: docker_manager.NewAutomaticPublishingSpec(),
	}

	// Mayyyybe better to take this in as an arg?
	enclaveObjLabelsProvider := object_labels_providers.NewEnclaveObjectLabelsProvider(enclaveId)
	containerLabels := enclaveObjLabelsProvider.ForApiContainer(apiContainerIpAddr)

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
		enclaveDataDirpathOnHostMachine: api_container_docker_consts.EnclaveDataDirMountpoint,
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



func (launcher ApiContainerLauncher) genApiContainerEnvVars(
	containerName string,
	logLevel logrus.Level,
	enclaveId string,
	networkId string,
	subnetMask string,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	otherTakenIpAddrsInEnclave []net.IP,
	isPartitioningEnabled bool,
	enclaveDataDirpathOnHostMachine string,
) (map[string]string, error) {
	takenIpAddrStrSet := map[string]bool{
		gatewayIpAddr.String(): true,
		apiContainerIpAddr.String(): true,
	}
	for _, takenIp := range otherTakenIpAddrsInEnclave {
		takenIpAddrStrSet[takenIp.String()] = true
	}
	args, err := NewAPIContainerArgs(
		containerName,
		logLevel.String(),
		enclaveId,
		networkId,
		subnetMask,
		apiContainerIpAddr.String(),
		takenIpAddrStrSet,
		isPartitioningEnabled,
		shouldPublishServicePorts,
		enclaveDataDirpathOnHostMachine,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the test execution args")
	}

	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing API container test execution args to JSON")
	}

	argsStr := string(argsBytes)

	// TODO switch to the envVars requiring a visitor to hit, so we get them all
	return map[string]string{
		serializedArgsEnvVar: argsStr,
	}, nil
}
