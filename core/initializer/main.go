/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller"
	"github.com/kurtosis-tech/kurtosis/initializer/docker_flag_parser"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_runner"
	"github.com/sirupsen/logrus"
	"os"
	"sort"
	"strings"
)

const (
	successExitCode = 0
	failureExitCode = 1

	testNameArgSeparator = ","

	defaultKurtosisApiImage = "kurtosistech/kurtosis-core_api:latest"

	// We don't want to overwhelm slow machines, since it becomes not-obvious what's happening
	defaultParallelism = 2

	// Web link shown to users who do not authenticate.
	licenseWebUrl = "https://kurtosistech.com/"

	// The location on the INITIALIZER container where the suite execution volume will be mounted
	// A user MUST mount a volume here
	suiteExecutionVolumeMountDirpath = "/suite-execution"

	// The location on the INITIALIZER container where the Kurtosis storage directory (containing things like JWT
	//  tokens) will be bind-mounted from the host filesystem
	storageDirectoryBindMountDirpath = "/kurtosis"


	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//            If you change any of these arguments, you need to update the Dockerfile!
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	doListArg = "DO_LIST"
	testSuiteImageArg = "TEST_SUITE_IMAGE"
	showHelpArg = "SHOW_HELP"
	testNamesArg = "TEST_NAMES"
	kurtosisLogLevelArg = "KURTOSIS_LOG_LEVEL"
	testSuiteLogLevelArg = "TEST_SUITE_LOG_LEVEL"
	clientIdArg = "CLIENT_ID"
	clientSecretArg = "CLIENT_SECRET"
	kurtosisApiImageArg = "KURTOSIS_API_IMAGE"
	parallelismArg = "PARALLELISM"
	customEnvVarsJsonArg = "CUSTOM_ENV_VARS_JSON"
	suiteExecutionVolumeArg = "SUITE_EXECUTION_VOLUME"
)

var flagConfigs = map[string]docker_flag_parser.FlagConfig{
	doListArg: {
		Required: false,
		Default:  false,
		HelpText: "Rather than running the tests, lists the tests available to run",
		Type:     docker_flag_parser.BoolFlagType,
	},
	testSuiteImageArg: {
		Required: true,
		Default:  "",
		HelpText: "The name of the Docker image of the test suite that will be run",
		Type:     docker_flag_parser.StringFlagType,
	},
	showHelpArg: {
		Required: false,
		Default:  false,
		HelpText: "Shows this help message",
		Type:     docker_flag_parser.BoolFlagType,
	},
	testNamesArg: {
		Required: false,
		Default:  "",
		HelpText: "List of test names to run, separated by '" + testNameArgSeparator + "' (default or empty: run all tests)",
		Type:     docker_flag_parser.StringFlagType,
	},
	kurtosisLogLevelArg: {
		Required: false,
		Default: "info",
		HelpText: fmt.Sprintf(
			"Log level to use for Kurtosis itself (%v)",
			strings.Join(logrus_log_levels.AcceptableLogLevels, "|"),
		),
		Type: docker_flag_parser.StringFlagType,
	},
	testSuiteLogLevelArg: {
		Required: false,
		Default:  "debug",
		HelpText: fmt.Sprintf("Log level string to use for the test suite (will be passed to the test suite container as-is)"),
		Type:     docker_flag_parser.StringFlagType,
	},
	clientIdArg: {
		Required: false,
		Default:  "",
		HelpText: fmt.Sprintf("Only needed when running in CI. Client ID from CI license."),
		Type:     docker_flag_parser.StringFlagType,
	},
	clientSecretArg: {
		Required: false,
		Default:  "",
		HelpText: fmt.Sprintf("Only needed when running in CI. Client Secret from CI license."),
		Type:     docker_flag_parser.StringFlagType,
	},
	kurtosisApiImageArg: {
		Required: true,
		Default:  "",
		HelpText: "The Docker image that will be used to run the Kurtosis API container",
		Type:     docker_flag_parser.StringFlagType,
	},
	parallelismArg: {
		Required: false,
		Default:  defaultParallelism,
		HelpText: "Number of tests to run concurrently (NOTE: should be set no higher than the number of cores on your machine!)",
		Type:     docker_flag_parser.IntFlagType,
	},
	customEnvVarsJsonArg: {
		Required: false,
		Default:  "{}",
		HelpText: "JSON containing key-value mappings of custom environment variables that will be set in the Docker environment when running the test suite container (e.g. '{\"MY_VAR\": \"/some/value\"}')",
		Type:     docker_flag_parser.StringFlagType,
	},
	suiteExecutionVolumeArg: {
		Required: true,
		Default:  "",
		HelpText: "The name of the Docker volume that will contain all the data for the test suite execution",
		Type:     docker_flag_parser.StringFlagType,
	},
}

