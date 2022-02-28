package enclave_liveness_validator

import (
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
)

// TODO Get rid of this when we merge the API container & engine server, so that we don't have these weird states of "enclave exists
//  but API container isn't running"
// Validates the enclave has a running API container, and returns the host machine IP & port for connecting to it
func ValidateEnclaveLiveness(enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo) (string, uint32, uint32, error) {
	if enclaveInfo.ContainersStatus != kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING {
		return "", 0, 0, stacktrace.NewError("The containers in the enclave aren't running")
	}
	if enclaveInfo.ApiContainerStatus != kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING {
		return "", 0, 0, stacktrace.NewError("The enclave's API container isn't running")
	}
	hostMachineInfo := enclaveInfo.GetApiContainerHostMachineInfo()
	return hostMachineInfo.IpOnHostMachine, hostMachineInfo.GrpcPortOnHostMachine, hostMachineInfo.GrpcProxyPortOnHostMachine, nil
}
