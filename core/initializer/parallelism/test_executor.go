package parallelism

import (
	"context"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"time"
)

/*
WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
No logging to the system-level logger is allowed in this file!!! Everything should use the specific
logger passed in at construction time!!
WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
 */

const (
	containerSuccessExitCode = 0

	controllerLogMountFilepath = "/test-controller.log"

	testVolumeMountpoint = "/shared"

	// These are an "API" of sorts - environment variables that are agreed to be set in the test controller's Docker environment
	testVolumeArg           = "TEST_VOLUME"
	testNameArg             = "TEST_NAME"
	networkNameArg          = "NETWORK_NAME"
	subnetMaskArg           = "SUBNET_MASK"
	gatewayIpArg            = "GATEWAY_IP"
	logFilepathArg          = "LOG_FILEPATH"
	logLevelArg             = "LOG_LEVEL"
	testControllerIpArg     = "TEST_CONTROLLER_IP"
	testVolumeMountpointArg = "TEST_VOLUME_MOUNTPOINT"

	// After we hard-timeout a test, how long we'll give the test to clean itself up (namely the Docker network & containers)
	//  before we call it lost and continue on
	networkTeardownGraceTime = 60 * time.Second

	// When we're tearing down a network after a test (either after normal exit or test timeout), this is the maximum
	//  time we'll wait for each container to stop
	networkTeardownContainerStopTimeout = 10 * time.Second
)

// Because a test is run in its own goroutine to allow us to time it out, we need to pass the results back
//  via a channel. Tihs struct is what's passed over the chanel.
type testResult struct {
	// Whether the test passed or not (undefined if an error occurred that prevented us from retrieving test results)
	testPassed   bool

	// If not nil, the error that prevented us from retrieving the test result
	executionErr error
}

// Executor responsible for running a test, with timeout, cleaning up after the test as needed
type testExecutor struct {
	log *logrus.Logger

	// Tests already declare timeouts, but that timeout only represents time spent running the actual test
	// This value represents how much time *on top of* the test timeout we'll give to the test goroutine for it to do
	//  do test setup and finalization (and if this total timeout is exceeded, we'll try to shut down everything in
	//  the test - the controller, the containers in the network, the Docker network itself, etc.)
	additionalTestTimeoutBuffer time.Duration
}

func newTestExecutor(log *logrus.Logger, additionalTestTimeoutBuffer time.Duration) *testExecutor {
	return &testExecutor{
		log: log,
		additionalTestTimeoutBuffer: additionalTestTimeoutBuffer,
	}
}

// TODO Just make all these params struct params - passing them through each to function is tedious and error-prone because
//  we have a lot of string params and it's easy to get them mixed up
/*
Returns:
	bool: A boolean indicating if the test passed (will be undefined if the test result couldn't be retrieved for any reason)
	error: If not nil, represents the error hit while running the test that prevented the retrieval of the test result
 */
func (executor testExecutor) runTest(
		executionInstanceId uuid.UUID,
		dockerClient *client.Client,
		subnetMask string,
		testControllerImageName string,
		testControllerLogLevel string,
		customTestControllerEnvVars map[string]string,
		testName string,
		test testsuite.Test) (bool, error) {
	testResultChan := make(chan testResult)

	// When this is breached, we'll try to tear down everything
	totalTimeout := test.GetTimeout() + executor.additionalTestTimeoutBuffer

	context, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// We run the test in a separate goroutine because we don't know if the test will even respect the context we pass in -
	//  we hope so, but (because this runs user-written code) we can't trust it so we give ourselves the option to move
	//  on if the test, e.g., infinite-loops
	go func() {
		testPassed, setupErr := executor.runTestGoroutine(
			context,
			executionInstanceId,
			dockerClient,
			subnetMask,
			testControllerImageName,
			testControllerLogLevel,
			customTestControllerEnvVars,
			testName)
		testResultChan <- testResult{
			testPassed:   testPassed,
			executionErr: setupErr,
		}
	}()

	var timedOut bool
	var testExecutionResult testResult
	select {
	case testExecutionResult = <- testResultChan:
		timedOut = false
	case <- time.After(totalTimeout):
		timedOut = true
	}

	if timedOut {
		executor.log.Tracef("Hit hard test timeout of %v; the context is being cancelled to give it the chance to exit gracefully...", totalTimeout)
		cancelFunc()

		// We've now cancelled the context so the test goroutine *should* exit gracefully soon
		select {
		case testExecutionResult = <- testResultChan:
			executor.log.Info("Test goroutine exited gracefully after context cancellation")
		case <- time.After(networkTeardownGraceTime):
			executor.log.Warnf(
				"Test goroutine didn't exit gracefully after context cancellation even after a grace period of %v; the test goroutine is being called lost",
				networkTeardownGraceTime,
			)
		}
		return false, stacktrace.NewError("Test hit hard timeout, %v", totalTimeout)
	} else {
		return testExecutionResult.testPassed, testExecutionResult.executionErr
	}
}

