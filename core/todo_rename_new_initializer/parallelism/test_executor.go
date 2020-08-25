package parallelism

import (
	"context"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/execution/exit_codes"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/todo_rename_new_initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/todo_rename_new_initializer/test_suite_env_vars"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"
)

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

No logging to the system-level logger is allowed in this file!!! Everything should use the specific logger passed
	in at construction time, which allows us to capture per-test log messages so they don't all get jumbled together!

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
 */

const (
	containerSuccessExitCode = 0

	testVolumeMountDirpath = "/test-volume" // TODO parameterize this!!
	bindMountsDirpath = "/bind-mounts"  // TODO Parameterize this!!
	testSuiteLogFilepath        = bindMountsDirpath + "/test-listing.log"

	// After we hard-timeout a test, how long we'll give the test to clean itself up (namely the Docker network & containers)
	//  before we call it lost and continue on
	networkTeardownGraceTime = 60 * time.Second

	// When we're tearing down a network after a test (either after normal exit or test timeout), this is the maximum
	//  time we'll wait for each container to stop
	networkTeardownContainerStopTimeout = 10 * time.Second

	dockerSocket = "/var/run/docker.sock"
)

/*
Because a test is run in its own goroutine to allow us to time it out, we need to pass the results back
	via a channel. This struct is what's passed over the channel.
 */
type testResult struct {
	// Whether the test passed or not (undefined if an error occurred that prevented us from retrieving test results)
	testPassed   bool

	// If not nil, the error that prevented us from retrieving the test result
	executionErr error
}

/*
Executor responsible for running a test with timeout, cleaning up after the test as needed.
 */
type testExecutor struct {
	// The logger to which all log statements must be sent
	log *logrus.Logger

	// The execution UUID that the test is running with
	executionInstanceId uuid.UUID

	// The Docker client to execute Docker actions with
	dockerClient *client.Client

	// The mask of the subnet that the test should run in
	subnetMask string

	// The name of the Docker image of the Kurtosis API container to run
	kurtosisApiImageName string

	// The name of the Docker image of the test suite to run
	testSuiteImageName string

	// The log level string to pass to the test suite (should be meaningful to the test suite image)
	testSuiteLogLevel string

	// TODO Use these, by validating that they don't collide with existing environment vars!!!
	// Mapping of user-defined custom environment variables that will also be passed to the test suite contianer on start
	customTestSuiteEnvVars map[string]string

	// Name of the test being run
	testName string
}

/*
Creates a new test executor with the given params, ready to execute a single test. Technically all these params
	could be passed to the runTest method, but it's much simpler to do these at a per-instance level since executors
	don't get reused.

Args:
	log: the logger to which all logging events during test execution will be sent
	executionInstanceId: The UUID representing an execution of the user's test suite, to which this test execution belongs
	dockerClient: The Docker client to use to manipulate the Docker engine
	subnetMask: The subnet mask of the Docker network that has been spun up for this test
	testSuiteImageName: The name of the Docker image of the test controller that will orchestrate execution of this test
	testSuiteLogLevel: A string representing the log level that the test controller should set for itself; this string
		should be meaningful to the user-defined controller code
	customTestSuiteEnvVars: A key-value mapping of custom Docker environment variables that will be passed to the
		controller image (as a method for the user to pass their own custom params between initializer and controller)
	testName: The name of the test the executor should execute
	test: The logic of the test being executed
 */
func newTestExecutor(
			log *logrus.Logger,
			executionInstanceId uuid.UUID,
			dockerClient *client.Client,
			subnetMask string,
			kurtoisApiImageName string,
			testControllerImageName string,
			testControllerLogLevel string,
			customTestControllerEnvVars map[string]string,
			testName string) *testExecutor {
	return &testExecutor{
		log:                    log,
		executionInstanceId:    executionInstanceId,
		dockerClient:           dockerClient,
		subnetMask:             subnetMask,
		kurtosisApiImageName: kurtoisApiImageName,
		testSuiteImageName:     testControllerImageName,
		testSuiteLogLevel:      testControllerLogLevel,
		customTestSuiteEnvVars: customTestControllerEnvVars,
		testName:               testName,
	}
}


