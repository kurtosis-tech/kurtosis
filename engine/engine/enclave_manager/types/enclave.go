package types

import (
	"github.com/docker/go-connections/nat"
	"net"
)

// TODO Refactor this a bit - an enclave will ALWAYS have a networkId, networkIpAndMask, and status but
//  it MAY NOT have an API container inside it (either because of a failed Destroy, or because it hasn't been
//  started yet)
type Enclave struct {
	networkId string
	networkIpAndMask *net.IPNet
	apiContainerId string
	apiContainerIpAddr *net.IP
	apiContainerHostPortBinding *nat.PortBinding
}

func NewEnclave(networkId string, networkIpAndMask *net.IPNet, apiContainerId string, apiContainerIpAddr *net.IP, apiContainerHostPortBinding *nat.PortBinding) *Enclave {
	return &Enclave{networkId: networkId, networkIpAndMask: networkIpAndMask, apiContainerId: apiContainerId, apiContainerIpAddr: apiContainerIpAddr, apiContainerHostPortBinding: apiContainerHostPortBinding}
}

func (e Enclave) GetNetworkId() string {
	return e.networkId
}

func (e Enclave) GetNetworkIpAndMask() *net.IPNet {
	return e.networkIpAndMask
}

func (e Enclave) GetApiContainerId() string {
	return e.apiContainerId
}

func (e Enclave) GetApiContainerIpAddr() *net.IP {
	return e.apiContainerIpAddr
}

func (e Enclave) GetApiContainerHostPortBinding() *nat.PortBinding {
	return e.apiContainerHostPortBinding
}
