package initializer

import (
	"context"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/kurtosis-tech/kurtosis/commons/testnet"
	"log"
	"os"

	"github.com/palantir/stacktrace"
)


type TestSuiteRunner struct {
	testSuite testsuite.TestSuite
	testImageName string
	testControllerImageName string
	startPortRange int
	endPortRange int
}

const DEFAULT_SUBNET_MASK = "172.23.0.0/16"

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

		volumeName := fmt.Sprintf("%v-%v", testName, testInstanceUuid.String())
		controllerIp, err := publicIpProvider.GetFreeIpAddr()
		if err != nil {
			// TODO we still need to shut down the service network if we get an error here!
			return stacktrace.Propagate(err, "Could not get IP address for controller")
		}

		// TODO allow passing in the path to the controller executable, as it can vary by image
		_, controllerContainerId, err := dockerManager.CreateAndStartControllerContainer(
			runner.testControllerImageName,
			// TODO change this to use FreeIpAddrTracker!!
			controllerIp,
			testName,
			volumeName)
		if err != nil {
			// TODO if we get an error here we still need to shut down the container network!
			return stacktrace.Propagate(err, "Could not start test controller")
		}

		// TODO add a timeout here if the test doesn't complete successfully
		waitAndGrabLogsOnExit(dockerCtx, dockerClient, controllerContainerId)

		// TODO gracefully shut down all the service containers we started
		for _, containerId := range serviceNetwork.ContainerIds {
			log.Printf("Waiting for containerId %v", containerId)
			waitAndGrabLogsOnExit(dockerCtx, dockerClient, containerId)
		}

	}
	return nil
}

// ======================== Private helper functions =====================================


func waitAndGrabLogsOnExit(dockerCtx context.Context, dockerClient *client.Client, containerId string) (err error) {
	statusCh, errCh := dockerClient.ContainerWait(dockerCtx, containerId, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return stacktrace.Propagate(err, "Failed to wait for container to return.")
		}
	case <-statusCh:
	}

	// Grab logs on container quit
	out, err := dockerClient.ContainerLogs(
		dockerCtx,
		containerId,
		types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to retrieve container logs.")
	}

	// Copy the logs to stdout.
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}
