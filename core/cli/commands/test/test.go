/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authenticators"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/session_cache"
	"github.com/kurtosis-tech/kurtosis/initializer/initializer_container_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_launcher"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_runner"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path"
	"sort"
	"strings"
)

const (
	clientIdArg          = "client-id"
	clientSecretArg      = "client-secret"
	customParamsJsonArg  = "custom-params"
	doListArg            = "list"
	isDebugModeArg       = "debug"
	kurtosisLogLevelArg  = "kurtosis-log-level"
	parallelismArg       = "parallelism"
	testNamesArg         = "tests"
	testSuiteLogLevelArg = "suite-log-level"

	// Positional args
	testsuiteImageArg = "testsuite-image"
	apiContainerImageArg = "api-container-image"

	defaultParallelism = uint32(4)
	testNameArgSeparator = ","
	defaultSuiteLogLevelStr = "info"

	// Debug mode forces parallelism == 1, since it doesn't make much sense without it
	debugModeParallelism = 1

	// Name of the file within the Kurtosis storage directory where the session cache will be stored
	kurtosisDirname = ".kurtosis"
	sessionCacheFilename = "session-cache"
	sessionCacheFileMode os.FileMode = 0600
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	testsuiteImageArg,
	apiContainerImageArg,
}

var TestCmd = &cobra.Command{
	Use:   "test [flags] testsuite_image api_container_image",
	Short: "Runs a Kurtosis testsuite using the specified Kurtosis Core version",
	RunE:  run,
}

var clientId string
var clientSecret string
var customParamsJson string
var isDebugMode bool
var kurtosisLogLevelStr string
var doList bool
var parallelism uint32
var delimitedTestNamesToRun string
var suiteLogLevelStr string

func init() {
	TestCmd.Flags().StringVar(
		&clientId,
		clientIdArg,
		"",
		"An OAuth client ID which is needed for running Kurtosis in CI, and should be left empty when running Kurtosis on a local machine",
	)
	TestCmd.Flags().StringVar(
		&clientSecret,
		clientSecretArg,
		"",
		"An OAuth client secret which is needed for running Kurtosis in CI, and should be left empty when running Kurtosis on a local machine",
	)
	TestCmd.Flags().StringVar(
		&customParamsJson,
		customParamsJsonArg,
		"{}",
		"JSON string containing arbitrary data that will be passed as-is to your testsuite, so it can modify its behaviour based on input",
	)
	TestCmd.Flags().BoolVar(
		&isDebugMode,
		isDebugModeArg,
		false,
		"Turns on debug mode, which will: 1) set parallelism == 1 (overriding any other parallelism argument) 2) connect a port on the local machine a port on the testsuite container, for use in running a debugger in the testsuite container 3) bind every used port for every service to a port on the local machine, for making requests to the services",
	)
	TestCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
	TestCmd.Flags().BoolVar(
		&doList,
		doListArg,
		false,
		"Rather than running the tests, lists the tests available to run",
	)
	TestCmd.Flags().Uint32Var(
		&parallelism,
		parallelismArg,
		defaultParallelism,
		"The number of tests to execute in parallel",
	)
	TestCmd.Flags().StringVar(
		&delimitedTestNamesToRun,
		testNamesArg,
		"",
		"List of test names to run, separated by '" + testNameArgSeparator + "' (default or empty: run all tests)",
	)
	TestCmd.Flags().StringVar(
		&suiteLogLevelStr,
		testSuiteLogLevelArg,
		defaultSuiteLogLevelStr,
		"A string that will be passed as-is to the test suite container to indicate what log level the test suite container should output at; this string should be meaningful to the test suite container because Kurtosis won't know what logging framework the testsuite uses",
	)
}

