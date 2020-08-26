package test_suite_runner

import (
	"encoding/binary"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/initializer/parallelism"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math"
	"net"
)

// =============================== Test Suite Runner =========================================
const (
	// This is the IP address that the first Docker subnet will be doled out from, with subsequent Docker networks doled out with
	//  increasing IPs corresponding to the NETWORK_WIDTH_BITS
	SUBNET_START_ADDR = "172.23.0.0"

	BITS_IN_IP4_ADDR = 32

	// TODO Pull this from the testsuite image!
	networkWidthBits = 8
)

/*
An executor to run one or more tests from a given test suite
 */
type TestSuiteRunner struct {
	// Docker client to use for interacting with the Docker engine
	dockerClient *client.Client

	// The name of image of test suite
	testSuiteImage               string

	// The name of the Kurtosis API Docker image
	kurtosisApiImage string

	// Key-value mapping that will be passed as-is to the test suite container on startup in the form of Docker
	// 	environment variables
	customTestSuiteEnvVars map[string]string

	// A string, meaningful only to the test controller, that represents the log level that the controller container should
	//	run with
	testSuiteLogLevel string
}

/*
Creates a new TestSuiteRunner with the given parameters.

Args:
	client: Docker client to use when interacting with the Docker engine
	testSuite: The test suite containing all the user's registered tests
	testControllerImageName: The name of the Docker image of the test controller that will orchestrate test execution
	testSuiteLogLevel: The string representing the loglevel of the test suite (the initializer won't be able
		to parse this, so this should be meaningful to the test suite image)
	networkWidthBits: Each test will get a Docker network with a number of available IP addresses = 2^network_width_bits.
		This parameter should be set high enough so that each test can fit all the services they want.
 */
func NewTestSuiteRunner(
			dockerClient *client.Client,
			testSuiteImage string,
			kurtosisApiImage string,
			testSuiteLogLevel string,
			testControllerEnvVars map[string]string) *TestSuiteRunner {
	return &TestSuiteRunner{
		dockerClient:           dockerClient,
		testSuiteImage:         testSuiteImage,
		kurtosisApiImage:       kurtosisApiImage,
		testSuiteLogLevel:      testSuiteLogLevel,
		customTestSuiteEnvVars: testControllerEnvVars,
	}
}

/*
Runs the tests with the given names and prints the results to STDOUT. If no tests are specifically defined, all tests are run.

Args:
	testNamesToRun: A "set" of test names to run
	testParallelism: How many tests to run in parallel

Returns:
	allTestsPassed: True if all tests passed, false otherwise
	executionErr: An error that will be non-nil if an error occurred that prevented the test from running and/or the result
		being retrieved. If this is non-nil, the allTestsPassed value is undefined!
 */
func (runner TestSuiteRunner) RunTests(testNamesToRun map[string]bool, testParallelism uint) (allTestsPassed bool, executionErr error) {
	stdoutDockerManager, err := commons.NewDockerManager(logrus.StandardLogger(), runner.dockerClient)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	allTestNames, err := test_suite_metadata_acquirer.GetAllTestNamesInSuite(runner.testSuiteImage, stdoutDockerManager)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the names of the tests in the test suite")
	}

	// If the user doesn't specify any test names to run, do all of them
	if len(testNamesToRun) == 0 {
		testNamesToRun = map[string]bool{}
		for testName := range allTestNames {
			testNamesToRun[testName] = true
		}
	}

	// Validate all the requested tests exist
	for testName := range testNamesToRun {
		if _, found := allTestNames[testName]; !found {
			return false, stacktrace.NewError("No test registered with name '%v'", testName)
		}
	}

	executionInstanceId := uuid.Generate()
	testParams, err := buildTestParams(executionInstanceId, testNamesToRun, networkWidthBits)
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
		runner.kurtosisApiImage,
		runner.testSuiteImage,
		runner.testSuiteLogLevel,
		runner.customTestSuiteEnvVars,
		testParallelism)

	logrus.Infof("Running %v tests with execution ID %v...", len(testNamesToRun), executionInstanceId.String())
	allTestsPassed = testExecutor.RunInParallelAndPrintResults(testParams)
	return allTestsPassed, nil
}

/*
Helper function to build, from the set of tests to run, the map of test params that we'll pass to the TestExecutorParallelizer

Args:
	testsToRun: A "set" of test names to run in parallel
 */
func buildTestParams(executionInstanceId uuid.UUID, testNamesToRun map[string]bool, networkWidthBits uint32) (map[string]parallelism.ParallelTestParams, error) {
	subnetMaskBits := BITS_IN_IP4_ADDR - networkWidthBits

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
	for testName, _ := range testNamesToRun {
		// Pick the next free available subnet IP, considering all the tests we've started previously
		subnetIpInt := subnetStartIpInt + uint32(testIndex) * uint32(math.Pow(2, float64(networkWidthBits)))
		subnetIp := make(net.IP, 4)
		binary.BigEndian.PutUint32(subnetIp, subnetIpInt)
		subnetCidrStr := fmt.Sprintf("%v/%v", subnetIp.String(), subnetMaskBits)

		testParams[testName] = *parallelism.NewParallelTestParams(testName, subnetCidrStr, executionInstanceId)
		testIndex++
	}
	return testParams, nil
}
