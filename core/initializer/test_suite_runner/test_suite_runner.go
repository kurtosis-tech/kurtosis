/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_runner

import (
	"encoding/binary"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/commons/artifact_cache"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller/permissions"
	"github.com/kurtosis-tech/kurtosis/initializer/test_execution/test_executor_parallelizer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math"
	"net"
	sort "sort"
)

// =============================== Test Suite Runner =========================================
const (
	// This is the IP address that the first Docker subnet will be doled out from, with subsequent Docker networks doled out with
	//  increasing IPs corresponding to the NETWORK_WIDTH_BITS
	SUBNET_START_ADDR = "172.23.0.0"

	BITS_IN_IP4_ADDR = 32

	// Extracted as a separate variable for testing
	suiteExecutionPermissionDeniedErrStr = "Permission denied to execute the test suite"
)

/*
Runs the tests with the given names and prints the results to STDOUT. If no tests are specifically defined, all tests are run.

Args:
	permissions: The permissions the user is running the test suite with
	executionInstanceId: The UUID  uniquely identifying this testsuite execution
	dockerClient: Docker client to use when interacting with the Docker engine
	suiteExecutionVolumeMountDirpath: The mount dirpath, ON THE INITIALIZER, where the suite execution volume is
		mounted.
	testSuiteMetadata: Metadata about the test suite - e.g. name of tests, network width bits, etc.
	testNamesToRun: A "set" of test names to run
	testParallelism: How many tests to run in parallel
	testSuiteImage: The Docker image that will be used to launch the test suite container
	testSuiteLogLevel: The string representing the loglevel of the test suite (the initializer won't be able
		to parse this, so this should be meaningful to the test suite image)
	customTestSuiteEnvVars: Key-value mapping of custom Docker environment variables that will be passed as-is to
		the test suite container

Returns:
	allTestsPassed: True if all tests passed, false otherwise
	executionErr: An error that will be non-nil if an error occurred that prevented the test from running and/or the result
		being retrieved. If this is non-nil, the allTestsPassed value is undefined!
 */
func RunTests(
		permissions *permissions.Permissions,
		executionInstanceId uuid.UUID,
		dockerClient *client.Client,
		suiteExecutionVolumeMountDirpath string,
		testSuiteMetadata test_suite_metadata_acquirer.TestSuiteMetadata,
		testNamesToRun map[string]bool,
		testParallelism uint,
		testsuiteLauncher *test_suite_constants.TestsuiteContainerLauncher,
		freeHostPortBindingSupplier *FreeHostPortBindingSupplier) (allTestsPassed bool, executionErr error) {
	numTestsInSuite := len(testSuiteMetadata.TestMetadata)
	if err := permissions.CanExecuteSuite(numTestsInSuite); err != nil {
		return false, stacktrace.Propagate(
			err,
			suiteExecutionPermissionDeniedErrStr)
	}

	// If the user doesn't specify any test names to run, do all of them
	if len(testNamesToRun) == 0 {
		testNamesToRun = map[string]bool{}
		for testName := range testSuiteMetadata.TestMetadata {
			testNamesToRun[testName] = true
		}
	}

	// Validate all the requested tests exist
	for testName := range testNamesToRun {
		if _, found := testSuiteMetadata.TestMetadata[testName]; !found {
			return false, stacktrace.NewError("No test registered with name '%v'", testName)
		}
	}

	orderedTestNames := []string{}
	for testName, _ := range testNamesToRun {
		orderedTestNames = append(orderedTestNames, testName)
	}
	sort.Strings(orderedTestNames)

	logrus.Infof("Running %v tests with execution ID '%v':", len(testNamesToRun), executionInstanceId.String())
	for _, testName := range orderedTestNames {
		logrus.Infof(" - %v", testName)
	}

	// TODO Switch this to be inside the SuiteExecutionVolume object
	// Download any required artifacts for the tests being run
	logrus.Debug("Downloading artifacts used by the tests...")
	if err := downloadUsedArtifacts(suiteExecutionVolumeMountDirpath, testNamesToRun, testSuiteMetadata); err != nil {
		return false, stacktrace.Propagate(
			err,
			"An error occurred downloading the artifacts needed by the tests being run")
	}
	logrus.Debug("Test artifacts downloaded successfully")

	// NOTE: To implement network partitioning we need to start sidecar containers, so we'll need 2N the IP addresses
	//  that the user requests to avoid running out. We use this crude method - double ALL testnet widths if even
	//  one test has network partitioning enabled - and if running out of IP address ranges is ever a problem we can make
	//  this more precise later ~ ktoday, 2021-01-15
	shouldDoubleNetworkWidth := false
	for testName, _ := range testNamesToRun {
		testMetadata, found := testSuiteMetadata.TestMetadata[testName]
		if !found {
			return false, stacktrace.NewError("Couldn't find test metadata for test '%v'", testName)
		}
		shouldDoubleNetworkWidth = shouldDoubleNetworkWidth || testMetadata.IsPartitioningEnabled
	}
	networkWidthBits := testSuiteMetadata.NetworkWidthBits
	if shouldDoubleNetworkWidth {
		networkWidthBits = networkWidthBits + 1
	}
	logrus.Debugf("Using network width bits: %v", networkWidthBits)

	testParams, err := buildTestParams(testNamesToRun, networkWidthBits, freeHostPortBindingSupplier, testSuiteMetadata)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred building the test params map")
	}

	allTestsPassed = test_executor_parallelizer.RunInParallelAndPrintResults(
		executionInstanceId,
		dockerClient,
		testParallelism,
		testParams,
		testsuiteLauncher)
	return allTestsPassed, nil
}

