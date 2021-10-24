/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_suite_runner

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/access_controller/permissions"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_execution/parallel_test_params"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_execution/test_executor_parallelizer"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/test_suite_launcher"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"sort"
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
	executionInstanceUuid: The UUID  uniquely identifying this testsuite execution
	dockerClient: Docker client to use when interacting with the Docker engine
	artifactCache: The artifact cache where artifacts needed by the tests-to-run will be downloaded
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
		engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
		testsuiteExObjNameProvider *object_name_providers.TestsuiteExecutionObjectNameProvider,
		kurtosisLogLevel logrus.Level,
	    apiContainerImage string,
		testSuiteMetadata *kurtosis_testsuite_rpc_api_bindings.TestSuiteMetadata,
		testNamesToRun map[string]bool,
		testParallelism uint,
		testsuiteLauncher *test_suite_launcher.TestsuiteContainerLauncher,
		isDebugModeEnabled bool,
	) (allTestsPassed bool, executionErr error) {
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

	logrus.Infof("Running %v tests:", len(testNamesToRun))
	for _, testName := range orderedTestNames {
		logrus.Infof(" - %v", testName)
	}

	testParams, err := buildTestParams(testNamesToRun, testSuiteMetadata)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred building the test params map")
	}

	allTestsPassed = test_executor_parallelizer.RunInParallelAndPrintResults(
		engineClient,
		testsuiteExObjNameProvider,
		kurtosisLogLevel,
		apiContainerImage,
		testParallelism,
		testParams,
		testsuiteLauncher,
		isDebugModeEnabled,
	)
	return allTestsPassed, nil
}

/*
Helper function to build, from the set of tests to run, the map of test params that we'll pass to the TestExecutorParallelizer

Args:
	testsToRun: A "set" of test names to run in parallel
 */
func buildTestParams(
		testNamesToRun map[string]bool,
		testSuiteMetadata *kurtosis_testsuite_rpc_api_bindings.TestSuiteMetadata) (map[string]parallel_test_params.ParallelTestParams, error) {
	testParams := make(map[string]parallel_test_params.ParallelTestParams)
	for testName, _ := range testNamesToRun {
		testMetadata, found := testSuiteMetadata.TestMetadata[testName]
		if !found {
			return nil, stacktrace.NewError("Could not find test metadata for test '%v'", testName)
		}

		testParamsForTest := *parallel_test_params.NewParallelTestParams(
			testName,
			testMetadata.TestSetupTimeoutInSeconds,
			testMetadata.TestRunTimeoutInSeconds,
			testMetadata.IsPartitioningEnabled,
		)
		logrus.Debugf(
			"Built parallel test param for test '%v' and test metadata '%v'",
			testName,
			testMetadata,
		)

		testParams[testName] = testParamsForTest
	}
	return testParams, nil
}
