package testnet

type TestNetworkConfigProvider interface {
	GetNetworkConfig() (*ServiceNetworkConfig, error)

	// TODO need to also return an enum that will tell the test controller what type of network
}
