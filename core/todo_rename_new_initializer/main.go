package main

import (
	"flag"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/todo_rename_new_initializer/test_suite_metadata_acquirer"
	"github.com/kurtosis-tech/kurtosis/todo_rename_new_initializer/test_suite_runner"
	"github.com/palantir/stacktrace"
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
	// TODO add a "list tests" flag
	flag.Parse()

	// TODO make this configurable
	logrus.SetLevel(logrus.TraceLevel)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker client: %v", err)
		os.Exit(failureExitCode)
	}

	dockerManager, err := docker.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker manager: %v", err)
		os.Exit(failureExitCode)
	}

	testNamesToRun, err := getTestNamesToRun(*testNamesArg, *testSuiteImageArg, dockerManager)
	if err != nil {
		logrus.Errorf("An error occurred when validating the list of tests to run:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}

	executionId := uuid.New()

	allTestsPassed, err := test_suite_runner.RunTests(
		executionId,
		*testSuiteImageArg,
		dockerManager,
		testNamesToRun)
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



/*
Helper function to translate the user-provided string that we receive from the CLI about which tests to run to a "set"
	of the test names to run, validating that all the test names are valid.
 */
func getTestNamesToRun(
			testsToRunStr string,
			testSuiteImage string,
			dockerManager *docker.DockerManager) (map[string]bool, error) {
	// Split user-input string into actual candidate test names
	testNamesArgStr := strings.TrimSpace(testsToRunStr)
	testNamesToRun := map[string]bool{}
	if len(testNamesArgStr) > 0 {
		testNamesList := strings.Split(testNamesArgStr, testNameArgSeparator)
		for _, name := range testNamesList {
			testNamesToRun[name] = true
		}
	}

	allTestNames, err := test_suite_metadata_acquirer.GetAllTestNamesInSuite(testSuiteImage, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the names of the tests in the test suite")
	}

	// If the user doesn't specify any test names to run, do all of them
	if len(testNamesToRun) == 0 {
		testNamesToRun = map[string]bool{}
		for testName := range allTestNames {
			testNamesToRun[testName] = true
		}
	}

	// Validate all the requested tests exist
	for testName := range testNamesToRun {
		if _, found := allTestNames[testName]; !found {
			return nil, stacktrace.NewError("No test registered with name '%v'", testName)
		}
	}
	return testNamesToRun, nil
}



