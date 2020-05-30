package initializer

import (
	"context"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/testnet"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)


type TestSuiteRunner struct {
	testSuite testsuite.TestSuite
	testImageName string
	testControllerImageName string
	startPortRange int
	endPortRange int
}

const (
	DEFAULT_SUBNET_MASK = "172.23.0.0/16"

	CONTAINER_NETWORK_INFO_VOLUME_MOUNTPATH = "/data/network"

	// These are an "API" of sorts - environment variables that are agreed to be set in the test controller's Docker environment
	TEST_NAME_BASH_ARG = "TEST_NAME"
	NETWORK_FILEPATH_ARG = "NETWORK_DATA_FILEPATH"
)


func NewTestSuiteRunner(
			testSuite testsuite.TestSuite,
			testImageName string,
			testControllerImageName string,
			startPortRange int,
			endPortRange int) *TestSuiteRunner {
	return &TestSuiteRunner{
		testSuite:               testSuite,
		testImageName:           testImageName,
		testControllerImageName: testControllerImageName,
		startPortRange:          startPortRange,
		endPortRange:            endPortRange,
	}
}

// Runs the tests whose names are defined in the given map (the map value is ignored - this is a hacky way to
// do a set implementation)
func (runner TestSuiteRunner) RunTests() (err error) {
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

	tests := runner.testSuite.GetTests()

	// TODO TODO TODO Support creating one network per testnet
	_, err = dockerManager.CreateNetwork(DEFAULT_SUBNET_MASK)
	if err != nil {
		return stacktrace.Propagate(err, "Error in creating docker subnet for testnet.")
	}

	// TODO implement parallelism and specific test selection here
	for testName, config := range tests {
		networkLoader := config.NetworkLoader
		testNetworkCfg, err := networkLoader.GetNetworkConfig(runner.testImageName)
		if err != nil {
			stacktrace.Propagate(err, "Unable to get network config from config provider")
		}

		testInstanceUuid := uuid.Generate()
		// TODO push the network name generation lower??
		networkName := fmt.Sprintf("%v-%v", testName, testInstanceUuid.String())
		publicIpProvider, err := testnet.NewFreeIpAddrTracker(networkName, DEFAULT_SUBNET_MASK)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
		serviceNetwork, err := testNetworkCfg.CreateAndRun(publicIpProvider, dockerManager)
		// TODO if we get an err back, we need to shut down whatever containers were started
		if err != nil {
			return stacktrace.Propagate(err, "Unable to create network for test '%v'", testName)
		}

		err = runControllerContainer(dockerManager, runner.testControllerImageName, publicIpProvider, testName, testInstanceUuid)
		if err != nil {
			// TODO we need to clean up the Docker network still!
			return stacktrace.Propagate(err, "An error occurred running the test controller")
		}


		// TODO gracefully shut down all the service containers we started
		for _, containerId := range serviceNetwork.ContainerIds {
			logrus.Infof("Waiting for containerId %v", containerId)
			dockerManager.WaitForExit(containerId)
			// TODO after the service containers have been changed to write logs to disk, print each container's logs here for convenience
		}

	}
	return nil
}

// ======================== Private helper functions =====================================



func runControllerContainer(
		manager *docker.DockerManager,
		dockerImage string,
		ipProvider *testnet.FreeIpAddrTracker,
		testName string,
		testInstanceUuid uuid.UUID) (err error){

	// Using tempfiles, is there a risk that for a verrrry long-running E2E test suite the filesystem cleans up the tempfile
	//  out from underneath the test??
	networkFilename := fmt.Sprintf("%v-%v", testName, testInstanceUuid.String())
	tmpfile, err := ioutil.TempFile("", networkFilename)
	if err != nil {
		return stacktrace.Propagate(err, "Could not create tempfile to store network info for passing to test controller")
	}

	// TODO debugging
	logrus.Debugf("Tempfile filepath: %v", tmpfile.Name())

	// TODO just for testing
	err = ioutil.WriteFile(tmpfile.Name(), []byte("JSON data would go here"), 0644)
	if err != nil {
		return stacktrace.Propagate(err, "Could not write data to testing file")
	}
	containerMountpoint := CONTAINER_NETWORK_INFO_VOLUME_MOUNTPATH + "/testing.txt"


	envVariables := map[string]string{
		TEST_NAME_BASH_ARG: testName,
		// TODO just for testing
		NETWORK_FILEPATH_ARG: containerMountpoint,
	}

	ipAddr, err := ipProvider.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get free IP address to assign the test controller")
	}

	_, controllerContainerId, err := manager.CreateAndStartContainer(
		dockerImage,
		true,
		ipAddr,
		make(map[int]bool),
		nil,
		envVariables,
		map[string]string{
			tmpfile.Name(): containerMountpoint,
		})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to run test controller container")
	}

	// TODO add a timeout here if the test doesn't complete successfully
	err = manager.WaitForExit(controllerContainerId)
	if err != nil {
		return stacktrace.Propagate(err, "Failed when waiting for controller to exit")
	}

	return nil
}
