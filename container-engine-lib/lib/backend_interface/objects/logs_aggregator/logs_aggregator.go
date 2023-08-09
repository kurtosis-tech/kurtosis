package logs_aggregator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"net"
)

// This component is responsible for:
// 1. aggregating logs from all enclaves (by listening for logs from each enclaves logs collector)
// 2. persistening logs to persistent storage so they can be retrieved, filtered, etc.
type LogsAggregator struct {
	status container_status.ContainerStatus

	// This will be nil if the container is not running
	maybePrivateIpAddr net.IP
}

func NewLogsAggregator(status container_status.ContainerStatus, maybePrivateIpAddr net.IP) *LogsAggregator {
	return &LogsAggregator{status: status, maybePrivateIpAddr: maybePrivateIpAddr}
}

func (logsAggregator *LogsAggregator) GetStatus() container_status.ContainerStatus {
	return logsAggregator.status
}

func (logsAggregator *LogsAggregator) GetMaybePrivateIpAddr() net.IP {
	return logsAggregator.maybePrivateIpAddr
}