func main() {
	// NOTE: we'll want to chnage the ForceColors to false if we ever want structured logging
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

	if parsedFlags.GetBool(showHelpArg) {
		flagParser.ShowUsage()
		os.Exit(failureExitCode)
	}

	kurtosisLevel, err := logrus.ParseLevel(parsedFlags.GetString(kurtosisLogLevelArg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred parsing the Kurtosis log level string: %v\n", err)
		os.Exit(failureExitCode)
	}
	logrus.SetLevel(kurtosisLevel)

	authenticated, authorized, err := access_controller.AuthenticateAndAuthorize(
		storageDirectoryBindMountDirpath,
		parsedFlags.GetString(clientIdArg),
		parsedFlags.GetString(clientSecretArg))
	if err != nil {
		logrus.Fatalf("An error occurred while attempting to authenticate user: %v\n", err)
		os.Exit(failureExitCode)
	}
	if !authenticated {
		logrus.Fatalf("Please register to use Kurtosis. To register, visit %v.\n", licenseWebUrl)
		os.Exit(failureExitCode)
	}
	if !authorized {
		logrus.Fatalf("Your Kurtosis license has expired. To purchase an extended license, visit %v.\n", licenseWebUrl)
		os.Exit(failureExitCode)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker client: %v", err)
		os.Exit(failureExitCode)
	}

	// Parse environment variables
	customEnvVarsJson := parsedFlags.GetString(customEnvVarsJsonArg)
	var customEnvVars map[string]string
	if err := json.Unmarshal([]byte(customEnvVarsJson), &customEnvVars); err != nil {
		logrus.Errorf("An error occurred parsing the custom environment variables JSON: %v", err)
		os.Exit(failureExitCode)
	}

	suiteMetadata, err := test_suite_metadata_acquirer.GetTestSuiteMetadata(
		parsedFlags.GetString(testSuiteImageArg),
		parsedFlags.GetString(suiteExecutionVolumeArg),
		suiteExecutionVolumeMountDirpath,
		dockerClient,
		parsedFlags.GetString(testSuiteLogLevelArg),
		customEnvVars)
	if err != nil {
		logrus.Errorf("An error occurred getting the test suite metadata: %v", err)
		os.Exit(failureExitCode)
	}

	// If any test names have our special test name arg separator, we won't be able to select the test so throw an
	//  error and loudly alert the user
	for testName, _ := range suiteMetadata.TestNames {
		if strings.Contains(testName, testNameArgSeparator) {
			logrus.Errorf(
				"Test '%v' contains illegal character '%v'; we use this character for delimiting when choosing which tests to run so test names cannot contain it!",
				testName,
				testNameArgSeparator)
			os.Exit(failureExitCode)
		}
	}

	if parsedFlags.GetBool(doListArg) {
		testNames := []string{}
		for name := range suiteMetadata.TestNames {
			testNames = append(testNames, name)
		}
		sort.Strings(testNames)

		fmt.Println("\nTests in test suite:")
		for _, name := range testNames {
			// We intentionally don't use Logrus here so that we always see the output, even with a misconfigured loglevel
			fmt.Println("- " + name)
		}
		os.Exit(successExitCode)
	}


	// Split user-input string into actual candidate test names
	testNamesArgStr := strings.TrimSpace(parsedFlags.GetString(testNamesArg))
	testNamesToRun := map[string]bool{}
	if len(testNamesArgStr) > 0 {
		testNamesList := strings.Split(testNamesArgStr, testNameArgSeparator)
		for _, name := range testNamesList {
			testNamesToRun[name] = true
		}
	}

	parallelismUint := uint(parsedFlags.GetInt(parallelismArg))
	allTestsPassed, err := test_suite_runner.RunTests(
		dockerClient,
		parsedFlags.GetString(suiteExecutionVolumeArg),
		suiteExecutionVolumeMountDirpath,
		*suiteMetadata,
		testNamesToRun,
		parallelismUint,
		parsedFlags.GetString(kurtosisApiImageArg),
		parsedFlags.GetString(kurtosisLogLevelArg),
		parsedFlags.GetString(testSuiteImageArg),
		parsedFlags.GetString(testSuiteLogLevelArg),
		customEnvVars)
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
