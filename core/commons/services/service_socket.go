package services

import "github.com/docker/go-connections/nat"

type ServiceSocket struct {
	ipAddr string
	port nat.Port
}

func NewServiceSocket(ipAddr string, port nat.Port) *ServiceSocket {
	return &ServiceSocket{
		ipAddr: ipAddr,
		port:   port,
	}
}

func (socket *ServiceSocket) GetIpAddr() string {
	return socket.ipAddr
}

func (socket *ServiceSocket) GetPort() nat.Port {
	return socket.port
}
