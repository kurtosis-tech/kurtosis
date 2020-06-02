package networks

import "github.com/kurtosis-tech/kurtosis/commons/services"

type ServiceNetwork struct {
	NetworkId string

	// If Go had generics, we'd make this object genericized and use that as the return type here
	Services map[int]services.Service

	ContainerIds map[int]string
}
