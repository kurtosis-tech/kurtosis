package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/gmarchetti/kurtosis/nodes"
)

func main() {
	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")
	
	// Define and parse command line flags.
	geckoImageNameArg := flag.String(
		"gecko-image-name", 
		"gecko-f290f73", // by default, pick commit that was on master May 14, 2020.
		"the name of a pre-built gecko image in your docker engine.",
	)
	flag.Parse()
	
	// Initialize a Docker client and panic if any error occurs in the process.
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	fmt.Println("I'm going to run a Gecko node, and hang while it's running! Kill me and then clear your docker containers.")

	geckoNode := &nodes.GeckoNode{
		GeckoImageName: *geckoImageNameArg,
		HttpPortOnHost: "9650",
		StakingPortOnHost: "9651",
	}

	
	// Create the container based on the configurations, but don't start it yet.
	ctx := context.Background()
	geckoNode.Create(ctx, cli)
	geckoNode.Run(ctx, cli)

	// Wait on the container to return
	statusCh, errCh := cli.ContainerWait(ctx, geckoNode.GetRespID(), container.WaitConditionNotRunning)
	select {
		case err := <-errCh:
			if err != nil {
				panic(err)
			}
		case <-statusCh:
	}

	// Once container has returned, grab the logs.
	out, err := cli.ContainerLogs(ctx, geckoNode.GetRespID(), types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	// Copy the logs to stdout.
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
