/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_constants

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_modes"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_mountpoints"
	"github.com/kurtosis-tech/kurtosis/api_container/server/api_container_server_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/server_core_creator"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_constants/test_suite_env_vars"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

const (
	debuggerPortProtocol = "tcp"

	dockerSocket = "/var/run/docker.sock"
)

// TODO TODO this is a hack! The metadata-acquiring testsuite container needs to know the IP of the corresponding
//  API container, but because both testsuite & API containers will be started in the bridge network (so that we
//  don't have to burn the time to create a network just for them), we don't know what IPs are free so we can't
//  use the freeIpAddrProvider to give a static IP for the API container. We hack around this by picking
//  a hardcoded IP for the API container, and praying that nobody else is using it.
//  The correct fix would be one of:
//   - Inspect the API container after starting in the bridge network to get its IP (requires adding a new function to DockerManager)
//   - Using a hostname rather than an IP for the API container (requires adding a new function argument to CreateAndStartContainer)
//   - Creating a separate network for the metadata-acquiring testsuite & API containers
var metadataAcquiringApiContainerIp net.IP = []byte{172, 17, 256, 256}

type TestsuiteContainerLauncher struct {
	executionInstanceId uuid.UUID

	kurtosisApiImage string

	kurtosisApiLogLevel logrus.Level

	testsuiteImage string

	// The log level string that will be passed as-is to the testsuite (should be meaningful to the testsuite)
	suiteLogLevel string

	// The JSON-serialized custom params object that will be passed as-is to the testsuite
	customParamsJson string

	// This is the port on the testsuite container that a debugger might be listening on, if any is running at all
	// We'll always bind this port on the testsuite container to a port on the user's machine, so they can attach
	//  a debugger if desired
	debuggerPort nat.Port
}

func NewTestsuiteContainerLauncher(
		executionInstanceId uuid.UUID,
		kurtosisApiImage string,
		kurtosisApiLogLevel logrus.Level,
		testsuiteImage string,
		testsuiteLogLevel string,
		customParamsJson string,
		debuggerPort int) (*TestsuiteContainerLauncher, error) {
	debuggerPortObj, err := nat.NewPort(debuggerPortProtocol, strconv.Itoa(debuggerPort))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the debugger port object from port int '%v'", debuggerPort)
	}
	return &TestsuiteContainerLauncher{
		executionInstanceId: executionInstanceId,
		kurtosisApiImage:    kurtosisApiImage,
		kurtosisApiLogLevel: kurtosisApiLogLevel,
		testsuiteImage:      testsuiteImage,
		suiteLogLevel:       testsuiteLogLevel,
		customParamsJson: customParamsJson,
		debuggerPort:        debuggerPortObj,
	}, nil
}

/*
Launches a new testsuite container to acquire testsuite metadata
 */
func (launcher TestsuiteContainerLauncher) LaunchMetadataAcquiringContainer(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		bridgeNetworkId string,
		suiteExecutionVolume string,
		debuggerPortBinding nat.PortBinding) (testsuiteContainerId string, kurtosisApiContainerId string, err error) {
	apiContainerEnvVars, err := launcher.genSuiteMetadataSerializationApiContainerEnvVars()
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the API container env vars")
	}

	logrus.Info("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/%v", api_container_server_consts.ListenPort, api_container_server_consts.ListenProtocol))
	kurtosisApiContainerId, err = dockerManager.CreateAndStartContainer(
		ctx,
		launcher.kurtosisApiImage,
		bridgeNetworkId,
		metadataAcquiringApiContainerIp,
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil,
		},
		nil,
		apiContainerEnvVars,
		map[string]string{},   // We don't need to bind mount the Docker socket because this API container won't interact with Docker
		map[string]string{
			suiteExecutionVolume: api_container_mountpoints.SuiteExecutionVolumeMountDirpath,
		},
	)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the Kurtosis API container")
	}
	logrus.Infof("Successfully launched the Kurtosis API container")

	testsuiteEnvVars, err := launcher.generateTestSuiteEnvVars(metadataAcquiringApiContainerIp)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the testsuite container env vars")
	}

	logrus.Infof("Launching testsuite container to send metadata to Kurotsis API container...")
	testsuiteContainerId, err = dockerManager.CreateAndStartContainer(
		ctx,
		launcher.testsuiteImage,
		bridgeNetworkId,
		nil,                                           // Nil because the bridge network will assign IPs on its own (and won't know what IPs are already used)
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for testsuite containers
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			launcher.debuggerPort: &debuggerPortBinding,
		},
		nil, // Nil start command args because we expect the test suite image to be parameterized with variables
		testsuiteEnvVars,
		map[string]string{},
		map[string]string{
			suiteExecutionVolume: TestsuiteContainerSuiteExVolMountpoint,
		})
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the testsuite container to send metadata to the Kurtosis API container")
	}
	logrus.Infof("Successfully launched testsuite container to send metadata to Kurtosis API container")

	return testsuiteContainerId, kurtosisApiContainerId, nil
}

