/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_executor

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_execution/output"
	"github.com/kurtosis-tech/kurtosis/initializer/test_execution/parallel_test_params"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_launcher"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
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
	waitForTestsuiteAvailabilityTimeout = 10 * time.Second

	// How long we'll give the API container & testsuite container to gracefully stop after we're done trying to run the test
	//  (either through success or error)
	containerTeardownTimeout = 10 * time.Second

	printTestsuiteLogSectionAsError = false

	dockerLogStreamerLogLineLabel = "DOCKER LOGS STREAMER"

	// For debugging, we'll sometimes want to print logs from the initializer during the section labelled "testsuite logs"
	// To distinguish initializer logs from testsuite logs, we add this prefix to loglines that come from the initializer
	//  rather than the testsuite
	initializerLogPrefix = "[INITIALIZER] "

	// During network removal, how long to wait after issuing the kill command to the containers and before
	//  trying to remove the network (which will fail if there are running containers)
	networkTeardownWaitTime = 2 * time.Second

	// How much time we'll give the testsuite to stop gracefully before killing it
	testsuiteGracefulStopTimeout = 10 * time.Second

	// This string is used to evaluate if a context cancel error has been encountered during a setup or run test execution
	contextDeadlineStringError = "context deadline exceeded"
)

/*
Runs a single test with the given name

Args:
	testSetupExecutionCtx: The Context that the test setup & execution is happening in, which can be canclled via Ctrl-C
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
		testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider,
		initializerContainerId string,
		log *logrus.Logger,
		enclaveManager *enclave_manager.EnclaveManager,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
		testParams parallel_test_params.ParallelTestParams) (bool, error) {

	testName := testParams.TestName
	isPartitioningEnabled := testParams.IsPartitioningEnabled

	// We'll use the test setup context for setting stuff up so that a cancellation (e.g. Ctrl-C)
	//  will prevent any new things from getting added to Docker. We still want to be able to retrieve exit codes
	//  and logs after a Ctrl-C though, so we use the background context for doing those tasks (rather than the
	//  potentially-cancelled setup context).
	testTeardownContext := context.Background()

	enclaveId := testsuiteExObjNameProvider.ForTestEnclave(testName)

	log.Debugf("Creating enclave for test '%v'....", testName)
	enclaveCtx, err := enclaveManager.CreateEnclave(
		testSetupExecutionCtx,
		log,
		map[string]bool{initializerContainerId: true},
		enclaveId,
		isPartitioningEnabled,
	)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating a Kurtosis enclave for test '%v'", testName)
	}
	defer func() {
		if err := enclaveManager.DestroyEnclave(testTeardownContext, log, enclaveCtx); err != nil {
			log.Errorf("An error occurred destroying enclave '%v':", enclaveId)
			fmt.Fprintln(log.Out, err)
			log.Errorf("ACTION REQUIRED: You'll need to manually clean up the containers and network of enclave '%v'!!!!!", enclaveId)
		}
	}()

	dockerManager := enclaveCtx.GetDockerManager()
	networkId := enclaveCtx.GetNetworkID()
	kurtosisApiIp := enclaveCtx.GetAPIContainerIPAddr()
	testsuiteContainerIp := enclaveCtx.GetTestsuiteContainerIPAddr()
	testsuiteContainerName := enclaveCtx.GetTestsuiteContainerName()

	testsuiteContainerId, err := testsuiteLauncher.LaunchTestRunningContainer(
		testSetupExecutionCtx,
		log,
		dockerManager,
		networkId,
		testsuiteContainerName,
		kurtosisApiIp,
		testsuiteContainerIp,
		enclaveId,
	)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred launching the test-running testsuite container")
	}
	// No need to have a deferred function that tears down the testsuite container if things don't work because the enclave
	//  teardown logic that we have earlier will handle it

	testsuiteEndpointUri := fmt.Sprintf("%v:%v", testsuiteContainerIp.String(), kurtosis_testsuite_rpc_api_consts.ListenPort)
	// TODO SECURITY: Use HTTPS
	conn, err := grpc.Dial(testsuiteEndpointUri, grpc.WithInsecure())
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred dialing the testsuite container endpoint")
	}
	defer conn.Close()
	testsuiteServiceClient := kurtosis_testsuite_rpc_api_bindings.NewTestSuiteServiceClient(conn)

	if err := waitUntilTestsuiteContainerIsAvailable(testSetupExecutionCtx, testsuiteServiceClient); err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while waiting for the testsuite container to become available")
	}

	if err := streamTestsuiteLogsWhileRunningTest(
		testSetupExecutionCtx,
		testTeardownContext,
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

// =========================== PRIVATE HELPER FUNCTIONS =========================================
func waitUntilTestsuiteContainerIsAvailable(ctx context.Context, client kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient) error {
	contextWithTimeout, cancelFunc := context.WithTimeout(ctx, waitForTestsuiteAvailabilityTimeout)
	defer cancelFunc()
	if _, err := client.IsAvailable(contextWithTimeout, &emptypb.Empty{}, grpc.WaitForReady(true)); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the testsuite container to become available")
	}
	return nil
}

// NOTE: This function makes heavy use of deferred functions, to simplify the code
func streamTestsuiteLogsWhileRunningTest(
		testSetupExecutionCtx context.Context,
		testTeardownCtx context.Context,
		log *logrus.Logger,
		dockerManager *docker_manager.DockerManager,
		testsuiteContainerId string,
		testParams parallel_test_params.ParallelTestParams,
		testsuiteServiceClient kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient) error {
	banner_printer.PrintSection(log, "Testsuite Logs", printTestsuiteLogSectionAsError)
	// After this point, we can go back to printing initializer logs
	defer banner_printer.PrintSection(log, "End Testsuite Logs", printTestsuiteLogSectionAsError)

	// NOTE: We use the testSetupExecutionContext so that the logstream from the testsuite container will be closed
	// if the user presses Ctrl-C.

	logStreamer := output.NewLogStreamer(dockerLogStreamerLogLineLabel, log)

	if startStreamingErr := logStreamer.StartStreamingFromDockerLogs(testSetupExecutionCtx, dockerManager,
		testsuiteContainerId); startStreamingErr != nil {
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

	// We need to stop the container BEFORE we stop the logstreamer to ensure that it flushes its logs and sends
	//  an EOF
	defer func() {
		if err := dockerManager.StopContainer(testTeardownCtx, testsuiteContainerId, containerTeardownTimeout); err != nil {
				log.Errorf("An error occurred stopping the testsuite container; this will likely cause a log streamer error:")
				fmt.Fprintln(log.Out, err)
		}
	}()

	testName := testParams.TestName

	// NOTE: We could add a timeout here if necessary
	log.Tracef("%vRegistering test files...", initializerLogPrefix)
	registerFilesArgs := &kurtosis_testsuite_rpc_api_bindings.RegisterFilesArgs{TestName: testName}
	if _, err := testsuiteServiceClient.RegisterFiles(testSetupExecutionCtx, registerFilesArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred registering the files for test '%v'", testName)
	}
	log.Tracef("%vTest files registered successfully", initializerLogPrefix)

	// Setup the test network before running the test
	setupTimeout := time.Duration(testParams.TestSetupTimeoutSeconds) * time.Second
	log.Tracef("%vSetting up test with setup timeout of %v...", initializerLogPrefix, setupTimeout)
	if err := setupTestWithTimeout(testSetupExecutionCtx, setupTimeout, testName, testsuiteServiceClient); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting up the test")
	}
	log.Tracef("%vTest setup completed successfully", initializerLogPrefix)

	// Run the test using the setup network
	runTimeout := time.Duration(testParams.TestRunTimeoutSeconds) * time.Second
	log.Tracef("%vRunning test with run timeout of %v...", initializerLogPrefix, runTimeout)
	if err := runTestWithTimeout(testSetupExecutionCtx, runTimeout, testName, testsuiteServiceClient); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the test")
	}
	log.Tracef("%vTest run completed successfully", initializerLogPrefix)

	return nil
}

func setupTestWithTimeout(
		testSetupExecutionCtx context.Context,
		setupTimeout time.Duration,
		testName string,
		testsuiteServiceClient kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient) error {
	testSetupCtx, testSetupCtxCancelFunc := context.WithTimeout(
		testSetupExecutionCtx,
		setupTimeout,
	)
	defer testSetupCtxCancelFunc()
	setupArgs := &kurtosis_testsuite_rpc_api_bindings.SetupTestArgs{
		TestName: testName,
	}
	if _, err := testsuiteServiceClient.SetupTest(testSetupCtx, setupArgs); err != nil {
		if strings.Contains(err.Error(), contextDeadlineStringError){
			return stacktrace.NewError("Test setup timeout exceeded")
		}
		return stacktrace.Propagate(err, "An error occurred setting up the test network before running the test")
	}
	return nil
}

func runTestWithTimeout(
		testSetupExecutionCtx context.Context,
		runTimeout time.Duration,
		testName string,
		testsuiteServiceClient kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient) error {
	testRunCtx, testRunCtxCancelFunc := context.WithTimeout(
		testSetupExecutionCtx,
		runTimeout,
	)
	defer testRunCtxCancelFunc()
	runArgs := &kurtosis_testsuite_rpc_api_bindings.RunTestArgs{TestName: testName}
	if _, err := testsuiteServiceClient.RunTest(testRunCtx, runArgs); err != nil {
		if strings.Contains(err.Error(), contextDeadlineStringError){
			return stacktrace.NewError("Test run timeout exceeded")
		}
		return stacktrace.Propagate(err, "An error occurred running the test")
	}
	return nil
}
