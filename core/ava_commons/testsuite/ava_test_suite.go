package testsuite

import (
	"github.com/kurtosis-tech/kurtosis/ava_commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.TestConfig {
	result := make(map[string]testsuite.TestConfig)

	// TODO register tests for the other networks here
	result["singleNodeGeckoNetwork"] = testsuite.TestConfig{
		Test: SingleNodeGeckoNetworkBasicTest{},
		NetworkLoader: networks.SingleNodeGeckoNetworkLoader{},
	}
	/*
	result["tenNodeGeckoNetwork"] = testsuite.TestConfig{
		Test: TenNodeGeckoNetworkBasicTest{},
		NetworkLoader: networks.TenNodeGeckoNetworkLoader{},
	}

	 */

	return result
}

