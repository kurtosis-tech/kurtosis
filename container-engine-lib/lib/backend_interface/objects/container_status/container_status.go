package container_status

// Represents the state of an engine
//go:generate go run github.com/dmarkham/enumer -trimprefix=ContainerStatus_ -transform=snake-upper -type=ContainerStatus
type ContainerStatus int
const (
	// The engine has been stopped (and cannot be restarted, as engines are single-use)
	ContainerStatus_Stopped ContainerStatus = iota

	// The engine is running
	ContainerStatus_Running
)