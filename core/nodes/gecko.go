/*

Contains types to represent nodes contained in Docker containers.

*/

package nodes

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/gmarchetti/kurtosis/utils"
)

type GeckoNode struct {
	GeckoImageName, HttpPortOnHost, StakingPortOnHost string
	respID string
}

func (node *GeckoNode) Create(ctx context.Context, cli *client.Client) {
	nodeConfig := utils.GetBasicGeckoNodeConfig(node.GeckoImageName)
	nodeToHostConfig:= utils.GetNodeToHostConfig(node.HttpPortOnHost, node.StakingPortOnHost)
	resp, err := cli.ContainerCreate(ctx, nodeConfig, nodeToHostConfig, nil, "")
	node.respID = resp.ID
	if err != nil {
		panic(err)
	}
}

func (node *GeckoNode) Run(ctx context.Context, cli *client.Client) {
	if err := cli.ContainerStart(ctx, node.respID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func (node *GeckoNode) GetRespID() string {
	return node.respID
}