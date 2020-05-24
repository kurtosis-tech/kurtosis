package ava_commons

import (
	"github.com/gmarchetti/kurtosis/commons"
	"github.com/palantir/stacktrace"
)

type SingleNodeAvaNetworkCfgProvider struct{
	GeckoImageName string
}
func (network SingleNodeAvaNetworkCfgProvider) GetNetworkConfig() (*commons.ServiceNetworkConfig, error) {
	factoryConfig := NewGeckoServiceFactoryConfig(
		network.GeckoImageName,
		1,
		1,
		false,
		LOG_LEVEL_DEBUG)
	factory := commons.NewServiceFactory(factoryConfig)

	builder := commons.NewServiceNetworkConfigBuilder()
	config1 := builder.AddServiceConfiguration(*factory)
	_, err := builder.AddService(config1, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add service")
	}
	return builder.Build(), nil
}

type TwoNodeAvaNetworkCfgProvider struct{
	GeckoImageName string
}
func (network TwoNodeAvaNetworkCfgProvider) GetNetworkConfig() (*commons.ServiceNetworkConfig, error) {
	factoryConfig := NewGeckoServiceFactoryConfig(
		network.GeckoImageName,
		1,
		1,
		false,
		LOG_LEVEL_DEBUG)
	factory := commons.NewServiceFactory(factoryConfig)

	builder := commons.NewServiceNetworkConfigBuilder()
	config1 := builder.AddServiceConfiguration(*factory)
	bootNode, err := builder.AddService(config1, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add bootnode service")
	}
	_, err = builder.AddService(
		config1,
		map[int]bool{
			bootNode: true,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add dependent service")
	}
	return builder.Build(), nil
}