/*
Runs a single test with the given name
*/
func (executor testExecutor) runSingleTest(ctx *context.Context) (bool, error) {
	uniqueTestIdentifier := fmt.Sprintf("%v-%v", executor.executionInstanceId.String(), executor.testName)

	executor.log.Info("Creating Docker manager from environment settings...")
	// NOTE: at this point, all Docker commands from here forward will be bound by the Context that we pass in here - we'll
	//  only need to cancel this context once
	dockerManager, err := docker.NewDockerManager(executor.log, executor.dockerClient)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the Docker manager for test %v", executor.testName)
	}
	executor.log.Info("Docker manager created successfully")

	volumeName := uniqueTestIdentifier
	executor.log.Debugf("Creating Docker volume %v which will be shared with the test network...", volumeName)
	if err := dockerManager.CreateVolume(ctx, volumeName); err != nil {
		return false, stacktrace.Propagate(err, "Error creating Docker volume to share amongst test nodes for test %v", executor.testName)
	}
	executor.log.Debugf("Docker volume %v created successfully", volumeName)

	executor.log.Infof("Creating Docker network for test with subnet mask %v...", executor.subnetMask)
	freeIpAddrTracker, err := networks.NewFreeIpAddrTracker(
		executor.log,
		executor.subnetMask,
		map[string]bool{})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker for test %v", executor.testName)
	}
	gatewayIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting a free IP for the gateway for test %v", executor.testName)
	}
	networkName := fmt.Sprintf("%v-%v", executor.executionInstanceId.String(), executor.testName)
	// TODO different context for parallelization?
	networkId, err := dockerManager.CreateNetwork(context.Background(), networkName, executor.subnetMask, gatewayIp)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error occurred creating Docker network %v for test %v", networkName, executor.testName)
	}
	defer removeNetworkDeferredFunc(executor.log, dockerManager, networkId)
	executor.log.Infof("Docker network %v created successfully", networkId)

	// TODO use hostnames rather than IPs, which makes things nicer and which we'll need for Docker swarm support
	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	kurtosisApiIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}
	testRunningContainerIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the test suite container running the test")
	}

	executor.log.Debugf(
		"Test suite container IP: %v; kurtosis API container IP: %v",
		testRunningContainerIp.String(),
		kurtosisApiIp.String())

	// TODO When we have the initializer run in a Docker container, transition to using Docker volumes to store logs
	containerLogFp, err := ioutil.TempFile("", "test-execution.log")
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the temporary file for holding the " +
			"logs of the test suite container running the test")
	}
	containerLogFp.Close()

	executor.log.Infof("Creating test suite container that will run the test...")
	testRunningContainerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		executor.testSuiteImageName,
		networkId,
		testRunningContainerIp,
		map[nat.Port]bool{},
		nil,
		map[string]string{
			test_suite_env_vars.TestNamesFilepathEnvVar:    "",                     // We leave this blank because we want test execution, not listing
			test_suite_env_vars.TestEnvVar:                 executor.testName,               // We leave this blank to signify that we want test listing, not test execution
			test_suite_env_vars.KurtosisApiIpEnvVar:        kurtosisApiIp.String(), // Because we're doing test listing, this can be blank
			test_suite_env_vars.TestSuiteLogFilepathEnvVar: testSuiteLogFilepath,
		},
		map[string]string{
			containerLogFp.Name(): testSuiteLogFilepath,
		},
		map[string]string{
			volumeName: testVolumeMountDirpath,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the test suite container to run the test")
	}
	executor.log.Info("Successfully created test suite container to run the test")

	executor.log.Info("Creating Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/tcp", api.KurtosisAPIContainerPort))
	kurtosisApiContainerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		executor.kurtosisApiImageName,
		networkId,
		kurtosisApiIp,
		map[nat.Port]bool{
			kurtosisApiPort: true,
		},
		nil,
		map[string]string{
			api_container_env_vars.TestSuiteContainerIdEnvVar: testRunningContainerId,
			api_container_env_vars.NetworkIdEnvVar: networkId,
			api_container_env_vars.SubnetMaskEnvVar: executor.subnetMask,
			api_container_env_vars.GatewayIpEnvVar: gatewayIp.String(),
			// TODO make this parameterizable
			api_container_env_vars.LogLevelEnvVar:                 "trace",
			// NOTE: We set this to some random file inside the volume because we don't expect it to be read
			api_container_env_vars.ApiLogFilepathEnvVar:           testVolumeMountDirpath + "/api.log",
			api_container_env_vars.ApiContainerIpAddrEnvVar:       kurtosisApiIp.String(),
			api_container_env_vars.TestSuiteContainerIpAddrEnvVar: testRunningContainerIp.String(),
		},
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{
			volumeName: testVolumeMountDirpath,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the Kurtosis API container")
	}
	executor.log.Infof("Successfully created Kurtosis API container")

	// TODO add a timeout waiting for Kurtosis API container to stop???
	// The Kurtosis API will be our indication of whether the test suite container stopped within the timeout or not
	executor.log.Info("Waiting for Kurtosis API container to exit...")
	kurtosisApiExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		kurtosisApiContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred waiting for the exit of the Kurtosis API container: %v", err)
	}

	var testStatusRetrievalError error
	switch kurtosisApiExitCode {
	case exit_codes.TestCompletedInTimeoutExitCode:
		testStatusRetrievalError = nil
		// TODO this is in a really crappy spot; move it
		banner_printer.PrintContainerLogsWithBanners(testRunningContainerDescription, containerLogFp.Name())
	case exit_codes.OutOfOrderTestStatusExitCode:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container received an out-of-order " +
			"test execution status update; this is a Kurtosis code bug")
	case exit_codes.TestHitTimeoutExitCode:
		testStatusRetrievalError = stacktrace.NewError("The test failed to complete within the hard test " +
			"timeout (test_execution_timeout + setup_buffer)")
	case exit_codes.NoTestSuiteRegisteredExitCode:
		testStatusRetrievalError = stacktrace.NewError("The test suite failed to register itself with the " +
			"Kurtosis API container; this is a bug with the test suite")
		// TODO this is in a really crappy spot; move it
		banner_printer.PrintContainerLogsWithBanners(testRunningContainerDescription, containerLogFp.Name())
	case exit_codes.ShutdownSignalExitCode:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container exited due to receiving " +
			"a shutdown signal; if this is not expected, it's a Kurtosis bug")
	default:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container exited with an unrecognized " +
			"exit code %v; this is a code bug in Kurtosis indicating that the new exit code needs to be mapped " +
			"for the initializer", kurtosisApiExitCode)
	}
	if testStatusRetrievalError != nil {
		executor.log.Error("An error occurred that prevented retrieval of the test completion status")
		return false, testStatusRetrievalError
	}
	executor.log.Info("The test suite container running the test exited within the hard test timeout")

	// The test suite container will already have stopped, so now we get the exit code
	testSuiteExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		testRunningContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred retrieving the test suite container exit code")
	}
	return testSuiteExitCode == 0, nil
}





