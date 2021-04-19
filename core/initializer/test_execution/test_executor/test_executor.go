/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_exit_codes"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_execution/output"
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
	networkNameTimestampFormat = "2006-01-02T15.04.05" // Go timestamp formatting is absolutely absurd...

	// When the user presses Ctrl-C on the initializer, we need to forward that on to the API container and tell
	// it to stop. When that happens, this is how long we'll give the API container to gracefully exit before we
	// hard-kill the container.
	apiContainerStopTimeoutAfterParentCtxCancellation = 10 * time.Second

	printTestsuiteLogSectionAsError = false

	// For debugging, we'll sometimes want to print logs from the initializer during the section labelled "testsuite logs"
	// To distinguish initializer logs from testsuite logs, we add this prefix to loglines that come from the initializer
	//  rather than the testsuite
	initializerLogPrefix = "[INITIALIZER] "
)

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
		testMetadata.TestSetupTimeoutInSeconds,
		testMetadata.TestRunTimeoutInSeconds,
		testMetadata.IsPartitioningEnabled)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred launching the testsuite & Kurtosis API containers for executing the test")
	}

	// ====================================== TRICKY CODE WARNING =====================================================
	// The code from this point on is very tricky! It is trying to stream the testsuite logs while still responding
	//  to what the API container is doing. Be VERY careful with editing it - there are a lot of things happening
	//  in parallel and it's easy to cause subtle bugs!
	// ====================================== TRICKY CODE WARNING =====================================================
	banner_printer.PrintSection(log, "Testsuite Logs", printTestsuiteLogSectionAsError)
	var logStreamer *output.LogStreamer = nil
	// NOTE: We use the testSetupExecutionContext so that the logstream from the testsuite container will be closed
	// if the user presses Ctrl-C.
	readCloser, err := dockerManager.GetContainerLogs(testSetupExecutionCtx, testsuiteContainerId)
	if err != nil {
		log.Errorf("An error occurred getting the testsuite container logs: %v", err)
	} else {
		newStreamer := output.NewLogStreamer(log)
		if startStreamingErr := newStreamer.StartStreamingFromDockerLogs(readCloser); err != nil {
			log.Errorf("The following error occurred when attempting to stream the testsuite logs: %v", startStreamingErr)
		} else {
			log.Tracef("%vTestsuite container log streamer started successfully", initializerLogPrefix)
			logStreamer = newStreamer

			// Catch-all to make sure we don't leave a thread hanging around in case this function exits abnormally
			defer logStreamer.StopStreaming()
		}
	}
	defer readCloser.Close()

	log.Tracef("%vWaiting for API container exit...", initializerLogPrefix)
	kurtosisApiContainerExitError := waitForApiContainerExit(
			dockerManager,
			kurtosisApiContainerId,
			testSetupExecutionCtx,
			testTeardownContext)
	log.Tracef("%vAPI container exited; resulting err is: %v", initializerLogPrefix, kurtosisApiContainerExitError)
	var stopLogStreamerErr error = nil
	if logStreamer != nil {
		log.Tracef("%vStopping testsuite container log streamer...", initializerLogPrefix)
		stopLogStreamerErr = logStreamer.StopStreaming()
		log.Tracef("%vStopped testsuite container log streamer", initializerLogPrefix)
	}
	// After this point, we can go back to printing initializer logs
	banner_printer.PrintSection(log, "End Testsuite Logs", printTestsuiteLogSectionAsError)
	if stopLogStreamerErr != nil {
		log.Warnf("An error occurred stopping the log streamer: %v", stopLogStreamerErr)
	}
	if kurtosisApiContainerExitError != nil {
		// NOTE: We haven't stopped the testsuite at this point, but that's okay - it'll get torn down
		return false, stacktrace.Propagate(err, "An error occurred that prevented retrieval of the test completion status")
	}
	log.Info("The test suite container exited as expected")
	// ====================================== END TRICKY CODE WARNING =================================================

	// If we got here, then the testsuite container exited as the API container expected, meaning the testsuite
	//  container is stopped, meaning it's okay to WaitForExit using testTeardownContext to get the testsuite
	//  container's exit code
	log.Tracef("Waiting for testsuite container to exit...")
	testSuiteExitCode, err := dockerManager.WaitForExit(
		testTeardownContext,
		testsuiteContainerId)
	if err != nil {
		log.Tracef("Received an error waiting for the testsuite container to exit: %v", err)
		return false, stacktrace.Propagate(err, "An error occurred retrieving the test suite container exit code")
	}
	log.Tracef("Testsuite container exited with exit code '%v'", testSuiteExitCode)
	return testSuiteExitCode == 0, nil
}