/*
Returns:
	error: If an error occurred that prevented us from running the test & retrieving the results (independent from whether the test itself passed)
	bool: A boolean indicating whether the test passed (undefined if an error occurred running the test)
*/
func (executor testExecutor) runTestGoroutine(
		context context.Context,
		executionInstanceId uuid.UUID,
		dockerClient *client.Client,
		subnetMask string,
		testControllerImageName string,
		testControllerLogLevel string,
		customTestControllerEnvVars map[string]string,
		testName string) (bool, error) {
	executor.log.Info("Creating Docker manager from environment settings...")
	// NOTE: at this point, all Docker commands from here forward will be bound by the Context that we pass in here - we'll
	//  only need to cancel this context once
	dockerManager, err := docker.NewDockerManager(executor.log, dockerClient)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the Docker manager for test %v", testName)
	}
	executor.log.Info("Docker manager created successfully")

	executor.log.Infof("Creating Docker network for test with subnet mask %v...", subnetMask)
	networkName := fmt.Sprintf("%v-%v", executionInstanceId.String(), testName)
	publicIpProvider, err := networks.NewFreeIpAddrTracker(executor.log, subnetMask, []string{})
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create the free IP address tracker")
	}
	gatewayIp, err := publicIpProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the gateway IP")
	}
	_, err = dockerManager.CreateNetwork(context, networkName, subnetMask, gatewayIp)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error occurred creating Docker network %v for test %v", networkName, testName)
	}
	defer removeNetworkDeferredFunc(executor.log, dockerManager, networkName)
	executor.log.Infof("Docker network %v created successfully", networkName)

	executor.log.Info("Running test controller...")
	controllerIp, err := publicIpProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.NewError("An error occurred getting an IP for the test controller")
	}
	testPassed, err := runControllerContainer(
		context,
		executor.log,
		dockerManager,
		networkName,
		subnetMask,
		gatewayIp,
		controllerIp,
		testControllerImageName,
		testControllerLogLevel,
		customTestControllerEnvVars,
		testName,
		executionInstanceId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while running the test, independent of test success")
	}
	executor.log.Info("The test controller ran and exited successfully")

	return testPassed, nil
}

