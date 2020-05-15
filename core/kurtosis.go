package main

import (
	"context"
	"os"
	"fmt"
	
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

var GECKO_IMAGE_NAME = "gecko-f290f73"
var GECKO_START_COMMAND = [5]string{
	"/gecko/build/ava",
	"--public-ip=127.0.0.1",
	"--snow-sample-size=1",
	"--snow-quorum-size=1",
	"--staking-tls-enabled=false",
}

func getNodeConfig(nodeImageName string, nodeStartCommand [5]string) *container.Config {
	nodeConfig := &container.Config{
		Image: nodeImageName,
		ExposedPorts: nat.PortSet{
			"9650/tcp": struct{}{},
			"9651/tcp": struct{}{},
		},
		Cmd:   nodeStartCommand[:len(nodeStartCommand)],
		Tty: false,
	}
	return nodeConfig
}

func getNodeToHostConfig(hostHttpPort string, hostStakingPort string) *container.HostConfig {
	nodeToHostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"9650/tcp": []nat.PortBinding{
				{
					HostIP: "0.0.0.0",
					HostPort: hostHttpPort,
				},
			},
			"9651/tcp": []nat.PortBinding{
				{
					HostIP: "0.0.0.0",
					HostPort: hostStakingPort,
				},
			},
		},
	}
	return nodeToHostConfig
}

func main() {
	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")

	ctx := context.Background()
	fmt.Println("Here are your containers that are currently running:")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}

	fmt.Println("I'm going to run a Gecko node, and hang while it's running!")

	nodeConfig := getNodeConfig(GECKO_IMAGE_NAME, GECKO_START_COMMAND)
	nodeToHostConfig := getNodeToHostConfig("9650", "9651")

	resp, err := cli.ContainerCreate(ctx, nodeConfig, nodeToHostConfig, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
