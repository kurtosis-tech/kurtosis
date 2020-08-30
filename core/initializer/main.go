/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_runner"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	successExitCode = 0
	failureExitCode = 1

	testNameArgSeparator = ","
)

func main() {
	testSuiteImageArg := flag.String(
		"test-suite-image",
		"",
		"The name of the Docker image of the test suite that will be run")
	testNamesArg := flag.String(
		"test-names",
		"",
		"List of test names to run, separated by '" + testNameArgSeparator + "' (default or empty: run all tests)",
	)
	kurtosisLogLevelArg := flag.String(
		"kurtosis-log-level",
		logrus_log_levels.Info,
		fmt.Sprintf("Log level to use for Kurtosis itself (%v)", logrus_log_levels.GetAcceptableStrings())
	)

	// TODO add a "list tests" flag
	flag.Parse()

	// TODO make this configurable
	kurtosisLevel, err := logrus_log_levels.LevelFromString(*kurtosisLogLevelArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred setting the Kurtosis log level: %v", err)
		os.Exit(1)
	}
	logrus.SetLevel(kurtosisLevel)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker client: %v", err)
		os.Exit(failureExitCode)
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

	testSuiteRunner := test_suite_runner.NewTestSuiteRunner(
		dockerClient,
		*testSuiteImageArg,
		// TODO parameterize this
		"kurtosistech/kurtosis-core_api",
		// TODO parameterize this
		"trace",
		// TODO parameterize this
		map[string]string{},
		*kurtosisLogLevelArg)

	allTestsPassed, err := testSuiteRunner.RunTests(
		testNamesToRun,
		// TODO parameterize this
		4)
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
