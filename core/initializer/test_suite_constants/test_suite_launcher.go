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
	"github.com/gogo/protobuf/protoc-gen-gogo/testdata/import_public/sub"
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

type TestsuiteContainerLauncher struct {
	executionInstanceId uuid.UUID

	kurtosisApiImage string

	kurtosisApiLogLevel logrus.Level

	testsuiteImage string

	// The log level string that will be passed as-is to the testsuite (should be meaningful to the testsuite)
	suiteLogLevel string

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
		// TODO add all the fields
		kurtosisApiImage: kurtosisApiImage,
		kurtosisApiLogLevel: kurtosisApiLogLevel,
		testsuiteImage: testsuiteImage,
		suiteLogLevel:  logLevel,
		customEnvVars:  customEnvVars,
		debuggerPort:   debuggerPortObj,
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
