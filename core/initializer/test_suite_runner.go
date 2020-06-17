package initializer

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"sort"
	"time"
)

type TestResult string
// "enum" for TestResult
const (
	PASSED TestResult = "PASSED"
	FAILED TestResult = "FAILED"
	ERRORED TestResult = "ERRORED"   // Indicates an error during setup that prevented the test from running
)

type TestSuiteRunner struct {
	testSuite               testsuite.TestSuite
	testServiceImageName    string
	testControllerImageName string

	// The test controller image-specific string representing the log level, that will be passed as-is to the test controller
	testControllerLogLevel	string
	startPortRange          int
	endPortRange            int
}

const (
	DEFAULT_SUBNET_MASK = "172.23.0.0/16"

	CONTAINER_NETWORK_INFO_MOUNTED_FILEPATH = "/data/network/network-info"
	CONTAINER_LOG_INFO_MOUNTED_FILEPATH = "/data/service/container-log"

	// These are an "API" of sorts - environment variables that are agreed to be set in the test controller's Docker environment
	TEST_NAME_BASH_ARG = "TEST_NAME"
	NETWORK_FILEPATH_ARG = "NETWORK_DATA_FILEPATH"
	LOG_FILEPATH_ARG = "LOG_FILEPATH"
	LOG_LEVEL_ARG = "LOG_LEVEL"

	// How long to wait before force-killing a container
	CONTAINER_STOP_TIMEOUT = 30 * time.Second

	SUCCESS_EXIT_CODE = 0
)


/*
Creates a new TestSuiteRunner with the following arguments
 */
func NewTestSuiteRunner(
			testSuite testsuite.TestSuite,
			testServiceImageName string,
			testControllerImageName string,
			testControllerLogLevel string,
			startPortRange int,
			endPortRange int) *TestSuiteRunner {
	return &TestSuiteRunner{
		testSuite:               testSuite,
		testServiceImageName:    testServiceImageName,
		testControllerImageName: testControllerImageName,
		testControllerLogLevel: testControllerLogLevel,
		startPortRange:          startPortRange,
		endPortRange:            endPortRange,
	}
}

/*
Runs the tests with the given names. If no tests are specifically defined, all tests are run.
 */
func (runner TestSuiteRunner) RunTests(testNamesToRun []string) (map[string]TestResult, error) {
	// Initialize default environment context.
	dockerCtx := context.Background()
	// Initialize a Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err,"Failed to initialize Docker client from environment.")
	}

	dockerManager, err := docker.NewDockerManager(dockerCtx, dockerClient, runner.startPortRange, runner.endPortRange)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in initializing Docker Manager.")
	}

	allTests := runner.testSuite.GetTests()
	if len(testNamesToRun) == 0 {
		testNamesToRun = make([]string, 0, len(runner.testSuite.GetTests()))
		for testName, _ := range allTests {
			testNamesToRun = append(testNamesToRun, testName)
		}
	}
	sort.Strings(testNamesToRun)

	// Validate all the requested tests exist
	testsToRun := make(map[string]testsuite.Test)
	for _, testName := range testNamesToRun {
		test, found := allTests[testName]
		if !found {
			return nil, stacktrace.NewError("No test registered with name '%v'", testName)
		}
		testsToRun[testName] = test
	}


	executionInstanceId := uuid.Generate()
	logrus.Infof("Running %v tests with execution ID: %v", len(testsToRun), executionInstanceId.String())

	// TODO TODO TODO Support creating one network per testnet
	_, err = dockerManager.CreateNetwork(DEFAULT_SUBNET_MASK)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error in creating docker subnet for testnet.")
	}

	// TODO break everything inside this for loop into its own function for readability
	// TODO implement parallelism
	testResults := make(map[string]TestResult)
	for testName, test := range testsToRun {
		logrus.Infof("---------------------------------- %v --------------------------------", testName)
		networkLoader, err := test.GetNetworkLoader()
		if err != nil {
			testResults[testName] = logTestResult(testName, err, false)
			continue
		}

		builder := networks.NewServiceNetworkConfigBuilder()
		if err := networkLoader.ConfigureNetwork(builder); err != nil {
			testResults[testName] = logTestResult(testName, err, false)
			continue
		}
		testNetworkCfg := builder.Build()

		logrus.Infof("Creating network for test...")
		publicIpProvider, err := networks.NewFreeIpAddrTracker(DEFAULT_SUBNET_MASK)
		if err != nil {
			testResults[testName] = logTestResult(testName, err, false)
			continue
		}
		serviceNetwork, err := testNetworkCfg.CreateNetwork(runner.testServiceImageName, publicIpProvider, dockerManager)
		if err != nil {
			testResults[testName] = logTestResult(testName, err, false)
			continue
		}
		logrus.Info("Network created successfully")

		testPassed, err := runControllerContainer(
			dockerManager,
			serviceNetwork,
			runner.testControllerImageName,
			runner.testControllerLogLevel,
			publicIpProvider,
			testName,
			executionInstanceId)
		if err != nil {
			testResults[testName] = logTestResult(testName, err, false)
			stopNetwork(dockerManager, serviceNetwork, CONTAINER_STOP_TIMEOUT)
			continue
		}
		stopNetwork(dockerManager, serviceNetwork, CONTAINER_STOP_TIMEOUT)
		testResults[testName] = logTestResult(testName, nil, testPassed)
		// TODO after the service containers have been changed to write logs to disk, print each container's logs here for convenience
		// TODO after printing logs, delete each container
	}

	return testResults, nil
}

