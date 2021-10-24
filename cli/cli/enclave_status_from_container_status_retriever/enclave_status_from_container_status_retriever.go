package enclave_status_from_container_status_retriever

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_statuses"
	"github.com/palantir/stacktrace"
)

// Helper function, used in the enclave commands, for getting an enclave state from the states of the containers
//  inside the enclave
func GetEnclaveStatus(containerStates []*types.Container) (enclave_statuses.EnclaveStatus, error) {
	result := enclave_statuses.Stopped
	for _, containerState := range containerStates {
		containerStatus := containerState.GetStatus()
		switch containerStatus {
		case types.Running, types.Restarting:
			result = enclave_statuses.Running
		case types.Paused, types.Removing, types.Dead, types.Created, types.Exited:
			continue
		default:
			return "", stacktrace.NewError("Unrecognized container status '%v'; this is a bug in Kurtosis", containerStatus)
		}
	}
	return result, nil
}