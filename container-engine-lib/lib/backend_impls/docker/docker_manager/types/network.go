package types

import "net"

type Network struct {
	name      string
	id        string
	ipAndMask *net.IPNet
	labels           map[string]string
}

func NewNetwork(name string, id string, ipAndMask *net.IPNet, labels map[string]string) *Network {
	return &Network{name: name, id: id, ipAndMask: ipAndMask, labels: labels}
}

func (n Network) GetName() string {
	return n.name
}

func (n Network) GetId() string {
	return n.id
}

func (n Network) GetIpAndMask() *net.IPNet {
	return n.ipAndMask
}

func (c Network) GetLabels() map[string]string {
	return c.labels
}
