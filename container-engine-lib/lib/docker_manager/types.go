package docker_manager

import (
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
)

type Status uint32

const (
	unknown Status = iota
	paused
	restarting
	running
	removing
	dead
	created
	exited
	numStatuses //it must be the last const in this list, it's a helper it's only used to get all values throw allStatuses()
)

type Container struct {
	id     string
	name   string
	labels map[string]string
	status Status
	hostPortBindings map[nat.Port]*nat.PortBinding
}

func NewContainer(id string, name string, labels map[string]string, status Status, hostPortBindings map[nat.Port]*nat.PortBinding) (*Container, error) {
	if status == numStatuses {
		return nil, stacktrace.NewError("It is not allowed to create a Container with status value = numStatuses")
	}
	return &Container{id: id, name: name, labels: labels, status: status, hostPortBindings: hostPortBindings}, nil
}

func (c Container) GetId() string {
	return c.id
}

func (c Container) GetName() string {
	return c.name
}

func (c Container) GetLabels() map[string]string {
	return c.labels
}

func (c Container) GetStatus() string {
	return c.status.string()
}

func (c Container) GetHostPortBindings() map[nat.Port]*nat.PortBinding {
	return c.hostPortBindings
}

func (s Status) string() string {
	switch s {
	case paused:
		return "paused"
	case restarting:
		return "restarting"
	case running:
		return "running"
	case removing:
		return "removing"
	case dead:
		return "dead"
	case created:
		return "created"
	case exited:
		return "exited"
	default:
		return "unknown"
	}
}

func getAllContainerStatuses() []Status {
	allStatuses := make([]Status, numStatuses)
	for i := 0; i < int(numStatuses); i++ {
		allStatuses[i] = Status(i)
	}
	return allStatuses
}
