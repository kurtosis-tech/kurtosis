package main

import (
	"context"
	"flag"
	"fmt"
	
	"github.com/docker/docker/client"

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
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	geckoNode := &nodes.GeckoNode{
		GeckoImageName: *geckoImageNameArg,
		HttpPortOnHost: "9650",
		StakingPortOnHost: "9651",
		Context: ctx,
		Client: cli,
	}

	
	// Create the container based on the configurations, but don't start it yet.
	fmt.Println("I'm going to run a Gecko node, and hang while it's running! Kill me and then clear your docker containers.")
	geckoNode.Create()
	geckoNode.Run()
	geckoNode.WaitAndGrabLogsOnError()
}
