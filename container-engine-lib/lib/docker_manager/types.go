package docker_manager

import "github.com/palantir/stacktrace"

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
	numStatus //it must be the last const in this list, it's a helper it's only used to get all values throw allStatus()
)

type Container struct {
	id     string
	names  []string
	labels map[string]string
	status Status
}

func NewContainer(id string, names []string, labels map[string]string, status Status) (*Container, error) {
	if status == numStatus {
		return nil, stacktrace.NewError("It is not allowed to create a Container with status value = numStatus")
	}
	return &Container{id: id, names: names, labels: labels, status: status}, nil
}

func (c Container) GetId() string {
	return c.id
}

func (c Container) GetNames() []string {
	return c.names
}

func (c Container) GetLabels() map[string]string {
	return c.labels
}

func (c Container) GetStatus() string {
	return c.status.string()
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

func getAllStatus() []Status {
	allStatus := make([]Status, numStatus)
	for i := 0; i < int(numStatus); i++ {
		allStatus[i] = Status(i)
	}
	return allStatus
}

func getContainerStatusByDockerContainerState(dockerContainerState string) Status {
	allStatus := getAllStatus()
	for _, status := range allStatus {
		if status.string() == dockerContainerState {
			return status
		}
	}
	 return unknown
}