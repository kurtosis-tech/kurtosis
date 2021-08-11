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
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_env_var_values"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

const (

	dockerSocket = "/var/run/docker.sock"
)

type ApiContainerLauncher struct {
	testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider
	containerImage          string
	kurtosisLogLevel        logrus.Level
	shouldPublishPorts		bool
}

func NewApiContainerLauncher(testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider, containerImage string, kurtosisLogLevel logrus.Level, shouldPublishPorts bool) *ApiContainerLauncher {
	return &ApiContainerLauncher{testsuiteExObjNameProvider: testsuiteExObjNameProvider, containerImage: containerImage, kurtosisLogLevel: kurtosisLogLevel, shouldPublishPorts: shouldPublishPorts}
}

func (launcher ApiContainerLauncher) Launch(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		containerName string,
		enclaveId string,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		apiContainerIpAddr net.IP,
		otherTakenIpAddrsInEnclave []net.IP,
		isPartitioningEnabled bool) (string, error){
	apiContainerEnvVars, err := launcher.genApiContainerEnvVars(
		enclaveId,
		networkId,
		subnetMask,
		gatewayIpAddr,
		apiContainerIpAddr,
		otherTakenIpAddrsInEnclave,
		isPartitioningEnabled,
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	log.Info("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf(
		"%v/%v",
		kurtosis_core_rpc_api_consts.ListenPort,
		kurtosis_core_rpc_api_consts.ListenProtocol,
	))
	containerId, _, err := dockerManager.CreateAndStartContainer(
		ctx,
		launcher.containerImage,
		containerName,
		networkId,
		apiContainerIpAddr,
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]bool{kurtosisApiPort: true},
		false, // For now, we don't publish the API container's port to the host machine (though maybe this will change in the future)
		nil, // Nil ENTRYPOINT args because the API container is launched by setting env vars
		nil, // Nil CMD args because the API container is launched by setting env vars
		apiContainerEnvVars,
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{
			enclaveId: api_container_mountpoints.EnclaveDataVolumeMountpoint,
		},
		false, // The API container doesn't need access to the host machine
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred launching the Kurtosis API container")
	}
	log.Infof("Successfully launched the Kurtosis API container")

	return containerId, nil

}

func (launcher ApiContainerLauncher) genApiContainerEnvVars(
		enclaveId string,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		apiContainerIpAddr net.IP,
		otherTakenIpAddrsInEnclave []net.IP,
		isPartitioningEnabled bool) (map[string]string, error) {
	takenIpAddrStrSet := map[string]bool{
		gatewayIpAddr.String(): true,
		apiContainerIpAddr.String(): true,
	}
	for _, takenIp := range otherTakenIpAddrsInEnclave {
		takenIpAddrStrSet[takenIp.String()] = true
	}
	args, err := api_container_env_var_values.NewApiContainerArgs(
		enclaveId,
		networkId,
		subnetMask,
		apiContainerIpAddr.String(),
		takenIpAddrStrSet,
		isPartitioningEnabled,
		launcher.shouldPublishPorts,
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
		api_container_env_vars.LogLevelEnvVar: launcher.kurtosisLogLevel.String(),
		api_container_env_vars.ParamsJsonEnvVar: argsStr,
	}, nil
}
