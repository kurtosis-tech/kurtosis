package types

import "net"

type Network struct {
	name      string
	id        string
	ipAndMask *net.IPNet
	gatewayIpStr   string
	// Set of used IP address strings on the network
	containerIpStrs map[string]bool
	labels           map[string]string

}

func NewNetwork(name string, id string, ipAndMask *net.IPNet, gatewayIpStr string, containerIpStrs map[string]bool, labels map[string]string) *Network {
	return &Network{name: name, id: id, ipAndMask: ipAndMask, gatewayIpStr: gatewayIpStr, containerIpStrs: containerIpStrs, labels: labels}
}

func (network Network) GetName() string {
	return network.name
}

func (network Network) GetId() string {
	return network.id
}

func (network Network) GetIpAndMask() *net.IPNet {
	return network.ipAndMask
}

func (network Network) GetGatewayIp() string {
	return network.gatewayIpStr
}

func (network Network) GetContainerIps() map[string]bool {
	return network.containerIpStrs
}

func (network Network) GetLabels() map[string]string {
	return network.labels
}
