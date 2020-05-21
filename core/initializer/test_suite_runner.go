package initializer

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/gmarchetti/kurtosis/commons"
	"github.com/gmarchetti/kurtosis/nodes"
	"os"
	"strconv"
)

// TODO replace these with FreeHostPortProvider in the future
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

	// TODO set up non-null nodes (indicating that they're not boot nodes)
	bootNodes := make(map[commons.JsonRpcServiceSocket]commons.JsonRpcRequest)
	geckoNodeConfig := nodes.NewGeckoServiceConfig(testSuiteRunner.imageName, bootNodes)

	// Create the container based on the configurations, but don't start it yet.
	fmt.Println("I'm going to run a Gecko node, and hang while it's running! Kill me and then clear your docker containers.")
	containerConfigPtr := getContainerCfgFromServiceCfg(*geckoNodeConfig)

	containerHostConfigPtr := getContainerHostConfig(geckoNodeConfig)
	resp, err := dockerClient.ContainerCreate(dockerCtx, containerConfigPtr, containerHostConfigPtr, nil, "")
	containerId := resp.ID
	if err != nil {
		panic(err)
	}
	if err := dockerClient.ContainerStart(dockerCtx, containerId, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// TODO add a timeout here
	waitAndGrabLogsOnError(dockerCtx, dockerClient, containerId)

	// TODO gracefully shut down all the Docker containers we started here
}

// ======================== Private helper functions =====================================

// TODO should I actually be passing sorta-complex objects like JsonRpcServiceConfig by value???
// Creates a more generalized Docker Container configuration for Gecko, with a 5-parameter initialization command.
// Gecko HTTP and Staking ports inside the Container are the standard defaults.
func getContainerCfgFromServiceCfg(serviceConfig commons.JsonRpcServiceConfig) *container.Config {
	jsonRpcPort, err := nat.NewPort("tcp", strconv.Itoa(serviceConfig.GetJsonRpcPort()))
	if err != nil {
		panic("Could not parse port int - this is VERY weird")
	}

	portSet := nat.PortSet{
		jsonRpcPort: struct{}{},
	}
	for _, port := range serviceConfig.GetOtherPorts() {
		otherPort, err := nat.NewPort("tcp", strconv.Itoa(port))
		if err != nil {
			panic("Could not parse port int - this is VERY weird")
		}
		portSet[otherPort] = struct{}{}
	}

	nodeConfigPtr := &container.Config{
		Image: serviceConfig.GetDockerImage(),
		// TODO allow modifying of protocol at some point
		ExposedPorts: portSet,
		Cmd: serviceConfig.GetContainerStartCommand(),
		Tty: false,
	}
	return nodeConfigPtr
}

// Creates a Docker-Container-To-Host Port mapping, defining how a Container's JSON RPC and service-specific ports are
// mapped to the host ports
func getContainerHostConfig(serviceConfig commons.JsonRpcServiceConfig) *container.HostConfig {
	// TODO right nwo this is hardcoded - replace these with FreeHostPortProvider in the future, so we can have
	//  arbitrary service-specific ports!
	jsonRpcPortBinding := []nat.PortBinding{
		{
			HostIP: LOCAL_HOST_IP,
			HostPort: strconv.Itoa(serviceConfig.GetJsonRpcPort()),
		},
	}
	// TODO this shouldn't be Ava-specific here
	stakingPortInt := strconv.Itoa(serviceConfig.GetOtherPorts()[nodes.STAKING_PORT_ID])
	stakingPortBinding := []nat.PortBinding{
		{
			HostIP: LOCAL_HOST_IP,
			HostPort: stakingPortInt,
		},
	}

	containerHostConfigPtr := &container.HostConfig{
		PortBindings: nat.PortMap{
			DEFAULT_GECKO_HTTP_PORT: jsonRpcPortBinding,
			DEFAULT_GECKO_STAKING_PORT: stakingPortBinding,
		},
	}
	return containerHostConfigPtr
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
