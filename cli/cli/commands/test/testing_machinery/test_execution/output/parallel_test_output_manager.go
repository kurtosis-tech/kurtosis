/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package output

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/banner_printer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"sync"
)

// =============================== "enum" for test result =========================================
type testStatus string
const (
	PASSED  testStatus = "PASSED"
	FAILED  testStatus = "FAILED"
	ERRORED testStatus = "ERRORED" // Indicates an error during setup that prevented the test from running

	ansiGoodResultColor = "\u001b[32;1m"
	ansiBadResultColor = "\u001b[31;1m"
	ansiResetColor = "\u001b[0m"
)

// =============================== Parallel Test Output =========================================
/*
Package struct containing the output of a single test that was run in parallel
*/
type parallelTestOutput struct {
	// Name of the test that was run
	testName string

	// Indicates whether an error occurred during the execution of the test that prevented it from running
	executionErr error

	// Indicates whether the test passed or failed (undefined if the test had a setup error)
	testPassed bool
}


// =============================== Running Test Info =========================================
type runningTestInfo struct {
	fp *os.File

	logger *logrus.Logger
}

// ================================ Output Manager ==================================================
const (
	logTestNameBannerAsError = false
	logAllTestResultsAsError = false
)

/*
A struct for managing the output of tests during parallel execution, so that we don't get jumbled log messages.
When possible, this struct will try to stream logs in realtime.
 */
type ParallelTestOutputManager struct {
	// A logger that will write to the output source, but which doesn't use the system-wide logger (e.g. logrus.Info)
	// so won't trigger the erroneous system logger capturer
	threadSafeOutputLogger *logrus.Logger

	// The maximum number of tests that can be running at a single time. This number will be
	// min(num_tests, parallelism)
	maxConcurrentTestsRunning uint

	// Lock to guard the internal state of this struct
	internalStateLock *sync.Mutex

	// Map of currently-running tests to information about the test
	runningTestInfo map[string]runningTestInfo

	// Captures the output of all finished tests
	testOutputs  		   map[string]parallelTestOutput

	// vvvvvvvvvvvvvvvvvvv Only one test running concurrently vvvvvvvvvvvvvvvvvvvvvvv
	// Streamer that is currently streaming the test's logs
	// This is ONLY used when a single test is running at a time
	logStreamer *LogStreamer
	// ^^^^^^^^^^^^^^^^^^^ Only one test running concurrently ^^^^^^^^^^^^^^^^^^^^^^^
}

/*
Creates a new output manager to handle the display of parallel test results.
 */
func NewParallelTestOutputManager(output io.Writer, numTestsToRun uint, parallelism uint) *ParallelTestOutputManager {
	// maxConcurrentTestsRunning = min(numTests, parallelism)
	var maxConcurrentTestsRunning uint
	if numTestsToRun < parallelism {
		maxConcurrentTestsRunning = numTestsToRun
	} else {
		maxConcurrentTestsRunning = parallelism
	}

	// We wrap the output in a thread-safe wrapper, to ensure that we can safely log to it
	threadSafeWriter := newThreadSafeWriter(output)
	logger := stdLoggerCloneWithOutput(threadSafeWriter)

	return &ParallelTestOutputManager{
		threadSafeOutputLogger:    logger,
		maxConcurrentTestsRunning: maxConcurrentTestsRunning,
		internalStateLock:         &sync.Mutex{},
		runningTestInfo: map[string]runningTestInfo{},
		testOutputs:               make(map[string]parallelTestOutput),
		logStreamer: nil,
	}
}

/*
Logs the launching of a new test, including any host-bound ports that the testsuite is using.

NOTE: This method panics, rather than returning an error, because any errors here aren't recoverable and
	instead indicate a bug in Kurtosis

Args:
	testName: Name of test being launched
	debuggerHostPortBinding: Binding on the host that the testsuite debugger port will have

Returns:
	The logger that the test should write to when doing logging
 */
