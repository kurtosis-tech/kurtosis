package initializer

import (
	"encoding/binary"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/kurtosis-tech/kurtosis/initializer/parallelism"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"net"
	"os"
	"sort"
	"time"
)

// =============================== "enum" for test result =========================================
type testResult string
const (
	PASSED  testResult = "PASSED"
	FAILED  testResult = "FAILED"
	ERRORED testResult = "ERRORED" // Indicates an error during setup that prevented the test from running
)

// =============================== Test Suite Runner =========================================
type TestSuiteRunner struct {
	testSuite               testsuite.TestSuite
	testServiceImageName    string
	testControllerImageName string

	// The test controller image-specific string representing the log level, that will be passed as-is to the test controller
	testControllerLogLevel	string

	// The additional time, on top of the per-test timeout, that's given to tests for setup & teardown
	additionalTestTimeoutBuffer time.Duration
}

const (
	// Each Docker network created will have 2^this_num free IP addresses available
	NETWORK_WIDTH_BITS = 8

	// This is the IP address that the first Docker subnet will be doled out from, with subsequent Docker networks doled out with
	//  increasing IPs corresponding to the NETWORK_WIDTH_BITS
	SUBNET_START_ADDR = "172.23.0.0"

	BITS_IN_IP4_ADDR = 32
)


/*
Creates a new TestSuiteRunner with the following arguments
 */
func NewTestSuiteRunner(
			testSuite testsuite.TestSuite,
			testServiceImageName string,
			testControllerImageName string,
			testControllerLogLevel string,
			additionalTestTimeoutBuffer time.Duration) *TestSuiteRunner {
	return &TestSuiteRunner{
		testSuite:               testSuite,
		testServiceImageName:    testServiceImageName,
		testControllerImageName: testControllerImageName,
		testControllerLogLevel: testControllerLogLevel,
		additionalTestTimeoutBuffer: additionalTestTimeoutBuffer,
	}
}

// TODO Change the argument of this to be a "set" of tests, so we don't have to deal with duplication here??
/*
Runs the tests with the given names and prints the results to STDOUT. If no tests are specifically defined, all tests are run.
 */
func (runner TestSuiteRunner) RunTests(testNamesToRun []string, testParallelism uint) (allTestsPassed bool, executionErr error) {
	allTests := runner.testSuite.GetTests()

	// If the user doesn't specify any test names to run, run all of them
	if len(testNamesToRun) == 0 {
		testNamesToRun = make([]string, 0, len(runner.testSuite.GetTests()))
		for testName, _ := range allTests {
			testNamesToRun = append(testNamesToRun, testName)
		}
	}

	// Validate all the requested tests exist
	testsToRun := make(map[string]testsuite.Test)
	for _, testName := range testNamesToRun {
		test, found := allTests[testName]
		if !found {
			return false, stacktrace.NewError("No test registered with name '%v'", testName)
		}
		testsToRun[testName] = test
	}

	executionInstanceId := uuid.Generate()
	testParams, err := buildTestParams(executionInstanceId, testsToRun)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred building the test params map")
	}

	// Initialize a Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, stacktrace.Propagate(err,"Failed to initialize Docker client from environment.")
	}

	testExecutor := parallelism.NewTestExecutorParallelizer(
		executionInstanceId,
		dockerClient,
		runner.testControllerImageName,
		runner.testControllerLogLevel,
		runner.testServiceImageName,
		testParallelism,
		runner.additionalTestTimeoutBuffer)

	logrus.Infof("Running %v tests with execution ID %v...", len(testsToRun), executionInstanceId.String())
	interceptor := parallelism.NewErroneousSystemLogCaptureWriter()
	testOutputs := testExecutor.RunInParallel(interceptor, testParams)

	logrus.Infof("Printing results for %v tests...", len(testsToRun))
	allTestsPassed = processTestOutputs(testsToRun, testOutputs)

	// If there was any erroneous system-level logging, loudly display that to the user
	capturedErroneousMessages := interceptor.GetCapturedMessages()
	if len(capturedErroneousMessages) > 0 {
		logrus.Error("")
		logrus.Error("There were log messages printed to the system-level logger during parallel test execution!")
		logrus.Error("Because the system-level logger is shared and the tests run in parallel, the messages cannot be")
		logrus.Error(" attributed to any specific test. This means either:")
		logrus.Error("   1) there's a bug in Kurtosis, and a system-level logger call was used when a test-specific logger")
		logrus.Error("       should have been (likely)")
		logrus.Error("   2) third-party code called logrus independently, and there's nothing we can do (unlikely, but possible)")
		logrus.Error("")
		logrus.Error("The log message(s) attempted, and the stacktrace(s) of origination, are as follows in the order they were logged:")

		for i, messageInfo := range capturedErroneousMessages {
			logrus.Errorf("----------------- Erroneous Message #%d -------------------", i+1)
			logrus.Error("Message:")
			logrus.StandardLogger().Out.Write(messageInfo.Message)
			logrus.Error("")
			logrus.Error("Stacktrace:")
			logrus.StandardLogger().Out.Write(messageInfo.Stacktrace)
		}
	}

	return allTestsPassed, nil
}

