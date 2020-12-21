/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_constants

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"net"
	"strconv"
)

const (
	debuggerPortProtocol = "tcp"
)

type TestsuiteContainerLauncher struct {
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
		dockerManager *commons.DockerManager,
		bridgeNetworkId string,
		suiteExecutionVolume string,
		metadataFilepathOnTestsuiteContainer string,
		debuggerPortBinding nat.PortBinding) (containerId string, err error) {
	envVars, err := launcher.generateTestSuiteEnvVars(
		metadataFilepathOnTestsuiteContainer,
		"", // We leave the test name blank to signify that we want test listing, not test execution
		"", // Because we're doing test listing, not test execution, the Kurtosis API IP can be blank
		"", // We leave the services dirpath blank because getting suite metadata doesn't require knowing this
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating the metadata-acquiring testsuite container env vars")
	}

	resultContainerId, err := dockerManager.CreateAndStartContainer(
		context,
		launcher.testsuiteImage,
		bridgeNetworkId,
		nil,  // Nil because the bridge network will assign IPs on its own (and won't know what IPs are already used)
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
		dockerManager *commons.DockerManager,
		networkId string,
		suiteExecutionVolume string,
		testName string,
		kurtosisApiIpStr string,
		testsuiteContainerIp net.IP,
		servicesRelativeDirpath string,
		debuggerPortBinding nat.PortBinding) (containerId string, err error){
	testSuiteEnvVars, err := launcher.generateTestSuiteEnvVars(
		"",  // We're executing a test, not getting metadata, so this should be blank
		testName,
		kurtosisApiIpStr,
		servicesRelativeDirpath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating the test-running testsuite container env vars")
	}

	resultContainerId, err := dockerManager.CreateAndStartContainer(
		context,
		launcher.testsuiteImage,
		networkId,
		testsuiteContainerIp,
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
		return "", stacktrace.Propagate(err, "An error occurred creating the test suite container to run the test")
	}
	return resultContainerId, nil
}

/*
Generates the map of environment variables needed to run a test suite container

NOTE: exactly one of metadata_filepath or test_name must be non-empty!
*/
func (launcher TestsuiteContainerLauncher) generateTestSuiteEnvVars(
		metadataFilepath string,
		testName string,
		kurtosisApiIp string,
		servicesRelativeDirpath string) (map[string]string, error){
	debuggerPortIntStr := strconv.Itoa(launcher.debuggerPort.Int())
	standardVars := map[string]string{
		metadataFilepathEnvVar:        metadataFilepath,
		testEnvVar:                    testName,
		kurtosisApiIpEnvVar:           kurtosisApiIp,
		servicesRelativeDirpathEnvVar: servicesRelativeDirpath,
		logLevelEnvVar:                launcher.logLevel,
		debuggerPort: 				   debuggerPortIntStr,
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
