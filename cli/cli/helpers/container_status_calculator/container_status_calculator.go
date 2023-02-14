package container_status_calculator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/stacktrace"
)

func IsContainerRunning(status types.ContainerStatus) (bool, error) {
	switch status {
	case types.ContainerStatus_Running, types.ContainerStatus_Restarting:
		return true, nil
	case types.ContainerStatus_Paused, types.ContainerStatus_Removing, types.ContainerStatus_Dead, types.ContainerStatus_Created, types.ContainerStatus_Exited:
		return false, nil
	default:
		return false, stacktrace.NewError("Unrecognized container status '%v'; this is a bug in Kurtosis", status)

	}
}
