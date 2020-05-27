package testsuite

import (
	"github.com/gmarchetti/kurtosis/ava_commons/networks"
	"github.com/gmarchetti/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) GetTests() map[string]testsuite.TestConfig {
	result := make(map[string]testsuite.TestConfig)

	// TODO register tests for the other networks here
	result["singleNodeGeckoNetwork"] = testsuite.TestConfig{
		Test: SingleNodeGeckoNetworkBasicTest{},
		NetworkLoader: networks.SingleNodeGeckoNetworkLoader{},
	}

	return result
}

