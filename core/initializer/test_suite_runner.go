package initializer

import (
	"context"
	"github.com/gmarchetti/kurtosis/commons"
	"os"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/palantir/stacktrace"
)

const START_HOST_PORT_RANGE = 9650
const END_HOST_PORT_RANGE = 9652


type TestSuiteRunner struct {
	tests map[string]commons.TestNetworkConfigProvider
}

func NewTestSuiteRunner() *TestSuiteRunner {
	return &TestSuiteRunner{
		tests: make(map[string]commons.TestNetworkConfigProvider),
	}
}

func (runner TestSuiteRunner) RegisterTest(name string, configProvider commons.TestNetworkConfigProvider) {
	// TODO check if the test already exists and throw an error if so (means an error in the user code)
	runner.tests[name] = configProvider
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

	dockerManager, err := commons.NewDockerManager(dockerCtx, dockerClient, START_HOST_PORT_RANGE, END_HOST_PORT_RANGE)
	if err != nil {
		return stacktrace.Propagate(err, "Error in initializing Docker Manager.")
	}

	// TODO implement parallelism and specific test selection here
	for testName, configProvider := range runner.tests {
		testNetworkCfg := configProvider.GetNetworkConfig()
		serviceNetwork, err := testNetworkCfg.CreateAndRun(dockerManager)
		if err != nil {
			return stacktrace.Propagate(err, "Unable to create network for test '%v'", testName)
		}
		for _, containerId := range serviceNetwork.ServiceContainerIds {
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
	out, err := dockerClient.ContainerLogs(dockerCtx, containerId, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return stacktrace.Propagate(err, "Failed to retrieve container logs.")
	}

	// Copy the logs to stdout.
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}
