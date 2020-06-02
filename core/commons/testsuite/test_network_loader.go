package testsuite

import "github.com/kurtosis-tech/kurtosis/commons/testnet"

/*
This interface should be implemented by the user and is responsible for:
1) creating a network config so that the initializer can use it to create a network and
2) parsing the RawServiceNetwork object passed over to the TestController to create the actual network the test will  receive
 */
type TestNetworkLoader interface {
	GetNetworkConfig(testImageName string) (*testnet.ServiceNetworkConfig, error)

	// If Go had generics, this return type would be parameterized to be the actual type of network a test will consume
	LoadNetwork(ipAddrs map[int]string) (interface{}, error)
}
