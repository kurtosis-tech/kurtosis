/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server_core_creator

import (
	"encoding/json"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_var_values/api_container_modes"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_var_values/api_container_params_json"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/suite_metadata_serialization"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_constants"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

func Create(mode api_container_modes.ApiContainerMode, paramsJson string) (server.ApiContainerServerCore, error) {
	paramsJsonBytes := []byte(paramsJson)

	suiteExecutionVolume := suite_execution_volume.NewSuiteExecutionVolume(api_container_mountpoints.SuiteExecutionVolumeMountDirpath)

	logrus.Debugf("Creating server core by parsing params JSON '%v' using mode '%v'...", paramsJson, mode)
	switch mode {
	case api_container_modes.SuiteMetadataSerializingMode:
		var args api_container_params_json.SuiteMetadataSerializationArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing the suite metadata serializing args JSON")
		}
		serializationOutputFilepath := suiteExecutionVolume.GetSuiteMetadataFile().GetAbsoluteFilepath()
		result := suite_metadata_serialization.NewSuiteMetadataSerializationServerCore(serializationOutputFilepath)
		logrus.Debugf("Successfully created suite metadata-serializing server core")
		return result,  nil
	case api_container_modes.TestExecutionMode:
		var args api_container_params_json.TestExecutionArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing the test execution args JSON")
		}
		result, err := createTestExecutionCore(suiteExecutionVolume, args)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the test execution core")
		}
		logrus.Debugf("Successfully created test execution server core")
		return result, nil
	default:
		return nil, stacktrace.NewError("Unrecognized API container mode '%v'", mode)
	}
}

// ===============================================================================================
//                                      Test Execution
// ===============================================================================================
func createTestExecutionCore(
		suiteExecutionVolume *suite_execution_volume.SuiteExecutionVolume,
		args api_container_params_json.TestExecutionArgs) (*test_execution.TestExecutionServerCore, error) {
	dockerManager, err := createDockerManager()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	freeIpAddrTracker, err := createFreeIpAddrTracker(
		args.SubnetMask,
		args.GatewayIpAddr,
		args.ApiContainerIpAddr,
		args.TestSuiteContainerIpAddr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}

	testExecutionDirectory, err := suiteExecutionVolume.GetTestExecutionDirectory(args.TestName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the test execution directory for test '%v'", args.TestName)
	}

	artifactCache, err := suiteExecutionVolume.GetArtifactCache()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the artifact cache for test '%v'", args.TestName)
	}

	var hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier = nil
	if args.HostPortBindingSupplierParams != nil {
		hostPortSupplierParams := args.HostPortBindingSupplierParams
		supplier, err := free_host_port_binding_supplier.NewFreeHostPortBindingSupplier(
			docker_constants.HostMachineDomainInsideContainer,
			hostPortSupplierParams.InterfaceIp,
			hostPortSupplierParams.Protocol,
			hostPortSupplierParams.PortRangeStart,
			hostPortSupplierParams.PortRangeEnd,
			hostPortSupplierParams.TakenPorts,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"Host port binding supplier params were non-null, but an error occurred creating the host port binding supplier",
			)
		}
		hostPortBindingSupplier = supplier
	}

	serviceNetwork := createServiceNetwork(
		args.ExecutionInstanceId,
		args.TestName,
		args.SuiteExecutionVolumeName,
		artifactCache,
		testExecutionDirectory,
		dockerManager,
		freeIpAddrTracker,
		args.NetworkId,
		args.IsPartitioningEnabled,
		hostPortBindingSupplier)

	core := test_execution.NewTestExecutionServerCore(
		dockerManager,
		serviceNetwork,
		args.TestSetupTimeoutInSeconds,
		args.TestRunTimeoutInSeconds,
		args.TestName,
		args.TestSuiteContainerId)

	return core, nil
}

func createDockerManager() (*docker_manager.DockerManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not initialize a Docker client from the environment")
	}

	dockerManager, err := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	return dockerManager, nil
}

func createFreeIpAddrTracker(
		networkSubnetMask string,
		gatewayIp string,
		apiContainerIp string,
		testSuiteContainerIp string) (*commons.FreeIpAddrTracker, error){
	freeIpAddrTracker, err := commons.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		networkSubnetMask,
		map[string]bool{
			gatewayIp:      true,
			apiContainerIp: true,
			testSuiteContainerIp: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}
	return freeIpAddrTracker, nil
}

func createServiceNetwork(
		executionInstanceId string,
		testName string,
		suiteExecutionVolName string,
		artifactCache *suite_execution_volume.ArtifactCache,
		testExecutionDirectory *suite_execution_volume.TestExecutionDirectory,
		dockerManager *docker_manager.DockerManager,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerNetworkId string,
		isPartitioningEnabled bool,
		hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier) *service_network.ServiceNetwork {

	filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
		suiteExecutionVolName,
		dockerManager,
		dockerNetworkId,
		freeIpAddrTracker)

	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		executionInstanceId,
		testName,
		dockerManager,
		freeIpAddrTracker,
		hostPortBindingSupplier,
		artifactCache,
		filesArtifactExpander,
		dockerNetworkId,
		suiteExecutionVolName)

	networkingSidecarManager := networking_sidecar.NewStandardNetworkingSidecarManager(
		dockerManager,
		freeIpAddrTracker,
		dockerNetworkId)

	serviceNetwork := service_network.NewServiceNetwork(
		isPartitioningEnabled,
		freeIpAddrTracker,
		dockerManager,
		testExecutionDirectory,
		userServiceLauncher,
		networkingSidecarManager)

	return serviceNetwork
}
