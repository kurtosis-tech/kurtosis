package initializer

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
)


type TestSuiteRunner struct {
	testSuite               testsuite.TestSuite
	testServiceImageName    string
	testControllerImageName string
	startPortRange          int
	endPortRange            int
}

const (
	DEFAULT_SUBNET_MASK = "172.23.0.0/16"

	CONTAINER_NETWORK_INFO_MOUNTED_FILEPATH = "/data/network/network-info"

	// These are an "API" of sorts - environment variables that are agreed to be set in the test controller's Docker environment
	TEST_NAME_BASH_ARG = "TEST_NAME"
	NETWORK_FILEPATH_ARG = "NETWORK_DATA_FILEPATH"

	// How long to wait before force-killing a container
	CONTAINER_STOP_TIMEOUT = 30 * time.Second
)


func NewTestSuiteRunner(
			testSuite testsuite.TestSuite,
			testServiceImageName string,
			testControllerImageName string,
			startPortRange int,
			endPortRange int) *TestSuiteRunner {
	return &TestSuiteRunner{
		testSuite:               testSuite,
		testServiceImageName:    testServiceImageName,
		testControllerImageName: testControllerImageName,
		startPortRange:          startPortRange,
		endPortRange:            endPortRange,
	}
}

/*
Runs the tests with the given names. If no tests are specifically defined, all tests are run.
 */
func (runner TestSuiteRunner) RunTests(testNamesToRun []string) (err error) {
	// Initialize default environment context.
	dockerCtx := context.Background()
	// Initialize a Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err,"Failed to initialize Docker client from environment.")
	}

	dockerManager, err := docker.NewDockerManager(dockerCtx, dockerClient, runner.startPortRange, runner.endPortRange)
	if err != nil {
		return stacktrace.Propagate(err, "Error in initializing Docker Manager.")
	}

	allTests := runner.testSuite.GetTests()
	if len(testNamesToRun) == 0 {
		testNamesToRun = make([]string, 0, len(runner.testSuite.GetTests()))
		for testName, _ := range allTests {
			testNamesToRun = append(testNamesToRun, testName)
		}
	}

	// Validate all the requested tests exist
	testsToRun := make(map[string]testsuite.Test)
	for _, testName := range testNamesToRun {
		test, found := allTests[testName]
		if !found {
			return stacktrace.NewError("No test registered with name '%v'", testName)
		}
		testsToRun[testName] = test
	}


	executionInstanceId := uuid.Generate()
	logrus.Infof("Running %v tests with execution ID: %v", len(testsToRun), executionInstanceId.String())

	// TODO TODO TODO Support creating one network per testnet
	_, err = dockerManager.CreateNetwork(DEFAULT_SUBNET_MASK)
	if err != nil {
		return stacktrace.Propagate(err, "Error in creating docker subnet for testnet.")
	}

	// TODO implement parallelism
	// TODO implement capturing test results
	for testName, test := range testsToRun {
		logrus.Infof("Running test: %v", testName)
		networkLoader := test.GetNetworkLoader()

		builder := networks.NewServiceNetworkConfigBuilder()
		if err := networkLoader.ConfigureNetwork(builder); err != nil {
			return stacktrace.Propagate(err, "Unable to configure test network")
		}
		testNetworkCfg := builder.Build()

		logrus.Infof("Creating network for test...")
		publicIpProvider, err := networks.NewFreeIpAddrTracker(DEFAULT_SUBNET_MASK)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		serviceNetwork, err := testNetworkCfg.CreateNetwork(runner.testServiceImageName, publicIpProvider, dockerManager)
		if err != nil {
			stopNetwork(dockerManager, serviceNetwork, CONTAINER_STOP_TIMEOUT)
			return stacktrace.Propagate(err, "Unable to create network for test '%v'", testName)
		}
		logrus.Info("Network created successfully")

		err = runControllerContainer(
			dockerManager,
			serviceNetwork,
			runner.testControllerImageName,
			publicIpProvider,
			testName,
			executionInstanceId)
		if err != nil {
			stopNetwork(dockerManager, serviceNetwork, CONTAINER_STOP_TIMEOUT)
			return stacktrace.Propagate(err, "An error occurred running the test controller")
		}
		stopNetwork(dockerManager, serviceNetwork, CONTAINER_STOP_TIMEOUT)

		// TODO after the service containers have been changed to write logs to disk, print each container's logs here for convenience
	}
	return nil
}