// =========================== PRIVATE HELPER FUNCTIONS =========================================
// Waits for the API container to exit, and returns an error (if any) describing the problem with API container exiting
func waitForApiContainerExit(
		dockerManager *docker_manager.DockerManager,
		kurtosisApiContainerId string,
		testSetupExecutionCtx context.Context,
		testTeardownCtx context.Context) error {
	var kurtosisApiExitCodeInt64 int64
	kurtosisApiExitCodeInt64, err := dockerManager.WaitForExit(testSetupExecutionCtx, kurtosisApiContainerId)
	if err != nil {
		if testSetupExecutionCtx.Err() != context.Canceled {
			return stacktrace.Propagate(err, "An unexpected error occurred waiting for the exit of the Kurtosis API container")
		}

		// The parent context was cancelled, indicating that the user pressed Ctrl-C. The API container will have no idea
		//  that this happened, so we need to stop it (which will result in an error code, in the same way as if
		//  the user had killed the API container) so the code below can work as expected
		if err := dockerManager.StopContainer(
			testTeardownCtx,
			kurtosisApiContainerId,
			apiContainerStopTimeoutAfterParentCtxCancellation); err != nil {
			return stacktrace.Propagate(
				err,
				"The test execution context was cancelled (likely due to Ctrl-C), which we forwarded to the API container " +
					"by requesting that it stop so that it would register the termination signal. However, the below error occurred stopping " +
					"the API container. This should never happen, and indicates a bug in Kurtosis.",
			)
		}

		// The API container will be stopped at this point, so grab the exit code (which should be TERM)
		//  so that all the code below here works as normal
		kurtosisApiExitCodeInt64, err = dockerManager.WaitForExit(
			testTeardownCtx,
			kurtosisApiContainerId)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred waiting for the API container to exit AFTER we'd successfully stopped it; this is very" +
					"unexpected and likely indicates a bug in Kurtosis",
			)
		}
	}

	// The Kurtosis API will be our indication of whether the test suite container stopped happily or not
	kurtosisApiExitCode := int(kurtosisApiExitCodeInt64)

	apiExitCodeVisitorAcceptFunc, found := api_container_exit_codes.ExitCodeErrorVisitorAcceptFuncs[kurtosisApiExitCode]
	if !found {
		return stacktrace.NewError(
			"The Kurtosis API container exited with an unrecognized " +
				"exit code '%v' that doesn't have an accept listener; this is a code bug in Kurtosis",
			kurtosisApiExitCode,
		)
	}
	visitor := testExecutionExitCodeErrorVisitor{}
	if err := apiExitCodeVisitorAcceptFunc(visitor); err != nil {
		return stacktrace.Propagate(err, "The API container exit code indicated an error")
	}
	return nil
}


/*
Helper function for making a best-effort attempt at removing a network and the containers inside after a test has
	exited (either normally or with error)
*/
func removeNetworkDeferredFunc(
		testTeardownContext context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		networkId string) {
	log.Debugf("Attempting to remove Docker network with ID %v...", networkId)
	if err := dockerManager.RemoveNetwork(testTeardownContext, networkId); err != nil {
		log.Errorf("An error occurred removing Docker network with ID %v:", networkId)
		log.Error(err.Error())
		log.Error("NOTE: This means you will need to clean up the Docker network manually!!")
	} else {
		log.Debugf("Successfully removed Docker network with ID %v", networkId)
	}
}
