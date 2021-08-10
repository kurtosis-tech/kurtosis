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
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
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
	hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier
}

func NewApiContainerLauncher(testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider, containerImage string, kurtosisLogLevel logrus.Level, hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier) *ApiContainerLauncher {
	return &ApiContainerLauncher{testsuiteExObjNameProvider: testsuiteExObjNameProvider, containerImage: containerImage, kurtosisLogLevel: kurtosisLogLevel, hostPortBindingSupplier: hostPortBindingSupplier}
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
		enclaveDataVolumeName string,
		isPartitioningEnabled bool) (string, error){
	enclaveId, enclaveObjNameProvider := launcher.testsuiteExObjNameProvider.ForTestEnclave(testName)
	apiContainerEnvVars, err := launcher.genApiContainerEnvVars(
		enclaveId,
		networkId,
		subnetMask,
		gatewayIpAddr,
		initializerContainerIpAddr,
		testSuiteContainerIpAddr,
		apiContainerIpAddr,
		enclaveDataVolumeName,
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
	containerName := enclaveObjNameProvider.ForApiContainer()
	containerId, err := dockerManager.CreateAndStartContainer(
		ctx,
		launcher.containerImage,
		containerName,
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
			enclaveDataVolumeName: api_container_mountpoints.EnclaveDataVolumeMountpoint,
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
		enclaveId string,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		initializerContainerIpAddr net.IP,
		testSuiteContainerIpAddr net.IP,
		apiContainerIpAddr net.IP,
		enclaveDataVolumeName string,
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
		enclaveId,
		networkId,
		subnetMask,
		gatewayIpAddr.String(),
		enclaveDataVolumeName,
		apiContainerIpAddr.String(),
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
