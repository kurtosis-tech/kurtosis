package initializer

import (
	"context"
	"github.com/docker/distribution/uuid"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gmarchetti/kurtosis/commons/docker"
	"github.com/gmarchetti/kurtosis/commons/testsuite"
	"os"

	"github.com/palantir/stacktrace"
)

const SUBNET_MASK = "172.18.0.0/16"


type TestSuiteRunner struct {
	testSuite testsuite.TestSuite
	testImageName string
	startPortRange int
	endPortRange int
}

func NewTestSuiteRunner(testSuite testsuite.TestSuite, testImageName string, startPortRange int, endPortRange int) *TestSuiteRunner {
	return &TestSuiteRunner{
		testSuite: testSuite,
		testImageName: testImageName,
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

	dockerManager, err := docker.NewDockerManager(dockerCtx, dockerClient, runner.startPortRange, runner.endPortRange)
	if err != nil {
		return stacktrace.Propagate(err, "Error in initializing Docker Manager.")
	}

	tests := runner.testSuite.GetTests()

	// TODO implement parallelism and specific test selection here
	for testName, config := range tests {
		networkLoader := config.NetworkLoader
		testNetworkCfg, err := networkLoader.GetNetworkConfig(runner.testImageName)
		if err != nil {
			stacktrace.Propagate(err, "Unable to get network config from config provider")
		}
		networkName := testName + uuid.Generate().String()
		serviceNetwork, err := testNetworkCfg.CreateAndRun(networkName, SUBNET_MASK, dockerManager)
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
