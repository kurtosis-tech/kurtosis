package testsuite

import (
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/services"
)

// This class is intended to provide an easy place to capture the specifics of configuring a network
type TestNetworkLoader interface {
	ConfigureNetwork(builder *networks.ServiceNetworkConfigBuilder) error

	// TODO When Go has generics, make the input and output types parameterized
	// Wraps the map of service_id -> service with a user-custom object representing the network, so the user can expose
	//  whatever methods they please so writing tests is as simple as possible
	WrapNetwork(services map[int]services.Service) (interface{}, error)
}