func (manager *ParallelTestOutputManager) RegisterTestLaunch(testName string) *logrus.Logger {
	manager.internalStateLock.Lock()
	defer manager.internalStateLock.Unlock()

	if uint(len(manager.runningTestInfo)) >= manager.maxConcurrentTestsRunning {
		panic(
			stacktrace.NewError(
				"Cannot register test launch because there are already %v tests running " +
					"in parallel, which is the maximum allowed. This is a bug with Kurtosis, where Kurtosis is trying to " +
					"start too many tests!",
				manager.maxConcurrentTestsRunning,
			),
		)
	}
	if _, found := manager.runningTestInfo[testName]; found {
		panic(
			stacktrace.NewError("Cannot register launch of test %v because it is already running", testName),
		)
	}
	if _, found := manager.testOutputs[testName]; found {
		panic(
			stacktrace.NewError(
				"Cannot register launch of test %v because it's already completed running, indicating it's being " +
					"run twice; this is a bug in Kurtosis",
				testName,
			),
		)
	}

	testOutputFp, err := ioutil.TempFile("", testName)
	if err != nil {
		panic(
			stacktrace.Propagate(
				err,
				"An error occurred creating temporary file to contain logs of test %v",
				testName,
			),
		)
	}
	testOutputLog := stdLoggerCloneWithOutput(testOutputFp)
	runningTestInfo := runningTestInfo{
		fp:     testOutputFp,
		logger: testOutputLog,
	}
	manager.runningTestInfo[testName] = runningTestInfo

	// It's safe to just do this logging because the underlying writer is thread-safe
	message := fmt.Sprintf("Launching test %v ...", testName)
	manager.threadSafeOutputLogger.Info(message)

	if manager.maxConcurrentTestsRunning == 1 {
		streamer := NewLogStreamer("TEST LOG STREAMER", manager.threadSafeOutputLogger)
		if err := streamer.StartStreamingFromFilepath(testOutputFp.Name()); err != nil {
			// An error occurred starting the streamer, so we'll fall back to non-streaming log-printing
			manager.threadSafeOutputLogger.Warn(
				"The following error occurred starting a streamer to print logs of test '%v' in realtime, meaning that test logs " +
					"will only be printed after the test returns: %v",
				testName,
				err,
			)
			manager.logStreamer = nil
		} else {
			// We started the streamer successfully, so print the test name banner before any log messages get outputted
			banner_printer.PrintBanner(testOutputLog, testName, logTestNameBannerAsError)
			manager.logStreamer = streamer
		}
	}

	return testOutputLog
}

func (manager *ParallelTestOutputManager) RegisterTestCompletion(
			testName string,
			executionErr error,
			testPassed bool) {
	manager.internalStateLock.Lock()
	defer manager.internalStateLock.Unlock()

	runningTestInfo, found := manager.runningTestInfo[testName]
	if !found {
		// We hijack whatever the actual test result was to ensure that the user gets notified of the error
		executionErr = stacktrace.NewError(
			"Completion of test %v is registered twice, indicating that it was run twice! This is a bug in Kurtosis that should be fixed!",
			testName)
		testPassed = false
	}

	manager.testOutputs[testName] = parallelTestOutput{
		testName:     testName,
		executionErr: executionErr,
		testPassed:   testPassed,
	}
	delete(manager.runningTestInfo, testName)

	// Since this method registers test completion, no more output should be written to the test logs so we're
	// safe to close the test log fp (and thereby flush any remaining bytes to disk before we print the output)
	testLogFp := runningTestInfo.fp
	testLogFp.Close()

	if manager.maxConcurrentTestsRunning == 1 && manager.logStreamer != nil {
		if err := manager.logStreamer.StopStreaming(); err != nil {
			// This isn't a huge deal if the streamer throws an error when stopping, because the file that the
			// streamer is reading from is already closed and we recreate a new streamer on every new test
			manager.threadSafeOutputLogger.Warnf(
				"The following error occurred stopping the streamer reading logs for test '%v': %v",
				testName,
				err,
			)
			manager.logStreamer = nil
		}
	} else {
		banner_printer.PrintBanner(manager.threadSafeOutputLogger, testName, logTestNameBannerAsError)

		// If we've had more than one test running at once (or we couldn't set up a log streamer), we need to copy
		//  all the test output at once
		// We can do this in a single call because the underlying output is thread-safe
		readerFp, err := os.Open(testLogFp.Name())
		if err != nil {
			// Need to print the banner because it's contained in the logs
			manager.threadSafeOutputLogger.Errorf("Cannot print test logs; the following error occurred opening the test log for reading: %v", err)
		} else {
			if _, err := io.Copy(manager.threadSafeOutputLogger.Out, readerFp); err != nil {
				manager.threadSafeOutputLogger.Errorf("Cannot print test logs; the following error occurred copying the test log to system out: %v", err)
			}
		}
	}

	status := getTestStatusFromResult(executionErr, testPassed)
	switch status {
	case ERRORED:
		manager.threadSafeOutputLogger.Error("Test " + testName + " " + colorizeStr(string(status), ansiBadResultColor))
		manager.threadSafeOutputLogger.Error(executionErr)
	case PASSED:
		manager.threadSafeOutputLogger.Infof("Test " + testName + " " + colorizeStr(string(status), ansiGoodResultColor))
	case FAILED:
		manager.threadSafeOutputLogger.Errorf("Test " + testName + " " + colorizeStr(string(status), ansiBadResultColor))
	}
}



