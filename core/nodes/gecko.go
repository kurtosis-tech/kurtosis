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

type GeckoNode struct {
	GeckoImageName, HttpPortOnHost, StakingPortOnHost string
	Context context.Context
	Client *client.Client
	respID string
}

func (node *GeckoNode) GetRespID() string {
	return node.respID
}

func (node *GeckoNode) Create() {
	nodeConfig := utils.GetBasicGeckoNodeConfig(node.GeckoImageName)
	nodeToHostConfig:= utils.GetNodeToHostConfig(node.HttpPortOnHost, node.StakingPortOnHost)
	resp, err := node.Client.ContainerCreate(node.Context, nodeConfig, nodeToHostConfig, nil, "")
	node.respID = resp.ID
	if err != nil {
		panic(err)
	}
}

func (node *GeckoNode) Run() {
	if err := node.Client.ContainerStart(node.Context, node.respID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

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