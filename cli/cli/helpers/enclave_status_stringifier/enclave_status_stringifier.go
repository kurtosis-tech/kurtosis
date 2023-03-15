package enclave_status_stringifier

import (
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
)

var (
	colorizeRunning = color.New(color.FgGreen).SprintFunc()
	colorizeStopped = color.New(color.FgYellow).SprintFunc()
	colorizeEmpty   = color.New(color.FgCyan).SprintFunc()
)

func EnclaveContainersStatusStringifier(enclaveStatus kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus) (string, error) {
	switch enclaveStatus {
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_EMPTY:
		return colorizeEmpty("EMPTY"), nil
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING:
		return colorizeRunning("RUNNING"), nil
	case kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED:
		return colorizeStopped("STOPPED"), nil
	default:
		return "", stacktrace.NewError("Unrecognized enclave status '%v'; this is a bug in Kurtosis", enclaveStatus)
	}
}
