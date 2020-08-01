package networks

import (
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

/*
Implementations of this class bridge the gap between Kurtosis' ServiceNetwork and developer-written, test-friendly
wrappers around it.
 */
type NetworkLoader interface {
	/*
	Hook for the user to set the service configurations that will be available for use in the network produced by this
		class, both for `InitializeNetwork` and in the test itself.
	 */
	ConfigureNetwork(builder *ServiceNetworkBuilder) error

	/*
	Hook for the user to initialize the network to whatever initial state they'd like to have when the test starts.

	Args:
		network: The network that the user should call AddService on to add nodes to the network.

	Returns:
		A map of serviceId -> availability checkers. The network will be considered available when all checkers return
			successful.
	 */
	InitializeNetwork(network *ServiceNetwork) (map[ServiceID]services.ServiceAvailabilityChecker, error)

	// GENERICS TOOD: When Go has generics, make the input and output types parameterized
	/*
	Gives the developer the opportunity to wrap the ServiceNetwork with a custom struct of their own creation, so that
		the developer can add custom test-specific methods so that writing tests is as simple as possible.
	 */
	WrapNetwork(network *ServiceNetwork) (Network, error)
}
