/*

Contains helpful functions for managing containers with containing Ava nodes.

*/

package utils

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

const DEFAULT_GECKO_HTTP_PORT = nat.Port("9650/tcp")
const DEFAULT_GECKO_STAKING_PORT = nat.Port("9651/tcp")
const LOCAL_HOST_IP = "0.0.0.0"

func GetGeckoNodeConfig(nodeImageName string) *container.Config {
	var geckoStartCommand = [5]string{
		"/gecko/build/ava",
		"--public-ip=127.0.0.1",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-tls-enabled=false",
	}
	return GetNodeConfig(nodeImageName, geckoStartCommand)
}

func GetNodeConfig(nodeImageName string, nodeStartCommand [5]string) *container.Config {
	nodeConfig := &container.Config{
		Image: nodeImageName,
		ExposedPorts: nat.PortSet{
			DEFAULT_GECKO_HTTP_PORT: struct{}{},
			DEFAULT_GECKO_STAKING_PORT: struct{}{},
		},
		Cmd:   nodeStartCommand[:len(nodeStartCommand)],
		Tty: false,
	}
	return nodeConfig
}

func GetNodeToHostConfig(hostHttpPort string, hostStakingPort string) *container.HostConfig {
	nodeToHostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			DEFAULT_GECKO_HTTP_PORT: []nat.PortBinding{
				{
					HostIP: LOCAL_HOST_IP,
					HostPort: hostHttpPort,
				},
			},
			DEFAULT_GECKO_STAKING_PORT: []nat.PortBinding{
				{
					HostIP: LOCAL_HOST_IP,
					HostPort: hostStakingPort,
				},
			},
		},
	}
	return nodeToHostConfig
}
