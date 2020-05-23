package ava_commons

import (
	"github.com/gmarchetti/kurtosis/commons"
	"github.com/palantir/stacktrace"
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
	verifySliceEquality(t, expectedNoDeps, actualNoDeps)

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
	verifySliceEquality(t, expectedWithDeps, actualWithDeps)

}

func verifySliceEquality(t *testing.T, expected []string, actual []string) {
	if (len(expected) != len(actual)) {
		t.Fatal(stacktrace.NewError("Expected slice of length %v but got %v", len(expected), len(actual)))
	}
	for idx, expectedElem := range expected {
		actualElem := actual[idx]
		if (expectedElem != actualElem) {
			t.Fatal(stacktrace.NewError("Expected elem '%v' at position %v but got '%v'", expectedElem, idx, actualElem))
		}
	}
}