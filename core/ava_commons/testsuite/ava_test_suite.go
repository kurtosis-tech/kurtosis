package testsuite

import (
	"github.com/gmarchetti/kurtosis/ava_commons/networks"
	"github.com/gmarchetti/kurtosis/commons/testsuite"
)

type AvaTestSuite struct {}

func (a AvaTestSuite) RegisterTests(builder testsuite.TestRegistryBuilder) {
	builder.AddTest("singleNodeGeckoNetwork", networks.SingleNodeGeckoNetworkLoader{}, SingleNodeGeckoNetworkBasicTest{})
	builder.AddTest("twoNodeGeckoNetwork", networks.SingleNodeGeckoNetworkLoader{}, SingleNodeGeckoNetworkBasicTest{})
	builder.AddTest("tenNodeGeckoNetwork", networks.SingleNodeGeckoNetworkLoader{}, SingleNodeGeckoNetworkBasicTest{})
}

