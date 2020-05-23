package commons

import (
	"fmt"
	"gotest.tools/assert"
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

	_, err := builder.AddService(config, dependencies)
	if err == nil {
		t.Fatal("Expected error when declaring a dependency on a service ID that doesn't exist")
	}
}

func TestIdsDifferent(t *testing.T) {
	builder := NewJsonRpcServiceNetworkConfigBuilder()
	config := TestJsonRpcServiceConfig{}
	svc1, err := builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}
	svc2, err := builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}
	assert.Assert(t, svc1 != svc2, "IDs should be different")
}

func TestDependencyBookkeeping(t *testing.T) {
	builder := NewJsonRpcServiceNetworkConfigBuilder()
	config := TestJsonRpcServiceConfig{}

	svc1, err := builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	svc2, err := builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	svc3Deps := map[int]bool{
		svc1: true,
		svc2: true,
	}
	svc3, err := builder.AddService(config, svc3Deps)
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	svc4Deps := map[int]bool{
		svc1: true,
		svc3: true,
	}
	svc4, err := builder.AddService(config, svc4Deps)
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	svc5Deps := map[int]bool{
		svc2: true,
	}
	svc5, err := builder.AddService(config, svc5Deps)
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}


	expectedOrder := []int{
		svc1,
		svc2,
		svc3,
		svc4,
		svc5,
	}
	assert.DeepEqual(t,
		expectedOrder,
		builder.servicesStartOrder)

	expectedDependents := map[int]bool{
		svc4: true,
		svc5: true,
	}
	if len(expectedDependents) != len(builder.onlyDependentServices) {
		t.Fatal("Size of dependent-only services didn't match expected")
	}
	for svcId := range builder.onlyDependentServices {
		if _, found := expectedDependents[svcId]; !found {
			t.Fatalf("ID %v should be marked as dependent-only, but wasn't", svcId)
		}
	}
}

func TestDefensiveCopies(t *testing.T) {
	builder := NewJsonRpcServiceNetworkConfigBuilder()
	config := TestJsonRpcServiceConfig{}

	dependencyMap := make(map[int]bool)
	svc1, err := builder.AddService(config, dependencyMap)
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}

	networkConfig := builder.Build()

	_, err = builder.AddService(config, make(map[int]bool))
	if err != nil {
		t.Fatal("Add service shouldn't return error here")
	}
	assert.Equal(t, 1, len(networkConfig.onlyDependentServices))
	assert.Equal(t, 1, len(networkConfig.serviceConfigs))
	assert.Equal(t, 1, len(networkConfig.servicesStartOrder))

	svcDependencies := networkConfig.serviceDependencies
	assert.Equal(t, 1, len(svcDependencies))
	dependencyMap[99] = true
	assert.Equal(t, 0, len(svcDependencies[svc1]))
}
