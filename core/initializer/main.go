/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/commons/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authenticators"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/session_cache"
	"github.com/kurtosis-tech/kurtosis/initializer/docker_flag_parser"
	"github.com/kurtosis-tech/kurtosis/initializer/initializer_container_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
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

	successExitCode = 0
	failureExitCode = 1

	// We don't want to overwhelm slow machines, since it becomes not-obvious what's happening
	defaultParallelism = 2

	// The location on the INITIALIZER container where the Kurtosis storage directory (containing things like JWT
	//  tokens) will be bind-mounted from the host filesystem
	storageDirectoryBindMountDirpath = "/kurtosis"

	// Name of the file within the Kurtosis storage directory where the session cache will be stored
	sessionCacheFilename = "session-cache"
	sessionCacheFileMode os.FileMode = 0600

	// Debug mode forces parallelism == 1, since it doesn't make much sense without it
	debugModeParallelism = 1

	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//                  If you change the below, you need to update the Dockerfile!
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	clientIdArg          = "CLIENT_ID"
	clientSecretArg      = "CLIENT_SECRET"
	customParamsJsonArg  = "CUSTOM_PARAMS_JSON"
	doListArg            = "DO_LIST"
	executionIdArg       = "EXECUTION_ID"
	isDebugModeArg       = "IS_DEBUG_MODE"
	kurtosisApiImageArg  = "KURTOSIS_API_IMAGE"
	kurtosisLogLevelArg  = "KURTOSIS_LOG_LEVEL"
	parallelismArg       = "PARALLELISM"
	testNamesArg         = "TEST_NAMES"
	testSuiteImageArg    = "TEST_SUITE_IMAGE"
	testSuiteLogLevelArg = "TEST_SUITE_LOG_LEVEL"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//                     If you change the above, you need to update the Dockerfile!
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
)


// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//          If you change default values below, you need to update the Dockerfile!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
var flagConfigs = map[string]docker_flag_parser.FlagConfig{
	clientIdArg: {
		Required: false,
		Default:  "",
		HelpText: fmt.Sprintf("An OAuth client ID which is needed for running Kurtosis in CI, and should be left empty when running Kurtosis on a local machine"),
		Type:     docker_flag_parser.StringFlagType,
	},
	clientSecretArg: {
		Required: false,
		Default:  "",
		HelpText: fmt.Sprintf("An OAuth client secret which is needed for running Kurtosis in CI, and should be left empty when running Kurtosis on a local machine"),
		Type:     docker_flag_parser.StringFlagType,
	},
	customParamsJsonArg: {
		Required: false,
		Default:  "{}",
		HelpText: "JSON string containing custom data that will be passed as-is to your testsuite, so your testsuite can modify its operation based on input",
		Type:     docker_flag_parser.StringFlagType,
	},
	doListArg: {
		Required: false,
		Default:  false,
		HelpText: "Rather than running the tests, lists the tests available to run",
		Type:     docker_flag_parser.BoolFlagType,
	},
	executionIdArg: {
		Required: true,
		Default:  "",
		HelpText: "ID used for identifying everything associated with this run of Kurtosis",
		Type:     docker_flag_parser.StringFlagType,
	},
	isDebugModeArg: {
		Required: false,
		Default:  false,
		HelpText: "Turns on debug mode, which will: 1) set parallelism == 1 (overriding any other parallelism argument) 2) connect a port on the local machine a port on the testsuite container, for use in running a debugger in the testsuite container 3) bind every used port for every service to a port on the local machine, for making requests to the services",
		Type:     docker_flag_parser.BoolFlagType,
	},
	kurtosisApiImageArg: {
		Required: true,
		Default:  "",
		HelpText: "The Docker image from the Kurtosis API image repo (https://hub.docker.com/repository/docker/kurtosistech/kurtosis-core_api) that will be used during operation",
		Type:     docker_flag_parser.StringFlagType,
	},
	kurtosisLogLevelArg: {
		Required: false,
		Default: "info",
		HelpText: fmt.Sprintf(
			"The log level that all output generated by the Kurtosis framework itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
		Type: docker_flag_parser.StringFlagType,
	},
	parallelismArg: {
		Required: false,
		Default:  defaultParallelism,
		HelpText: "A positive integer telling Kurtosis how many tests to run concurrently (should be set no higher than the number of cores on your machine, else you'll slow down your tests and potentially hit test timeouts!)",
		Type:     docker_flag_parser.IntFlagType,
	},
	testNamesArg: {
		Required: false,
		Default:  "",
		HelpText: "List of test names to run, separated by '" + initializer_container_constants.TestNameArgSeparator + "' (default or empty: run all tests)",
		Type:     docker_flag_parser.StringFlagType,
	},
	testSuiteImageArg: {
		Required: true,
		Default:  "",
		HelpText: "The name of the Docker image containing your test suite to run",
		Type:     docker_flag_parser.StringFlagType,
	},
	testSuiteLogLevelArg: {
		Required: false,
		Default:  "debug",
		HelpText: fmt.Sprintf("A string that will be passed as-is to the test suite container to indicate " +
			"what log level the test suite container should output at; this string should be meaningful to " +
			"the test suite container because Kurtosis won't know what logging framework the testsuite uses"),
		Type:     docker_flag_parser.StringFlagType,
	},
}
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//             If you change default values above, you need to update the Dockerfile!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!


func main() {
	// TODO DEBUGGING
	/*
	fmt.Println(watermark)

	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	flagParser := docker_flag_parser.NewFlagParser(flagConfigs)
	parsedFlags, err := flagParser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred parsing the initializer CLI flags: %v\n", err)
		os.Exit(failureExitCode)
	}

	kurtosisLogLevel, err := logrus.ParseLevel(parsedFlags.GetString(kurtosisLogLevelArg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred parsing the Kurtosis log level string: %v\n", err)
		os.Exit(failureExitCode)
	}
	logrus.SetLevel(kurtosisLogLevel)

	clientId := parsedFlags.GetString(clientIdArg)
	clientSecret := parsedFlags.GetString(clientSecretArg)
	accessController := getAccessController(
		clientId,
		clientSecret,
	)
	permissions, err := accessController.Authenticate()
	if err != nil {
		logrus.Fatalf(
			"The following error occurred when authenticating this Kurtosis instance: %v",
			err)
		os.Exit(failureExitCode)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker client: %v", err)
		os.Exit(failureExitCode)
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
	*/
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
