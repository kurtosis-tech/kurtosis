package service_status_stringifier

import (
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
)

var (
	colorizeRunning = color.New(color.FgGreen).SprintFunc()
	colorizeStopped = color.New(color.FgYellow).SprintFunc()
)

func ServiceStatusStringifier(serviceStatus kurtosis_core_rpc_api_bindings.ServiceStatus) string {
	serviceStatusStr := kurtosis_core_rpc_api_bindings.ServiceStatus_name[int32(serviceStatus)]
	switch serviceStatus {
	case kurtosis_core_rpc_api_bindings.ServiceStatus_STOPPED:
		return colorizeStopped(serviceStatusStr)
	case kurtosis_core_rpc_api_bindings.ServiceStatus_RUNNING:
		return colorizeRunning(serviceStatusStr)
	case kurtosis_core_rpc_api_bindings.ServiceStatus_UNKNOWN:
		return serviceStatusStr
	default:
		return serviceStatusStr
	}
}
