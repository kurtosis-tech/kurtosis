/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test

import (
	"encoding/binary"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/commons/container_own_id_finder"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/enclave_context"
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
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"time"
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

	defaultParallelism = uint32(4)
	testNameArgSeparator = ","
	defaultSuiteLogLevelStr = "info"
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs a Kurtosis testsuite using the specified Kurtosis Core version",
	RunE:  run,
	Args: cobra.ExactArgs(len(positionalArgsPointers)),
}

var clientId *string
var clientSecret *string
var customParamsJson *string
var isDebugMode *bool
var kurtosisLogLevelStr *string
var doList *bool
var parallelism *uint32
var commaSeparatedTests *string
var suiteLogLevelStr *string

// TODO Positional args???
var testsuiteImage string
var kurtosisCoreVersionStr string
var positionalArgsPointers = []*string{
	&testsuiteImage,
	&kurtosisCoreVersionStr,
}

func init() {
	TestCmd.Flags().StringVar(
		clientId,
		clientIdArg,
		"",
		"An OAuth client ID which is needed for running Kurtosis in CI, and should be left empty when running Kurtosis on a local machine",
	)
	TestCmd.Flags().StringVar(
		clientSecret,
		clientSecretArg,
		"",
		"An OAuth client secret which is needed for running Kurtosis in CI, and should be left empty when running Kurtosis on a local machine",
	)
	TestCmd.Flags().StringVar(
		customParamsJson,
		customParamsJsonArg,
		"{}",
		"JSON string containing arbitrary data that will be passed as-is to your testsuite, so it can modify its behaviour based on input",
	)
	isDebugMode = TestCmd.Flags().Bool(
		isDebugModeArg,
		false,
		"Turns on debug mode, which will: 1) set parallelism == 1 (overriding any other parallelism argument) 2) connect a port on the local machine a port on the testsuite container, for use in running a debugger in the testsuite container 3) bind every used port for every service to a port on the local machine, for making requests to the services",
	)
	TestCmd.Flags().StringVarP(
		kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
	isDebugMode = TestCmd.Flags().Bool(
		doListArg,
		false,
		"Rather than running the tests, lists the tests available to run",
	)
	TestCmd.Flags().Uint32(
		parallelismArg,
		defaultParallelism,
		"The number of tests to execute in parallel",
	)
	TestCmd.Flags().StringVar(
		commaSeparatedTests,
		testNamesArg,
		"",
		"List of test names to run, separated by '" + testNameArgSeparator + "' (default or empty: run all tests)",
	)
	TestCmd.Flags().StringVar(
		suiteLogLevelStr,
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

	accessController := getAccessController(
		*clientId,
		*clientSecret,
	)
	permissions, err := accessController.Authenticate()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when authenticating this Kurtosis instance")
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}

	executionId := parsedFlags.GetString(executionIdArg)

	testsuiteExObjNameProvider := object_name_providers.NewTestsuiteExecutionObjectNameProvider(executionId)

	// We need the initializer container's ID so that we can connect it to the subnetworks so that it can
	//  call the testsuite containers, and the least fragile way we have to find it is to use the execution UUID
	// We must do this before starting any other containers though, else we won't know which one is the initializer
	//  (since using the image name is very fragile)
	initializerContainerId, err := container_own_id_finder.GetOwnContainerIdByName(
		context.Background(),
		docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient),
		executionId,
	)
	if err != nil {
		logrus.Errorf("An error occurred getting the initializer container's ID: %v", err)
		os.Exit(failureExitCode)
	}

	isDebugMode := parsedFlags.GetBool(isDebugModeArg)

	customParamsJson := parsedFlags.GetString(customParamsJsonArg)
	logrus.Infof("Using custom params: \n%v", customParamsJson)
	testsuiteLauncher := test_suite_launcher.NewTestsuiteContainerLauncher(
		testsuiteExObjNameProvider,
		parsedFlags.GetString(testSuiteImageArg),
		parsedFlags.GetString(testSuiteLogLevelArg),
		customParamsJson,
		isDebugMode,
	)

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient, parsedFlags.GetString(kurtosisApiImageArg))

	suiteMetadata, err := test_suite_metadata_acquirer.GetTestSuiteMetadata(
		dockerClient,
		testsuiteLauncher,
	)
	if err != nil {
		logrus.Errorf("An error occurred getting the test suite metadata: %v", err)
		os.Exit(failureExitCode)
	}

	if err := verifyNoDelimiterCharInTestNames(suiteMetadata); err != nil {
		logrus.Errorf("An error occurred verifying no delimiter in the test names: %v", err)
		os.Exit(failureExitCode)
	}

	if parsedFlags.GetBool(doListArg) {
		printTestsInSuite(suiteMetadata)
		os.Exit(successExitCode)
	}

	testNamesToRun := splitTestsStrIntoTestsSet(parsedFlags.GetString(testNamesArg))

	var parallelismUint uint
	if isDebugMode {
		logrus.Infof("NOTE: Due to debug mode being set to true, parallelism is set to %v", debugModeParallelism)
		parallelismUint = debugModeParallelism
	} else {
		parallelismUint = uint(parsedFlags.GetInt(parallelismArg))
	}

	logrus.Infof("Running testsuite with execution ID '%v'...", executionId)
	allTestsPassed, err := test_suite_runner.RunTests(
		permissions,
		testsuiteExObjNameProvider,
		initializerContainerId,
		enclaveManager,
		kurtosisLogLevel,
		suiteMetadata,
		testNamesToRun,
		parallelismUint,
		testsuiteLauncher,
		isDebugMode,
	)
	if err != nil {
		logrus.Errorf("An error occurred running the tests:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}

	var exitCode int
	if allTestsPassed {
		exitCode = successExitCode
	} else {
		exitCode = failureExitCode
	}
	os.Exit(exitCode)
}

func getAccessController(
	clientId string,
	clientSecret string) access_controller.AccessController {
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
		sessionCacheFilepath := path.Join(storageDirectoryBindMountDirpath, sessionCacheFilename)
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
	return accessController
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
