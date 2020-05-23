package ava_commons

import (
	"github.com/gmarchetti/kurtosis/commons"
	"gotest.tools/assert"
	"testing"
)

func TestGetContainerStartCommand(t *testing.T) {
	svcConfig := GeckoServiceConfig{geckoImageName: "TEST"}

	expectedNoDeps := []string{
		"/gecko/build/ava",
		"--public-ip=127.0.0.1",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-tls-enabled=false",
	}
	actualNoDeps := svcConfig.GetContainerStartCommand(make(map[commons.JsonRpcServiceSocket]commons.JsonRpcRequest))
	assert.DeepEqual(t, expectedNoDeps, actualNoDeps)

	socket := commons.JsonRpcServiceSocket{
		IPAddress: "dep1",
		Port:      1234,
	}
	testDependencyReqs := map[commons.JsonRpcServiceSocket]commons.JsonRpcRequest{
		socket: {
			Endpoint:   "test-endpoint",
			Method:     "testMethod",
			RpcVersion: "1.0",
			Params:     nil,
			ID:         0,
		},
	}
	expectedWithDeps := append(expectedNoDeps, "--bootstrap-ips=dep1:1234")
	actualWithDeps := svcConfig.GetContainerStartCommand(testDependencyReqs)
	assert.DeepEqual(t, expectedWithDeps, actualWithDeps)

}
