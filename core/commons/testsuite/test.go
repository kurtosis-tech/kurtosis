package testsuite

import "github.com/kurtosis-tech/kurtosis/controller"

type Test interface {
	// NOTE: if Go had generics, interface{} would be a parameterized type representing the network that this test consumes
	// as produced by the TestNetworkLoader
	Run(network interface{}, context controller.TestContext)

	// TODO add a GetNetworkLoader function
}
