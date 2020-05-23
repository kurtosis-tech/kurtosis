package ava_commons

import "github.com/gmarchetti/kurtosis/commons"

type SingleNodeAvaNetworkCfgProvider struct{
	GeckoImageName string
}
func (network SingleNodeAvaNetworkCfgProvider) GetNetworkConfig() *commons.JsonRpcServiceNetworkConfig {
	// TODO set up non-null nodes (indicating that they're not boot nodes)
	geckoNodeConfig := NewGeckoServiceConfig(network.GeckoImageName)

	builder := commons.NewJsonRpcServiceNetworkConfigBuilder()
	builder.AddService(geckoNodeConfig, make(map[int]bool))
	return builder.Build()
}
