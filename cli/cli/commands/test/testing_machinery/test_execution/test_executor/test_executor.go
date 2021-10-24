/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_executor

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/banner_printer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_execution/output"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_execution/parallel_test_params"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_suite_launcher"
	enclave_liveness_validator "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
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
		dockerClient *client.Client,
		engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
		testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider,
		log *logrus.Logger,
	    apiContainerImage string,
		kurtosisLogLevel logrus.Level,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
		testParams parallel_test_params.ParallelTestParams,
		isDebugModeEnabled bool) (bool, error) {

	testName := testParams.TestName
	isPartitioningEnabled := testParams.IsPartitioningEnabled

	// We'll use the test setup context for setting stuff up so that a cancellation (e.g. Ctrl-C)
	//  will prevent any new things from getting added to Docker. We still want to be able to retrieve exit codes
	//  and logs after a Ctrl-C though, so we use the background context for doing those tasks (rather than the
	//  potentially-cancelled setup context).
	testTeardownContext := context.Background()

	enclaveId := testsuiteExObjNameProvider.ForTestEnclave(testName)

	log.Debugf("Creating enclave for test '%v'....", testName)

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveId: enclaveId,
		ApiContainerImage: apiContainerImage,
		ApiContainerLogLevel: kurtosisLogLevel.String(),
		IsPartitioningEnabled: isPartitioningEnabled,
		ShouldPublishAllPorts: isDebugModeEnabled,
	}

	response, err := engineClient.CreateEnclave(testSetupExecutionCtx, createEnclaveArgs)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveId)
	}
	defer func() {
		stopEnclaveArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{
			EnclaveId: enclaveId,
		}
		if _, err := engineClient.StopEnclave(testSetupExecutionCtx, stopEnclaveArgs); err != nil {
			log.Errorf("An error occurred stopping enclave '%v':", enclaveId)
			fmt.Fprintln(log.Out, err)
			log.Errorf("ACTION REQUIRED: You'll need to manually stop enclave '%v'!", enclaveId)
		}
	}()
	enclaveInfo := response.GetEnclaveInfo()

	apicHostMachineIp, apicHostMachinePort, err := enclave_liveness_validator.ValidateEnclaveLiveness(enclaveInfo)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred verifying the liveness of the enclave created for the test")
	}

	dockerManager := docker_manager.NewDockerManager(
		log,
		dockerClient,
	)

	networkId := enclaveInfo.GetNetworkId()
	kurtosisApiIp := net.ParseIP(enclaveInfo.GetApiContainerInfo().GetIpInsideEnclave())

	enclaveObjNameProvider := object_name_providers.NewEnclaveObjectNameProvider(enclaveId)
	enclaveObjLabelsProvider := object_labels_providers.NewEnclaveObjectLabelsProvider(enclaveId)
	testsuiteContainerName := enclaveObjNameProvider.ForTestRunningTestsuiteContainer()

	apiContainerUrlOnHostMachine := fmt.Sprintf(
		"%v:%v",
		apicHostMachineIp,
		apicHostMachinePort,
	)
	apiContainerConn, err := grpc.Dial(apiContainerUrlOnHostMachine, grpc.WithInsecure())
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred dialling the API container using its host machine port binding '%v'", apiContainerUrlOnHostMachine)
	}
	defer apiContainerConn.Close()
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(apiContainerConn)

	startTestsuiteContainerRegistrationResp, err := apiContainerClient.StartExternalContainerRegistration(
		testSetupExecutionCtx,
		&emptypb.Empty{},
	)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred when starting the registration of the testsuite container")
	}
	testsuiteRegistrationKey := startTestsuiteContainerRegistrationResp.RegistrationKey
	testsuiteIpAddrStr := startTestsuiteContainerRegistrationResp.IpAddr
	testsuiteIpAddr := net.ParseIP(testsuiteIpAddrStr)
	if testsuiteIpAddr == nil {
		return false, stacktrace.NewError("The API container returned an IP address string, '%v', for the testsuite container, but it wasn't parseable to an IP", testsuiteIpAddrStr)
	}

	testsuiteContainerId, hostMachineRpcPortBinding, err := testsuiteLauncher.LaunchTestRunningContainer(
		testSetupExecutionCtx,
		log,
		dockerManager,
		networkId,
		testsuiteContainerName,
		kurtosisApiIp,
		testsuiteIpAddr,
		enclaveId,
		enclaveObjLabelsProvider.ForTestRunningTestsuiteContainer(),
	)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred launching the test-running testsuite container")
	}
	finishRegistrationArgs := &kurtosis_core_rpc_api_bindings.FinishExternalContainerRegistrationArgs{
		RegistrationKey: testsuiteRegistrationKey,
		ContainerId:     testsuiteContainerId,
	}
	if _, err := apiContainerClient.FinishExternalContainerRegistration(testSetupExecutionCtx, finishRegistrationArgs); err != nil {
		return false, stacktrace.Propagate(err, "An error occurred finishing the testsuite container registration with the API container")
	}

	testsuiteEndpointUri := fmt.Sprintf(
		"%v:%v",
		hostMachineRpcPortBinding.HostIP,
		hostMachineRpcPortBinding.HostPort,
	)
	// TODO SECURITY: Use HTTPS
	testsuiteContainerConn, err := grpc.Dial(testsuiteEndpointUri, grpc.WithInsecure())
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred dialing the testsuite container endpoint")
	}
	defer testsuiteContainerConn.Close()
	testsuiteServiceClient := kurtosis_testsuite_rpc_api_bindings.NewTestSuiteServiceClient(testsuiteContainerConn)

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
		isDebugModeEnabled,
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
		testsuiteServiceClient kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient,
		isDebugModeEnabled bool) error {
	banner_printer.PrintSection(log, "Testsuite Logs", printTestsuiteLogSectionAsError)
	// After this point, we can go back to printing initializer logs
	defer banner_printer.PrintSection(log, "End Testsuite Logs", printTestsuiteLogSectionAsError)

	// NOTE: We use the testSetupExecutionContext so that the logstream from the testsuite container will be closed
	// if the user presses Ctrl-C.

	logStreamer := output.NewLogStreamer(dockerLogStreamerLogLineLabel, log)
	if isDebugModeEnabled {
		log.Infof("Debug mode is enabled, test setup and run method will be executed without timeouts")
	}

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
	registerFilesArgs := &kurtosis_testsuite_rpc_api_bindings.RegisterFilesArtifactsArgs{TestName: testName}
	if _, err := testsuiteServiceClient.RegisterFilesArtifacts(testSetupExecutionCtx, registerFilesArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred registering the files for test '%v'", testName)
	}
	log.Tracef("%vTest files registered successfully", initializerLogPrefix)

	// Setup the test network before running the test
	setupTimeout := time.Duration(testParams.TestSetupTimeoutSeconds) * time.Second
	if isDebugModeEnabled {
		setupTimeout = 0
	}
	log.Tracef("%vSetting up test with setup timeout of %v...", initializerLogPrefix, setupTimeout)
	if err := setupTestWithTimeout(testSetupExecutionCtx, setupTimeout, testName, testsuiteServiceClient); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting up the test")
	}
	log.Tracef("%vTest setup completed successfully", initializerLogPrefix)

	// Run the test using the setup network
	runTimeout := time.Duration(testParams.TestRunTimeoutSeconds) * time.Second
	if isDebugModeEnabled {
		runTimeout = 0
	}
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

	testSetupCtx := testSetupExecutionCtx
	if setupTimeout > 0{
		ctxWithTimeout, testSetupCtxCancelFunc := context.WithTimeout(
			testSetupExecutionCtx,
			setupTimeout,
		)
		defer testSetupCtxCancelFunc()

		testSetupCtx = ctxWithTimeout
	}

	if err := setupTest(testSetupCtx, testName, testsuiteServiceClient); err != nil {
		if strings.Contains(err.Error(), contextDeadlineStringError){
			return stacktrace.NewError("Test setup timeout exceeded")
		}
		return stacktrace.Propagate(err, "An error occurred setting up the test network with timeout '%v'", setupTimeout)
	}
	return nil
}

