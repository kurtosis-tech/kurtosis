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
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_rpc_api/api_container_rpc_api_consts"
	api_container_env_var_values2 "github.com/kurtosis-tech/kurtosis/api_container/docker_api/api_container_env_var_values"
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
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/%v", api_container_rpc_api_consts.ListenPort, api_container_rpc_api_consts.ListenProtocol))
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
			kurtosisApiPort: nil,
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

/*
Launches a new testsuite container to acquire testsuite metadata
*/
/*
func (launcher TestsuiteContainerLauncher) LaunchMetadataAcquiringContainers(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager) (testsuiteContainerId string, kurtosisApiContainerId string, err error) {
	functionCompletedSuccessfully := false

	bridgeNetworkIds, err := dockerManager.GetNetworkIdsByName(ctx, bridgeNetworkName)
	if err != nil {
		return "", "", stacktrace.Propagate(
			err,
			"An error occurred getting the network IDs matching the '%v' network",
			bridgeNetworkName)
	}
	if len(bridgeNetworkIds) == 0 || len(bridgeNetworkIds) > 1 {
		return "", "", stacktrace.NewError(
			"%v Docker network IDs were returned for the '%v' network - this is very strange!",
			len(bridgeNetworkIds),
			bridgeNetworkName)
	}
	bridgeNetworkId := bridgeNetworkIds[0]

	apiContainerEnvVars, err := launcher.genSuiteMetadataSerializationApiContainerEnvVars()
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the API container env vars")
	}

	log.Debug("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/%v", api_container_server_consts.ListenPort, api_container_server_consts.ListenProtocol))
	kurtosisApiContainerNameElems := []string{
		launcher.executionInstanceUuid,
		metadataAcquisitionContainerNameLabel,
		apiContainerNameSuffix,
	}
	kurtosisApiContainerId, err = dockerManager.CreateAndStartContainer(
		ctx,
		launcher.kurtosisApiImage,
		kurtosisApiContainerNameElems,
		bridgeNetworkId,
		nil,	// We're connecting to the bridge network, which will assign an IP automatically
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil,
		},
		nil, // No ENTRYPOINT overriding needed because the API container is launched via env vars
		nil, // No CMD overriding needed for the same reason
		apiContainerEnvVars,
		map[string]string{},   // We don't need to bind mount the Docker socket because this API container won't interact with Docker
		map[string]string{
			launcher.suiteExecutionVolName: api_container_mountpoints.SuiteExecutionVolumeMountDirpath,
		},
		false,	// During metadata-acquisition, the API container doesn't need access to the Docker host machine
	)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the Kurtosis API container")
	}
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		kurtosisApiContainerId,
		func() bool { return functionCompletedSuccessfully },
	)
	log.Debug("Successfully launched the Kurtosis API container")

	apiContainerIp, err := dockerManager.GetContainerIP(ctx, bridgeNetworkName, kurtosisApiContainerId)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred getting the API container's IP on network '%v'", bridgeNetworkName)
	}

	testsuiteEnvVars, err := launcher.generateTestSuiteEnvVars(apiContainerIp)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the testsuite container env vars")
	}

	suiteContainerDesc := "metadata-reporting testsuite container"
	log.Infof("Launching %v...", suiteContainerDesc)
	testsuiteContainerNameElems := []string{
		launcher.executionInstanceUuid,
		metadataAcquisitionContainerNameLabel,
		testsuiteContainerNameSuffix,
	}
	testsuiteContainerId, debuggerPortHostBinding, err := launcher.createAndStartTestsuiteContainerWithDebuggingPortIfNecessary(
		ctx,
		dockerManager,
		testsuiteContainerNameElems,
		bridgeNetworkId,
		nil,   // Nil because the bridge network will assign IPs on its own (and won't know what IPs are already used)
		testsuiteEnvVars,
	)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the testsuite container to send metadata to the Kurtosis API container")
	}
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		testsuiteContainerId,
		func() bool { return functionCompletedSuccessfully},
	)
	logSuccessfulSuiteContainerLaunch(log, suiteContainerDesc, debuggerPortHostBinding)

	functionCompletedSuccessfully = true
	return testsuiteContainerId, kurtosisApiContainerId, nil
}
*/

