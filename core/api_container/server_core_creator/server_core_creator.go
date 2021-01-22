/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server_core_creator

import (
	"encoding/json"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/suite_metadata_serialization"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"path"
)

func Create(mode api_container_env_vars.ApiContainerMode, paramsJson string) (server.ApiContainerServerCore, error) {
	paramsJsonBytes := []byte(paramsJson)

	switch mode {
	case api_container_env_vars.SuiteMetadataSerializingMode:
		var args SuiteMetadataSerializingArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing the suite metadata serializing args JSON")
		}
		result := createSuiteMetadataSerializationCore(args)
		return result,  nil
	case api_container_env_vars.TestExecutionMode:
		var args TestExecutionArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred deserializing the test execution args JSON")
		}
		result, err := createTestExecutionCore(args)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the test execution core")
		}
		return result, nil
	default:
		return nil, stacktrace.NewError("Unrecognized API container mode '%v'", mode)
	}
}

// ===============================================================================================
//                                 Suite Metadata Serilaization
// ===============================================================================================
func createSuiteMetadataSerializationCore(args SuiteMetadataSerializingArgs) *suite_metadata_serialization.SuiteMetadataSerializationServerCore {
	serializationOutputFilepath := path.Join(
		api_container_mountpoints.SuiteExecutionVolumeMountDirpath,
		args.SuiteMetadataRelativeFilepath)
	return suite_metadata_serialization.NewSuiteMetadataSerializationServerCore(serializationOutputFilepath)
}

// ===============================================================================================
//                                      Test Execution
// ===============================================================================================
func createTestExecutionCore(args TestExecutionArgs) (*test_execution.TestExecutionServerCore, error) {
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

	serviceNetwork := createServiceNetwork(
		args.ExecutionInstanceId,
		args.TestName,
		args.SuiteExecutionVolumeName,
		dockerManager,
		freeIpAddrTracker,
		args.NetworkId,
		args.IsPartitioningEnabled)

	core := test_execution.NewTestExecutionServerCore(
		dockerManager,
		serviceNetwork,
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
		dockerManager *docker_manager.DockerManager,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		dockerNetworkId string,
		isPartitioningEnabled bool) *service_network.ServiceNetwork {

	filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
		suiteExecutionVolName,
		dockerManager,
		dockerNetworkId,
		freeIpAddrTracker)

	suiteExecutionVolume := suite_execution_volume.NewSuiteExecutionVolume(
		api_container_mountpoints.SuiteExecutionVolumeMountDirpath)

	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		executionInstanceId,
		testName,
		dockerManager,
		freeIpAddrTracker,
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
		suiteExecutionVolume,
		userServiceLauncher,
		networkingSidecarManager)

	return serviceNetwork
}
