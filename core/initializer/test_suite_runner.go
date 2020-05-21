package initializer

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/gmarchetti/kurtosis/nodes"
	"os"
)

const DEFAULT_GECKO_HTTP_PORT = nat.Port("9650/tcp")
const DEFAULT_GECKO_STAKING_PORT = nat.Port("9651/tcp")
const LOCAL_HOST_IP = "0.0.0.0"

type TestSuiteRunner struct {
	// TODO later, when we have a proper node config class, get rid of this from here
	imageName string

	// TODO value type needs to be changed to a test config type eventually
	tests map[string]string
}

// TODO Instead of passing in the imageName here, pass it in with the config!
func NewTestSuiteRunner(geckoImage string) *TestSuiteRunner {
	return &TestSuiteRunner{
		imageName: geckoImage,
		tests: make(map[string]string),
	}
}

// TODO implement a RegisterTest function here

// Runs the tests whose names are defined in the given map (the map value is ignored - this is a hacky way to
// do a set implementation)
func (testSuiteRunner TestSuiteRunner) RunTests() () {
	// Initialize default environment context.
	dockerCtx := context.Background()
	// Initialize a Docker client and panic if any error occurs in the process.
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	geckoNode := &nodes.GeckoNode{
		GeckoImageName: testSuiteRunner.imageName,
		HttpPortOnHost: "9650",
		StakingPortOnHost: "9651",
	}


	// Create the container based on the configurations, but don't start it yet.
	fmt.Println("I'm going to run a Gecko node, and hang while it's running! Kill me and then clear your docker containers.")
	nodeConfig := getBasicGeckoNodeConfig(geckoNode.GeckoImageName)
	nodeToHostConfig:= getNodeToHostConfig(geckoNode.HttpPortOnHost, geckoNode.StakingPortOnHost)
	resp, err := dockerClient.ContainerCreate(dockerCtx, nodeConfig, nodeToHostConfig, nil, "")
	containerId := resp.ID
	if err != nil {
		panic(err)
	}
	if err := dockerClient.ContainerStart(dockerCtx, containerId, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// TODO add a timeout here
	waitAndGrabLogsOnError(dockerCtx, dockerClient, containerId)
}

// ======================== Private helper functions =====================================

// Creates a basic Gecko Node Docker container configuration
func getBasicGeckoNodeConfig(nodeImageName string) *container.Config {
	var geckoStartCommand = [5]string{
		"/gecko/build/ava",
		"--public-ip=127.0.0.1",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-tls-enabled=false",
	}
	return getGeckoNodeConfig(nodeImageName, geckoStartCommand)
}

// Creates a more generalized Docker Container configuration for Gecko, with a 5-parameter initialization command.
// Gecko HTTP and Staking ports inside the Container are the standard defaults.
func getGeckoNodeConfig(nodeImageName string, nodeStartCommand [5]string) *container.Config {
	nodeConfig := &container.Config{
		Image: nodeImageName,
		ExposedPorts: nat.PortSet{
			DEFAULT_GECKO_HTTP_PORT: struct{}{},
			DEFAULT_GECKO_STAKING_PORT: struct{}{},
		},
		Cmd:   nodeStartCommand[:len(nodeStartCommand)],
		Tty: false,
	}
	return nodeConfig
}

// Creates a Docker-Container-To-Host Port mapping, defining how a Container's HTTP and Staking ports
// are exposed on Host ports.
func getNodeToHostConfig(hostHttpPort string, hostStakingPort string) *container.HostConfig {
	nodeToHostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			DEFAULT_GECKO_HTTP_PORT: []nat.PortBinding{
				{
					HostIP: LOCAL_HOST_IP,
					HostPort: hostHttpPort,
				},
			},
			DEFAULT_GECKO_STAKING_PORT: []nat.PortBinding{
				{
					HostIP: LOCAL_HOST_IP,
					HostPort: hostStakingPort,
				},
			},
		},
	}
	return nodeToHostConfig
}

func waitAndGrabLogsOnError(dockerCtx context.Context, dockerClient *client.Client, containerId string) {
	statusCh, errCh := dockerClient.ContainerWait(dockerCtx, containerId, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	// If the container stops and there is an error, grab the logs.
	out, err := dockerClient.ContainerLogs(dockerCtx, containerId, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	// Copy the logs to stdout.
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
