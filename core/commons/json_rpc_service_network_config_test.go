package commons

import (
	"fmt"
	"testing"
)

type TestJsonRpcServiceConfig struct {}

func (t TestJsonRpcServiceConfig) GetDockerImage() string {
	return "testImage"
}

func (t TestJsonRpcServiceConfig) GetJsonRpcPort() int {
	return 1234
}

func (t TestJsonRpcServiceConfig) GetOtherPorts() map[ServiceSpecificPort]int {
	return make(map[ServiceSpecificPort]int)
}

func (t TestJsonRpcServiceConfig) GetContainerStartCommand(dependencyLivenessReqs map[JsonRpcServiceSocket]JsonRpcRequest) []string {
	cmdArgs := []string{
		"arg1",
		"arg2",
	}
	for socket, _ := range dependencyLivenessReqs {
		cmdArgs = append(cmdArgs, fmt.Sprintf("%v:%v", socket.IPAddress, socket.Port))
	}
	return cmdArgs
}

func (t TestJsonRpcServiceConfig) GetLivenessRequest() JsonRpcRequest {
	return JsonRpcRequest{
		Endpoint:   "testEndpoint",
		Method:     "testMethod",
		RpcVersion: "1.0",
		Params:     nil,
		ID:         0,
	}
}

func TestDisallowingNonexistentDependencies(t *testing.T) {
	builder := NewJsonRpcServiceNetworkConfigBuilder()
	config := TestJsonRpcServiceConfig{}

	dependencies := map[int]bool{
		0: true,
	}

	builder.AddService(config, dependencies)
}
