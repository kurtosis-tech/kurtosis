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

// Creates a basic Gecko Node Docker container configuration
func GetBasicGeckoNodeConfig(nodeImageName string) *container.Config {
	var geckoStartCommand = [5]string{
		"/gecko/build/ava",
		"--public-ip=127.0.0.1",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-tls-enabled=false",
	}
	return getGeckoNodeConfig(nodeImageName, geckoStartCommand)
}

// Creates a more generalized Docker Container configuration for Gecko, with a 5-parameter initialization command.
// Gecko HTTP and Staking ports inside the Container are the standard defaults.
func getGeckoNodeConfig(nodeImageName string, nodeStartCommand [5]string) *container.Config {
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

// Creates a Docker-Container-To-Host Port mapping, defining how a Container's HTTP and Staking ports
// are exposed on Host ports.
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
