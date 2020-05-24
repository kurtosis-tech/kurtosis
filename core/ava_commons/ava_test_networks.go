package ava_commons

import (
	"github.com/gmarchetti/kurtosis/commons"
	"github.com/palantir/stacktrace"
)

type SingleNodeAvaNetworkCfgProvider struct{
	GeckoImageName string
}
func (network SingleNodeAvaNetworkCfgProvider) GetNetworkConfig() (*commons.JsonRpcServiceNetworkConfig, error) {
	geckoNodeConfig := NewGeckoServiceConfig(network.GeckoImageName, 1, 1, false)

	builder := commons.NewJsonRpcServiceNetworkConfigBuilder()
	_, err := builder.AddService(geckoNodeConfig, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add service")
	}
	return builder.Build(), nil
}

type TwoNodeAvaNetworkCfgProvider struct{
	GeckoImageName string
}
func (network TwoNodeAvaNetworkCfgProvider) GetNetworkConfig() (*commons.JsonRpcServiceNetworkConfig, error) {
	geckoNodeConfig := NewGeckoServiceConfig(network.GeckoImageName, 2, 2, false)

	builder := commons.NewJsonRpcServiceNetworkConfigBuilder()
	bootNode, err := builder.AddService(geckoNodeConfig, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add bootnode service")
	}
	_, err = builder.AddService(
		geckoNodeConfig,
		map[int]bool{
			bootNode: true,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add dependent service")
	}
	return builder.Build(), nil
}
