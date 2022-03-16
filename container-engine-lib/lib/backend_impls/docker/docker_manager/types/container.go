package types

import (
	"github.com/docker/go-connections/nat"
	"net"
)

type Container struct {
	id     string
	name   string
	labels           map[string]string
	status           ContainerStatus
	hostPortBindings map[nat.Port]*nat.PortBinding
	networksIPAdrresses map[string]net.IP
}

func NewContainer(
	id string,
	name string,
	labels map[string]string,
	status ContainerStatus,
	hostPortBindings map[nat.Port]*nat.PortBinding,
	networksIPAdrresses map[string]net.IP,
) *Container {
	return &Container{
		id: id,
		name: name,
		labels: labels,
		status: status,
		hostPortBindings: hostPortBindings,
		networksIPAdrresses: networksIPAdrresses,
	}
}

func (c *Container) GetId() string {
	return c.id
}

func (c *Container) GetName() string {
	return c.name
}

func (c *Container) GetLabels() map[string]string {
	return c.labels
}

func (c *Container) GetStatus() ContainerStatus {
	return c.status
}

func (c Container) GetHostPortBindings() map[nat.Port]*nat.PortBinding {
	return c.hostPortBindings
}

func (c *Container) GetNetworksIPAddresses() map[string]net.IP {
	return c.networksIPAdrresses
}
