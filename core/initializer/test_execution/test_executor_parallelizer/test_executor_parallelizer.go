/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor_parallelizer

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_execution/test_executor"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_launcher"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const (
	logErroneousSystemLogsAsError = true
)

/*
Runs the given tests in parallel, printing:
1) the output of tests as they finish
2) a summary of all tests once all tests have finished

Args:
	executionId: The UUID uniquely identifying this execution of the tests
	dockerClient: The handle to manipulating the Docker environment
	parallelism: The number of tests to run concurrently
	numTestsToRun: The number of tests that will be run
	allTestParams: A mapping of test_name -> parameters for running the test
	testSuiteImageName: The name of the Docker image that will be used to run the test controller
	testSuiteLogLevel: A string, meaningful to the test controller, that represents the user's desired log level
	customTestSuiteEnvVars: A custom user-defined map from <env variable name> -> <env variable value> that will be
		passed via Docker environment variables to the test controller

Returns:
	True if all tests passed, false otherwise
 */
func RunInParallelAndPrintResults(
		executionId uuid.UUID,
		dockerClient *client.Client,
		parallelism uint,
		numTestsToRun uint,
		allTestParams map[string]ParallelTestParams,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher) bool {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	// Set up listener for exit signals so we handle it nicely
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// Asynchronously handle graceful exit signals by cancelling context.
	go func() {
		sig, ok := <-sigs
		// signal channel was closed with no syscall signal
		if !ok { return }
		fmt.Printf("\nReceived signal: %v. Cleaning up tests and exiting gracefully...\n", sig)
		cancelFunc()
		// TODO send message to all the parallel threads that they should tear down immediately
	}()
	// These need to be buffered else sending to the channel will be blocking
	testParamsChan := make(chan ParallelTestParams, len(allTestParams))

	logrus.Debug("Loading test params into work queue...")
	for _, testParams := range allTestParams {
		testParamsChan <- testParams
	}
	close(testParamsChan) // We close the channel so that when all params are consumed, the worker threads won't block on waiting for more params
	logrus.Debug("All test params loaded into work queue")

	// This is where erroneous usages of the system-wide logger will be captured so we can warn the user about them
	// (e.g. using logrus.Info, when testSpecificLogger.Info should have been used)
	erroneousSystemLogCaptureWriter := newErroneousSystemLogCaptureWriter()
	outputManager := newParallelTestOutputManager(logrus.StandardLogger().Out, numTestsToRun, parallelism)

	logrus.Infof("Launching %v tests with parallelism %v...", len(allTestParams), parallelism)
	shouldStreamToStdout := numTestsToRun == 1 || parallelism == 1
	disableSystemLogAndRunTestThreads(
		executionId,
		ctx,
		erroneousSystemLogCaptureWriter,
		outputManager,
		testParamsChan,
		parallelism,
		dockerClient,
		testsuiteLauncher,
		shouldStreamToStdout)
	logrus.Info("All tests exited")

	allTestsPassed, err := outputManager.printSummary()
	if err != nil {
		logrus.Error("An error occurred printing the test summary: %v", err)
		return false
	}

	capturedMessages := erroneousSystemLogCaptureWriter.getCapturedMessages()
	logErroneousSystemLogging(capturedMessages)

	return allTestsPassed
}


func disableSystemLogAndRunTestThreads(
		executionId uuid.UUID,
		parentContext context.Context,
		erroneousSystemLogWriter *erroneousSystemLogCaptureWriter,
		outputManager *ParallelTestOutputManager,
		testParamsChan chan ParallelTestParams,
		parallelism uint,
		dockerClient *client.Client,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
		shouldStreamToStdout bool) {
	// When we're running tests in parallel, each test needs to have its logs written to an independent file to avoid getting logs all mixed up.
	// We therefore need to make sure that all code beyond this point uses the per-test logger rather than the systemwide logger.
	// However, it's very difficult for  a coder to remember to use 'log.Info' when they're used to doing 'logrus.Info'.
	// To enforce this, we capture any systemwide logger usages during this function so we can show them later.
	standardLoggerOut := logrus.StandardLogger().Out
	logrus.StandardLogger().SetOutput(erroneousSystemLogWriter)
	defer func() {
		logrus.StandardLogger().SetOutput(standardLoggerOut)
	}()

	var waitGroup sync.WaitGroup
	for i := uint(0); i < parallelism; i++ {
		waitGroup.Add(1)
		go runTestWorkerGoroutine(
			executionId,
			parentContext,
			&waitGroup,
			testParamsChan,
			outputManager,
			dockerClient,
			testsuiteLauncher,
			shouldStreamToStdout)
	}
	waitGroup.Wait()
}