func run(cmd *cobra.Command, args []string) error {
	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the Kurtosis log level string '%v'", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	parsedPositionalArgs, err := parsePositionalArgs(args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	testsuiteImage := parsedPositionalArgs[testsuiteImageArg]
	apiContainerImage := parsedPositionalArgs[apiContainerImageArg]

	accessController, err := getAccessController(
		clientId,
		clientSecret,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the access controller")
	}
	permissions, err := accessController.Authenticate()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when authenticating this Kurtosis instance")
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}

	executionId := execution_ids.GetExecutionID()

	testsuiteExObjNameProvider := object_name_providers.NewTestsuiteExecutionObjectNameProvider(executionId)

	logrus.Infof("Using custom params: \n%v", customParamsJson)
	testsuiteLauncher := test_suite_launcher.NewTestsuiteContainerLauncher(
		testsuiteExObjNameProvider,
		testsuiteImage,
		suiteLogLevelStr,
		customParamsJson,
	)

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient, apiContainerImage)

	suiteMetadata, err := test_suite_metadata_acquirer.GetTestSuiteMetadata(
		dockerClient,
		testsuiteLauncher,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test suite metadata")
	}

	if err := verifyNoDelimiterCharInTestNames(suiteMetadata); err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying no test name delimiter in the test names")
	}

	if doList {
		printTestsInSuite(suiteMetadata)
		return nil
	}

	testNamesToRun := splitTestsStrIntoTestsSet(delimitedTestNamesToRun)

	var parallelismUint uint
	if isDebugMode {
		logrus.Infof("NOTE: Due to debug mode being set to true, parallelism is set to %v", debugModeParallelism)
		parallelismUint = debugModeParallelism
	} else {
		parallelismUint = uint(parallelism)
	}

	logrus.Infof("Running testsuite with execution ID '%v'...", executionId)
	allTestsPassed, err := test_suite_runner.RunTests(
		permissions,
		testsuiteExObjNameProvider,
		enclaveManager,
		kurtosisLogLevel,
		suiteMetadata,
		testNamesToRun,
		parallelismUint,
		testsuiteLauncher,
		isDebugMode,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running the tests")
	}

	if !allTestsPassed {
		return stacktrace.Propagate(err, "One or more tests didn't pass")
	}
	return nil
}

// Parses the args into a map of positional_arg_name -> value
func parsePositionalArgs(args []string) (map[string]string, error) {
	if len(args) != len(positionalArgs) {
		return nil, stacktrace.NewError("Expected %v positional arguments but got %v", len(positionalArgs), len(args))
	}

	result := map[string]string{}
	for idx, argValue := range args {
		arg := positionalArgs[idx]
		result[arg] = argValue
	}
	return result, nil
}

func getAccessController(
	clientId string,
	clientSecret string) (access_controller.AccessController, error) {
	var accessController access_controller.AccessController
	if len(clientId) > 0 && len(clientSecret) > 0 {
		logrus.Debugf("Running CI machine-to-machine auth flow...")
		accessController = access_controller.NewClientAuthAccessController(
			auth0_constants.RsaPublicKeyCertsPem,
			auth0_authenticators.NewStandardClientCredentialsAuthenticator(),
			clientId,
			clientSecret)
	} else {
		logrus.Debugf("Running developer device auth flow...")
		homeDirpath, err := os.UserHomeDir()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the home directory")
		}
		sessionCacheFilepath := path.Join(homeDirpath, kurtosisDirname, sessionCacheFilename)
		sessionCache := session_cache.NewEncryptedSessionCache(
			sessionCacheFilepath,
			sessionCacheFileMode,
		)
		accessController = access_controller.NewDeviceAuthAccessController(
			auth0_constants.RsaPublicKeyCertsPem,
			sessionCache,
			auth0_authenticators.NewStandardDeviceCodeAuthenticator(),
		)
	}
	return accessController, nil
}

func verifyNoDelimiterCharInTestNames(suiteMetadata *kurtosis_testsuite_rpc_api_bindings.TestSuiteMetadata) error {
	// If any test names have our special test name arg separator, we won't be able to select the test so throw an
	//  error and loudly alert the user
	for testName, _ := range suiteMetadata.TestMetadata {
		if strings.Contains(testName, initializer_container_constants.TestNameArgSeparator) {
			return stacktrace.NewError(
				"Test '%v' contains illegal character '%v'; we use this character for delimiting when choosing which tests to run so test names cannot contain it!",
				testName,
				initializer_container_constants.TestNameArgSeparator)
		}
	}
	return nil
}

func printTestsInSuite(suiteMetadata *kurtosis_testsuite_rpc_api_bindings.TestSuiteMetadata) {
	testNames := []string{}
	for name := range suiteMetadata.TestMetadata {
		testNames = append(testNames, name)
	}
	sort.Strings(testNames)

	logrus.Info("\nTests in test suite:")
	for _, name := range testNames {
		// We intentionally don't use Logrus here so that we always see the output, even with a misconfigured loglevel
		fmt.Println("- " + name)
	}
}

// Split user-input string into actual candidate test names
func splitTestsStrIntoTestsSet(testsStr string) map[string]bool {
	testNamesArgStr := strings.TrimSpace(testsStr)
	testNamesToRun := map[string]bool{}
	if len(testNamesArgStr) > 0 {
		testNamesList := strings.Split(testNamesArgStr, initializer_container_constants.TestNameArgSeparator)
		for _, name := range testNamesList {
			testNamesToRun[name] = true
		}
	}
	return testNamesToRun
}
