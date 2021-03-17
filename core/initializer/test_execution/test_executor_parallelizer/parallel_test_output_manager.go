/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_executor_parallelizer

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
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

// ================================ Output Manager ==================================================
const (
	logTestNameBannerAsError = false
	logAllTestResultsAsError = false
)

/*
A SINGLE-USE struct for managing the output of tests during parallel execution, such that:
- Once activated, any system logs will get captured by the given interceptor (system logging should never be used while parallel test execution is happening)
- Logging of test output is done synchronously, as tests finish

NOTE: ONLY the logTestOutput method is thread-safe!!
 */
type ParallelTestOutputManager struct {
	// Capture writer that will store any erroneous system logs during (all test logs should be printed through the
	//  test-specific logger)
	interceptor             *erroneousSystemLogCaptureWriter
	writerBeforeManagement  io.Writer

	// Mutex gating access to the internal state and the logger, to ensure that tests trying to log at the same time
	//  don't get their messages jumbled
	mutex *sync.Mutex

	// During management, the the system-level logs - e.g. logrus.Info, logrus.Debug, etc. - will get disabled. However,
	//  we need to log test output in realtime so we still need to log to the same output source. Thus, we
	//  create a new logger with the same characteristics as the logrus standard logger and use that for printing test
	//  information.
	sideChannelLogger	   *logrus.Logger

	// Captures all test output sent through the output manager
	testOutputs  		   map[string]parallelTestOutput

	// The number of tests that we expect to run
	numTestsToRun uint

	// The parallelism being used
	parallelism uint
}

/*
Creates a new output manager to handle the display of parallel test results.
 */
func newParallelTestOutputManager(numTestsToRun uint, parallelism uint) *ParallelTestOutputManager {

	return &ParallelTestOutputManager{
		interceptor:             newErroneousSystemLogCaptureWriter(),
		writerBeforeManagement:  nil,
		sideChannelLogger:       nil,
		testOutputs:             make(map[string]parallelTestOutput),
		numTestsToRun: numTestsToRun,
		parallelism: parallelism,
	}
}

func (manager *ParallelTestOutputManager) logTestLaunch(
		testName string,
		debuggerHostPortBinding nat.PortBinding) *logrus.Logger {

}

	/*
Logs the launching of a new test, including any host-bound ports that the testsuite is using

Args:
	testName: Name of test being launched
	debuggerHostPortBinding: Binding on the host that the testsuite debugger port will have
 */
func (manager *ParallelTestOutputManager) logTestLaunch(
			testName string,
			debuggerHostPortBinding nat.PortBinding) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	var outputLogger *logrus.Logger
	if !manager.isInterceptingStdLogger {
		outputLogger = logrus.StandardLogger()
	} else {
		outputLogger = manager.sideChannelLogger
	}

	message := fmt.Sprintf(
		"Launching test %v ... (testsuite debugger port binding on host: %v:%v)",
		testName,
		debuggerHostPortBinding.HostIP,
		debuggerHostPortBinding.HostPort)
	outputLogger.Info(message)
}

/*
Thread-safe method to log test output, to provide parallel tests a way to print their log messages in real time as
	they finish.
 */
func (manager *ParallelTestOutputManager) logTestOutput(
			testName string,
			executionErr error,
			testPassed bool,
			testLogs io.Reader) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if _, found := manager.testOutputs[testName]; found {
		// We hijack whatever the actual test output was to ensure that the user gets notification of the test failing
		executionErr = stacktrace.NewError(
			"Test %v is logged twice, indicating that it was run twice! This is a bug in Kurtosis that should be fixed!",
			testName)
		testPassed = false
	}
	manager.testOutputs[testName] = parallelTestOutput{
		testName:     testName,
		executionErr: executionErr,
		testPassed:   testPassed,
	}

	var outputLogger *logrus.Logger
	if !manager.isInterceptingStdLogger {
		outputLogger = logrus.StandardLogger()
	} else {
		outputLogger = manager.sideChannelLogger
	}

	printBanner(outputLogger, testName, logTestNameBannerAsError)
	_, err := io.Copy(outputLogger.Out, testLogs)
	if err != nil {
		outputLogger.Error("An error occurred copying the test's logfile to STDOUT; the logs above may not be complete!")
		fmt.Fprintln(outputLogger.Out, err) // Logrus will escape newlines so we don't actually log this
	}

	status := getTestStatusFromResult(executionErr, testPassed)
	switch status {
	case ERRORED:
		outputLogger.Error("Test " + testName + " " + colorizeStr(string(status), ansiBadResultColor))
		outputLogger.Error(executionErr)
	case PASSED:
		outputLogger.Infof("Test " + testName + " " + colorizeStr(string(status), ansiGoodResultColor))
	case FAILED:
		outputLogger.Errorf("Test " + testName + " " + colorizeStr(string(status), ansiBadResultColor))
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
		outputLogger = manager.sideChannelLogger
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


func colorizeStr(str string, ansiColorStr string) string {
	return ansiColorStr + str + ansiResetColor
}
