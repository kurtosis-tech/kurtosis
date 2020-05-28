package initializer

import (
	"context"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"os"

	"github.com/palantir/stacktrace"
)


type TestSuiteRunner struct {
	testSuite testsuite.TestSuite
	testImageName string
	subnetMask string
	startPortRange int
	endPortRange int
}

func NewTestSuiteRunner(testSuite testsuite.TestSuite, testImageName string, subnetMask string, startPortRange int, endPortRange int) *TestSuiteRunner {
	return &TestSuiteRunner{
		testSuite: testSuite,
		testImageName: testImageName,
		subnetMask: subnetMask,
		startPortRange: startPortRange,
		endPortRange: endPortRange,
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

	// TODO: Right now all tests run using the same subnetMask. We should have a different subnetMask per test to achieve subnet isolation between tests
	dockerManager, err := docker.NewDockerManager(dockerCtx, dockerClient,  runner.subnetMask, runner.startPortRange, runner.endPortRange)
	if err != nil {
		return stacktrace.Propagate(err, "Error in initializing Docker Manager.")
	}

	tests := runner.testSuite.GetTests()

	// TODO implement parallelism and specific test selection here
	for testName, config := range tests {
		networkLoader := config.NetworkLoader
		testNetworkCfg, err := networkLoader.GetNetworkConfig(runner.testImageName, runner.subnetMask)
		if err != nil {
			stacktrace.Propagate(err, "Unable to get network config from config provider")
		}
		networkName := testName + uuid.Generate().String()
		serviceNetwork, err := testNetworkCfg.CreateAndRun(networkName, dockerManager)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to create network for test '%v'", testName)
		}

		// TODO Actually spin up TestController and run the tests here
		for _, containerId := range serviceNetwork.ContainerIds {
			waitAndGrabLogsOnError(dockerCtx, dockerClient, containerId)
		}
	}

	return nil
	// TODO add a timeout here
	// TODO gracefully shut down all the Docker containers we started here
}

// ======================== Private helper functions =====================================


func waitAndGrabLogsOnError(dockerCtx context.Context, dockerClient *client.Client, containerId string) (err error) {
	statusCh, errCh := dockerClient.ContainerWait(dockerCtx, containerId, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return stacktrace.Propagate(err, "Failed to wait for container to return.")
		}
	case <-statusCh:
	}

	// If the container stops and there is an error, grab the logs.
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
