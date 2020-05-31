package testsuite

import (
	"github.com/kurtosis-tech/kurtosis/ava_commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.TestConfig {
	result := make(map[string]testsuite.TestConfig)

	singleNodeNetworkLoaer := networks.SingleNodeGeckoNetworkLoader{}

	result["singleNodeBasicTest"] = testsuite.TestConfig{
		Test: SingleNodeGeckoNetworkBasicTest{},
		NetworkLoader: singleNodeNetworkLoaer,
	}
	/*
	result["singleNodeGetValidatorsTest"] = testsuite.TestConfig{
		Test: SingleNodeGeckoNetworkBasicTest{},
		NetworkLoader: singleNodeNetworkLoaer,
	}
	result["tenNodeGeckoNetwork"] = testsuite.TestConfig{
		Test: TenNodeGeckoNetworkBasicTest{},
		NetworkLoader: networks.TenNodeGeckoNetworkLoader{},
	}

	 */

	return result
}