/*
Runs the test that was configured at time of construction in a separate goroutine, and set a timeout that - if breached -
	will trigger the hard teardown of the test network.

Args:
	ctx: the context of the calling function, used to handle graceful shutdowns

Returns:
	bool: A boolean indicating if the test passed (will be undefined if the test result couldn't be retrieved for any reason)
	error: If not nil, represents the error hit while running the test that prevented the retrieval of the test result
 */
func (executor testExecutor) runTest(ctx *context.Context) (bool, error) {
	testResultChan := make(chan testResult)

	// When this is breached, we'll try to tear down everything
	totalTimeout := executor.test.GetExecutionTimeout() + executor.test.GetSetupBuffer()

	context, cancelFunc := context.WithCancel(*ctx)
	defer cancelFunc()

	// We run the test in a separate goroutine because we don't know if the test will even respect the context we pass in -
	//  we hope so, but (because this runs user-written code) we can't trust it so we give ourselves the option to move
	//  on if the test, e.g., infinite-loops
	go func() {
		testPassed, setupErr := executor.runTestGoroutine(context)
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


// =========================== INSTANCE HELPER FUNCTIONS =========================================
/*
A helper function for running the actual test logic, intended to be run inside a goroutine.

Args:
	ctx: the context of the calling function, used to handle graceful shutdowns

Returns:
	error: If an error occurred that prevented us from running the test & retrieving the results (independent from whether the test itself passed)
	bool: A boolean indicating whether the test passed (undefined if an error occurred running the test)
*/
func (executor testExecutor) runTestGoroutine(context context.Context) (bool, error) {
	executor.log.Info("Creating Docker manager from environment settings...")
	// NOTE: at this point, all Docker commands from here forward will be bound by the Context that we pass in here - we'll
	//  only need to cancel this context once
	dockerManager, err := docker.NewDockerManager(executor.log, executor.dockerClient)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the Docker manager for test %v", executor.testName)
	}
	executor.log.Info("Docker manager created successfully")

	executor.log.Infof("Creating Docker network for test with subnet mask %v...", executor.subnetMask)
	networkName := fmt.Sprintf("%v-%v", executor.executionInstanceId.String(), executor.testName)
	publicIpProvider, err := networks.NewFreeIpAddrTracker(executor.log, executor.subnetMask, map[string]bool{})
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create the free IP address tracker")
	}
	gatewayIp, err := publicIpProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the gateway IP")
	}
	networkId, err := dockerManager.CreateNetwork(context, networkName, executor.subnetMask, gatewayIp)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error occurred creating Docker network %v for test %v", networkName, executor.testName)
	}
	defer removeNetworkDeferredFunc(executor.log, dockerManager, networkId)
	executor.log.Infof("Docker network %v created successfully", networkId)

	executor.log.Info("Running test controller...")
	controllerIp, err := publicIpProvider.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.NewError("An error occurred getting an IP for the test controller")
	}
	testPassed, err := executor.runControllerContainer(
		context,
		dockerManager,
		networkId,
		gatewayIp,
		controllerIp)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while running the test, independent of test success")
	}
	executor.log.Info("The test controller ran and exited successfully")

	return testPassed, nil
}

