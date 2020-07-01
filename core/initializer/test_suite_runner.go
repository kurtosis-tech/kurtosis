package initializer

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"net"
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
	// Each Docker network created will have 2^this_num free IP addresses available
	NETWORK_WIDTH_BITS = 8

	// This is the IP address that the first Docker subnet will be doled out from, with subsequent Docker networks doled out with
	//  increasing IPs corresponding to the NETWORK_WIDTH_BITS
	SUBNET_START_ADDR = "172.23.0.0"

	BITS_IN_IP4_ADDR = 32

	CONTROLLER_LOG_MOUNT_FILEPATH = "/test-controller.log"

	TEST_VOLUME_MOUNTPOINT = "/shared"

	// These are an "API" of sorts - environment variables that are agreed to be set in the test controller's Docker environment
	TEST_VOLUME_ARG            = "TEST_VOLUME"
	TEST_NAME_BASH_ARG         = "TEST_NAME"
	NETWORK_NAME_ARG		   = "NETWORK_NAME"
	SUBNET_MASK_ARG            = "SUBNET_MASK"
	GATEWAY_IP_ARG             = "GATEWAY_IP"
	LOG_FILEPATH_ARG           = "LOG_FILEPATH"
	LOG_LEVEL_ARG              = "LOG_LEVEL"
	TEST_IMAGE_NAME_ARG        = "TEST_IMAGE_NAME"
	TEST_CONTROLLER_IP_ARG     = "TEST_CONTROLLER_IP"
	TEST_VOLUME_MOUNTPOINT_ARG = "TEST_VOLUME_MOUNTPOINT"

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

	subnetStartIp := net.ParseIP(SUBNET_START_ADDR)
	if subnetStartIp == nil {
		return nil, stacktrace.NewError("Subnet start IP %v was not a valid IP address; this is a code problem", SUBNET_START_ADDR)
	}
	logrus.Tracef("Subnet start IP: %v", subnetStartIp.String())
	subnetMaskBits := BITS_IN_IP4_ADDR - NETWORK_WIDTH_BITS

	executionInstanceId := uuid.Generate()
	logrus.Infof("Running %v tests with execution ID: %v", len(testsToRun), executionInstanceId.String())

	// TODO implement parallelism
	testResults := make(map[string]TestResult)
	testIndex := 0
	for testName, test := range testsToRun {
		logrus.Infof("---------------------------------- %v --------------------------------", testName)
		// Pick the next free available subnet IP, considering all the tests we've started previously
		subnetIpInt := binary.BigEndian.Uint32(subnetStartIp) + uint32(testIndex) * uint32(math.Pow(2, NETWORK_WIDTH_BITS))
		logrus.Tracef("subnetIpInt: %v", subnetIpInt)
		subnetIp := make(net.IP, 4)
		binary.BigEndian.PutUint32(subnetIp, subnetIpInt)
		logrus.Tracef("Subnet IP bytes after PutUint32: %v", subnetIp)
		subnetCidrStr := fmt.Sprintf("%v/%v", subnetIp.String(), subnetMaskBits)

		logrus.Debugf("Running test %v with subnet CIDR %v..", testName, subnetCidrStr)

		testPassed, err := runTest(
			dockerManager,
			subnetCidrStr,
			runner.testControllerImageName,
			runner.testControllerLogLevel,
			executionInstanceId,
			runner.testServiceImageName,
			testName,
			test)
		testResults[testName] = logTestResult(testName, err, testPassed)
		testIndex++
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
			subnetMask string,
			testControllerImageName string,
			testControllerLogLevel string,
			executionInstanceId uuid.UUID,
			testServiceImageName string,
			testName string,
			test testsuite.Test) (bool, error) {

	logrus.Infof("Creating Docker network for test...")
	networkName := fmt.Sprintf("%v-%v", executionInstanceId.String(), testName)
	publicIpProvider, err := networks.NewFreeIpAddrTracker(subnetMask, []string{})
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create the free IP addr tracker")
	}
	gatewayIp, err := publicIpProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the gateway IP")
	}
	_, err = dockerManager.CreateNetwork(networkName, subnetMask, gatewayIp)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error occurred creating docker network for testnet")
	}
	defer removeNetworkDeferredFunc(dockerManager, networkName)
	logrus.Infof("Docker network %v created successfully", networkName)

	logrus.Info("Running test controller...")
	controllerIp, err := publicIpProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.NewError("An error occurred getting an IP for the test controller")
	}
	testPassed, err := runControllerContainer(
		dockerManager,
		networkName,
		subnetMask,
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
	// TODO after printing logs, delete each container???
}



