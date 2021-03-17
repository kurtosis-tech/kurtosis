/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_exit_codes"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_launcher"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
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
	testsuiteContainerDescription = "Testsuite Container"

	networkNameTimestampFormat = "2006-01-02T15.04.05" // Go timestamp formatting is absolutely absurd...

	// When the user presses Ctrl-C on the initializer, we need to forward that on to the API container and tell
	// it to stop. When that happens, this is how long we'll give the API container to gracefully exit before we
	// hard-kill the container.
	apiContainerStopTimeoutAfterParentCtxCancellation = 10 * time.Second
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
Runs a single test with the given name

Args:
	executionInstanceId: The UUID representing an execution of the user's test suite, to which this test execution belongs
	testSetupExecutionCtx: The context in which test setup & execution runs, which can potentially be cancelled on a Ctrl-C
	ctx: The Context that the test execution is happening in
	log: the logger to which all logging events during test execution will be sent
	dockerClient: The Docker client to use to manipulate the Docker engine
	subnetMask: The subnet mask of the Docker network that has been spun up for this test
	testsuiteLauncher: Launcher for running the test-running testsuite instances
	testsuiteDebuggerHostPortBinding: The port binding on the host machine that the testsuite debugger port should be tied to
	testName: The name of the test the executor should execute
	testMetadata: Metadata declared by the test itslef (e.g. if partitioning is enabled)

Returns:
	bool: True if the test passed, false otherwise
	error: Non-nil if an error occurred that prevented the test pass/fail status from being retrieved