/*
Launches a new testsuite container to run a test
*/
/*
func (launcher TestsuiteContainerLauncher) LaunchTestRunningContainers(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		testName string,
		kurtosisApiContainerIp net.IP,
		testsuiteContainerIp net.IP,
		testSetupTimeoutInSeconds uint32,
		testRunTimeoutInSeconds uint32,
		isPartitioningEnabled bool) (testsuiteContainerId string, kurtosisApiContainerId string, resultErr error){
	log.Debugf(
		"Test suite container IP: %v; kurtosis API container IP: %v",
		testsuiteContainerIp.String(),
		kurtosisApiContainerIp.String())

	functionCompletedSuccessfully := false

	testSuiteEnvVars, err := launcher.generateTestSuiteEnvVars(kurtosisApiContainerIp.String())
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the test-running testsuite container env vars")
	}

	suiteContainerDesc := "test-running testsuite container"
	log.Infof("Launching %v....", suiteContainerDesc)
	suiteContainerNameElems := []string{
		launcher.executionInstanceUuid,
		testName,
		testsuiteContainerNameSuffix,
	}
	suiteContainerId, debuggerPortHostBinding, err := launcher.createAndStartTestsuiteContainerWithDebuggingPortIfNecessary(
		ctx,
		dockerManager,
		suiteContainerNameElems,
		networkId,
		testsuiteContainerIp,
		testSuiteEnvVars)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred creating the test-running testsuite container")
	}
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		suiteContainerId,
		func() bool { return functionCompletedSuccessfully },
	)
	logSuccessfulSuiteContainerLaunch(log, suiteContainerDesc, debuggerPortHostBinding)


	apiContainerEnvVars, err := launcher.genTestExecutionApiContainerEnvVars(
		networkId,
		subnetMask,
		gatewayIpAddr,
		testName,
		suiteContainerId,
		testsuiteContainerIp,
		kurtosisApiContainerIp,
		testSetupTimeoutInSeconds,
		testRunTimeoutInSeconds,
		isPartitioningEnabled)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	log.Info("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/%v", api_container_server_consts.ListenPort, api_container_server_consts.ListenProtocol))
	kurtosisApiContainerNameElems := []string{
		launcher.executionInstanceUuid,
		testName,
		apiContainerNameSuffix,
	}
	kurtosisApiContainerId, err = dockerManager.CreateAndStartContainer(
		ctx,
		launcher.kurtosisApiImage,
		kurtosisApiContainerNameElems,
		networkId,
		kurtosisApiContainerIp,
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil,
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
	defer killContainerIfNotFunctionSuccess(
		ctx,
		log,
		dockerManager,
		kurtosisApiContainerId,
		func() bool { return functionCompletedSuccessfully })
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the Kurtosis API container")
	}
	log.Infof("Successfully launched the Kurtosis API container")

	functionCompletedSuccessfully = true
	return suiteContainerId, kurtosisApiContainerId, nil
}
*/

func (launcher ApiContainerLauncher) genApiContainerEnvVars(
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		initializerContainerIpAddr net.IP,
		testSuiteContainerIpAddr net.IP,
		apiContainerIpAddr net.IP,
		testName string,
		isPartitioningEnabled bool) (map[string]string, error) {
	var hostPortBindingSupplierParams *api_container_env_var_values2.HostPortBindingSupplierParams = nil
	hostPortBindingSupplier := launcher.hostPortBindingSupplier
	if hostPortBindingSupplier != nil {
		hostPortBindingSupplierParams = api_container_env_var_values2.NewHostPortBindingSupplierParams(
			hostPortBindingSupplier.GetInterfaceIp(),
			hostPortBindingSupplier.GetProtocol(),
			hostPortBindingSupplier.GetPortRangeStart(),
			hostPortBindingSupplier.GetPortRangeEnd(),
			hostPortBindingSupplier.GetTakenPorts(),
		)
	}

	args, err := api_container_env_var_values2.NewTestExecutionArgs(
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
