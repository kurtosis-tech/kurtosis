package networks

import (
	"github.com/kurtosis-tech/kurtosis/ava_commons/services"
	"github.com/kurtosis-tech/kurtosis/commons/testnet"
	"github.com/palantir/stacktrace"
)

type SingleNodeGeckoNetwork struct{
	node services.GeckoService
}
func (network SingleNodeGeckoNetwork) GetNode() services.GeckoService {
	return network.node
}

type SingleNodeGeckoNetworkLoader struct {}
func (loader SingleNodeGeckoNetworkLoader) GetNetworkConfig(testImageName string) (*testnet.ServiceNetworkConfig, error) {
	factoryConfig := services.NewGeckoServiceFactoryConfig(
		testImageName,
		1,
		1,
		false,
		services.LOG_LEVEL_DEBUG)
	factory := testnet.NewServiceFactory(factoryConfig)

	builder := testnet.NewServiceNetworkConfigBuilder()
	config1 := builder.AddServiceConfiguration(*factory)
	_, err := builder.AddService(config1, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add service")
	}
	return builder.Build(), nil
}
func (loader SingleNodeGeckoNetworkLoader) LoadNetwork(ipAddrs map[int]string) (interface{}, error) {
	return SingleNodeGeckoNetwork{
		node: *services.NewGeckoService(ipAddrs[0]),
	}, nil
}
