package services

import (
	"github.com/gmarchetti/kurtosis/commons/testnet"
	"gotest.tools/assert"
	"testing"
)

func TestGetContainerStartCommand(t *testing.T) {
	factoryConfig := GeckoServiceFactoryConfig{
		dockerImage:       "TEST",
		snowSampleSize:    1,
		snowQuorumSize:    1,
		stakingTlsEnabled: false,
		logLevel:          LOG_LEVEL_INFO,
	}

	expectedNoDeps := []string{
		"/gecko/build/ava",
		"--public-ip=172.17.0.2",
		"--network-id=local",
		"--http-port=9650",
		"--staking-port=9651",
		"--log-level=info",
		"--snow-sample-size=1",
		"--snow-quorum-size=1",
		"--staking-tls-enabled=false",
	}
	actualNoDeps := factoryConfig.GetStartCommand(0, make([]testnet.Service, 0))
	assert.DeepEqual(t, expectedNoDeps, actualNoDeps)

	testDependency := GeckoService{ipAddr: "1.2.3.4"}
	testDependencySlice := []testnet.Service{
		testDependency,
	}
	expectedWithDeps := append(expectedNoDeps, "--bootstrap-ips=1.2.3.4:9651")
	actualWithDeps := factoryConfig.GetStartCommand(0, testDependencySlice)
	assert.DeepEqual(t, expectedWithDeps, actualWithDeps)

}
