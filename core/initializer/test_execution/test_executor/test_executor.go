/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/initializer/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_execution/output"
	"github.com/kurtosis-tech/kurtosis/initializer/test_execution/parallel_test_params"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_launcher"
	"github.com/kurtosis-tech/kurtosis/test_suite/test_suite_rpc_api/bindings"
	"github.com/kurtosis-tech/kurtosis/test_suite/test_suite_rpc_api/test_suite_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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

	// How long we'll give the API container & testsuite container to gracefully stop after we're done trying to run the test
	//  (either through success or error)
	containerTeardownTimeout = 10 * time.Second

	printTestsuiteLogSectionAsError = false

	// For debugging, we'll sometimes want to print logs from the initializer during the section labelled "testsuite logs"
	// To distinguish initializer logs from testsuite logs, we add this prefix to loglines that come from the initializer
	//  rather than the testsuite
	initializerLogPrefix = "[INITIALIZER] "

	shouldFollowTestsuiteLogs = true

	// During network removal, how long to wait after issuing the kill command to the containers and before
	//  trying to remove the network (which will fail if there are running containers)
	networkTeardownWaitTime = 2 * time.Second
)

/*
Runs a single test with the given name

Args:
	ctx: The Context that the test execution is happening in
	executionInstanceUuid: The UUID representing an execution of the user's test suite, to which this test execution belongs
	initializerContainerId: The ID of the intiializer container
	testSetupExecutionCtx: The context in which test setup & execution runs, which can potentially be cancelled on a Ctrl-C
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
		testSetupExecutionCtx context.Context,
		executionInstanceUuid string,
		initializerContainerId string,
		log *logrus.Logger,
		dockerClient *client.Client,
		subnetMask string,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
		apiContainerLauncher *api_container_launcher.ApiContainerLauncher,
		testParams parallel_test_params.ParallelTestParams) (bool, error) {

	testName := testParams.TestName

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
	initializerContainerIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting a free IP for mounting the initializer container in the test network")
	}
	networkName := fmt.Sprintf(
		"%v_%v_%v",
		time.Now().Format(networkNameTimestampFormat),
		executionInstanceUuid,
		testName)
	networkId, err := dockerManager.CreateNetwork(testSetupExecutionCtx, networkName, subnetMask, gatewayIp)
	if err != nil {
		// TODO If the user Ctrl-C's while the CreateNetwork call is ongoing then the CreateNetwork will error saying
		//  that the Context was cancelled as expected, but *the Docker engine will still create the networks!!! We'll
		//  need to parse the log message for the string "context canceled" and, if found, do another search for
		//  networks with our network name and delete them
		return false, stacktrace.Propagate(err, "Error occurred creating Docker network %v for test %v", networkName, testName)
	}
	// TODO Have whitelist of IDs that won't get removed
	defer removeNetworkDeferredFunc(testTeardownContext, log, dockerManager, networkId, initializerContainerId)
	log.Infof("Docker network %v created successfully", networkId)

	if err := dockerManager.ConnectContainerToNetwork(testSetupExecutionCtx, networkId, initializerContainerId, initializerContainerIp); err != nil {
		return false, stacktrace.Propagate(err, "An error occurred connecting the initializer container to the test network, which is required to communicate with the testsuite")
	}

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

	apiContainerId, err := apiContainerLauncher.Launch(
		testSetupExecutionCtx,
		log,
		dockerManager,
		testName,
		networkId,
		subnetMask,
		gatewayIp,
		initializerContainerIp,
		testRunningContainerIp,
		kurtosisApiIp,
		testParams.IsPartitioningEnabled,
	)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred launching the API container")
	}
	defer func() {
		if err := dockerManager.StopContainer(testTeardownContext, apiContainerId, containerTeardownTimeout); err !=  nil {
			// This is a warning because the network teardown will try again
			log.Warnf("An error occurred stopping the API container during teardown; stopping & removal will be attmepted again during network teardown")
		}
	}()

	testsuiteContainerId, err := testsuiteLauncher.LaunchTestRunningContainer(
		testSetupExecutionCtx,
		log,
		dockerManager,
		networkId,
		testName,
		kurtosisApiIp,
		testRunningContainerIp,
	)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred launching the testsuite & Kurtosis API containers for executing the test")
	}
	defer func() {
		if err := dockerManager.StopContainer(testTeardownContext, testsuiteContainerId, containerTeardownTimeout); err !=  nil {
			// This is a warning because the network teardown will try again
			log.Warnf("An error occurred stopping the testsuite container during teardown; stopping & removal will be attmepted again during network teardown")
		}
	}()

	testsuiteEndpointUri := fmt.Sprintf("%v:%v", testRunningContainerIp.String(), test_suite_rpc_api_consts.ListenPort)
	// TODO SECURITY: Use HTTPS
	conn, err := grpc.Dial(testsuiteEndpointUri, grpc.WithInsecure())
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred dialing the testsuite container endpoint")
	}
	defer conn.Close()
	testsuiteServiceClient := bindings.NewTestSuiteServiceClient(conn)

	if err := streamTestsuiteLogsWhileRunningTest(
		testSetupExecutionCtx,
		log,
		dockerManager,
		testsuiteContainerId,
		testParams,
		testsuiteServiceClient,
	); err != nil {
		return false, stacktrace.Propagate(err, "An error occurred running the test")
	}

	return true, nil
}

// NOTE: This function makes heavy use of deferred functions, to simplify the code
func streamTestsuiteLogsWhileRunningTest(
		testSetupExecutionCtx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		testsuiteContainerId string,
		testParams parallel_test_params.ParallelTestParams,
		testsuiteServiceClient bindings.TestSuiteServiceClient) error {
	banner_printer.PrintSection(log, "Testsuite Logs", printTestsuiteLogSectionAsError)
	// After this point, we can go back to printing initializer logs
	defer banner_printer.PrintSection(log, "End Testsuite Logs", printTestsuiteLogSectionAsError)

	// NOTE: We use the testSetupExecutionContext so that the logstream from the testsuite container will be closed
	// if the user presses Ctrl-C.
	readCloser, err := dockerManager.GetContainerLogs(testSetupExecutionCtx, testsuiteContainerId, shouldFollowTestsuiteLogs)
	if err != nil {
		log.Errorf("An error occurred getting the testsuite container logs for streaming: %v", err)
	} else {
		defer readCloser.Close()

		logStreamer := output.NewLogStreamer("DOCKER LOGS STREAMER", log)
		if startStreamingErr := logStreamer.StartStreamingFromDockerLogs(readCloser); err != nil {
			log.Errorf("The following error occurred when attempting to stream the testsuite logs: %v", startStreamingErr)
		} else {
			log.Tracef("%vTestsuite container log streamer started successfully", initializerLogPrefix)

			// Catch-all to make sure we don't leave a thread hanging around in case this function exits abnormally
			defer func() {
				if err := logStreamer.StopStreaming(); err != nil {
					log.Warnf("An error occurred stopping the log streamer: %v", err)
				}
			}()
		}
	}


	// TODO Probably need to connect the initializer container inside the testnet

	log.Tracef("%vSetting up test...", initializerLogPrefix)
	testSetupCtx, testSetupCtxCancelFunc := context.WithTimeout(
		testSetupExecutionCtx,
		time.Duration(testParams.TestSetupTimeoutSeconds) * time.Second,
	)
	defer testSetupCtxCancelFunc()
	setupArgs := &bindings.SetupTestArgs{
		TestName: testParams.TestName,
	}
	if _, err := testsuiteServiceClient.SetupTest(testSetupCtx, setupArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting up the test network before running the test")
	}
	log.Tracef("%vTest setup completed successfully", initializerLogPrefix)

	log.Tracef("%vRunning test...", initializerLogPrefix)
	testRunCtx, testRunCtxCancelFunc := context.WithTimeout(
		testSetupExecutionCtx,
		time.Duration(testParams.TestRunTimeoutSeconds) * time.Second,
	)
	defer testRunCtxCancelFunc()
	runArgs := &bindings.RunTestArgs{
		TestName: testParams.TestName,
	}
	if _, err := testsuiteServiceClient.RunTest(testRunCtx, runArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the test")
	}
	log.Tracef("%vTest run completed successfully", initializerLogPrefix)
	return nil
}


// =========================== PRIVATE HELPER FUNCTIONS =========================================
/*
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
 */


