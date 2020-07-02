package initializer

import (
	"encoding/binary"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"net"
	"os"
	"sort"
)

type testResult string
// "enum" for testResult
const (
	PASSED  testResult = "PASSED"
	FAILED  testResult = "FAILED"
	ERRORED testResult = "ERRORED" // Indicates an error during setup that prevented the test from running
)

type TestSuiteRunner struct {
	testSuite               testsuite.TestSuite
	testServiceImageName    string
	testControllerImageName string

	// The test controller image-specific string representing the log level, that will be passed as-is to the test controller
	testControllerLogLevel	string
}

const (
	// Each Docker network created will have 2^this_num free IP addresses available
	NETWORK_WIDTH_BITS = 8

	// This is the IP address that the first Docker subnet will be doled out from, with subsequent Docker networks doled out with
	//  increasing IPs corresponding to the NETWORK_WIDTH_BITS
	SUBNET_START_ADDR = "172.23.0.0"

	BITS_IN_IP4_ADDR = 32

	CONTROLLER_LOG_MOUNT_FILEPATH = "/test-controller.log"

	TEST_VOLUME_MOUNTPOINT = "/shared"

	// These are an "API" of sorts - environment variables that are agreed to be set in the test controller's Docker environment
	TEST_VOLUME_ARG            = "TEST_VOLUME"
	TEST_NAME_BASH_ARG         = "TEST_NAME"
	NETWORK_NAME_ARG		   = "NETWORK_NAME"
	SUBNET_MASK_ARG            = "SUBNET_MASK"
	GATEWAY_IP_ARG             = "GATEWAY_IP"
	LOG_FILEPATH_ARG           = "LOG_FILEPATH"
	LOG_LEVEL_ARG              = "LOG_LEVEL"
	TEST_IMAGE_NAME_ARG        = "TEST_IMAGE_NAME"
	TEST_CONTROLLER_IP_ARG     = "TEST_CONTROLLER_IP"
	TEST_VOLUME_MOUNTPOINT_ARG = "TEST_VOLUME_MOUNTPOINT"

	SUCCESS_EXIT_CODE = 0
)


/*
Creates a new TestSuiteRunner with the following arguments
 */
func NewTestSuiteRunner(
			testSuite testsuite.TestSuite,
			testServiceImageName string,
			testControllerImageName string,
			testControllerLogLevel string) *TestSuiteRunner {
	return &TestSuiteRunner{
		testSuite:               testSuite,
		testServiceImageName:    testServiceImageName,
		testControllerImageName: testControllerImageName,
		testControllerLogLevel: testControllerLogLevel,
	}
}

/*
Runs the tests with the given names. If no tests are specifically defined, all tests are run.
 */
func (runner TestSuiteRunner) RunTests(testNamesToRun []string, parallelism int) (allTestsPassed bool, executionErr error) {
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

	subnetMaskBits := BITS_IN_IP4_ADDR - NETWORK_WIDTH_BITS

	subnetStartIp := net.ParseIP(SUBNET_START_ADDR)
	if subnetStartIp == nil {
		return false, stacktrace.NewError("Subnet start IP %v was not a valid IP address; this is a code problem", SUBNET_START_ADDR)
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

	executionInstanceId := uuid.Generate()

	testIndex := 0
	testParams := make(map[string]ParallelTestParams)
	for testName, _ := range testsToRun {
		// Pick the next free available subnet IP, considering all the tests we've started previously
		subnetIpInt := subnetStartIpInt + uint32(testIndex) * uint32(math.Pow(2, NETWORK_WIDTH_BITS))
		subnetIp := make(net.IP, 4)
		binary.BigEndian.PutUint32(subnetIp, subnetIpInt)
		subnetCidrStr := fmt.Sprintf("%v/%v", subnetIp.String(), subnetMaskBits)

		tempFilename := fmt.Sprintf("%v-%v", executionInstanceId, testName)
		tempFp, err := ioutil.TempFile("", tempFilename)
		if err != nil {
			return false, stacktrace.Propagate(err, "An error occurred creating temporary file to contain logs of test %v", testName)
		}
		// We're responsible for cleaning up our own tempfiles
		logrus.Tracef("Temp logfile: %v", tempFp.Name()) // TODO DEBUGGING
		// defer os.Remove(tempFp.Name()) // TODO DEBUGGING
		defer tempFp.Close()

		testParams[testName] = ParallelTestParams{
			testName:            testName,
			logFp:               tempFp,
			subnetMask:          subnetCidrStr,
			executionInstanceId: executionInstanceId,
		}
		testIndex++
	}

	// Initialize a Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, stacktrace.Propagate(err,"Failed to initialize Docker client from environment.")
	}


	testExecutor := NewParallelTestExecutor(
		executionInstanceId,
		dockerClient,
		runner.testControllerImageName,
		runner.testControllerLogLevel,
		runner.testServiceImageName,
		parallelism)

	logrus.Infof("Running %v tests with execution ID: %v", len(testsToRun), executionInstanceId.String())
	testOutputs := testExecutor.RunTestsInParallel(testParams)

	allTestResults := make(map[string]testResult)
	sort.Strings(testNamesToRun)
	for _, name := range testNamesToRun {
		output := testOutputs[name]
		passed := output.Passed
		executionErr := output.ExecutionErr
		logFp := output.LogFp

		logrus.Infof("---------------------------------- %v --------------------------------", name)
		bytesWritten, err := io.Copy(os.Stdout, logFp)
		logrus.Tracef("Wrote %v bytes to STDOUT from test logfile", bytesWritten)
		if err != nil {
			logrus.Error("An error occurred copying the test's logfile to STDOUT; the logs above may not be complete!")
			logrus.Error(err)
		}

		result := logTestResult(name, executionErr, passed)
		allTestResults[name] = result
	}

	logrus.Info("================================== TEST RESULTS ================================")
	allTestsPassed = true
	for testName, result := range allTestResults {
		logrus.Infof("- %v: %v", testName, result)
		allTestsPassed = allTestsPassed && result == PASSED
	}

	return allTestsPassed, nil
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
		logrus.Warnf("Test %v %v", testName, result)
		logrus.Warnf("Error reason: %v", executionErr)
	case PASSED:
		logrus.Infof("Test %v %v", testName, result)
	case FAILED:
		logrus.Warnf("Test %v %v", testName, result)
	}
	return result
}

