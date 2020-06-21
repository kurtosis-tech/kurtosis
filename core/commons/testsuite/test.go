package testsuite

import "time"

type Test interface {
	// NOTE: if Go had generics, interface{} would be a parameterized type representing the network that this test consumes
	// as produced by the TestNetworkLoader
	Run(network interface{}, context TestContext)

	GetNetworkLoader() (TestNetworkLoader, error)

	// The runtime after which the test will be killed and an error message reported
	GetTimeout() time.Duration
}