// Downloads only the artifacts that are needed by the tests being run (i.e. not any artifacts used by
// 	tests which aren't being run)
func downloadUsedArtifacts(
		suiteExecutionVolumeMountDirpath string,
		testNames map[string]bool,
		suiteMetadata test_suite_metadata_acquirer.TestSuiteMetadata) error {
	artifactCache := artifact_cache.NewArtifactCache(suiteExecutionVolumeMountDirpath)
	allTestMetadata := suiteMetadata.TestMetadata
	artifactUrlsToDownloadById := map[string]string{}
	for testName := range testNames {
		testMetadata := allTestMetadata[testName]
		for artifactId, artifactUrl := range testMetadata.UsedArtifacts {
			artifactUrlsToDownloadById[artifactId] = artifactUrl
		}
	}
	if err := artifactCache.DownloadArtifacts(artifactUrlsToDownloadById); err != nil {
		return stacktrace.Propagate(err, "An error occurred downloading the artifacts used by the following tests")
	}
	return nil
}

/*
Helper function to build, from the set of tests to run, the map of test params that we'll pass to the TestExecutorParallelizer

Args:
	testsToRun: A "set" of test names to run in parallel
 */
func buildTestParams(
		testNamesToRun map[string]bool,
		networkWidthBits uint32,
		freeHostPortBindingSupplier *FreeHostPortBindingSupplier,
		testSuiteMetadata test_suite_metadata_acquirer.TestSuiteMetadata) (map[string]test_executor_parallelizer.ParallelTestParams, error) {
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
	testParams := make(map[string]test_executor_parallelizer.ParallelTestParams)
	for testName, _ := range testNamesToRun {
		// Pick the next free available subnet IP, considering all the tests we've started previously
		subnetIpInt := subnetStartIpInt + uint32(testIndex) * uint32(math.Pow(2, float64(networkWidthBits)))
		subnetIp := make(net.IP, 4)
		binary.BigEndian.PutUint32(subnetIp, subnetIpInt)
		subnetCidrStr := fmt.Sprintf("%v/%v", subnetIp.String(), subnetMaskBits)

		freeHostPortBinding, err := freeHostPortBindingSupplier.GetFreePortBinding()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting a free host port binding for test '%v'", testName)
		}

		testMetadata, found := testSuiteMetadata.TestMetadata[testName]
		if !found {
			return nil, stacktrace.NewError("Could not find test metadata for test '%v'", testName)
		}

		testParamsForTest := *test_executor_parallelizer.NewParallelTestParams(testName, subnetCidrStr, freeHostPortBinding, testMetadata)
		logrus.Debugf(
			"Built parallel test param for test '%v' with subnet CIDR string '%v', free host port binding '%v', and test metadata '%v'",
			testName,
			subnetCidrStr,
			freeHostPortBinding,
			testMetadata)

		testParams[testName] = testParamsForTest
		testIndex++
	}
	return testParams, nil
}
