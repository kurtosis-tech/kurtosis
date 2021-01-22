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
)

type TestsuiteContainerLauncher struct {
	kurtosisApiImage string

	testsuiteImage string

	// The log level string that will be passed as-is to the testsuite (should be meaningful to the testsuite)
	logLevel string

	// The testsuite-custom Docker environment variables that should be set in the testsuite container
	customEnvVars map[string]string

	// This is the port on the testsuite container that a debugger might be listening on, if any is running at all
	// We'll always bind this port on the testsuite container to a port on the user's machine, so they can attach
	//  a debugger if desired
	debuggerPort nat.Port
}

func NewTestsuiteContainerLauncher(
		testsuiteImage string,
		logLevel string,
		customEnvVars map[string]string,
		debuggerPort int) (*TestsuiteContainerLauncher, error) {
	debuggerPortObj, err := nat.NewPort(debuggerPortProtocol, strconv.Itoa(debuggerPort))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the debugger port object from port int '%v'", debuggerPort)
	}
	return &TestsuiteContainerLauncher{
		testsuiteImage: testsuiteImage,
		logLevel: logLevel,
		customEnvVars: customEnvVars,
		debuggerPort: debuggerPortObj,
	}, nil
}

/*
Launches a new testsuite container to acquire testsuite metadata
 */
func (launcher TestsuiteContainerLauncher) LaunchMetadataAcquiringContainer(
		context context.Context,
		dockerManager *docker_manager.DockerManager,
		bridgeNetworkId string,
		suiteExecutionVolume string,
		debuggerPortBinding nat.PortBinding,
		kurtosisApiIpAddr net.IP) (containerId string, err error) {
	envVars, err := launcher.generateTestSuiteEnvVars(kurtosisApiIpAddr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating the testsuite container env vars")
	}


	resultContainerId, err := dockerManager.CreateAndStartContainer(
		context,
		launcher.testsuiteImage,
		bridgeNetworkId,
		nil,                                           // Nil because the bridge network will assign IPs on its own (and won't know what IPs are already used)
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for testsuite containers
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			launcher.debuggerPort: &debuggerPortBinding,
		},
		nil, // Nil start command args because we expect the test suite image to be parameterized with variables
		envVars,
		map[string]string{},
		map[string]string{
			suiteExecutionVolume: TestsuiteContainerSuiteExVolMountpoint,
		})
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating the test suite container to acquire testsuite metadata")
	}
	return resultContainerId, nil
}

/*
Launches a new testsuite container to acquire testsuite metadata
*/
func (launcher TestsuiteContainerLauncher) LaunchTestRunningContainer(
		context context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		networkId string,
		suiteExecutionVolume string,
		testName string,
		kurtosisApiContainerIp net.IP,
		testsuiteContainerIp net.IP,
		servicesRelativeDirpath string,
		debuggerPortBinding nat.PortBinding) (containerId string, err error){
	log.Debugf(
		"Test suite container IP: %v; kurtosis API container IP: %v",
		testsuiteContainerIp.String(),
		kurtosisApiContainerIp.String())

	testSuiteEnvVars, err := launcher.generateTestSuiteEnvVars(kurtosisApiContainerIp)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating the test-running testsuite container env vars")
	}

	log.Info("Launching test-running testsuite container with debugger port bound to host port %v....", debuggerPortBinding)
	resultContainerId, err := dockerManager.CreateAndStartContainer(
		context,
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
		return "", stacktrace.Propagate(err, "An error occurred creating the test-running testsuite container")
	}
	log.Infof("Successfully created test-running testsuite container")



	log.Info("Creating Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/tcp", api_container_server_consts.ListenPort))
	kurtosisApiContainerEnvVars, err := buildApiContainerEnvVarsMap(
		kurtosisApiIp,
		apiLogFilepathOnApiContainer,
		executionInstanceId,
		gatewayIp,
		testMetadata.IsPartitioningEnabled,
		apiContainerLogLevel,
		networkId,
		subnetMask,
		testName,
		testRunningContainerId,
		testRunningContainerIp,
		suiteExecutionVolume)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the API container envvars map")
	}
	kurtosisApiContainerId, err := dockerManager.CreateAndStartContainer(
		testSetupContext,
		kurtosisApiImageName,
		networkId,
		kurtosisApiIp,
		map[docker_manager.ContainerCapability]bool{}, // No extra capabilities needed for the API container
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{
			kurtosisApiPort: nil,
		},
		nil,
		kurtosisApiContainerEnvVars,
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{
			suiteExecutionVolume: api_container_mountpoints.SuiteExecutionVolumeMountDirpath,
		},
	)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the Kurtosis API container")
	}
	log.Infof("Successfully created Kurtosis API container")

	return resultContainerId, nil
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
		test_suite_env_vars.LogLevelEnvVar:          launcher.logLevel,
		test_suite_env_vars.DebuggerPortEnvVar:      debuggerPortIntStr,
	}
	for key, val := range launcher.customEnvVars {
		if _, ok := standardVars[key]; ok {
			return nil, stacktrace.NewError(
				"Custom test suite environment variable binding %s=%s requested, but is not allowed because key is " +
					"already being used by Kurtosis.",
				key,
				val)
		}
		standardVars[key] = val
	}
	return standardVars, nil
}

func generateBaseApiContainerEnvVars(
		logLevel logrus.Level,
		mode api_container_env_vars.ApiContainerMode,
		paramsJsonStr string) {
	return map[string]string{
		api_container_env_vars.LogLevelEnvVar: logLevel.String(),

	}
}

func genTestExecutionApiContainerEnvVars(

		executionInstanceId uuid.UUID,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		testName string,
		suiteExecutionVolName string,
		testSuiteContainerId string,
		testSuiteContainerIpAddr net.IP,
		apiContainerIpAddr net.IP,
		isPartitioningEnabled bool) error {
	args := server_core_creator.NewTestExecutionArgs(
		executionInstanceId.String(),
		networkId,
		subnetMask,
		gatewayIpAddr.String(),
		testName,
		suiteExecutionVolName,
		testSuiteContainerId,
		testSuiteContainerIpAddr.String(),
		apiContainerIpAddr.String(),
		isPartitioningEnabled)
	argsStr, err := json.Marshal(args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing API container args to JSON")
	}
}

func genSuiteMetadataSerializationApiContainerEnvVars() error {

}

func buildApiContainerEnvVarsMap(
		apiContainerIp net.IP,
		logFilepathOnContainer string,
		executionInstanceId uuid.UUID,
		gatewayIp net.IP,
		isPartitioningEnabled bool,
		apiContainerLogLevel string,
		networkId string,
		subnetMask string,
		testName string,
		testRunningContainerId string,
		testRunningContainerIp net.IP,
		suiteExecutionVolumeName string) (map[string]string, error) {
	args := server_core_creator.NewTestExecutionArgs(
		executionInstanceId.String(),
		networkId,
		subnetMask,
		gatewayIp.String(),
		testName,
		suiteExecutionVolumeName,
		testRunningContainerId,
		testRunningContainerIp.String(),
		apiContainerIp.String(),
		isPartitioningEnabled)
	serializedArgsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the test execution args to JSON")
	}
	return map[string]string{
		api_container_env_vars.LogLevelEnvVar:   apiContainerLogLevel,
		api_container_env_vars.ModeEnvVar:       api_container_env_vars.TestExecutionMode,
		api_container_env_vars.ParamsJsonEnvVar: string(serializedArgsBytes),
	}, nil
}