// ======================== Private helper functions =====================================

/*
Handles determining the proper test status and logging it.
Returns the TestResult for convenience.
 */
func logTestResult(testName string, err error, testPassed bool) TestResult {
	var result TestResult
	if err != nil {
		result = ERRORED
	} else {
		if testPassed {
			result = PASSED
		} else {
			result = FAILED
		}
	}

	switch result {
	case ERRORED:
		logrus.Warnf("Test %v %v", testName, result)
		logrus.Warnf("Error reason: %v", err)
	case PASSED:
		logrus.Infof("Test %v %v", testName, result)
	case FAILED:
		logrus.Warnf("Test %v %v", testName, result)
	}
	return result
}


/*
Runs the controller container against the given test network.

Args:
	manager: the Docker manager, used for starting container & waiting for it to finish
	rawServiceNetwork: the network to run against
	dockerImage: the Docker image of the controller that will be started
	ipProvider: IP provider to give the controller container its address
	testName: name of test to run
	executionUuid: UUID tying together all the tests in this run of the test suite

Returns:
	bool: true if the test succeeded, false if not
	error: if any error occurred during the execution of the controller (independent of the test itself)
 */
func runControllerContainer(
		manager *docker.DockerManager,
		rawServiceNetwork *networks.RawServiceNetwork,
		dockerImage string,
		logLevel string,
		ipProvider *networks.FreeIpAddrTracker,
		testName string,
		executionUuid uuid.UUID) (bool, error){
	logrus.Info("Launching controller container...")

	// Using tempfiles, is there a risk that for a verrrry long-running E2E test suite the filesystem cleans up the tempfile
	//  out from underneath the test??
	testNetworkInfoFilename := fmt.Sprintf("%v-%v", testName, executionUuid.String())
	networkInfoTmpFile, err := ioutil.TempFile("", testNetworkInfoFilename)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create tempfile to store network info for passing to test controller")
	}

	logrus.Debugf("Temp filepath to write network file to: %v", networkInfoTmpFile.Name())
	logrus.Debugf("Raw service network contents: %v", rawServiceNetwork)

	encoder := gob.NewEncoder(networkInfoTmpFile)
	err = encoder.Encode(rawServiceNetwork)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not write service network state to file")
	}
	// Apparently, per https://www.joeshaw.org/dont-defer-close-on-writable-files/ , file.Close() can return errors too,
	//  but because this is a networkInfoTmpFile we don't fuss about checking them
	defer networkInfoTmpFile.Close()

	testControllerLogFilename := fmt.Sprintf("%v-%v-%s", testName, executionUuid.String(), "logs")
	logTmpFile, err := ioutil.TempFile("", testControllerLogFilename)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create tempfile to store log info for passing to test controller")
	}
	logrus.Debugf("Temp filepath to write log file to: %v", logTmpFile.Name())

	containerNetworkInfoMountpoint := CONTAINER_NETWORK_INFO_MOUNTED_FILEPATH
	containerLogInfoMountpoint := CONTAINER_LOG_INFO_MOUNTED_FILEPATH
	envVariables := map[string]string{
		TEST_NAME_BASH_ARG: testName,
		NETWORK_FILEPATH_ARG: containerNetworkInfoMountpoint,
		LOG_FILEPATH_ARG: containerLogInfoMountpoint,
		LOG_LEVEL_ARG: logLevel,
	}

	ipAddr, err := ipProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not get free IP address to assign the test controller")
	}

	_, controllerContainerId, err := manager.CreateAndStartContainer(
		dockerImage,
		ipAddr,
		make(map[int]bool),
		nil, // Use the default image CMD (which is parameterized)
		envVariables,
		map[string]string{
			networkInfoTmpFile.Name(): containerNetworkInfoMountpoint,
			logTmpFile.Name():         containerLogInfoMountpoint,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to run test controller container")
	}
	logrus.Infof("Controller container started successfully with id %s", controllerContainerId)

	logrus.Info("Waiting for controller container to exit...")
	// TODO add a timeout here if the test doesn't complete successfully
	exitCode, err := manager.WaitForExit(controllerContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed when waiting for controller to exit")
	}

	logrus.Info("Controller container exited successfully")
	logTmpFile.Close()
	buf, err := ioutil.ReadFile(logTmpFile.Name())
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to read log file from controller.")
	}
	logrus.Infof("Printing Controller logs:")
	logrus.Info(string(buf))

	return exitCode == SUCCESS_EXIT_CODE, nil
}

/*
Makes a best-effort attempt to stop the containers in the given network, waiting for the given timeout.
It is safe to pass in nil to the networkToStop
 */
func stopNetwork(manager *docker.DockerManager, networkToStop *networks.RawServiceNetwork, stopTimeout time.Duration) {
	logrus.Info("Stopping service container network...")
	for _, containerId := range networkToStop.ContainerIds {
		logrus.Debugf("Stopping container with ID '%v'", containerId)
		err := manager.StopContainer(containerId, &stopTimeout)
		if err != nil {
			logrus.Errorf("An error occurred stopping container ID '%v'; proceeding to stop other containers:", containerId)
			logrus.Error(err)
		}
		logrus.Debugf("Container with ID '%v' successfully stopped", containerId)
	}
	logrus.Info("Finished stopping service container network")
}