*/
func RunTest(
		executionInstanceId uuid.UUID,
		testSetupExecutionCtx context.Context,
		log *logrus.Logger,
		dockerClient *client.Client,
		subnetMask string,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
		testsuiteDebuggerHostPortBinding nat.PortBinding,
		testName string,
		testMetadata test_suite_metadata_acquirer.TestMetadata) (bool, error) {
	log.Info("Creating Docker manager from environment settings...")
	// NOTE: at this point, all Docker commands from here forward will be bound by the Context that we pass in here - we'll
	//  only need to cancel this context once
	dockerManager, err := docker_manager.NewDockerManager(log, dockerClient)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the Docker manager for test %v", testName)
	}
	log.Info("Docker manager created successfully")

	// We'll use the test setup context for setting stuff up so that a cancellation (e.g. Ctrl-C)
	//  will prevent any new things from getting added to Docker. We still want to be able to retrieve exit codes
	//  and logs after a Ctrl-C though, so we use the background context for doing those tasks (rather than the
	//  potentially-cancelled setup context).
	testTeardownContext := context.Background()

	log.Infof("Creating Docker network for test with subnet mask %v...", subnetMask)
	freeIpAddrTracker, err := commons.NewFreeIpAddrTracker(
		log,
		subnetMask,
		map[string]bool{})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker for test %v", testName)
	}
	gatewayIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting a free IP for the gateway for test %v", testName)
	}
	networkName := fmt.Sprintf(
		"%v_%v_%v",
		time.Now().Format(networkNameTimestampFormat),
		executionInstanceId.String(),
		testName)
	networkId, err := dockerManager.CreateNetwork(testSetupExecutionCtx, networkName, subnetMask, gatewayIp)
	if err != nil {
		// TODO If the user Ctrl-C's while the CreateNetwork call is ongoing then the CreateNetwork will error saying
		//  that the Context was cancelled as expected, but *the Docker engine will still create the networks!!! We'll
		//  need to parse the log message for the string "context canceled" and, if found, do another search for
		//  networks with our network name and delete them
		return false, stacktrace.Propagate(err, "Error occurred creating Docker network %v for test %v", networkName, testName)
	}
	defer removeNetworkDeferredFunc(testTeardownContext, log, dockerManager, networkId)
	log.Infof("Docker network %v created successfully", networkId)

	// TODO use hostnames rather than IPs, which makes things nicer and which we'll need for Docker swarm support
	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	kurtosisApiIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}
	defer freeIpAddrTracker.ReleaseIpAddr(kurtosisApiIp)
	testRunningContainerIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the test suite container running the test")
	}
	defer freeIpAddrTracker.ReleaseIpAddr(testRunningContainerIp)

	testsuiteContainerId, kurtosisApiContainerId, err := testsuiteLauncher.LaunchTestRunningContainers(
		testSetupExecutionCtx,
		log,
		dockerManager,
		networkId,
		subnetMask,
		gatewayIp,
		testName,
		kurtosisApiIp,
		testRunningContainerIp,
		testsuiteDebuggerHostPortBinding,
		testMetadata.TestSetupTimeoutInSeconds,
		testMetadata.TestExecutionTimeoutInSeconds,
		testMetadata.IsPartitioningEnabled)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred launching the testsuite & Kurtosis API containers for executing the test")
	}

	// The Kurtosis API will be our indication of whether the test suite container stopped within the timeout or not
	log.Info("Waiting for Kurtosis API container to exit...")
	var kurtosisApiExitCodeInt64 int64
	kurtosisApiExitCodeInt64, err = dockerManager.WaitForExit(
		testSetupExecutionCtx,
		kurtosisApiContainerId)
	if err != nil {
		if testSetupExecutionCtx.Err() != context.Canceled {
			return false, stacktrace.Propagate(err, "An unexpected error occurred waiting for the exit of the Kurtosis API container")
		}

		// The parent context was cancelled, indicating that the user pressed Ctrl-C. The API container will have no idea
		//  that this happened, so we need to stop it (which will result in an error code, in the same way as if
		//  the user had killed the API container) so the code below can work as expected
		if err := dockerManager.StopContainer(
				testTeardownContext,
				kurtosisApiContainerId,
				apiContainerStopTimeoutAfterParentCtxCancellation); err != nil {
			return false, stacktrace.Propagate(err, "The test execution context was cancelled (likely due to Ctrl-C), which we forwarded to the API container " +
				"by requesting that it stop so that it would register the termination signal. However, the below error occurred stopping " +
				"the API container. This should never happen, and indicates a bug in Kurtosis.")
		}

		// The API container will be stopped at this point, so grab the exit code (which should be TERM)
		//  so that all the code below here works as normal
		kurtosisApiExitCodeInt64, err = dockerManager.WaitForExit(
			testTeardownContext,
			kurtosisApiContainerId)
		if err != nil {
			return false, stacktrace.Propagate(
				err,
				"An error occurred waiting for the API container to exit AFTER we'd successfully stopped it; this is very" +
					"unexpected and likely indicates a bug in Kurtosis",
			)
		}
	}
	kurtosisApiExitCode := int(kurtosisApiExitCodeInt64)

	// At this point, we may be printing the logs of a stopped test suite container, or we may be printing the logs of
	//  still-running container because the test exceeded the hard test timeout or we've received a Ctrl-C. Regardless,
	//  we want to print these so the user gets more information about what's going on, and the user will learn the exact error below
	banner_printer.PrintContainerLogsWithBanners(dockerManager, testTeardownContext, testsuiteContainerId, log, testsuiteContainerDescription)

	apiExitCodeVisitorAcceptFunc, found := api_container_exit_codes.ExitCodeErrorVisitorAcceptFuncs[kurtosisApiExitCode]
	if !found {
		return false, stacktrace.NewError("The Kurtosis API container exited with an unrecognized " +
				"exit code '%v' that doesn't have an accept listener; this is a code bug in Kurtosis",
			kurtosisApiExitCode)
	}
	visitor := testExecutionExitCodeErrorVisitor{}
	testStatusRetrievalError := apiExitCodeVisitorAcceptFunc(visitor)

	if testStatusRetrievalError != nil {
		log.Error("An error occurred that prevented retrieval of the test completion status")
		return false, testStatusRetrievalError
	}
	log.Info("The test suite container exited as expected")

	// If we got here, then the testsuite container exited as the API container expected, meaning the testsuite
	//  container is stopped, meaning it's okay to WaitForExit using testTeardownContext to get the testsuite
	//  container's exit code
	testSuiteExitCode, err := dockerManager.WaitForExit(
		testTeardownContext,
		testsuiteContainerId)
	if err != nil {
			  return false, stacktrace.Propagate(err, "An error occurred retrieving the test suite container exit code")
			  }
	return testSuiteExitCode == 0, nil
}


// =========================== PRIVATE HELPER FUNCTIONS =========================================


/*
Helper function for making a best-effort attempt at removing a network and the containers inside after a test has
	exited (either normally or with error)
*/
func removeNetworkDeferredFunc(
		testTeardownContext context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		networkId string) {
	log.Infof("Attempting to remove Docker network with id %v...", networkId)
	if err := dockerManager.RemoveNetwork(testTeardownContext, networkId); err != nil {
		log.Errorf("An error occurred removing Docker network with ID %v:", networkId)
		log.Error(err.Error())
		log.Error("NOTE: This means you will need to clean up the Docker network manually!!")
	} else {
		log.Infof("Successfully removed Docker network with ID %v", networkId)
	}
}
