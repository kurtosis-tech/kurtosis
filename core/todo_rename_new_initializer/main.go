package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/api_container/api"
	"github.com/kurtosis-tech/kurtosis/api_container/execution/exit_codes"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"net"
	"os"
	"strings"
	"time"
)

const (
	successExitCode = 0
	failureExitCode = 1

	testNamesFilepath = "/shared/test-names-filepath"
	dockerSocket = "/var/run/docker.sock"
	testNameArgSeparator = ","

	// This is the IP address that the first Docker subnet will be doled out from, with subsequent Docker networks doled out with
	//  increasing IPs corresponding to the NETWORK_WIDTH_BITS
	subnetStartAddr = "172.23.0.0"
	betsInIPv4Addr  = 32

	bridgeNetworkName = "bridge"

	// When we're tearing down the network after running a test, how long we'll give each contaienr to gracefully stop itself
	//  before hard-killing it
	networkTeardownContainerStopTimeout = 30 * time.Second

	testListingContainerDescription = "Test-Listing Container"
	testRunningContainerDescription = "Test-Running Container"

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
	apiContainerIpAddrEnvVar   = "API_CONTAINER_IP"
	testSuiteContainerIpAddrEnvVar   = "TEST_SUITE_CONTAINER_IP"

	// TODO Parameterize this! (will require getting information from testsuite)
	defaultNetworkWidthBits = 8

	testVolumeMountDirpath = "/test-volume" // TODO parameterize this!!
	bindMountsDirpath = "/bind-mounts"

	testSuiteLogFilepath        = bindMountsDirpath + "/test-listing.log"

	// TODO parameterize
	kurtosisApiImage        = "kurtosistech/kurtosis-core_api"
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

	subnetStartIpInt, err := getSubnetStartIpInt()
	if err != nil {
		logrus.Errorf("An error getting the subnet start IP integer:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}

	executionId := uuid.New()

	// TODO sort these test names so we're doing things deterministically
	// TODO Parallelize these bitches
	allTestsPassed := true
	testIndex := 0
	for testName, _ := range testNamesToRun {
		testPassed, testExecutionErr := runSingleTest(
			executionId,
			*testSuiteImageArg,
			dockerManager,
			testName,
			subnetStartIpInt,
			testIndex)
		if testExecutionErr != nil {
			logrus.Errorf("An error occurred that prevented retrieving test pass/fail status for test '%v':", testName)
			fmt.Fprintln(logrus.StandardLogger().Out, testExecutionErr)
			logrus.Errorf("%v: ERRORED", testName)
		} else {
			if testPassed {
				logrus.Infof("%v: PASSED", testName)
			} else {
				logrus.Errorf("%v: FAILED", testName)
			}
		}
		allTestsPassed = allTestsPassed && (testExecutionErr == nil && testPassed)
		testIndex++
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
Spins up a testsuite container in test-listing mode and returns the "set" of tests that it spits out
 */
func getAllTestNamesInSuite(
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
			testNamesFilepathEnvVar: testNamesFilepath,
			testEnvVar:              "", // We leave this blank to signify that we want test listing, not test execution
			kurtosisApiIpEnvVar:     "", // Because we're doing test listing, this can be blank
			testSuiteLogFilepathEnvVar: testSuiteLogFilepath,
		},
		map[string]string{
			testNamesFp.Name():    testNamesFilepath,
			containerLogFp.Name(): testSuiteLogFilepath,
		},
		map[string]string{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the test suite container to list the tests")
	}

	testListingExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		testListingContainerId)
	if err != nil {
		printContainerLogs(testListingContainerDescription, containerLogFp.Name())
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the exit of the testsuite container to list the tests")
	}
	if testListingExitCode != 0 {
		printContainerLogs(testListingContainerDescription, containerLogFp.Name())
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

/*
Little helper function to print the file with banners indicating the name of the test suite container

Args:
	containerDescription: Short, human-readable description of the container whose logs are being printed
	logFilepath: Filepath of the file containing the container's logs
 */
func printContainerLogs(containerDescription string, logFilepath string) {
	containerDescUppercase := strings.ToUpper(containerDescription)
	logrus.Info("- - - - - - - - - - - - - " + containerDescUppercase + " LOGS - - - - - - - - - - - - -")
	fp, err := os.Open(logFilepath)
	if err != nil {
		logrus.Errorf("Could not print the test suite container's logs due to the following error when opening the file:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
	}
	defer fp.Close()
	if _, err := io.Copy(logrus.StandardLogger().Out, fp); err != nil {
		logrus.Errorf("Could not print the test suite container's logs due to the following error when copying logfile contents:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
	}
	logrus.Info("- - - - - - - - - - - - " + containerDescUppercase + " LOGS - - - - - - - - - - - - -")
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

	allTestNames, err := getAllTestNamesInSuite(testSuiteImage, dockerManager)
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

/*
Runs a single test with the given name
 */
func runSingleTest(
		executionId uuid.UUID,
		testSuiteImage string,
		dockerManager *docker.DockerManager,
		testName string,
		subnetStartIpInt uint32,
		testIndex int) (bool, error){

	// TODO pull netwokrwidthbits from the testsuite
	subnetMask, err := getTestSubnetMask(subnetStartIpInt, defaultNetworkWidthBits, testIndex)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting subnet mask for test '%v':", testName)
	}

	uniqueTestIdentifier := fmt.Sprintf("%v-%v", executionId.String(), testName)

	volumeName := uniqueTestIdentifier
	logrus.Debugf("Creating Docker volume %v which will be shared with the test network...", volumeName)
	if err := dockerManager.CreateVolume(context.Background(), volumeName); err != nil {
		return false, stacktrace.Propagate(err, "Error creating Docker volume to share amongst test nodes for test %v", testName)
	}
	logrus.Debugf("Docker volume %v created successfully", volumeName)

	logrus.Infof("Creating Docker network for test with subnet mask %v...", subnetMask)
	freeIpAddrTracker, err := networks.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		subnetMask,
		map[string]bool{})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker for test %v", testName)
	}
	gatewayIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting a free IP for the gateway for test %v", testName)
	}
	networkName := fmt.Sprintf("%v-%v", executionId.String(), testName)
	// TODO different context for parallelization?
	networkId, err := dockerManager.CreateNetwork(context.Background(), networkName, subnetMask, gatewayIp)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error occurred creating Docker network %v for test %v", networkName, testName)
	}
	defer removeNetworkDeferredFunc(dockerManager, networkId)
	logrus.Infof("Docker network %v created successfully", networkId)

	// TODO use hostnames rather than IPs, which makes things nicer and which we'll need for Docker swarm support
	// We need to create the IP addresses for BOTH containers because the testsuite needs to know the IP of the API
	//  container which will only be started after the testsuite container
	kurtosisApiIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the Kurtosis API container")
	}
	testRunningContainerIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting an IP for the test suite container running the test")
	}

	logrus.Debugf(
		"Test suite container IP: %v; kurtosis API container IP: %v",
		testRunningContainerIp.String(),
		kurtosisApiIp.String())

	// TODO When we have the initializer run in a Docker container, transition to using Docker volumes to store logs
	containerLogFp, err := ioutil.TempFile("", "test-execution.log")
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the temporary file for holding the " +
			"logs of the test suite container running the test")
	}
	containerLogFp.Close()

	logrus.Infof("Creating test suite container that will run the test...")
	testRunningContainerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		testSuiteImage,
		// TODO parameterize network stuff
		networkId,
		testRunningContainerIp,
		map[nat.Port]bool{},
		nil,
		map[string]string{
			testNamesFilepathEnvVar:    "", // We leave this blank because we want test execution, not listing
			testEnvVar:                 testName,                  // We leave this blank to signify that we want test listing, not test execution
			kurtosisApiIpEnvVar:        kurtosisApiIp.String(), // Because we're doing test listing, this can be blank
			testSuiteLogFilepathEnvVar: testSuiteLogFilepath,
		},
		map[string]string{
			containerLogFp.Name(): testSuiteLogFilepath,
		},
		map[string]string{
			volumeName: testVolumeMountDirpath,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the test suite container to run the test")
	}
	logrus.Info("Successfully created test suite container to run the test")

	logrus.Info("Creating Kurtosis API container...")
	kurtosisApiPort := nat.Port(fmt.Sprintf("%v/tcp", api.KurtosisAPIContainerPort))
	kurtosisApiContainerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		// TODO parameterize these
		kurtosisApiImage,
		networkId,
		kurtosisApiIp,
		map[nat.Port]bool{
			kurtosisApiPort: true,
		},
		nil,
		map[string]string{
			testSuiteContainerIdEnvVar: testRunningContainerId,
			// TODO Change all of these
			networkIdEnvVar: networkId,
			subnetMaskEnvVar: subnetMask,
			gatewayIpEnvVar: gatewayIp.String(),
			// TODO make this parameterizable
			logLevelEnvVar:                 "trace",
			// NOTE: We set this to some random file inside the volume because we don't expect it to be read
			apiLogFilepathEnvVar:           testVolumeMountDirpath + "/api.log",
			apiContainerIpAddrEnvVar:       kurtosisApiIp.String(),
			testSuiteContainerIpAddrEnvVar: testRunningContainerIp.String(),
		},
		map[string]string{
			dockerSocket: dockerSocket,
		},
		map[string]string{
			volumeName: testVolumeMountDirpath,
		})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred creating the Kurtosis API container")
	}
	logrus.Infof("Successfully created Kurtosis API container")

	// TODO add a timeout waiting for Kurtosis API container to stop???
	// The Kurtosis API will be our indication of whether the test suite container stopped within the timeout or not
	logrus.Info("Waiting for Kurtosis API container to exit...")
	kurtosisApiExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		kurtosisApiContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred waiting for the exit of the Kurtosis API container: %v", err)
	}

	var testStatusRetrievalError error
	switch kurtosisApiExitCode {
	case exit_codes.TestCompletedInTimeoutExitCode:
		testStatusRetrievalError = nil
		// TODO this is in a really crappy spot; move it
		printContainerLogs(testRunningContainerDescription, containerLogFp.Name())
	case exit_codes.OutOfOrderTestStatusExitCode:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container received an out-of-order " +
			"test execution status update; this is a Kurtosis code bug")
	case exit_codes.TestHitTimeoutExitCode:
		testStatusRetrievalError = stacktrace.NewError("The test failed to complete within the hard test " +
			"timeout (test_execution_timeout + setup_buffer)")
	case exit_codes.NoTestSuiteRegisteredExitCode:
		testStatusRetrievalError = stacktrace.NewError("The test suite failed to register itself with the " +
			"Kurtosis API container; this is a bug with the test suite")
		// TODO this is in a really crappy spot; move it
		printContainerLogs(testRunningContainerDescription, containerLogFp.Name())
	case exit_codes.ShutdownSignalExitCode:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container exited due to receiving " +
			"a shutdown signal; if this is not expected, it's a Kurtosis bug")
	default:
		testStatusRetrievalError = stacktrace.NewError("The Kurtosis API container exited with an unrecognized " +
			"exit code %v; this is a code bug in Kurtosis indicating that the new exit code needs to be mapped " +
			"for the initializer", kurtosisApiExitCode)
	}
	if testStatusRetrievalError != nil {
		logrus.Error("An error occurred that prevented retrieval of the test completion status")
		return false, testStatusRetrievalError
	}
	logrus.Info("The test suite container running the test exited within the hard test timeout")

	// The test suite container will already have stopped, so now we get the exit code
	testSuiteExitCode, err := dockerManager.WaitForExit(
		context.Background(),
		testRunningContainerId)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred retrieving the test suite container exit code")
	}
	return testSuiteExitCode == 0, nil
}

