/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller"
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
	defaultParallelism = 4

	// Web link shown to users who do not authenticate.
	licenseWebUrl = "https://kurtosistech.com/"

	// The location on the INITIALIZER container where the suite execution volume will be mounted
	// A user MUST mount a volume here
	suiteExecutionVolumeMountDirpath = "/suite-execution"

	// The location on the INITIALIZER container where the Kurtosis storage directory (containing things like JWT
	//  tokens) will be bind-mounted from the host filesystem
	storageDirectoryBindMountDirpath = "/kurtosis"
)

func main() {
	// NOTE: we'll want to chnage the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	testSuiteImageArg := flag.String(
		"test-suite-image",
		"",
		"The name of the Docker image of the test suite that will be run")

	doListArg := flag.Bool(
		"list",
		false,
		"Rather than running the tests, lists the tests available to run",
	)

	testNamesArg := flag.String(
		"test-names",
		"",
		"List of test names to run, separated by '" + testNameArgSeparator + "' (default or empty: run all tests)",
	)
	kurtosisLogLevelArg := flag.String(
		"kurtosis-log-level",
		"debug",
		fmt.Sprintf("Log level to use for Kurtosis itself (%v)", logrus_log_levels.AcceptableLogLevels),
	)
	testSuiteLogLevelArg := flag.String(
		"test-suite-log-level",
		"debug",
		fmt.Sprintf("Log level string to use for the test suite (will be passed to the test suite container as-is"),
	)

	clientIdArg := flag.String(
		"client-id",
		"",
		fmt.Sprintf("Only needed when running in CI. Client ID from CI license."))

	clientSecretArg := flag.String(
		"client-secret",
		"",
		fmt.Sprintf("Only needed when running in CI. Client Secret from CI license."))

	kurtosisApiImageArg := flag.String(
		"kurtosis-api-image",
		defaultKurtosisApiImage,
		"The Docker image that will be used to run the Kurtosis API container")

	parallelismArg := flag.Int(
		"parallelism",
		defaultParallelism,
		"Number of tests to run concurrently (NOTE: should be set no higher than the number of cores on your machine!)")

	customEnvVarsJsonArg := flag.String(
		"custom-env-vars-json",
		"{}",
		"JSON containing key-value mappings of custom environment variables that will be set in " +
			"the Docker environment when running the test suite container (e.g. '{\"MY_VAR\": \"/some/value\"}')")

	suiteExecutionVolumeArg := flag.String(
		"suite-execution-volume",
		"",
		"The name of the Docker volume that will contain all the data for the test suite execution")
	flag.Parse()

	kurtosisLevel, err := logrus.ParseLevel(*kurtosisLogLevelArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred parsing the Kurtosis log level string: %v\n", err)
		os.Exit(failureExitCode)
	}
	logrus.SetLevel(kurtosisLevel)

	authenticated, authorized, err := access_controller.AuthenticateAndAuthorize(
		storageDirectoryBindMountDirpath,
		*clientIdArg,
		*clientSecretArg)
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
	var customEnvVars map[string]string
	if err := json.Unmarshal([]byte(*customEnvVarsJsonArg), &customEnvVars); err != nil {
		logrus.Errorf("An error occurred parsing the custom environment variables JSON: %v", err)
		os.Exit(failureExitCode)
	}

	suiteMetadata, err := test_suite_metadata_acquirer.GetTestSuiteMetadata(
		*testSuiteImageArg,
		*suiteExecutionVolumeArg,
		suiteExecutionVolumeMountDirpath,
		dockerClient,
		*testSuiteLogLevelArg,
		customEnvVars)
	if err != nil {
		logrus.Errorf("An error occurred getting the test suite metadata: %v", err)
		os.Exit(failureExitCode)
	}

	if *doListArg {
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
	testNamesArgStr := strings.TrimSpace(*testNamesArg)
	testNamesToRun := map[string]bool{}
	if len(testNamesArgStr) > 0 {
		testNamesList := strings.Split(testNamesArgStr, testNameArgSeparator)
		for _, name := range testNamesList {
			testNamesToRun[name] = true
		}
	}

	parallelismUint := uint(*parallelismArg)
	allTestsPassed, err := test_suite_runner.RunTests(
		dockerClient,
		*suiteExecutionVolumeArg,
		suiteExecutionVolumeMountDirpath,
		*suiteMetadata,
		testNamesToRun,
		parallelismUint,
		*kurtosisApiImageArg,
		*kurtosisLogLevelArg,
		*testSuiteImageArg,
		*testSuiteLogLevelArg,
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
