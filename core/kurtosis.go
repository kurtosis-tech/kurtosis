package main

import (
	"context"
	"fmt"
	
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")

	fmt.Println("Here are your containers that are currently running:")
	cli, err := client.NewClientWithOpts(client.WithVersion("1.40"))
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
}