/*
Runs the controller container against the given test network.

Args:
	manager: the Docker manager, used for starting container & waiting for it to finish
	networkName: The name of the Docker network that the controller container will run in
	subnetMask: The CIDR representation of the network that the Docker network that the controller container is running in
	gatewayIp: The IP of the gateway on the Docker network that the controller is running in
	controllerIpAddr: The IP address that should be used for the container that the controller is running in
	controllerImageName: The name of the Docker image that should be used to run the controller container
	logLevel: A string representing the log level that the controller should use (will be passed as-is to the controller;
		should be semantically meaningful to the given controller image)
	testServiceImageName: The name of the Docker image that's being tested, and will be used for spinning up "test" nodes
		on the controller
	testName: Name of the test to tell the controller to run
	executionUuid: A UUID representing this specific execution of the test suite

Returns:
	bool: true if the test succeeded, false if not
	error: if any error occurred during the execution of the controller (independent of the test itself)
 */
func runControllerContainer(
		manager *docker.DockerManager,
		networkName string,
		subnetMask string,
		gatewayIp string,
		controllerIpAddr string,
		controllerImageName string,
		logLevel string,
		testServiceImageName string,
		testName string,
		executionUuid uuid.UUID) (bool, error){
	volumeName := fmt.Sprintf("%v-%v", executionUuid.String(), testName)
	_, err := manager.CreateVolume(volumeName)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error creating Docker volume to share amongst test nodes")
	}

	testControllerLogFilename := fmt.Sprintf("%v-%v-controller-logs", executionUuid.String(), executionUuid.String())
	logTmpFile, err := ioutil.TempFile("", testControllerLogFilename)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create tempfile to store log info for passing to test controller")
	}
	logTmpFile.Close()
	logrus.Debugf("Temp filepath to write log file to: %v", logTmpFile.Name())

	envVariables := generateTestControllerEnvVariables(
		networkName,
		subnetMask,
		gatewayIp,
		controllerIpAddr,
		testName,
		logLevel,
		testServiceImageName,
		volumeName)
	logrus.Debugf("Environment variables that are being passed to the controller: %v", envVariables)

	_, controllerContainerId, err := manager.CreateAndStartContainer(
		controllerImageName,
		networkName,
		controllerIpAddr,
		make(map[int]bool),
		nil, // Use the default image CMD (which is parameterized)
		envVariables,
		map[string]string{
			// Because the test controller will need to spin up new images, we need to bind-mount the host Docker engine into the test controller
			"/var/run/docker.sock": "/var/run/docker.sock",
			logTmpFile.Name(): CONTROLLER_LOG_MOUNT_FILEPATH,
		},
		map[string]string{
			volumeName: TEST_VOLUME_MOUNTPOINT,
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
	buf, err := ioutil.ReadFile(logTmpFile.Name())
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to read log file from controller.")
	}
	logrus.Infof("Printing Controller logs:")
	logrus.Info(string(buf))

	// TODO Clean up the volumeFilepath we created!
	return exitCode == SUCCESS_EXIT_CODE, nil
}

/*
Helper function for making a best-effort attempt at removing a network and logging any error states; intended to be run
as a deferred function.
 */
func removeNetworkDeferredFunc(dockerManager *docker.DockerManager, networkName string) {
	logrus.Infof("Attempting to remove Docker network with name %v...", networkName)
	err := dockerManager.RemoveNetwork(networkName)
	if err != nil {
		logrus.Errorf("An error occurred removing Docker network with name %v:", networkName)
		logrus.Error(err.Error())
	} else {
		logrus.Infof("Docker network %v successfully removed", networkName)
	}
}

/*
NOTE: This is a separate function because it provides a nice documentation reference point, where we can say to users,
"to see the latest special environment variables that will be passed to the test controller, see this function". Do not
put anything else in this function!!!

Args:
	networkName: The name of the Docker network that the test controller is running in, and which all services should be started in
	subnetMask: The subnet mask used to create the Docker network that the test controller, and all services it starts, are running in
	gatewayIp: The IP of the gateway of the Docker network that the test controller will run inside
	controllerIpAddr: The IP address of the container running the test controller
	testName: The name of the test that the test controller should run
	logLevel: A string representing the controller's loglevel (NOTE: this should be interpretable by the controller; the
		initializer will not know what to do with this!)
	testServiceImageName: The name of the Docker image of the service that we're testing
	testVolumeName: The name of the Docker volume that has been created for this particular test execution, and that the
		test controller can share with the services that it spins up to read and write data to them
*/
func generateTestControllerEnvVariables(
			networkName string,
			subnetMask string,
			gatewayIp string,
			controllerIpAddr string,
			testName string,
			logLevel string,
			testServiceImageName string,
			testVolumeName string) map[string]string {
	return map[string]string{
		TEST_NAME_BASH_ARG:         testName,
		SUBNET_MASK_ARG:            subnetMask,
		NETWORK_NAME_ARG:           networkName,
		GATEWAY_IP_ARG:             gatewayIp,
		LOG_FILEPATH_ARG:           CONTROLLER_LOG_MOUNT_FILEPATH,
		LOG_LEVEL_ARG:              logLevel,
		TEST_IMAGE_NAME_ARG:        testServiceImageName,
		TEST_CONTROLLER_IP_ARG:     controllerIpAddr,
		TEST_VOLUME_ARG:            testVolumeName,
		TEST_VOLUME_MOUNTPOINT_ARG: TEST_VOLUME_MOUNTPOINT,
	}
}

