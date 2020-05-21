package ava_commons

import "github.com/gmarchetti/kurtosis/commons"

type SingleNodeAvaNetworkCfgProvider struct{
	GeckoImageName string
}
func (network SingleNodeAvaNetworkCfgProvider) GetNetworkConfig() *commons.JsonRpcServiceNetworkConfig {
	// TODO set up non-null nodes (indicating that they're not boot nodes)
	bootNodes := make(map[commons.JsonRpcServiceSocket]commons.JsonRpcRequest)
	geckoNodeConfig := NewGeckoServiceConfig(network.GeckoImageName, bootNodes)
	serviceConfigs := map[int]commons.JsonRpcServiceConfig {
		// TODO just a meaningless dummy value here; we'll want 10 of these as soon as we have node deps hooked up
		0: geckoNodeConfig,
	}

	// TODO once we have a builder here that allows declaring node dependencies, we'd declar deps on boot nodes here
	return commons.NewJsonRpcServiceNetworkConfig(serviceConfigs)
}
