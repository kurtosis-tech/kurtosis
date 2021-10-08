/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_executor_parallelizer

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/banner_printer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_execution/output"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_execution/parallel_test_params"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_execution/test_executor"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_suite_launcher"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sort"
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

Returns:
	True if all tests passed, false otherwise
 */
func RunInParallelAndPrintResults(
		testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider,
		enclaveManager *enclave_manager.EnclaveManager,
		kurtosisLogLevel logrus.Level,
		parallelism uint,
		allTestParams map[string]parallel_test_params.ParallelTestParams,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
		isDebugModeEnabled bool) bool {
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
	testParamsChan := make(chan parallel_test_params.ParallelTestParams, len(allTestParams))

	allTestParamsOrderedKeys := getAllTestParamsOrderedKeys(allTestParams)

	logrus.Debug("Loading test params into work queue...")
	for _, allTestParamsKey := range allTestParamsOrderedKeys {
		testParamsChan <- allTestParams[allTestParamsKey]
	}
	close(testParamsChan) // We close the channel so that when all params are consumed, the worker threads won't block on waiting for more params
	logrus.Debug("All test params loaded into work queue")

	// This is where erroneous usages of the system-wide logger will be captured so we can warn the user about them
	// (e.g. using logrus.Info, when testSpecificLogger.Info should have been used)
	erroneousSystemLogCaptureWriter := output.NewErroneousSystemLogCaptureWriter()
	outputManager := output.NewParallelTestOutputManager(logrus.StandardLogger().Out, uint(len(allTestParams)), parallelism)

	logrus.Infof("Launching %v tests with parallelism %v...", len(allTestParams), parallelism)
	disableSystemLogAndRunTestThreads(
		ctx,
		testsuiteExObjNameProvider,
		erroneousSystemLogCaptureWriter,
		outputManager,
		testParamsChan,
		parallelism,
		enclaveManager,
		kurtosisLogLevel,
		testsuiteLauncher,
		isDebugModeEnabled)
	logrus.Info("All tests exited")

	allTestsPassed, err := outputManager.PrintSummary()
	if err != nil {
		logrus.Errorf("An error occurred printing the test summary: %v", err)
		return false
	}

	capturedMessages := erroneousSystemLogCaptureWriter.GetCapturedMessages()
	logErroneousSystemLogging(capturedMessages)

	return allTestsPassed
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func disableSystemLogAndRunTestThreads(
		parentContext context.Context,
		testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider,
		erroneousSystemLogWriter *output.ErroneousSystemLogCaptureWriter,
		outputManager *output.ParallelTestOutputManager,
		testParamsChan chan parallel_test_params.ParallelTestParams,
		parallelism uint,
		enclaveManager *enclave_manager.EnclaveManager,
		kurtosisLogLevel logrus.Level,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
		isDebugModeEnabled bool) {
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
			parentContext,
			testsuiteExObjNameProvider,
			&waitGroup,
			testParamsChan,
			outputManager,
			enclaveManager,
			kurtosisLogLevel,
			testsuiteLauncher,
			isDebugModeEnabled,
		)
	}
	waitGroup.Wait()
}

/*
A function, designed to be run inside a worker thread, that will pull test params from the given test params channel, execute the test, and
push the result to the test results channel
 */
func runTestWorkerGoroutine(
			parentContext context.Context,
			testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider,
			waitGroup *sync.WaitGroup,
			testParamsChan chan parallel_test_params.ParallelTestParams,
			outputManager *output.ParallelTestOutputManager,
			enclaveManager *enclave_manager.EnclaveManager,
			kurtosisLogLevel logrus.Level,
			testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
			isDebugModeEnabled bool) {
	// IMPORTANT: make sure that we mark a thread as done!
	defer waitGroup.Done()

	for testParams := range testParamsChan {
		testName := testParams.TestName
		testLog := outputManager.RegisterTestLaunch(testName)
		passed, executionErr := test_executor.RunTest(
			parentContext,
			testsuiteExObjNameProvider,
			testLog,
			enclaveManager,
			kurtosisLogLevel,
			testsuiteLauncher,
			testParams,
			isDebugModeEnabled,
		)
		outputManager.RegisterTestCompletion(testName, executionErr, passed)
	}
}

/*
Helper function to print a big warning if there was logging to the system-level logging when there should only have been
 logging to the test-specific logger
*/
func logErroneousSystemLogging(capturedErroneousMessages []output.ErroneousSystemLogInfo) {
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
		banner_printer.PrintSection(logrus.StandardLogger(), fmt.Sprintf("Erroneous Message #%d", i+1), logErroneousSystemLogsAsError)
		logrus.Error("Message:")
		logrus.StandardLogger().Out.Write(messageInfo.GetMessage())
		logrus.StandardLogger().Out.Write([]byte("\n")) // The message likely won't come with a newline so we add it
		logrus.Error("")
		logrus.Error("Stacktrace:")
		logrus.StandardLogger().Out.Write(messageInfo.GetStacktrace())
		logrus.StandardLogger().Out.Write([]byte("\n")) // The stacktrace likely won't end with a newline so we add it
	}
}
/*
This Helper function receives the allTestParams maps and takes its keys to create and return an ordered
slice of strings useful to iterate the allTestParams map in an alphabetical order
 */
func getAllTestParamsOrderedKeys(allTestParams map[string]parallel_test_params.ParallelTestParams) []string {
	allTestParamsOrderedKeys := make([]string, 0, len(allTestParams))
	for allTestParamKey := range allTestParams {
		allTestParamsOrderedKeys = append(allTestParamsOrderedKeys, allTestParamKey)
	}
	sort.Strings(allTestParamsOrderedKeys)
	return allTestParamsOrderedKeys
}
