package container_status_calculator

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/palantir/stacktrace"
)

func IsContainerRunning(status types.ContainerStatus) (bool, error) {
	switch status {
	case types.Running, types.Restarting:
		return true, nil
	case types.Paused, types.Removing, types.Dead, types.Created, types.Exited:
		return false, nil
	default:
		return false, stacktrace.NewError("Unrecognized container status '%v'; this is a bug in Kurtosis", status)

	}
}
