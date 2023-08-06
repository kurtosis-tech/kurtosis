package logs_aggregator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"net"
)

type LogsAggregator struct {
	status container_status.ContainerStatus

	// This information will be nil if the logs aggregator container isn't running
	maybePrivateIpAddr net.IP
}

func NewLogsAggregator(
	status container_status.ContainerStatus,
	maybePrivateIpAddr net.IP,
) *LogsAggregator {
	return &LogsAggregator{status: status, maybePrivateIpAddr: maybePrivateIpAddr}
}

func (logsAggregator *LogsAggregator) GetStatus() container_status.ContainerStatus {
	return logsAggregator.status
}

func (logsAggregator *LogsAggregator) GetMaybePrivateIpAddr() net.IP {
	return logsAggregator.maybePrivateIpAddr
}