/*
Prints a summary of:
1) the status of all the tests that have completed
2) any erroneous log messages that were captured while the standard logger was being intercepted

Returns:
	allTestsPassed: True if all tests passed, false otherwise
	resultErr: An error (if any) that occurred during summary-printing
 */
func (manager *ParallelTestOutputManager) PrintSummary() (allTestsPassed bool, resultErr error) {
	manager.internalStateLock.Lock()
	defer manager.internalStateLock.Unlock()

	// Any still-running tests at time of summary-printing indicates a bug in Kurtosis
	if len(manager.runningTestInfo) > 0 {
		return false, stacktrace.NewError("Could not print test summary; not all tests have finished (this is a Kurtosis bug)")
	}

	// We sort tests by name because we want normalized output between runs of the suite
	testPrintOrder := []string{}
	for testName, _ := range manager.testOutputs {
		testPrintOrder = append(testPrintOrder, testName)
	}
	sort.Strings(testPrintOrder)

	banner_printer.PrintBanner(manager.threadSafeOutputLogger, "TEST RESULTS", logAllTestResultsAsError)
	allTestsPassed = true
	for _, testName := range testPrintOrder {
		output := manager.testOutputs[testName]
		passed := output.testPassed
		executionErr := output.executionErr
		status := getTestStatusFromResult(executionErr, passed)

		var colorStr string
		var logFunc func(args ...interface{})
		if status == ERRORED || status == FAILED {
			colorStr = ansiBadResultColor
			logFunc = manager.threadSafeOutputLogger.Error
			allTestsPassed = false
		} else {
			colorStr = ansiGoodResultColor
			logFunc = manager.threadSafeOutputLogger.Info
		}
		logFunc("- " + testName + ": " + colorizeStr(string(status), colorStr))
	}
	return allTestsPassed, nil
}

// ================================== Private helper messages ==========================================
func getTestStatusFromResult(executionErr error, testPassed bool) testStatus {
	var result testStatus
	if executionErr != nil {
		result = ERRORED
	} else {
		if testPassed {
			result = PASSED
		} else {
			result = FAILED
		}
	}
	return result
}

// Returns a new logger cloned from the standard logger, but with a different output
// The reason we clone the standard logger is that this will be the one the user configures with
//  their logging level, format, etc. preferences
func stdLoggerCloneWithOutput(output io.Writer) *logrus.Logger {
	// Sadly no copy constructor on this :(
	result := logrus.New()
	result.SetOutput(output)
	result.SetFormatter(logrus.StandardLogger().Formatter)
	result.SetLevel(logrus.StandardLogger().Level)
	// NOTE: we don't copy hooks here because we don't use them - if we ever use hooks, copy them here!

	return result
}

func colorizeStr(str string, ansiColorStr string) string {
	return ansiColorStr + str + ansiResetColor
}
