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
	Hook for the user to set the service configurations that they're going to use, both in `InitializeNetwork` and in
	the test itself.
	 */
	ConfigureNetwork(builder *ServiceNetworkBuilder) error

	/*
	Hook for the user to initialize the network to whatever initial state they'd like to have when the test starts.

	Args:
		network: The network that the user should call AddService on

	Returns:
		A map of serviceId -> availability checkers where the network is considered available when all checkers return
	 */
	InitializeNetwork(network *ServiceNetwork) (map[ServiceID]services.ServiceAvailabilityChecker, error)

	// GENERICS TOOD: When Go has generics, make the input and output types parameterized
	// Wraps the network with a user-custom object representing the network, so the user can expose
	//  whatever methods they please so writing tests is as simple as possible
	WrapNetwork(network *ServiceNetwork) (Network, error)
}
