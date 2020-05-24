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
		"--public-ip=172.17.0.2",
		"--network-id=local",
		"--http-port=9650",
		"--staking-port=9651",
		"--log-level=verbo",
		"--snow-sample-size=0",
		"--snow-quorum-size=0",
		"--staking-tls-enabled=false",
	}
	actualNoDeps := svcConfig.GetContainerStartCommand(0, make(map[commons.JsonRpcServiceSocket]commons.JsonRpcRequest))
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
	expectedWithDeps := append(expectedNoDeps, "--bootstrap-ips=dep1:9651")
	actualWithDeps := svcConfig.GetContainerStartCommand(0, testDependencyReqs)
	assert.DeepEqual(t, expectedWithDeps, actualWithDeps)

}
