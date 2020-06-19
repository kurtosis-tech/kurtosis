package testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

// This class is intended to provide an easy place to capture the specifics of configuring a network
type TestNetworkLoader interface {
	// Hook for the user to set the service configurations that they're going to use
	ConfigureNetwork(builder *networks.ServiceNetworkBuilder) error

	/*
	Hook for the user to initialize the network to whatever state they'd like the Test to have
	Args:
		network: The network that the user should call AddService on
	Returns: A slice of availability checkers that, should they all return successfully, mean the network is counted as available
	 */
	BootstrapNetwork(network *networks.ServiceNetwork) ([]services.ServiceAvailabilityChecker, error)

	// TODO When Go has generics, make the input and output types parameterized
	// Wraps the map of service_id -> service with a user-custom object representing the network, so the user can expose
	//  whatever methods they please so writing tests is as simple as possible
	WrapNetwork(services map[int]services.Service) (interface{}, error)
}
