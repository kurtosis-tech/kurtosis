/*

Contains types to represent nodes contained in Docker containers.

*/

package ava_commons

import (
	"fmt"
	"github.com/gmarchetti/kurtosis/commons"
	"strings"
)

// Type representing a Gecko Node and which ports on the host machine it will use for HTTP and Staking.
type GeckoServiceConfig struct {
	geckoImageName string
}

const (
	STAKING_PORT_ID commons.ServiceSpecificPort = 0

	HTTP_PORT = 9650
	STAKING_PORT = 9651
)

// TODO implement more Ava-specific params here, like snow quorum
func NewGeckoServiceConfig(dockerImage string) *GeckoServiceConfig {
	return &GeckoServiceConfig{
		geckoImageName: dockerImage,
	}
}

func (g GeckoServiceConfig) GetDockerImage() string {
	return g.geckoImageName
}

func (g GeckoServiceConfig) GetJsonRpcPort() int {
	return HTTP_PORT
}

func (g GeckoServiceConfig) GetOtherPorts() map[commons.ServiceSpecificPort]int {
	result := make(map[commons.ServiceSpecificPort]int)
	result[STAKING_PORT_ID] = STAKING_PORT
	return result
}

// Argument will be a map of (IP,port) -> request to make to check if a node is up
func (g GeckoServiceConfig) GetContainerStartCommand(dependencyLivenessReqs map[commons.JsonRpcServiceSocket]commons.JsonRpcRequest) []string {
	commandList := []string{
		"/gecko/build/ava",
		"--public-ip=127.0.0.1",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-tls-enabled=false",
	}

	// If bootstrap nodes are down then Gecko will wait until they are, so we don't actually need to busy-loop making
	// requests to the nodes
	if dependencyLivenessReqs != nil && len(dependencyLivenessReqs) > 0 {
		socketStrs := make([]string, 0, len(dependencyLivenessReqs))
		for socket, _ := range dependencyLivenessReqs {
			socketStrs = append(socketStrs, fmt.Sprintf("%s:%d", socket.IPAddress, socket.Port))
		}
		joinedSockets := strings.Join(socketStrs, ",")
		commandList = append(commandList, "--bootstrap-ips=" + joinedSockets)
	}
	return commandList
}

func (g GeckoServiceConfig) GetLivenessRequest() commons.JsonRpcRequest {
	return commons.JsonRpcRequest{
		Endpoint: "/ext/P",
		Method: "platform.getCurrentValidators",
		RpcVersion: commons.RPC_VERSION_1_0,
		Params: make(map[string]string),
		ID: 1,   // Not really sure if we'd ever need to change this
	}
}

