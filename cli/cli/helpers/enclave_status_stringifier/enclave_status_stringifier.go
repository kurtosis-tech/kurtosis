package enclave_status_stringifier

import (
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
)

var (
	colorizeRunning = color.New(color.FgGreen).SprintFunc()
	colorizeStopped = color.New(color.FgYellow).SprintFunc()
)

func EnclaveContainersStatusStringifier(enclaveStatus kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus) (string, error) {
	switch enclaveStatus {
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_EMPTY:
		return "EMPTY", nil
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING:
		return colorizeRunning("RUNNING"), nil
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED:
		return colorizeStopped("STOPPED"), nil
	default:
		return "", stacktrace.NewError("Unrecognized enclave status '%v'; this is a bug in Kurtosis", enclaveStatus)
	}
}

func EnclaveAPIContainersStatusStringifier(enclaveStatus kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus) (string, error) {
	switch enclaveStatus {
	case kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING:
		return colorizeRunning("RUNNING"), nil
	case kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_NONEXISTENT:
		return "NONEXISTENT", nil
	case kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_STOPPED:
		return colorizeStopped("STOPPED"), nil
	default:
		return "", stacktrace.NewError("Unrecognized enclave API status '%v'; this is a bug in Kurtosis", enclaveStatus)
	}
}