/*
Helper function to build, from the set of tests to run, the map of test params that we'll pass to the TestExecutorParallelizer

Args:
	testsToRun: A "set" of test names to run in parallel
 */
func buildTestParams(executionInstanceId uuid.UUID, testsToRun map[string]testsuite.Test) (map[string]parallelism.ParallelTestParams, error) {

	subnetMaskBits := BITS_IN_IP4_ADDR - NETWORK_WIDTH_BITS

	subnetStartIp := net.ParseIP(SUBNET_START_ADDR)
	if subnetStartIp == nil {
		return nil, stacktrace.NewError("Subnet start IP %v was not a valid IP address; this is a code problem", SUBNET_START_ADDR)
	}

	// The IP can be either 4 bytes or 16 bytes long; we need to handle both
	//  else we'll get a silent 0 value for the int!
	// See https://gist.github.com/ammario/649d4c0da650162efd404af23e25b86b
	var subnetStartIpInt uint32
	if len(subnetStartIp) == 16 {
		subnetStartIpInt = binary.BigEndian.Uint32(subnetStartIp[12:16])
	} else {
		subnetStartIpInt = binary.BigEndian.Uint32(subnetStartIp)
	}

	testIndex := 0
	testParams := make(map[string]parallelism.ParallelTestParams)
	for testName, test := range testsToRun {
		// Pick the next free available subnet IP, considering all the tests we've started previously
		subnetIpInt := subnetStartIpInt + uint32(testIndex) * uint32(math.Pow(2, NETWORK_WIDTH_BITS))
		subnetIp := make(net.IP, 4)
		binary.BigEndian.PutUint32(subnetIp, subnetIpInt)
		subnetCidrStr := fmt.Sprintf("%v/%v", subnetIp.String(), subnetMaskBits)

		tempFilename := fmt.Sprintf("%v-%v", executionInstanceId, testName)
		tempFp, err := ioutil.TempFile("", tempFilename)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating temporary file to contain logs of test %v", testName)
		}
		logrus.Tracef("Temp file to write logs for test %v to: %v", testName, tempFp.Name())

		testParams[testName] = *parallelism.NewParallelTestParams(testName, test, tempFp, subnetCidrStr, executionInstanceId)
		testIndex++
	}
	return testParams, nil
}

/*
Helper function to process the TestExecutorParallelizer's output to print the necessary information to STDOUT

Args:
	testsToRun: A "set" of tests that were run
	testOutputs: The output of the

Returns:
	True if all tests passed, false otherwise
 */
func processTestOutputs(testsToRun map[string]testsuite.Test, testOutputs map[string]parallelism.ParallelTestOutput) bool {
	// We want normalized output between runs of the tests suite so we sort the tests by name
	testPrintOrder := []string{}
	for testName, _ := range testsToRun {
		testPrintOrder = append(testPrintOrder, testName)
	}
	sort.Strings(testPrintOrder)

	allTestResults := make(map[string]testResult)
	for _, name := range testPrintOrder {
		logrus.Infof("---------------------------------- %v --------------------------------", name)

		output := testOutputs[name]
		passed := output.TestPassed
		executionErr := output.ExecutionErr
		logFp := output.LogFp

		// Close our log FP now that we're done writing, to switch to read mode
		logFp.Close()
		readLogFp, err := os.Open(logFp.Name())
		if err != nil {
			logrus.Error("An error occurred opening the test's logfile for reading; logs for this test are unavailable")
			fmt.Fprintln(logrus.StandardLogger().Out, err) // Logrus will escape newlines so we don't actually log this
		} else {
			bytesWritten, err := io.Copy(logrus.StandardLogger().Out, readLogFp)
			logrus.Tracef("Wrote %v bytes to STDOUT from test logfile", bytesWritten)
			if err != nil {
				logrus.Error("An error occurred copying the test's logfile to STDOUT; the logs above may not be complete!")
				fmt.Fprintln(logrus.StandardLogger().Out, err) // Logrus will escape newlines so we don't actually log this
			}
		}
		readLogFp.Close()
		os.Remove(logFp.Name()) // We're responsible for cleaning up the temp file we created

		result := logTestResult(name, executionErr, passed)
		allTestResults[name] = result
	}

	logrus.Info("================================== TEST RESULTS ================================")
	allTestsPassed := true
	for _, testName := range testPrintOrder {
		result := allTestResults[testName]
		logrus.Infof("- %v: %v", testName, result)
		allTestsPassed = allTestsPassed && result == PASSED
	}

	return allTestsPassed
}

/*
Handles determining the proper test status and logging it.
Returns the testResult for convenience.
*/
func logTestResult(testName string, executionErr error, testPassed bool) testResult {
	var result testResult
	if executionErr != nil {
		result = ERRORED
	} else {
		if testPassed {
			result = PASSED
		} else {
			result = FAILED
		}
	}

	switch result {
	case ERRORED:
		logrus.Errorf("Test %v %v", testName, result)
		logrus.Errorf("Error reason: %v", executionErr)
	case PASSED:
		logrus.Infof("Test %v %v", testName, result)
	case FAILED:
		logrus.Errorf("Test %v %v", testName, result)
	}
	return result
}

