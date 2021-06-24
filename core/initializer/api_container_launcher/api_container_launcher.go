/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-client/golang/core_api_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_env_var_values"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	apiContainerNameSuffix       = "kurtosis-api"

	dockerSocket = "/var/run/docker.sock"
)

type ApiContainerLauncher struct {
	executionInstanceUuid string
	containerImage string
	suiteExecutionVolName string
	kurtosisLogLevel logrus.Level
	hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier
}

func NewApiContainerLauncher(executionInstanceUuid string, containerImage string, suiteExecutionVolName string, kurtosisLogLevel logrus.Level, hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier) *ApiContainerLauncher {
	return &ApiContainerLauncher{executionInstanceUuid: executionInstanceUuid, containerImage: containerImage, suiteExecutionVolName: suiteExecutionVolName, kurtosisLogLevel: kurtosisLogLevel, hostPortBindingSupplier: hostPortBindingSupplier}
}

func (launcher ApiContainerLauncher) Launch(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		testName string,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		initializerContainerIpAddr net.IP,
		testSuiteContainerIpAddr net.IP,
		apiContainerIpAddr net.IP,
		isPartitioningEnabled bool) (string, error){
	apiContainerEnvVars, err := launcher.genApiContainerEnvVars(
		networkId,
		subnetMask,
		gatewayIpAddr,
		initializerContainerIpAddr,
		testSuiteContainerIpAddr,
		apiContainerIpAddr,
		testName,
		isPartitioningEnabled,
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	log.Info("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf(
		"%v/%v",
		core_api_consts.ListenPort,
		core_api_consts.ListenProtocol,
	))
	kurtosisApiContainerNameElems := []string{
		launcher.executionInstanceUuid,
		testName,
		apiContainerNameSuffix,
	}
	containerId, err := dockerManager.CreateAndStartContainer(
		ctx,
		launcher.containerImage,
		kurtosisApiContainerNameElems,
		networkId,
		apiContainerIpAddr,
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil, // We don't expose the API container's port to the host machine for now
		},
		nil, // Nil ENTRYPOINT args because the API container is launched by setting env vars
		nil, // Nil CMD args because the API container is launched by setting env vars
		apiContainerEnvVars,
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{
			launcher.suiteExecutionVolName: api_container_mountpoints.SuiteExecutionVolumeMountDirpath,
		},
		launcher.hostPortBindingSupplier != nil, // If we're expecting ot dole out host ports, the API container WILL need access to the host machine running Docker
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred launching the Kurtosis API container")
	}
	log.Infof("Successfully launched the Kurtosis API container")

	return containerId, nil

}

func (launcher ApiContainerLauncher) genApiContainerEnvVars(
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		initializerContainerIpAddr net.IP,
		testSuiteContainerIpAddr net.IP,
		apiContainerIpAddr net.IP,
		testName string,
		isPartitioningEnabled bool) (map[string]string, error) {
	var hostPortBindingSupplierParams *api_container_env_var_values.HostPortBindingSupplierParams = nil
	hostPortBindingSupplier := launcher.hostPortBindingSupplier
	if hostPortBindingSupplier != nil {
		hostPortBindingSupplierParams = api_container_env_var_values.NewHostPortBindingSupplierParams(
			hostPortBindingSupplier.GetInterfaceIp(),
			hostPortBindingSupplier.GetProtocol(),
			hostPortBindingSupplier.GetPortRangeStart(),
			hostPortBindingSupplier.GetPortRangeEnd(),
			hostPortBindingSupplier.GetTakenPorts(),
		)
	}

	args, err := api_container_env_var_values.NewApiContainerArgs(
		launcher.executionInstanceUuid,
		networkId,
		subnetMask,
		gatewayIpAddr.String(),
		launcher.suiteExecutionVolName,
		[]string{
			launcher.executionInstanceUuid,
			testName,
		},
		map[string]bool{
			gatewayIpAddr.String(): true,
			initializerContainerIpAddr.String(): true,
			apiContainerIpAddr.String(): true,
			testSuiteContainerIpAddr.String(): true,
		},
		isPartitioningEnabled,
		hostPortBindingSupplierParams)
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