func setupTest(
	testSetupExecutionCtx context.Context,
	testName string,
	testsuiteServiceClient kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient) error {

	setupArgs := &kurtosis_testsuite_rpc_api_bindings.SetupTestArgs{TestName: testName}
	if _, err := testsuiteServiceClient.SetupTest(testSetupExecutionCtx, setupArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting up the test network before running the test")
	}
	return nil
}

func runTestWithTimeout(
		testSetupExecutionCtx context.Context,
		runTimeout time.Duration,
		testName string,
		testsuiteServiceClient kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient) error {

	testRunCtx := testSetupExecutionCtx
	if runTimeout > 0{
		ctxWithTimeout, testRunCtxCancelFunc := context.WithTimeout(
			testSetupExecutionCtx,
			runTimeout,
		)
		defer testRunCtxCancelFunc()
		testRunCtx = ctxWithTimeout
	}

	runArgs := &kurtosis_testsuite_rpc_api_bindings.RunTestArgs{TestName: testName}
	if _, err := testsuiteServiceClient.RunTest(testRunCtx, runArgs); err != nil {
		if strings.Contains(err.Error(), contextDeadlineStringError){
			return stacktrace.NewError("Test run timeout exceeded")
		}
		return stacktrace.Propagate(err, "An error occurred running the test with timeout '%v'", runTimeout)
	}
	return nil
}

func runTest(
	testSetupExecutionCtx context.Context,
	testName string,
	testsuiteServiceClient kurtosis_testsuite_rpc_api_bindings.TestSuiteServiceClient) error {

	runArgs := &kurtosis_testsuite_rpc_api_bindings.RunTestArgs{TestName: testName}
	if _, err := testsuiteServiceClient.RunTest(testSetupExecutionCtx, runArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the test")
	}
	return nil
}
