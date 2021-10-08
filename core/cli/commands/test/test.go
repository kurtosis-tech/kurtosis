/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/auth/access_controller"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/auth/auth0_authenticators"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/auth/auth0_constants"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/auth/session_cache"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/initializer_container_constants"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/test_suite_launcher"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/test_suite_metadata_acquirer"
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/test_suite_runner"
	"github.com/kurtosis-tech/kurtosis-core/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-core/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis-core/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-core/commons/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/user_support_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path"
	"sort"
	"strings"
)

const (
	// Logo generated from https://www.ascii-art-generator.org/, and then hand-modified to look better
	watermark = `
         :@X88@.        .888888888888%  
       :X888888X:     .888S@XSS8@88@;   
      tS%S; 't8@8:   :X8X%'  :8@88:     
    .X88X  .8888t  .888X%  .888:X.      
    't8' .88@8'  .8888%  .8888X                   KURTOSIS TESTING SUPPORT
        .t8XX   :SXX'   :S8SX;          Documentation: ` + user_support_constants.DocumentationUrl +`
      .888S   .888t   .8888X            Github Issues: ` + user_support_constants.GithubIssuesUrl + `
     :88@:  .8@8S'  .8@@8;                    Discord: ` + user_support_constants.DiscordUrl + `
    .88;:  :X@S;   .%tSt.  @88:.                Email: ` + user_support_constants.SupportEmail + `
    .8X   :88t'   8888XX    @8@8:       
    .t8X   ''  .88;8:88@X:    %S@8.     
     ';%X..  .:t8@X:'  8@8.    888S.    
       X888tXX88X:      %8'XS    %XX8:  
        .%8888X'          ;888888@888S  
`

	clientIdArg           = "client-id"
	clientSecretArg       = "client-secret"
	customParamsJsonArg   = "custom-params"
	doListArg             = "list"
	isDebugModeArg        = "debug"
	kurtosisLogLevelArg   = "kurtosis-log-level"
	parallelismArg        = "parallelism"
	delimitedTestNamesArg = "tests"
	testSuiteLogLevelArg  = "suite-log-level"
	kurtosisApiImageArg = "kurtosis-api-image"

	// Positional args
	testsuiteImageArg = "testsuite_image"

	// We don't want to overwhelm slow machines, since it becomes not-obvious what's happening
	defaultParallelism      = uint32(4)
	testNamesDelimiter      = ","
	defaultSuiteLogLevelStr = "info"

	// Debug mode forces parallelism == 1, since it doesn't make much sense without it
	debugModeParallelism = 1

	kurtosisDirname = ".kurtosis"
	// Name of the file within the Kurtosis storage directory where the session cache will be stored
	sessionCacheFilename = "session-cache"
	sessionCacheFileMode os.FileMode = 0600
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	testsuiteImageArg,
}

var TestCmd = &cobra.Command{
	Use:   "test [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
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
var delimitedTestNames string
var suiteLogLevelStr string
var kurtosisApiImage string

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
		&delimitedTestNames,
		delimitedTestNamesArg,
		"",
		"List of test names to run, separated by '" +testNamesDelimiter+ "' (default or empty: run all tests)",
	)
	TestCmd.Flags().StringVar(
		&suiteLogLevelStr,
		testSuiteLogLevelArg,
		defaultSuiteLogLevelStr,
		"A string that will be passed as-is to the test suite container to indicate what log level the test suite container should output at; this string should be meaningful to the test suite container because Kurtosis won't know what logging framework the testsuite uses",
	)
	TestCmd.Flags().StringVar(
		&kurtosisApiImage,
		kurtosisApiImageArg,
		defaults.DefaultApiContainerImage,
		"The image of the Kurtosis API container that should be used inside the enclave that the tests run inside",
	)
}

func run(cmd *cobra.Command, args []string) error {
	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the Kurtosis log level string '%v'", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgs(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	testsuiteImage := parsedPositionalArgs[testsuiteImageArg]

	fmt.Println(watermark)

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

	dockerManager := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)

	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, testsuiteImage)
	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, kurtosisApiImage)

	executionId := execution_ids.GetExecutionID()

	testsuiteExObjNameProvider := object_name_providers.NewTestsuiteExecutionObjectNameProvider(executionId)

	logrus.Infof("Using custom params: \n%v", customParamsJson)
	testsuiteLauncher := test_suite_launcher.NewTestsuiteContainerLauncher(
		testsuiteExObjNameProvider,
		testsuiteImage,
		suiteLogLevelStr,
		customParamsJson,
	)

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient, kurtosisApiImage)

	suiteMetadata, err := test_suite_metadata_acquirer.GetTestSuiteMetadata(
		dockerManager,
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

	testNamesToRun := splitTestsStrIntoTestsSet(delimitedTestNames)

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
		kurtosisDirpath := path.Join(homeDirpath, kurtosisDirname)
		if err := ensureKurtosisDirpathExists(kurtosisDirpath); err != nil {
			logrus.Warnf(
				"An error occurred creating the Kurtosis directory at '%v', which will likely impede saving your session:\n%v",
				kurtosisDirpath,
				err,
			)
		}
		sessionCacheFilepath := path.Join(kurtosisDirpath, sessionCacheFilename)
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

func ensureKurtosisDirpathExists(kurtosisDirpath string) error {
	if _, err := os.Stat(kurtosisDirpath); os.IsNotExist(err) {
		if err := os.Mkdir(kurtosisDirpath, 0777); err != nil {
			return stacktrace.Propagate(
				err,
				"Kurtosis directory '%v' didn't exist, and an error occurred trying to create it",
				kurtosisDirpath,
			)
		}
	}
	return nil
}
