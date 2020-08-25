package test_suite_metadata_acquirer

import (
	"bufio"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/todo_rename_new_initializer/banner_printer"
	"github.com/kurtosis-tech/kurtosis/todo_rename_new_initializer/test_suite_env_vars"
	"github.com/palantir/stacktrace"
	"io/ioutil"
	"os"
)

const (
	bridgeNetworkName = "bridge"
	testListingContainerDescription = "Test-Listing Container"
	bindMountsDirpath = "/bind-mounts"
	testNamesFilepath = bindMountsDirpath + "/test-names-filepath.txt"
	logFilepath = bindMountsDirpath + "/test-listing.log"
)

/*
Spins up a testsuite container in test-listing mode and returns the "set" of tests that it spits out
*/
func GetAllTestNamesInSuite(
		testSuiteImage string,
		dockerManager *docker.DockerManager) (map[string]bool, error) {
	// Create the tempfile that the testsuite image will write test names to
	testNamesFp, err := ioutil.TempFile("", "test-names")
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the temp filepath to write test names to")
	}
	testNamesFp.Close()

	containerLogFp, err := ioutil.TempFile("", "test-listing.log")
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the temp filepath to store the " +
			"test suite container logs")
	}
	containerLogFp.Close()
	defer os.Remove(containerLogFp.Name())

	bridgeNetworkIds, err := dockerManager.GetNetworkIdsByName(context.Background(), bridgeNetworkName)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting the network IDs matching the '%v' network",
			bridgeNetworkName)
	}
	if len(bridgeNetworkIds) == 0 || len(bridgeNetworkIds) > 1 {
		return nil, stacktrace.NewError(
			"%v Docker network IDs were returned for the '%v' network - this is very strange!",
			len(bridgeNetworkIds),
			bridgeNetworkName)
	}
	bridgeNetworkId := bridgeNetworkIds[0]

	testListingContainerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		testSuiteImage,
		// TODO parameterize these
		bridgeNetworkId,
		nil,  // Nil because the bridge network will assign IPs on its own; we don't need to (and won't know what IPs are already used)
		map[nat.Port]bool{},
		nil,
		map[string]string{
			test_suite_env_vars.TestNamesFilepathEnvVar:    testNamesFilepath,
			test_suite_env_vars.TestEnvVar:                 "", // We leave this blank to signify that we want test listing, not test execution
			test_suite_env_vars.KurtosisApiIpEnvVar:        "", // Because we're doing test listing, this can be blank
			test_suite_env_vars.TestSuiteLogFilepathEnvVar: logFilepath,
		},
		map[string]string{
			testNamesFp.Name():    testNamesFilepath,
			containerLogFp.Name(): logFilepath,
		},
		map[string]string{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the test suite container to list the tests")
	}

	testListingExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		testListingContainerId)
	if err != nil {
		banner_printer.PrintContainerLogsWithBanners(testListingContainerDescription, containerLogFp.Name())
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the testsuite container to list the tests")
	}
	if testListingExitCode != 0 {
		banner_printer.PrintContainerLogsWithBanners(testListingContainerDescription, containerLogFp.Name())
		return nil, stacktrace.NewError("The testsuite container for listing tests exited with a nonzero exit code")
	}

	tempFpReader, err := os.Open(testNamesFp.Name())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred opening the temp filename containing test names for reading")
	}
	defer tempFpReader.Close()
	scanner := bufio.NewScanner(tempFpReader)

	testNames := map[string]bool{}
	for scanner.Scan() {
		testNames[scanner.Text()] = true
	}

	return testNames, nil
}
