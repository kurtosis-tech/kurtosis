/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/api_container/server/optional_host_port_binding_supplier"
	"github.com/kurtosis-tech/kurtosis/commons/docker_constants"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	"github.com/kurtosis-tech/kurtosis/initializer/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authenticators"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/session_cache"
	"github.com/kurtosis-tech/kurtosis/initializer/docker_flag_parser"
	"github.com/kurtosis-tech/kurtosis/initializer/initializer_container_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_launcher"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_runner"
	"github.com/kurtosis-tech/kurtosis/test_suite/test_suite_rpc_api/bindings"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

const (

	successExitCode = 0
	failureExitCode = 1

	// We don't want to overwhelm slow machines, since it becomes not-obvious what's happening
	defaultParallelism = 2

	// The location on the INITIALIZER container where the suite execution volume will be mounted
	// A user MUST mount a volume here
	initializerContainerSuiteExVolMountDirpath = "/suite-execution"

	// The location on the INITIALIZER container where the Kurtosis storage directory (containing things like JWT
	//  tokens) will be bind-mounted from the host filesystem
	storageDirectoryBindMountDirpath = "/kurtosis"

	// Name of the file within the Kurtosis storage directory where the session cache will be stored
	sessionCacheFilename = "session-cache"
	sessionCacheFileMode os.FileMode = 0600

	// Debug mode forces parallelism == 1, since it doesn't make much sense without it
	debugModeParallelism = 1

	// Can make these configurable if needed
	hostPortTrackerInterfaceIp = "127.0.0.1"
	hostPortTrackerStartRange = 8000
	hostPortTrackerEndRange = 10000
	// TODO This is wrong - we shouldn't actually specify the protocol at FreeHostPortBindingSupplier construction
	//  time, but instead as a parameter to GetFreePort (so the protocol matches)
	protocolForDebuggerPorts = "tcp"

	maxTimesTryingToFindInitializerContainerId = 5
	timeBetweenTryingToFindInitializerContainerId = 1 * time.Second

	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//                  If you change the below, you need to update the Dockerfile!
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	clientIdArg                 = "CLIENT_ID"
	clientSecretArg             = "CLIENT_SECRET"
	customParamsJson            = "CUSTOM_PARAMS_JSON"
	doListArg                   = "DO_LIST"
	executionUuidArg            = "EXECUTION_UUID"
	isDebugModeArg              = "IS_DEBUG_MODE"
	kurtosisApiImageArg         = "KURTOSIS_API_IMAGE"
	kurtosisLogLevelArg         = "KURTOSIS_LOG_LEVEL"
	parallelismArg              = "PARALLELISM"
	suiteExecutionVolumeNameArg = "SUITE_EXECUTION_VOLUME"
	testNamesArg                = "TEST_NAMES"
	testSuiteImageArg           = "TEST_SUITE_IMAGE"
	testSuiteLogLevelArg        = "TEST_SUITE_LOG_LEVEL"
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
	customParamsJson: {
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
	executionUuidArg: {
		Required: true,
		Default:  "",
		HelpText: "UUID used for identifying everything associated with this run of Kurtosis",
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
	suiteExecutionVolumeNameArg: {
		Required: true,
		Default:  "",
		HelpText: "The name of the Docker volume that will contain all the data for the test suite execution (should be a new volume for each execution!)",
		Type:     docker_flag_parser.StringFlagType,
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

	executionInstanceId := parsedFlags.GetString(executionUuidArg)

	// We need the initializer container's ID so that we can connect it to the subnetworks so that it can
	//  call the testsuite containers, and the least fragile way we have to find it is to use the execution UUID
	// We must do this before starting any other containers though, else we won't know which one is the initializer
	//  (since using the image name is very fragile)
	initializerContainerId, err := getInitializerContainerId(dockerClient, executionInstanceId)
	if err != nil {
		logrus.Errorf("An error occurred getting the initializer container's ID: %v", err)
		os.Exit(failureExitCode)
	}

	suiteExecutionVolume := suite_execution_volume.NewSuiteExecutionVolume(initializerContainerSuiteExVolMountDirpath)

	isDebugMode := parsedFlags.GetBool(isDebugModeArg)

	suiteExecutionVolName := parsedFlags.GetString(suiteExecutionVolumeNameArg)

	var hostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier = nil
	if isDebugMode {
		supplier, err := free_host_port_binding_supplier.NewFreeHostPortBindingSupplier(
			docker_constants.HostMachineDomainInsideContainer,
			hostPortTrackerInterfaceIp,
			protocolForDebuggerPorts,
			hostPortTrackerStartRange,
			hostPortTrackerEndRange,
			map[uint32]bool{}, // We don't know of any taken ports at this point
		)
		if err != nil {
			logrus.Fatalf(
				"An error occurred instanting the free host port binding supplier: %v",
				err,
			)
			os.Exit(failureExitCode)
		}
		hostPortBindingSupplier = supplier
	}

	optionalHostPortBindingSupplier := optional_host_port_binding_supplier.NewOptionalHostPortBindingSupplier(hostPortBindingSupplier)

	testsuiteLauncher := test_suite_launcher.NewTestsuiteContainerLauncher(
		executionInstanceId,
		suiteExecutionVolName,
		parsedFlags.GetString(testSuiteImageArg),
		parsedFlags.GetString(testSuiteLogLevelArg),
		parsedFlags.GetString(customParamsJson),
		optionalHostPortBindingSupplier,
	)

	logrus.Infof("Custom param loaded: \n'%v'",testsuiteLauncher.GetCustomParams())

	apiContainerLauncher := api_container_launcher.NewApiContainerLauncher(
		executionInstanceId,
		parsedFlags.GetString(kurtosisApiImageArg),
		suiteExecutionVolName,
		kurtosisLogLevel,
		hostPortBindingSupplier,
	)

	suiteMetadata, err := test_suite_metadata_acquirer.GetTestSuiteMetadata(
		dockerClient,
		testsuiteLauncher)
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

	artifactCache, err := suiteExecutionVolume.GetArtifactCache()
	if err != nil {
		logrus.Errorf("An error occurred getting the artifact cache: %v", err)
		os.Exit(failureExitCode)
	}

	var parallelismUint uint
	if isDebugMode {
		logrus.Infof("NOTE: Due to debug mode being set to true, parallelism is set to %v", debugModeParallelism)
		parallelismUint = debugModeParallelism
	} else {
		parallelismUint = uint(parsedFlags.GetInt(parallelismArg))
	}

	allTestsPassed, err := test_suite_runner.RunTests(
		permissions,
		executionInstanceId,
		initializerContainerId,
		dockerClient,
		artifactCache,
		suiteMetadata,
		testNamesToRun,
		parallelismUint,
		testsuiteLauncher,
		apiContainerLauncher,
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

func getInitializerContainerId(dockerClient *client.Client, executionInstanceId string) (string, error) {
	logrus.Debugf("Getting initializer container ID...")
	dockerManager := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)

	// For some reason, Docker very occasionally will report 0 containers matching the execution instance ID even
	//  though this container definitely has the right name, so we therefore retry a couple times as a workaround
	// See: https://github.com/moby/moby/issues/42354)
	timesTried := 0
	for timesTried < maxTimesTryingToFindInitializerContainerId {
		matchingIds, err := dockerManager.GetContainerIdsByName(context.Background(), executionInstanceId)
		if err != nil {
			logrus.Debugf("Got an error while trying to get the initializer container ID: %v", err)
		} else if len(matchingIds) != 1 {
			logrus.Debugf("Expected exactly 1 container ID matching execution instance ID '%v' but got %v", executionInstanceId, len(matchingIds))
		} else {
			result := matchingIds[0]
			logrus.Debugf("Got initializer container ID: %v", result)
			return result, nil
		}
		timesTried = timesTried + 1
		if timesTried < maxTimesTryingToFindInitializerContainerId {
			logrus.Debugf("Sleeping for %v then trying to get initializer container ID again...", timeBetweenTryingToFindInitializerContainerId)
			time.Sleep(timeBetweenTryingToFindInitializerContainerId)
		}
	}
	return "", stacktrace.NewError("Couldn't get the ID of the initializer container despite trying %v times with %v between tries", maxTimesTryingToFindInitializerContainerId, timeBetweenTryingToFindInitializerContainerId)
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

func verifyNoDelimiterCharInTestNames(suiteMetadata *bindings.TestSuiteMetadata) error {
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

func printTestsInSuite(suiteMetadata *bindings.TestSuiteMetadata) {
	testNames := []string{}
	for name := range suiteMetadata.TestMetadata {
		testNames = append(testNames, name)
	}
	sort.Strings(testNames)

	fmt.Println("\nTests in test suite:")
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