/*
Helper function to run the controller container against the given test network.

Args:
	context: The context in which the test is being run, such that the test should be cancelled if the context is cancelled
	manager: the Docker manager, used for starting container & waiting for it to finish
	networkId: The id of the Docker network that the controller container will run in
	gatewayIp: The IP of the gateway on the Docker network that the controller is running in
	controllerIpAddr: The IP address that should be used for the container that the controller is running in

Returns:
	bool: true if the test succeeded, false if not
	error: if any error occurred during the execution of the controller (independent of the test itself)
*/
func (executor testExecutor) runControllerContainer(
			context context.Context,
			manager *docker.DockerManager,
			networkId string,
			gatewayIp net.IP,
			controllerIpAddr net.IP) (bool, error){
	uniqueTestIdentifier := fmt.Sprintf("%v-%v", executor.executionInstanceId.String(), executor.testName)

	volumeName := uniqueTestIdentifier
	executor.log.Debugf("Creating Docker volume %v which will be shared with the test network...", volumeName)
	if err := manager.CreateVolume(context, volumeName); err != nil {
		return false, stacktrace.Propagate(err, "Error creating Docker volume to share amongst test nodes")
	}
	executor.log.Debugf("Docker volume %v created successfully", volumeName)

	testControllerLogFilename := fmt.Sprintf("%v-controller-logs", uniqueTestIdentifier)
	executor.log.Debugf("Creating temporary file with name %v to store controller logs...", testControllerLogFilename)
	logTmpFile, err := ioutil.TempFile("", testControllerLogFilename)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not create tempfile to store log info for passing to test controller")
	}
	logTmpFile.Close()
	executor.log.Debugf("Successfully created temporary file to store controller logs at path %v", logTmpFile.Name())

	envVariables, err := generateTestControllerEnvVariables(
		networkId,
		executor.subnetMask,
		gatewayIp,
		controllerIpAddr,
		executor.testName,
		executor.testSuiteLogLevel,
		volumeName,
		executor.customTestSuiteEnvVars)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to map test controller environment variables.")
	}
	executor.log.Debugf("Environment variables that are being passed to the controller: %v", envVariables)

	bindMounts := map[string]string{
		// Because the test controller will need to spin up new images, we need to bind-mount the host Docker engine into the test controller
		"/var/run/docker.sock": "/var/run/docker.sock",
		logTmpFile.Name():      controllerLogMountFilepath,
	}

	volumeMounts := map[string]string{
		volumeName: testVolumeMountpoint,
	}

	controllerContainerId, err := manager.CreateAndStartContainer(
		context,
		executor.testSuiteImageName,
		networkId,
		controllerIpAddr,
		make(map[nat.Port]bool),
		nil, // The controller image's CMD should be parameterized, so we don't specify a start command here
		envVariables,
		bindMounts,
		volumeMounts)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to run test controller container")
	}
	executor.log.Infof("Controller container started successfully with id %s", controllerContainerId)

	executor.log.Info("Waiting for controller container to exit...")
	exitCode, err := manager.WaitForExit(context, controllerContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed when waiting for controller to exit")
	}
	executor.log.Info("Controller container exited successfully")

	// We open a new fp for reading because our original FP is only for writing
	executor.log.Info("- - - - - - - - - - - - - - - - - - - CONTROLLER LOGS - - - - - - - - - - - - - - - - - -")
	logReadFp, err := os.Open(logTmpFile.Name())
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to open controller log file for reading")
	}
	io.Copy(executor.log.Out, logReadFp)
	executor.log.Info("- - - - - - - - - - - - - - - - - - END CONTROLLER LOGS - - - - - - - - - - - - - - - - - -")
	logReadFp.Close()
	os.Remove(logTmpFile.Name()) // We're responsible for removing the tempfile we created

	return exitCode == containerSuccessExitCode, nil
}