/*
Translates the subnet start address into a uint32 representing the IP address
 */
func getSubnetStartIpInt() (uint32, error) {
	subnetStartIp := net.ParseIP(subnetStartAddr)
	if subnetStartIp == nil {
		return 0, stacktrace.NewError("Subnet start IP %v was not a valid IP address; this is a code problem", subnetStartAddr)
	}

	// The IP can be either 4 bytes or 16 bytes long; we need to handle both
	//  else we'll get a silent 0 value for the int!
	// See https://gist.github.com/ammario/649d4c0da650162efd404af23e25b86b
	var subnetStartIpInt uint32
	if len(subnetStartIp) == 16 {
		subnetStartIpInt = binary.BigEndian.Uint32(subnetStartIp[12:16])
	} else {
		subnetStartIpInt = binary.BigEndian.Uint32(subnetStartIp)
	}
	return subnetStartIpInt, nil
}

/*
Gets the subnet mask that Docker should use for a given test, with each network offset by testIndex * networkWidthBits
 */
func getTestSubnetMask(subnetStartIpInt uint32, networkWidthBits uint32, testIndex int) (string, error) {
	subnetMaskBits := betsInIPv4Addr - networkWidthBits

	// Pick the next free available subnet IP, considering all the tests we've started previously
	subnetIpInt := subnetStartIpInt + uint32(testIndex) * uint32(math.Pow(2, float64(networkWidthBits)))
	subnetIp := make(net.IP, 4)
	binary.BigEndian.PutUint32(subnetIp, subnetIpInt)
	subnetCidrStr := fmt.Sprintf("%v/%v", subnetIp.String(), subnetMaskBits)

	return subnetCidrStr, nil

}

/*
Helper function for making a best-effort attempt at removing a network and logging any error states; intended to be run
as a deferred function.
*/
func removeNetworkDeferredFunc(dockerManager *docker.DockerManager, networkId string) {
	logrus.Infof("Attempting to remove Docker network with id %v...", networkId)
	// We use the background context here because we want to try and tear down the network even if the context the test was running in
	//  was cancelled. This might not be right - the right way to do it might be to pipe a separate context for the network teardown to here!
	if err := dockerManager.RemoveNetwork(context.Background(), networkId, networkTeardownContainerStopTimeout); err != nil {
		logrus.Errorf("An error occurred removing Docker network with ID %v:", networkId)
		logrus.Error(err.Error())
		logrus.Error("NOTE: This means you will need to clean up the Docker network manually!!")
	} else {
		logrus.Infof("Docker network with ID %v successfully removed", networkId)
	}
}
