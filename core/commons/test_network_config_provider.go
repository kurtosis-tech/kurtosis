package commons

type TestNetworkConfigProvider interface {
	GetNetworkConfig() (*JsonRpcServiceNetworkConfig, error)

	// TODO need to also return an enum that will tell the test controller what type of network
}