// =========================== "STATIC" HELPER FUNCTIONS =========================================

/*
Helper function for making a best-effort attempt at removing a network and logging any error states; intended to be run
as a deferred function.
*/
func removeNetworkDeferredFunc(log *logrus.Logger, dockerManager *docker.DockerManager, networkId string) {
	log.Infof("Attempting to remove Docker network with id %v...", networkId)
	// We use the background context here because we want to try and tear down the network even if the context the test was running in
	//  was cancelled. This might not be right - the right way to do it might be to pipe a separate context for the network teardown to here!
	if err := dockerManager.RemoveNetwork(context.Background(), networkId, networkTeardownContainerStopTimeout); err != nil {
		log.Errorf("An error occurred removing Docker network with ID %v:", networkId)
		log.Error(err.Error())
		log.Error("NOTE: This means you will need to clean up the Docker network manually!!")
	} else {
		log.Infof("Docker network with ID %v successfully removed", networkId)
	}
}

/*
NOTE: This is a separate function because it provides a nice documentation reference point, where we can say to users,
"to see the latest special environment variables that will be passed to the test controller, see this function". Do not
put anything else in this function!!!

Args:
	networkId: The id of the Docker network that the test controller is running in, and which all services should be started in
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
			networkId string,
			subnetMask string,
			gatewayIp net.IP,
			controllerIpAddr net.IP,
			testName string,
			logLevel string,
			testVolumeName string,
			customEnvVars map[string]string) (map[string]string, error) {
	standardVars := map[string]string{
		testNameArg:             testName,
		subnetMaskArg:           subnetMask,
		networkIdArg:            networkId,
		gatewayIpArg:            gatewayIp.String(),
		logFilepathArg:          controllerLogMountFilepath,
		logLevelArg:             logLevel,
		testControllerIpArg:     controllerIpAddr.String(),
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
