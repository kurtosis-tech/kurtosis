/*

Contains helpful functions for managing containers with containing Ava nodes.

*/

package utils

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

func GetNodeConfig(nodeImageName string, nodeStartCommand [5]string) *container.Config {
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

func GetNodeToHostConfig(hostHttpPort string, hostStakingPort string) *container.HostConfig {
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
