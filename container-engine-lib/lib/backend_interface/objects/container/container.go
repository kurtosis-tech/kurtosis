package container

// Represents a service container
type Container struct {
	status         ContainerStatus
	imageName      string
	entrypointArgs []string
	cmdArgs        []string
	envVars        map[string]string
}

func NewContainer(status ContainerStatus, imageName string, entrypointArgs []string, cmdArgs []string, envVars map[string]string) *Container {
	return &Container{status: status, imageName: imageName, entrypointArgs: entrypointArgs, cmdArgs: cmdArgs, envVars: envVars}
}

func (container *Container) GetStatus() ContainerStatus {
	return container.status
}

func (container *Container) GetImageName() string {
	return container.imageName
}

func (container *Container) GetEntrypointArgs() []string {
	return container.entrypointArgs
}

func (container *Container) GetCmdArgs() []string {
	return container.cmdArgs
}

func (container *Container) GetEnvVars() map[string]string {
	return container.envVars
}