// TODO Move many of these args to the struct itself to simplify the params of this function
/*
Helper function to run the controller container against the given test network.

Args:
	log: The test-specific logger to write log messages to
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
	customTestControllerEnvVars: Custom user-defined environment variables that will be passed to the test controller
	testName: Name of the test to tell the controller to run
	executionUuid: A UUID representing this specific execution of the test suite

Returns:
	bool: true if the test succeeded, false if not
	error: if any error occurred during the execution of the controller (independent of the test itself)
*/
func runControllerContainer(
			context context.Context,
			log *logrus.Logger,
			manager *docker.DockerManager,
			networkName string,
			subnetMask string,
			gatewayIp string,
			controllerIpAddr string,
			controllerImageName string,
			logLevel string,
			customTestControllerEnvVars map[string]string,
			testName string,
			executionUuid uuid.UUID) (bool, error){
	volumeName := fmt.Sprintf("%v-%v", executionUuid.String(), testName)
	log.Debugf("Creating Docker volume %v which will be shared with the test network...", volumeName)
	if err := manager.CreateVolume(context, volumeName); err != nil {
		return false, stacktrace.Propagate(err, "Error creating Docker volume to share amongst test nodes")
	}
	log.Debugf("Docker volume %v created successfully", volumeName)

	testControllerLogFilename := fmt.Sprintf("%v-%v-controller-logs", executionUuid.String(), executionUuid.String())
	log.Debugf("Creating temporary file with name %v to store controller logs...", testControllerLogFilename)
	logTmpFile, err := ioutil.TempFile("", testControllerLogFilename)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create tempfile to store log info for passing to test controller")
	}
	logTmpFile.Close()
	log.Debugf("Successfully created temporary file to store controller logs at path %v", logTmpFile.Name())

	envVariables, err := generateTestControllerEnvVariables(
		networkName,
		subnetMask,
		gatewayIp,
		controllerIpAddr,
		testName,
		logLevel,
		volumeName,
		customTestControllerEnvVars)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to map test controller environment variables.")
	}
	log.Debugf("Environment variables that are being passed to the controller: %v", envVariables)

	bindMounts := map[string]string{
		// Because the test controller will need to spin up new images, we need to bind-mount the host Docker engine into the test controller
		"/var/run/docker.sock": "/var/run/docker.sock",
		logTmpFile.Name():      controllerLogMountFilepath,
	}

	volumeMounts := map[string]string{
		volumeName: testVolumeMountpoint,
	}

	_, controllerContainerId, err := manager.CreateAndStartContainer(
		context,
		controllerImageName,
		networkName,
		controllerIpAddr,
		make(map[nat.Port]bool),
		nil, // The controller image's CMD should be parameterized, so we don't specify a start command here
		envVariables,
		bindMounts,
		volumeMounts)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to run test controller container")
	}
	log.Infof("Controller container started successfully with id %s", controllerContainerId)

	log.Info("Waiting for controller container to exit...")
	exitCode, err := manager.WaitForExit(context, controllerContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed when waiting for controller to exit")
	}
	log.Info("Controller container exited successfully")

	// We open a new fp for reading because our original FP is only for writing
	log.Info("- - - - - - - - - - - - - - - - - - - CONTROLLER LOGS - - - - - - - - - - - - - - - - - -")
	logReadFp, err := os.Open(logTmpFile.Name())
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to open controller log file for reading")
	}
	io.Copy(log.Out, logReadFp)
	log.Info("- - - - - - - - - - - - - - - - - - END CONTROLLER LOGS - - - - - - - - - - - - - - - - - -")
	logReadFp.Close()
	os.Remove(logTmpFile.Name()) // We're responsible for removing the tempfile we created

	return exitCode == containerSuccessExitCode, nil
}

/*
Helper function for making a best-effort attempt at removing a network and logging any error states; intended to be run
as a deferred function.
*/
func removeNetworkDeferredFunc(log *logrus.Logger, dockerManager *docker.DockerManager, networkName string) {
	log.Infof("Attempting to remove Docker network with name %v...", networkName)
	// We use the background context here because we want to try and tear down the network even if the context the test was running in
	//  was cancelled. This might not be right - the right way to do it might be to pipe a separate context for the network teardown to here!
	if err := dockerManager.RemoveNetwork(context.Background(), networkName, networkTeardownContainerStopTimeout); err != nil {
		log.Errorf("An error occurred removing Docker network with name %v:", networkName)
		log.Error(err.Error())
		log.Error("NOTE: This means you will need to clean up the Docker network manually!!")
	} else {
		log.Infof("Docker network %v successfully removed", networkName)
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
	testVolumeName: The name of the Docker volume that has been created for this particular test execution, and that the
		test controller can share with the services that it spins up to read and write data to them
	customEnvVars: A custom user-defined map from <env variable name> -> <env variable value> that will be set for test controller
*/
func generateTestControllerEnvVariables(
			networkName string,
			subnetMask string,
			gatewayIp string,
			controllerIpAddr string,
			testName string,
			logLevel string,
			testVolumeName string,
			customEnvVars map[string]string) (map[string]string, error) {
	standardVars := map[string]string{
		testNameArg:             testName,
		subnetMaskArg:           subnetMask,
		networkNameArg:          networkName,
		gatewayIpArg:            gatewayIp,
		logFilepathArg:          controllerLogMountFilepath,
		logLevelArg:             logLevel,
		testControllerIpArg:     controllerIpAddr,
		testVolumeArg:           testVolumeName,
		testVolumeMountpointArg: testVolumeMountpoint,
	}
	for key, val := range customEnvVars {
		if _, ok := standardVars[key]; ok {
			return nil, stacktrace.NewError(
				"Tried to manually add custom environment variable %s to the test controller container, but it is already being used by Kurtosis.",
				key)
		}
		standardVars[key] = val
	}
	return standardVars, nil
}
