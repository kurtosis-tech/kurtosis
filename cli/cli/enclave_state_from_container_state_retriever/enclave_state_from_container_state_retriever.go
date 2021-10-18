package enclave_state_from_container_state_retriever

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager/enclave_states"
	"github.com/palantir/stacktrace"
)

// Helper function, used in the enclave commands, for getting an enclave state from the states of the containers
//  inside the enclave
func GetEnclaveState(containerStates []*types.Container) (enclave_states.EnclaveState, error) {
	result := enclave_states.Stopped
	for _, containerState := range containerStates {
		containerStatus := containerState.GetStatus()
		switch containerStatus {
		case types.Running, types.Restarting:
			result = enclave_states.Running
		case types.Paused, types.Removing, types.Dead, types.Created, types.Exited:
			continue
		default:
			return "", stacktrace.NewError("Unrecognized container status '%v'; this is a bug in Kurtosis", containerStatus)
		}
	}
	return result, nil
}