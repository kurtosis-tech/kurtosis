package types

import (
	"github.com/docker/go-connections/nat"
)

type Container struct {
	id               string
	name             string
	labels           map[string]string
	status           ContainerStatus
	hostPortBindings map[nat.Port]*nat.PortBinding
	imageName        string
	entrypointArgs   []string
	cmdArgs          []string
	envVars          map[string]string
	defaultIpAddress string
}

func NewContainer(
	id string,
	name string,
	labels map[string]string,
	status ContainerStatus,
	hostPortBindings map[nat.Port]*nat.PortBinding,
	imageName string,
	entrypointArgs []string,
	cmdArgs []string,
	envVars map[string]string,
	defaultIpAddress string,
) *Container {
	return &Container{
		id:               id,
		name:             name,
		labels:           labels,
		status:           status,
		hostPortBindings: hostPortBindings,
		imageName:        imageName,
		entrypointArgs:   entrypointArgs,
		cmdArgs:          cmdArgs,
		envVars:          envVars,
		defaultIpAddress: defaultIpAddress,
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

func (c *Container) GetImageName() string {
	return c.imageName
}

func (c *Container) GetEntrypointArgs() []string {
	return c.entrypointArgs
}

func (c *Container) GetCmdArgs() []string {
	return c.cmdArgs
}

func (c *Container) GetEnvVars() map[string]string {
	return c.envVars
}
