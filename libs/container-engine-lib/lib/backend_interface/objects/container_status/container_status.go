package container_status

// Represents the state of a container within Kurtosis
//go:generate go run github.com/dmarkham/enumer -trimprefix=ContainerStatus_ -transform=snake-upper -type=ContainerStatus
type ContainerStatus int
const (
	ContainerStatus_Stopped ContainerStatus = iota
	ContainerStatus_Running
)