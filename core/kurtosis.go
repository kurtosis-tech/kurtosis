package main

import (
	"context"
	"os"
	"fmt"
	
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/gmarchetti/kurtosis/utils"
)

const GECKO_IMAGE_NAME = "gecko-f290f73"

func main() {
	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")

	ctx := context.Background()
	
	// Initialize a Docker client and panic if any error occurs in the process.
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	fmt.Println("I'm going to run a Gecko node, and hang while it's running! Kill me and then clear your docker containers.")

	// Creates a configuration object representing the container itself, based on a prebuilt image.
	nodeConfig := utils.GetGeckoNodeConfig(GECKO_IMAGE_NAME)

	// Creates a configuration object representing the mappings between the container and the host.
	nodeToHostConfig := utils.GetNodeToHostConfig("9650", "9651")

	// Create the container based on the configurations, but don't start it yet.
	resp, err := cli.ContainerCreate(ctx, nodeConfig, nodeToHostConfig, nil, "")
	if err != nil {
		panic(err)
	}

	// Start the container.
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// Wait on the container to return
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
		case err := <-errCh:
			if err != nil {
				panic(err)
			}
		case <-statusCh:
	}

	// Once container has returned, grab the logs.
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	// Copy the logs to stdout.
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
