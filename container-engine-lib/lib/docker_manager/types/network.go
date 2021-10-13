package types

import "net"

type Network struct {
	name      string
	id        string
	ipAndMask *net.IPNet
}

func NewNetwork(name string, id string, ipAndMask *net.IPNet) *Network {
	return &Network{name: name, id: id, ipAndMask: ipAndMask}
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
