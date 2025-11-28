package container_status_stringifier

import (
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
)

var (
	colorizeRunning = color.New(color.FgGreen).SprintFunc()
	colorizeStopped = color.New(color.FgYellow).SprintFunc()
)

func ContainerStatusStringifier(containerStatus kurtosis_core_rpc_api_bindings.Container_Status) string {
	containerStatusStr := kurtosis_core_rpc_api_bindings.Container_Status_name[int32(containerStatus)]
	switch containerStatus {
	case kurtosis_core_rpc_api_bindings.Container_STOPPED:
		return colorizeStopped(containerStatusStr)
	case kurtosis_core_rpc_api_bindings.Container_RUNNING:
		return colorizeRunning(containerStatusStr)
	case kurtosis_core_rpc_api_bindings.Container_UNKNOWN:
		return containerStatusStr
	default:
		return containerStatusStr
	}
}
