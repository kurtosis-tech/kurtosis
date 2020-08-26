package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

const (
	testNamesFilepath = "/shared/test-names-filepath"
	kurtosisApiStopContainerTimeout = 30 * time.Second
	dockerSocket = "/var/run/docker.sock"

	// Test suite environment variables
	testNamesFilepathEnvVar    = "TEST_NAMES_FILEPATH"
	testEnvVar                 = "TEST"
	kurtosisApiIpEnvVar        = "KURTOSIS_API_IP"
	testSuiteLogFilepathEnvVar = "LOG_FILEPATH"

	// Kurtosis API environment variables
	testSuiteContainerIdEnvVar = "TEST_SUITE_CONTAINER_ID"
	networkIdEnvVar            = "NETWORK_ID"
	subnetMaskEnvVar           = "SUBNET_MASK"
	gatewayIpEnvVar            = "GATEWAY_IP"
	logLevelEnvVar             = "LOG_LEVEL"
	apiLogFilepathEnvVar       = "LOG_FILEPATH"


	// TODO parameterize
	kurtosisApiImage        = "kurtosistech/kurtosis-core_api"
	bridgeNetworkId         = "b453ce4bac01"
	bridgeNetworkSubnetMask = "172.17.0.0/16"
	bridgeNetworkGatewayIp  = "172.17.0.1"
)

func main() {
	testSuiteImageArg := flag.String(
		"test-suite-image",
		"",
		"The name of the Docker image of the test suite that will be run")
	flag.Parse()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker client: %v", err)
		os.Exit(1)
	}

	dockerManager, err := docker.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker manager: %v", err)
		os.Exit(1)
	}

	freeIpAddrTracker, err := networks.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		bridgeNetworkSubnetMask,
		map[string]bool{
			bridgeNetworkGatewayIp: true,	// gateway IP
		})
	if err != nil {
		logrus.Errorf("An error occurred creating the free IP address tracker: %v", err)
		os.Exit(1)
	}

	// TODO actually use tests
	_, err = getTests(*testSuiteImageArg, dockerManager, freeIpAddrTracker)
	if err != nil {
		logrus.Errorf("An error occurred getting the names of the tests in the test suite: %v", err)
		os.Exit(1)
	}

	// TODO parameterize test name
	testPassed, err := runTest(*testSuiteImageArg, dockerManager, freeIpAddrTracker, "FOO")
	if err != nil {
		logrus.Errorf("An error occurred that prevented retrieving test pass/fail status:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(1)
	}

	if !testPassed {
		// TODO insert test name
		logrus.Error("The test failed")
		os.Exit(1)
	} else {
		// TODO insert test name
		logrus.Info("The test passed")
		os.Exit(0)
	}
}

/*
Spins up a testsuite container in test-listing mode and reads the tests that it spits out
 */
func getTests(testSuiteImage string, dockerManager *docker.DockerManager, freeIpAddrTracker *networks.FreeIpAddrTracker) (map[string]bool, error) {
	// Create the tempfile that the testsuite image will write test names to
	tempFp, err := ioutil.TempFile("", "test-names")
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the temp filepath to write test names to")
	}
	tempFp.Close()

	testListingContainerIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a free IP address for the testsuite container")
	}

	testListingContainerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		testSuiteImage,
		// TODO parameterize these
		bridgeNetworkId,
		testListingContainerIp,
		map[nat.Port]bool{},
		nil,
		map[string]string{
			testNamesFilepathEnvVar: testNamesFilepath,
			testEnvVar:              "", // We leave this blank to signify that we want test listing, not test execution
			kurtosisApiIpEnvVar:     "", // Because we're doing test listing, this can be blank
			testSuiteLogFilepathEnvVar: "/tmp/log.txt",
		},
		map[string]string{
			tempFp.Name(): testNamesFilepath,
		},
		map[string]string{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the test suite container to list the tests")
	}

	testListingExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		testListingContainerId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the testsuite container to list the tests")
	}
	if testListingExitCode != 0 {
		return nil, stacktrace.NewError("The testsuite container for listing tests exited with a nonzero exit code")
	}

	tempFpReader, err := os.Open(tempFp.Name())
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

// TODO make output status  more meaningful
/*
Runs the given test by spinning up test suite & Kurtosis API containers
 */
func runTest(testSuiteImage string, dockerManager *docker.DockerManager, freeIpAddrTracker *networks.FreeIpAddrTracker, test string) (bool, error){
	// TODO segregate these into their own subnet
	kurtosisApiIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}
	testRunningContainerIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the test suite container running the test")
	}

	testRunningContainerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		testSuiteImage,
		// TODO parameterize network stuff
		bridgeNetworkId,
		testRunningContainerIp,
		map[nat.Port]bool{},
		nil,
		map[string]string{
			testNamesFilepathEnvVar: "", // We leave this blank because we want test execution, not listing
			testEnvVar:          test,                  // We leave this blank to signify that we want test listing, not test execution
			kurtosisApiIpEnvVar: kurtosisApiIp.String(), // Because we're doing test listing, this can be blank
			// TODO pipe this to a proper volume location
			testSuiteLogFilepathEnvVar: "/tmp/log.txt",
		},
		map[string]string{},
		map[string]string{})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the test suite container to run the test")
	}

	// TODO replace this with actually running tests
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/tcp", api.KurtosisAPIContainerPort))
	kurtosisApiContainerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		// TODO parameterize these
		kurtosisApiImage,
		bridgeNetworkId,
		kurtosisApiIp,
		map[nat.Port]bool{
			kurtosisApiPort: true,
		},
		nil,
		map[string]string{
			testSuiteContainerIdEnvVar: testRunningContainerId,
			networkIdEnvVar:            bridgeNetworkId,
			subnetMaskEnvVar: bridgeNetworkSubnetMask,
			gatewayIpEnvVar: bridgeNetworkGatewayIp,
			// TODO change this
			logLevelEnvVar:             "trace",
			apiLogFilepathEnvVar: "/tmp/logfile",
		},
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the Kurtosis API container")
	}

	// The Kurtosis API will be our indication of whether the test suite container stopped within the timeout or not
	// TODO add a timeout waiting for Kurtosis API container to stop???
	kurtosisApiExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		kurtosisApiContainerId)
	if err != nil {
		logrus.Errorf("An error occurred waiting for the exit of the Kurtosis API container: %v", err)
	}

	if kurtosisApiExitCode != 0 {
		// TODO change this with tearing down the entire network
		// The testsuite container didn't stop within the tiemout; make a best-effort attempt to stop the testsuite container
		if err := dockerManager.StopContainer(context.Background(), testRunningContainerId, 30 * time.Second); err != nil {
			logrus.Error("An error occurred during our best-effort attempt at stopping the testsuite container which exceeded its test timeout:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		}
		return false, stacktrace.NewError("The testsuite container didn't stop within the hard test timeout")
	}

	// Test stopped within the timeout; examine the testsuite container for actual test result
	testRunningExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		testRunningContainerId)
	if err != nil {
		return false, stacktrace.Propagate(
			err,
			"The testsuite container running the test stopped within the timeout, but an error occurred retrieving the exit code")
	}
	return testRunningExitCode == 0, nil
}