/*
Helper function for making a best-effort attempt at removing a network and the containers inside after a test has
	exited (either normally or with error)
*/
func removeNetworkDeferredFunc(
		testTeardownContext context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		networkId string,
		initializerContainerId string) {
	log.Debugf("Attempting to remove Docker network with ID %v...", networkId)
	containerIds, err := dockerManager.GetContainerIdsConnectedToNetwork(testTeardownContext, networkId)
	if err != nil {
		errorDesc := fmt.Sprintf("An error occurred getting the containers connected to network '%v' so we can stop them:", networkId)
		logErrorAndRecommendManualIntervention(log, errorDesc, err, networkId)
		return
	}

	for _, containerId := range containerIds {
		if containerId == initializerContainerId {
			// We don't want to kill the initializer, but we need it gone from the network before we can delete the network
			if err := dockerManager.DisconnectContainerFromNetwork(testTeardownContext, initializerContainerId, networkId); err != nil {
				errorDesc := fmt.Sprintf("An error occurred disconnecting the initializer container from the network, which prevents the network from being deleted:")
				logErrorAndRecommendManualIntervention(log, errorDesc, err, networkId)
				return
			}
		} else {
			if err := dockerManager.KillContainer(testTeardownContext, containerId); err != nil {
				errorDesc := fmt.Sprintf("An error occurred killing container '%v', which prevents the network from being deleted:", containerId)
				logErrorAndRecommendManualIntervention(log, errorDesc, err, networkId)
				return
			}
		}
	}

	if err := dockerManager.RemoveNetwork(testTeardownContext, networkId); err != nil {
		errorDesc := fmt.Sprintf("An error occurred removing Docker network with ID %v:", networkId)
		logErrorAndRecommendManualIntervention(log, errorDesc, err, networkId)
		return
	}
	log.Debugf("Successfully removed Docker network with ID %v", networkId)
}

func logErrorAndRecommendManualIntervention(log *logrus.Logger, humanReadableErrorDesc string, err error, networkId string) {
	log.Errorf(humanReadableErrorDesc)
	log.Error(err.Error())
	log.Errorf("ACTION REQUIRED: You'll need to delete network with ID '%v' manually!!!", networkId)
}
