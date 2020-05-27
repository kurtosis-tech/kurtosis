package testnet

/*
This interface is responsible for creating a network config so that the
test suite runner can use it to create a network, and then after the test
controller is initialized and the RawServiceNetwork object is passed over to
it this class will be used to create the actual network object
 */
type TestNetworkLoader interface {
	GetNetworkConfig(testImageName string) (*ServiceNetworkConfig, error)

	// If Go had generics, this return type would be parameterized to be the actual type of network a test will consume
	LoadNetwork(network RawServiceNetwork) (interface{}, error)
}
