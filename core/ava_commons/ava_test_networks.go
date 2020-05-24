package ava_commons

import (
	"github.com/gmarchetti/kurtosis/commons"
	"github.com/palantir/stacktrace"
)

type SingleNodeAvaNetworkCfgProvider struct{
	GeckoImageName string
}
func (network SingleNodeAvaNetworkCfgProvider) GetNetworkConfig() (*commons.JsonRpcServiceNetworkConfig, error) {
	geckoNodeConfig := NewGeckoServiceConfig(network.GeckoImageName, 1, 1, false, LOG_LEVEL_INFO)

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
	geckoNodeConfig := NewGeckoServiceConfig(network.GeckoImageName, 2, 2, false, LOG_LEVEL_INFO)

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

type TenNodeAvaNetworkCfgProvider struct{
	GeckoImageName string
}
func (network TenNodeAvaNetworkCfgProvider) GetNetworkConfig() (*commons.JsonRpcServiceNetworkConfig, error) {
	geckoNodeConfig := NewGeckoServiceConfig(network.GeckoImageName, 2, 2, false, LOG_LEVEL_INFO)

	builder := commons.NewJsonRpcServiceNetworkConfigBuilder()
	bootNode0, err := builder.AddService(geckoNodeConfig, make(map[int]bool))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add bootnode service")
	}
	bootNode1, err := builder.AddService(
		geckoNodeConfig,
		map[int]bool{
			bootNode0: true,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add dependent service")
	}
	bootNode2, err := builder.AddService(
		geckoNodeConfig,
		map[int]bool{
			bootNode0: true,
			bootNode1: true,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not add dependent service")
	}
	bootNodeMap := map[int]bool{
		bootNode0: true,
		bootNode1: true,
		bootNode2: true,
	}
	for i:=3; i < 10; i++ {
		_, err := builder.AddService(
			geckoNodeConfig,
			bootNodeMap,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not add dependent service")
		}
	}

	return builder.Build(), nil
}
