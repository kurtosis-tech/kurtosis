package types

import "net"

type Network struct {
	name      string
	id        string
	ipAndMask *net.IPNet
	gatewayIpStr   string
	labels           map[string]string

}

func NewNetwork(name string, id string, ipAndMask *net.IPNet, gatewayIpStr string, labels map[string]string) *Network {
	return &Network{name: name, id: id, ipAndMask: ipAndMask, gatewayIpStr: gatewayIpStr, labels: labels}
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

func (network Network) GetLabels() map[string]string {
	return network.labels
}
