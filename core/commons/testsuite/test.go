package testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"time"
)

/*
An interface encapsulating a test to run against a test network.
 */
type Test interface {
	// NOTE: if Go had generics, 'network' would be a parameterized type representing the network that this test consumes
	// as produced by the NetworkLoader
	/*
	Runs test logic against the given network, with failures reported using the given context.
	 */
	Run(network networks.Network, context TestContext)

	// Gets the network loader that will be used to spin up the test network that the test will run against
	GetNetworkLoader() (networks.NetworkLoader, error)


	// The amount of time the test will be allowed to execute for before it's killed and marked as failed
	GetTimeout() time.Duration
}