/*
Launches a new testsuite container to run a test
*/
func (launcher TestsuiteContainerLauncher) LaunchTestRunningContainer(
		ctx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		suiteExecutionVolume string,
		testName string,
		kurtosisApiContainerIp net.IP,
		testsuiteContainerIp net.IP,
		debuggerPortBinding nat.PortBinding,
		isPartitioningEnabled bool) (testsuiteContainerId string, kurtosisApiContainerId string, resultErr error){
	log.Debugf(
		"Test suite container IP: %v; kurtosis API container IP: %v",
		testsuiteContainerIp.String(),
		kurtosisApiContainerIp.String())

	testSuiteEnvVars, err := launcher.generateTestSuiteEnvVars(kurtosisApiContainerIp)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the test-running testsuite container env vars")
	}

	log.Info("Launching test-running testsuite container with debugger port bound to host port %v....", debuggerPortBinding)
	suiteContainerId, err := dockerManager.CreateAndStartContainer(
		ctx,
		launcher.testsuiteImage,
		networkId,
		testsuiteContainerIp,
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for testsuite container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			launcher.debuggerPort: &debuggerPortBinding,
		},
		nil,
		testSuiteEnvVars,
		map[string]string{},
		map[string]string{
			suiteExecutionVolume: TestsuiteContainerSuiteExVolMountpoint,
		})
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred creating the test-running testsuite container")
	}
	log.Infof("Successfully created test-running testsuite container")


	apiContainerEnvVars, err := launcher.genTestExecutionApiContainerEnvVars(
		networkId,
		subnetMask,
		gatewayIpAddr,
		suiteExecutionVolume,
		testName,
		suiteContainerId,
		testsuiteContainerIp,
		kurtosisApiContainerIp,
		isPartitioningEnabled)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	log.Info("Launching Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/%v", api_container_server_consts.ListenPort, api_container_server_consts.ListenProtocol))
	kurtosisApiContainerId, err = dockerManager.CreateAndStartContainer(
		ctx,
		launcher.kurtosisApiImage,
		networkId,
		kurtosisApiContainerIp,
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil,
		},
		nil,
		apiContainerEnvVars,
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{
			suiteExecutionVolume: api_container_mountpoints.SuiteExecutionVolumeMountDirpath,
		},
	)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred launching the Kurtosis API container")
	}
	log.Infof("Successfully launched the Kurtosis API container")

	return suiteContainerId, kurtosisApiContainerId, nil
}

/*
Generates the map of environment variables needed to run a test suite container

NOTE: exactly one of metadata_filepath or test_name must be non-empty!
*/
func (launcher TestsuiteContainerLauncher) generateTestSuiteEnvVars(kurtosisApiIp net.IP) (map[string]string, error) {
	debuggerPortIntStr := strconv.Itoa(launcher.debuggerPort.Int())
	// TODO Use ListenProtocol
	kurtosisApiSocket := fmt.Sprintf("%v:%v", kurtosisApiIp, api_container_server_consts.ListenPort)
	standardVars := map[string]string{
		test_suite_env_vars.KurtosisApiSocketEnvVar: kurtosisApiSocket,
		test_suite_env_vars.LogLevelEnvVar:          launcher.suiteLogLevel,
		test_suite_env_vars.DebuggerPortEnvVar:      debuggerPortIntStr,
		test_suite_env_vars.CustomParamsJson: launcher.customParamsJson,
	}
	return standardVars, nil
}

func genApiContainerEnvVars(
		logLevel logrus.Level,
		mode api_container_modes.ApiContainerMode,
		paramsJsonStr string) map[string]string {
	// TODO switch to the envVars requiring a visitor to hit, so we get them all
	return map[string]string{
		api_container_env_vars.LogLevelEnvVar: logLevel.String(),
		api_container_env_vars.ModeEnvVar: string(mode),
		api_container_env_vars.ParamsJsonEnvVar: paramsJsonStr,
	}
}

func (launcher TestsuiteContainerLauncher) genTestExecutionApiContainerEnvVars(
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		suiteExecutionVolName string,
		testName string,
		testSuiteContainerId string,
		testSuiteContainerIpAddr net.IP,
		apiContainerIpAddr net.IP,
		isPartitioningEnabled bool) (map[string]string, error) {
	args := server_core_creator.NewTestExecutionArgs(
		launcher.executionInstanceId.String(),
		networkId,
		subnetMask,
		gatewayIpAddr.String(),
		testName,
		suiteExecutionVolName,
		testSuiteContainerId,
		testSuiteContainerIpAddr.String(),
		apiContainerIpAddr.String(),
		isPartitioningEnabled)
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing API container test execution args to JSON")
	}
	argsStr := string(argsBytes)
	return genApiContainerEnvVars(
		launcher.kurtosisApiLogLevel,
		api_container_modes.TestExecutionMode,
		argsStr), nil
}

func (launcher TestsuiteContainerLauncher) genSuiteMetadataSerializationApiContainerEnvVars() (map[string]string, error) {
	args := server_core_creator.NewSuiteMetadataSerializationArgs()
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing API container suite metadata-serializing args to JSON")
	}
	argsStr := string(argsBytes)
	return genApiContainerEnvVars(
		launcher.kurtosisApiLogLevel,
		api_container_modes.SuiteMetadataSerializingMode,
		argsStr), nil

}
