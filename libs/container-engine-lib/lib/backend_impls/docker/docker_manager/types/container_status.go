package types

// TODO Make this private once nobody outside container-engine-lib uses DockerManager directly!

//go:generate go run github.com/dmarkham/enumer -transform=lower -trimprefix=ContainerStatus_ -type=ContainerStatus
type ContainerStatus int
const (
	// WARNING: The XXXXX in ContainerStatus_XXXX must, when lowercased, correspond to a Docker container status value!
	// https://github.com/moby/moby/blob/master/container/state.go#L140
	ContainerStatus_Paused ContainerStatus = iota
	ContainerStatus_Restarting
	ContainerStatus_Running
	ContainerStatus_Removing
	ContainerStatus_Dead
	ContainerStatus_Created
	ContainerStatus_Exited
)