// ======================== Private helper functions =====================================



func runControllerContainer(
		manager *docker.DockerManager,
		rawServiceNetwork *networks.RawServiceNetwork,
		dockerImage string,
		ipProvider *networks.FreeIpAddrTracker,
		testName string,
		testInstanceUuid uuid.UUID) (err error){
	logrus.Info("Launching controller container...")

	// Using tempfiles, is there a risk that for a verrrry long-running E2E test suite the filesystem cleans up the tempfile
	//  out from underneath the test??
	testNetworkInfoFilename := fmt.Sprintf("%v-%v", testName, testInstanceUuid.String())
	tmpfile, err := ioutil.TempFile("", testNetworkInfoFilename)
	if err != nil {
		return stacktrace.Propagate(err, "Could not create tempfile to store network info for passing to test controller")
	}

	logrus.Debugf("Temp filepath to write network file to: %v", tmpfile.Name())
	logrus.Debugf("Raw service network contents: %v", rawServiceNetwork)

	encoder := gob.NewEncoder(tmpfile)
	err = encoder.Encode(rawServiceNetwork)
	if err != nil {
		return stacktrace.Propagate(err, "Could not write service network state to file")
	}
	// Apparently, per https://www.joeshaw.org/dont-defer-close-on-writable-files/ , file.Close() can return errors too,
	//  but because this is a tmpfile we don't fuss about checking them
	defer tmpfile.Close()

	containerMountpoint := CONTAINER_NETWORK_INFO_MOUNTED_FILEPATH
	envVariables := map[string]string{
		TEST_NAME_BASH_ARG: testName,
		// TODO just for testing; replace with a dynamic filename
		NETWORK_FILEPATH_ARG: containerMountpoint,
	}

	ipAddr, err := ipProvider.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get free IP address to assign the test controller")
	}

	_, controllerContainerId, err := manager.CreateAndStartContainer(
		dockerImage,
		ipAddr,
		make(map[int]bool),
		nil, // Use the default image CMD (which is parameterized)
		envVariables,
		map[string]string{
			tmpfile.Name(): containerMountpoint,
		})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to run test controller container")
	}
	logrus.Info("Controller container started successfully")

	logrus.Info("Waiting for controller container to exit...")
	// TODO add a timeout here if the test doesn't complete successfully
	err = manager.WaitForExit(controllerContainerId)
	if err != nil {
		return stacktrace.Propagate(err, "Failed when waiting for controller to exit")
	}
	logrus.Info("Controller container exited")

	return nil
}

/*
Makes a best-effort attempt to stop the containers in the given network, waiting for the given timeout.
It is safe to pass in nil to the networkToStop
 */
func stopNetwork(manager *docker.DockerManager, networkToStop *networks.RawServiceNetwork, stopTimeout time.Duration) {
	logrus.Info("Stopping service container network...")
	for _, containerId := range networkToStop.ContainerIds {
		logrus.Debugf("Stopping container with ID '%v'", containerId)
		err := manager.StopContainer(containerId, stopTimeout)
		if err != nil {
			logrus.Warnf("An error occurred stopping container ID '%v'; continuing on with stopping other containers...", containerId)
			logrus.Warn(err)
		}
		logrus.Debugf("Container with ID '%v' successfully stopped", containerId)
	}
	logrus.Info("Finished stopping service container network")
}
