/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package parallelism

import (
	"context"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/execution/exit_codes"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_env_vars"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_mount_locations"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
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
	testSuiteLogFilename = "test-execution.log"

	// When we're tearing down a network after a test (either after normal exit or test timeout), this is the maximum
	//  time we'll wait for each container to stop
	networkTeardownContainerStopTimeout = 10 * time.Second

	dockerSocket = "/var/run/docker.sock"

	testRunningContainerDescription = "Test-Running Container"
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

	// The string representing the log level that the API container should run with
	apiContainerLogLevel string
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
 */
func newTestExecutor(
			log *logrus.Logger,
			executionInstanceId uuid.UUID,
			dockerClient *client.Client,
			subnetMask string,
			kurtoisApiImageName string,
			testSuiteImageName string,
			testSuiteLogLevel string,
			customTestControllerEnvVars map[string]string,
			testName string,
			apiContainerLogLevel string) *testExecutor {
	return &testExecutor{
		// TODO sort alphabetically
		log:                             log,
		executionInstanceId:             executionInstanceId,
		dockerClient:                    dockerClient,
		subnetMask:                      subnetMask,
		kurtosisApiImageName:            kurtoisApiImageName,
		testSuiteImageName:              testSuiteImageName,
		testSuiteLogLevel:               testSuiteLogLevel,
		customTestSuiteEnvVars:          customTestControllerEnvVars,
		testName:                        testName,
		apiContainerLogLevel: apiContainerLogLevel,
	}
}


/*
Runs a single test with the given name
*/
func (executor testExecutor) runTest(ctx context.Context) (bool, error) {
	uniqueTestIdentifier := fmt.Sprintf("%v-%v", executor.executionInstanceId.String(), executor.testName)

	executor.log.Info("Creating Docker manager from environment settings...")
	// NOTE: at this point, all Docker commands from here forward will be bound by the Context that we pass in here - we'll
	//  only need to cancel this context once
	dockerManager, err := commons.NewDockerManager(executor.log, executor.dockerClient)
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
	freeIpAddrTracker, err := commons.NewFreeIpAddrTracker(
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
	networkId, err := dockerManager.CreateNetwork(ctx, networkName, executor.subnetMask, gatewayIp)
	if err != nil {
		// TODO If the user Ctrl-C's while the CreateNetwork call is ongoing then the CreateNetwork will error saying
		//  that the Context was cancelled as expected, but *the Docker engine will still create the networks!!! We'll
		//  need to parse the log message for the string "context canceled" and, if found, do another search for
		//  networks with our network name and delete them
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

	containerLogFp, err := ioutil.TempFile("", "test-execution.log")
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the temporary file for holding the " +
			"logs of the test suite container running the test")
	}
	containerLogFp.Close()

	testSuiteLogFilepath := path.Join(test_suite_mount_locations.BindMountsDirpath, testSuiteLogFilename)
	testSuiteEnvVars, err := generateTestSuiteEnvVars(
		executor.testName,
		kurtosisApiIp.String(),
		testSuiteLogFilepath,
		executor.testSuiteLogLevel,
		executor.customTestSuiteEnvVars)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred generating the map of test suite environment variables")
	}

	executor.log.Infof("Creating test suite container that will run the test...")
	testRunningContainerId, err := dockerManager.CreateAndStartContainer(
		ctx,
		executor.testSuiteImageName,
		networkId,
		testRunningContainerIp,
		map[nat.Port]bool{},
		nil,
		testSuiteEnvVars,
		map[string]string{
			containerLogFp.Name(): testSuiteLogFilepath,
		},
		map[string]string{
			volumeName: test_suite_mount_locations.TestVolumeDirpath,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the test suite container to run the test")
	}
	executor.log.Info("Successfully created test suite container to run the test")

	executor.log.Info("Creating Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/tcp", api_container_docker_consts.ContainerPort))
	kurtosisApiContainerId, err := dockerManager.CreateAndStartContainer(
		ctx,
		executor.kurtosisApiImageName,
		networkId,
		kurtosisApiIp,
		map[nat.Port]bool{
			kurtosisApiPort: true,
		},
		nil,
		map[string]string{
			api_container_env_vars.ApiContainerIpAddrEnvVar:       kurtosisApiIp.String(),
			api_container_env_vars.ApiLogFilepathEnvVar:           api_container_docker_consts.LogMountFilepath,
			api_container_env_vars.GatewayIpEnvVar: gatewayIp.String(),
			api_container_env_vars.LogLevelEnvVar: executor.apiContainerLogLevel,
			api_container_env_vars.NetworkIdEnvVar: networkId,
			api_container_env_vars.SubnetMaskEnvVar: executor.subnetMask,
			api_container_env_vars.TestSuiteContainerIdEnvVar: testRunningContainerId,
			api_container_env_vars.TestSuiteContainerIpAddrEnvVar: testRunningContainerIp.String(),
			api_container_env_vars.TestVolumeName: volumeName,
		},
		map[string]string{
			dockerSocket: dockerSocket,
			// TODO bind-mount the API container logfile and print its output ONLY IF the user's token is an admin
			//  token, so we don't leak information about the API container
		},
		map[string]string{
			volumeName: api_container_docker_consts.TestVolumeDirpath,
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
		banner_printer.PrintContainerLogsWithBanners(executor.log, testRunningContainerDescription, containerLogFp.Name())
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
		banner_printer.PrintContainerLogsWithBanners(executor.log, testRunningContainerDescription, containerLogFp.Name())
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


// =========================== PRIVATE HELPER FUNCTIONS =========================================

/*
Helper function for making a best-effort attempt at removing a network and logging any error states; intended to be run
as a deferred function.
*/
func removeNetworkDeferredFunc(log *logrus.Logger, dockerManager *commons.DockerManager, networkId string) {
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

func generateTestSuiteEnvVars(
			testName string,
			kurtosisApiIp string,
			logFilepath string,
			logLevel string,
			customEnvVars map[string]string) (map[string]string, error) {
	standardVars := map[string]string{
		test_suite_env_vars.MetadataFilepathEnvVar: "", // We leave this blank because we want test execution, not listing
		test_suite_env_vars.TestEnvVar:             testName,
		test_suite_env_vars.KurtosisApiIpEnvVar:    kurtosisApiIp,
		test_suite_env_vars.LogFilepathEnvVar:      logFilepath,
		test_suite_env_vars.LogLevelEnvVar:         logLevel,
	}
	for key, val := range customEnvVars {
		if _, ok := standardVars[key]; ok {
			return nil, stacktrace.NewError(
				"Tried to manually add custom environment variable %s to the test controller container, but it is " +
					"already being used by Kurtosis.",
				key)
		}
		standardVars[key] = val
	}
	return standardVars, nil
}
