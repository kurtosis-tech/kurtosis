package types

import "github.com/kurtosis-tech/stacktrace"

type ContainerStatus string

const (
	Paused     ContainerStatus = "paused"
	Restarting ContainerStatus = "restarting"
	Running    ContainerStatus = "running"
	Removing   ContainerStatus = "removing"
	Dead       ContainerStatus = "dead"
	Created    ContainerStatus = "created"
	Exited     ContainerStatus = "exited"
)

var allContainerStatusesSet = map[ContainerStatus]bool{
	Paused:     true,
	Restarting: true,
	Running:    true,
	Removing:   true,
	Dead:       true,
	Created:    true,
	Exited:     true,
}

func GetContainerStatusFromString(containerStatusStr string) (ContainerStatus, error) {
	containerStatus := ContainerStatus(containerStatusStr)
	if _, found := allContainerStatusesSet[containerStatus]; !found {
		return "", stacktrace.NewError("No container status matches string '%v'", containerStatusStr)
	}
	return containerStatus, nil
}
