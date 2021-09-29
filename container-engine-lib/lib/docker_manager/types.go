package docker_manager

type Container struct {
	id     string
	names  []string
	labels map[string]string
	status string
}

func NewContainer(id string, names []string, labels map[string]string, status string) *Container {
	return &Container{id: id, names: names, labels: labels, status: status}
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
	return c.status
}
