package enclave_liveness_validator

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
)

// Validates the enclave has a running API container, and returns the host machine IP & port for connecting to it
func ValidateEnclaveLiveness(enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo) (string, uint32, error) {
	if enclaveInfo.ContainersStatus != kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING {
		return "", 0, stacktrace.NewError("The containers in the enclave aren't running")
	}
	if enclaveInfo.ApiContainerStatus != kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING {
		return "", 0, stacktrace.NewError("The enclave's API container isn't running")
	}
	hostMachineInfo := enclaveInfo.GetApiContainerHostMachineInfo()

	return hostMachineInfo.IpOnHostMachine, hostMachineInfo.GrpcPortOnHostMachine, nil
}
