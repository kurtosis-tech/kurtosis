package types

//go:generate go run github.com/dmarkham/enumer -transform=lower -trimprefix=ContainerStatus_ -type=ContainerStatus
type ContainerStatus int
const (
	// The names, as lowercase correspond to the Docker API values!
	ContainerStatus_Paused     ContainerStatus = iota
	ContainerStatus_Restarting
	ContainerStatus_Running
	ContainerStatus_Removing
	ContainerStatus_Dead
	ContainerStatus_Created
	ContainerStatus_Exited
)
