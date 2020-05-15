/*

Contains types to represent nodes contained in Docker containers.

*/

package nodes

import (
	"context"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"


	"github.com/gmarchetti/kurtosis/utils"
)


// Type representing a Gecko Node and which ports on the host machine it will use for HTTP and Staking.
type GeckoNode struct {
	GeckoImageName, HttpPortOnHost, StakingPortOnHost string
	Context context.Context
	Client *client.Client
	respID string
}

func (node *GeckoNode) GetRespID() string {
	return node.respID
}

// Creates a Docker container for the Gecko Node.
func (node *GeckoNode) Create() {
	nodeConfig := utils.GetBasicGeckoNodeConfig(node.GeckoImageName)
	nodeToHostConfig:= utils.GetNodeToHostConfig(node.HttpPortOnHost, node.StakingPortOnHost)
	resp, err := node.Client.ContainerCreate(node.Context, nodeConfig, nodeToHostConfig, nil, "")
	node.respID = resp.ID
	if err != nil {
		panic(err)
	}
}

// Runs the Docker container with default options.
func (node *GeckoNode) Run() {
	if err := node.Client.ContainerStart(node.Context, node.respID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

// Waits on the Docker container, and if it exits, exposes logs to stdout.
func (node *GeckoNode) WaitAndGrabLogsOnError() {
	statusCh, errCh := node.Client.ContainerWait(node.Context, node.respID, container.WaitConditionNotRunning)

	select {
		case err := <-errCh:
			if err != nil {
				panic(err)
			}
		case <-statusCh:
	}

	// If the container stops and there is an error, grab the logs.
	out, err := node.Client.ContainerLogs(node.Context, node.respID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	// Copy the logs to stdout.
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}