package docker_manager

const (
	AppArmorUnconfined ContainerSecurityOpt = "apparmor=unconfined"
)

type ContainerSecurityOpt string
