package enclave_status_from_container_status_retriever

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/container_status_calculator"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/enclave_statuses"
	"github.com/kurtosis-tech/stacktrace"
)

// Helper function, used in the enclave commands, for getting an enclave state from the states of the containers
//  inside the enclave
func GetEnclaveStatus(containerStates []*types.Container) (enclave_statuses.EnclaveStatus, error) {
	result := enclave_statuses.Stopped
	for _, containerState := range containerStates {
		containerName := containerState.GetName()
		containerStatus := containerState.GetStatus()
		isRunning, err := container_status_calculator.IsContainerRunning(containerStatus)
		if err != nil {
			return "", stacktrace.Propagate(
				err,
				"An error occurred if container '%v' with status '%v' is running",
				containerName,
				containerStatus,
			)
		}
		if isRunning {
			result = enclave_statuses.Running
		}
	}
	return result, nil
}