/*
A function, designed to be run inside a worker thread, that will pull test params from the given test params channel, execute the test, and
push the result to the test results channel
 */
func runTestWorkerGoroutine(
			executionId uuid.UUID,
			parentContext context.Context,
			waitGroup *sync.WaitGroup,
			testParamsChan chan ParallelTestParams,
			outputManager *ParallelTestOutputManager,
			dockerClient *client.Client,
			testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
			shouldStreamToStdout bool) {
	// IMPORTANT: make sure that we mark a thread as done!
	defer waitGroup.Done()

	for testParams := range testParamsChan {
		testName := testParams.TestName

		tempFilename := fmt.Sprintf("%v-%v", executionId, testName)
		writingTempFp, err := ioutil.TempFile("", tempFilename)
		if err != nil {
			emptyOutputReader := &strings.Reader{}
			executionErr := stacktrace.Propagate(err, "An error occurred creating temporary file to contain logs of test %v", testName)
			outputManager.registerTestCompletion(testName, executionErr, false, emptyOutputReader)
			continue
		}
		defer writingTempFp.Close()

		// Create a separate logger just for this test that writes to the test execution logfile
		log := logrus.New()
		log.SetLevel(logrus.GetLevel())
		log.SetOutput(writingTempFp)
		log.SetFormatter(logrus.StandardLogger().Formatter)

		testsuiteDebuggerHostPortBinding := testParams.DebuggerHostPortBinding

		outputManager.registerTestLaunch(testName, testsuiteDebuggerHostPortBinding)
		passed, executionErr := test_executor.RunTest(
			executionId,
			parentContext,
			log,
			dockerClient,
			testParams.SubnetMask,
			testsuiteLauncher,
			testsuiteDebuggerHostPortBinding,
			testName,
			testParams.TestMetadata)
		writingTempFp.Close() // Close to flush out anything remaining in the buffer

		// Create a new FP to read the logfile from the start
		var testOutputReader io.Reader
		readingTempFp, err := os.Open(writingTempFp.Name())
		if err != nil {
			errorMsg := fmt.Sprintf("An error occurred opening the test's logfile for reading; logs for this test are unavailable:\n%s", err)
			testOutputReader = strings.NewReader(errorMsg)
		} else {
			defer readingTempFp.Close()
			testOutputReader = readingTempFp
		}
		outputManager.registerTestCompletion(testName, executionErr, passed, testOutputReader)
	}
}

/*
Helper function to print a big warning if there was logging to the system-level logging when there should only have been
 logging to the test-specific logger
*/
func logErroneousSystemLogging(capturedErroneousMessages []erroneousSystemLogInfo) {
	if len(capturedErroneousMessages) == 0 {
		return
	}

	banner_printer.PrintBanner(logrus.StandardLogger(), "Erroneous Logs", logErroneousSystemLogsAsError)
	logrus.Error("There were log messages printed to the system-level logger during parallel test execution!")
	logrus.Error("Because the system-level logger is shared and the tests run in parallel, the messages cannot be")
	logrus.Error(" attributed to any specific test. This is:")
	logrus.Error("   1) A bug in Kurtosis, and a system-level logger call was used when a test-specific logger")
	logrus.Error("       should have been used (likely)")
	logrus.Error("   2) Third-party code calling logrus independently, and there's nothing we can do (unlikely, but possible)")
	logrus.Error("")
	logrus.Error("The log message(s) attempted, and the stacktrace(s) of origination, are as follows in the order they were logged:")
	logrus.Error("")

	for i, messageInfo := range capturedErroneousMessages {
		logrus.Errorf("----------------- Erroneous Message #%d -------------------", i+1)
		logrus.Error("Message:")
		logrus.StandardLogger().Out.Write(messageInfo.message)
		logrus.StandardLogger().Out.Write([]byte("\n")) // The message likely won't come with a newline so we add it
		logrus.Error("")
		logrus.Error("Stacktrace:")
		logrus.StandardLogger().Out.Write(messageInfo.stacktrace)
		logrus.StandardLogger().Out.Write([]byte("\n")) // The stacktrace likely won't end with a newline so we add it
	}
}
