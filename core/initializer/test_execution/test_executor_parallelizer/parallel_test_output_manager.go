/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor_parallelizer

import (
	"bytes"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/initializer/banner_printer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
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


// =============================== Parallel Test Output =========================================
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
	// Manager for managing a background thread that will stream logs to system out
	// This is ONLY used when a single test is running at a time
	streamerManager *logStreamerManager

	// If an error occurs when creating the streamer, we'll fall back to non-streaming log
	// printing
	didErrorOccurCreatingStreamer bool
	// ^^^^^^^^^^^^^^^^^^^ Only one test running concurrently ^^^^^^^^^^^^^^^^^^^^^^^
}

/*
Creates a new output manager to handle the display of parallel test results.
 */
func newParallelTestOutputManager(output io.Writer, testsToRun uint, parallelism uint) *ParallelTestOutputManager {
	var maxConcurrentTestsRunning uint
	if testsToRun < parallelism {
		maxConcurrentTestsRunning = testsToRun
	} else {
		maxConcurrentTestsRunning = parallelism
	}
	// We wrap the output in a thread-safe wrapper, to ensure that we can safely log to it
	threadSafeWriter := newThreadSafeWriter(output)
	logger := stdLoggerCloneWithOutput(threadSafeWriter)
	return &ParallelTestOutputManager{
		threadSafeOutputLogger:    logger,
		maxConcurrentTestsRunning: maxConcurrentTestsRunning,
		testOutputFilepaths:       map[string]string{},
		streamerShutdownChan:      nil,
		internalStateLock:         &sync.Mutex{},
		testOutputs:               make(map[string]parallelTestOutput),
	}
}

/*
Logs the launching of a new test, including any host-bound ports that the testsuite is using.

NOTE: logTestCompletion

Args:
	testName: Name of test being launched
	debuggerHostPortBinding: Binding on the host that the testsuite debugger port will have

Returns:
	The logger that the test should write to when doing logging
 */
func (manager *ParallelTestOutputManager) registerTestLaunch(
			testName string,
			debuggerHostPortBinding nat.PortBinding) (*logrus.Logger, error) {
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

	if manager.maxConcurrentTestsRunning == 1 {
		if err := manager.streamerManager.startStreamer(testOutputFp.Name()); err != nil {
			manager.threadSafeOutputLogger.Warn(
				"The following error occurred starting a thread to stream logs of test '%v' in realtime, meaning that test logs " +
					"will only be printed after the test returns: %v",
				testName,
				err,
			)
			manager.didErrorOccurCreatingStreamer = true
		} else {
			manager.didErrorOccurCreatingStreamer = false
		}
	}

	// It's safe to just do this logging because the underlying writer is thread-safe
	message := fmt.Sprintf(
		"Launching test %v ... (testsuite debugger port binding on host: %v:%v)",
		testName,
		debuggerHostPortBinding.HostIP,
		debuggerHostPortBinding.HostPort)
	manager.threadSafeOutputLogger.Info(message)

	banner_printer.PrintBanner(testOutputLog, testName, false)
	return testOutputLog, nil
}

func (manager *ParallelTestOutputManager) registerTestCompletion(
			testName string,
			executionErr error,
			testPassed bool) error {
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

	// Since method registers test completion, no more output should be written to the test logs so we're
	// safe to close the test log fp (and thereby flush any remaining bytes to disk before we print the output)
	testLogFp := runningTestInfo.fp
	testLogFp.Close()

	if manager.maxConcurrentTestsRunning == 1 && !manager.didErrorOccurCreatingStreamer {
		if err := manager.streamerManager.stopStreamer(); err != nil {
			manager.threadSafeOutputLogger.Error(
				"The following error occurred stopping the streamer reading logs for test '%v'; this likely means that log-streamign "
				)
			return stacktrace.Propagate(err, "An error occurred stopping the streamer reading logs for test '%v'", testName)
		}
	} else {
		// If we've had more than one test running at once, we need to copy all the output
		// We can do this in a single call because the underlying output is thread-safe
		readerFp, err := os.Open(testLogFp.Name())
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred opening the test log for reading")
		}
		if _, err := io.Copy(manager.threadSafeOutputLogger.Out, readerFp); err != nil {
			return stacktrace.Propagate(err, "An error occurred copying the test logs from the temporary file to the system log")
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
1) the status of all the tests that have been logged to the logger so far
2) any erroneous log messages that were captured while the standard logger was being intercepted
 */
func (manager *ParallelTestOutputManager) printSummary() {
	manager.mutex.Lock()
	manager.mutex.Unlock()

	// We sort tests by name because we want normalized output between runs of the suite
	testPrintOrder := []string{}
	for testName, _ := range manager.testOutputs {
		testPrintOrder = append(testPrintOrder, testName)
	}
	sort.Strings(testPrintOrder)

	var outputLogger *logrus.Logger
	if !manager.isInterceptingStdLogger {
		outputLogger = logrus.StandardLogger()
	} else {
		outputLogger = manager.threadSafeOutputLogger
	}

	printBanner(outputLogger, "TEST RESULTS", logAllTestResultsAsError)
	for _, testName := range testPrintOrder {
		output := manager.testOutputs[testName]
		passed := output.testPassed
		executionErr := output.executionErr
		status := getTestStatusFromResult(executionErr, passed)

		var colorStr string
		var logFunc func(args ...interface{})
		if status == ERRORED || status == FAILED {
			colorStr = ansiBadResultColor
			logFunc = outputLogger.Error
		} else {
			colorStr = ansiGoodResultColor
			logFunc = outputLogger.Info
		}
		logFunc("- " + testName + ": " + colorizeStr(string(status), colorStr))
	}
}

/*
Returns true if all tests captured so far have passed, false otherwise
 */
func (manager *ParallelTestOutputManager) getAllTestsPassed() bool {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	allTestsPassed := true
	for _, output := range manager.testOutputs {
		testHadNoIssues := PASSED == getTestStatusFromResult(output.executionErr, output.testPassed)
		allTestsPassed = allTestsPassed && testHadNoIssues
	}
	return allTestsPassed
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
