package initializer

import (
	"context"
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
}

const (
	DEFAULT_SUBNET_MASK = "172.23.0.0/16"

	CONTAINER_LOG_INFO_MOUNTED_FILEPATH = "/data/service/container-log"

	// These are an "API" of sorts - environment variables that are agreed to be set in the test controller's Docker environment
	TEST_NAME_BASH_ARG = "TEST_NAME"
	SUBNET_MASK_ARG = "SUBNET_MASK"
	GATEWAY_IP_ARG = "GATEWAY_IP_ARG"
	LOG_FILEPATH_ARG = "LOG_FILEPATH"
	LOG_LEVEL_ARG = "LOG_LEVEL"
	TEST_IMAGE_NAME_ARG = "TEST_IMAGE_NAME"
	TEST_CONTROLLER_IP_ARG = "TEST_CONTROLLER_IP"

	SUCCESS_EXIT_CODE = 0
)


/*
Creates a new TestSuiteRunner with the following arguments
 */
func NewTestSuiteRunner(
			testSuite testsuite.TestSuite,
			testServiceImageName string,
			testControllerImageName string,
			testControllerLogLevel string) *TestSuiteRunner {
	return &TestSuiteRunner{
		testSuite:               testSuite,
		testServiceImageName:    testServiceImageName,
		testControllerImageName: testControllerImageName,
		testControllerLogLevel: testControllerLogLevel,
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

	dockerManager, err := docker.NewDockerManager(dockerCtx, dockerClient)
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

	// TODO break everything inside this for loop into its own function for readability
	// TODO implement parallelism
	testResults := make(map[string]TestResult)
	for testName, test := range testsToRun {
		logrus.Infof("---------------------------------- %v --------------------------------", testName)
		testPassed, err := runTest(
			dockerManager,
			runner.testControllerImageName,
			runner.testControllerLogLevel,
			executionInstanceId,
			runner.testServiceImageName,
			testName,
			test)
		testResults[testName] = logTestResult(testName, err, testPassed)
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

func runTest(
			dockerManager *docker.DockerManager,
			testControllerImageName string,
			testControllerLogLevel string,
			executionInstanceId uuid.UUID,
			testServiceImageName string,
			testName string,
			test testsuite.Test) (bool, error) {

	logrus.Infof("Creating Docker network for test...")
	publicIpProvider, err := networks.NewFreeIpAddrTracker(DEFAULT_SUBNET_MASK, []string{})
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create the free IP addr tracker")
	}
	gatewayIp, err := publicIpProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the gateway IP")
	}
	// TODO TODO TODO Support creating one network per testnet
	_, err = dockerManager.CreateNetwork(DEFAULT_SUBNET_MASK, gatewayIp)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error occurred creating docker network for testnet")
	}
	// TODO When we have one-network-per-testnet, defer a deletion of the Docker network
	logrus.Info("Docker network created successfully")

	logrus.Info("Running test controller...")
	controllerIp, err := publicIpProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.NewError("An error occurred getting an IP for the test controller")
	}
	testPassed, err := runControllerContainer(
		dockerManager,
		gatewayIp,
		controllerIp,
		testControllerImageName,
		testControllerLogLevel,
		testServiceImageName,
		testName,
		executionInstanceId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while running the test")
	}
	return testPassed, nil
	// TODO after printing logs, delete each container
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
		gatewayIp string,
		controllerIpAddr string,
		controllerImageName string,
		logLevel string,
		testServiceImageName string,
		testName string,
		executionUuid uuid.UUID) (bool, error){
	testControllerLogFilename := fmt.Sprintf("%v-%v-%s", testName, executionUuid.String(), "logs")
	logTmpFile, err := ioutil.TempFile("", testControllerLogFilename)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create tempfile to store log info for passing to test controller")
	}
	logrus.Debugf("Temp filepath to write log file to: %v", logTmpFile.Name())

	containerLogInfoMountpoint := CONTAINER_LOG_INFO_MOUNTED_FILEPATH
	envVariables := map[string]string{
		TEST_NAME_BASH_ARG:  testName,
		SUBNET_MASK_ARG:     DEFAULT_SUBNET_MASK,
		GATEWAY_IP_ARG:      gatewayIp,
		LOG_FILEPATH_ARG:    containerLogInfoMountpoint,
		LOG_LEVEL_ARG:       logLevel,
		TEST_IMAGE_NAME_ARG: testServiceImageName,
		TEST_CONTROLLER_IP_ARG: controllerIpAddr,
	}

	_, controllerContainerId, err := manager.CreateAndStartContainer(
		controllerImageName,
		controllerIpAddr,
		make(map[int]bool),
		nil, // Use the default image CMD (which is parameterized)
		envVariables,
		map[string]string{
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

