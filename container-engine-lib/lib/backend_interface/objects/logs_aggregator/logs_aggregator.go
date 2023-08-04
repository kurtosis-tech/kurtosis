package logs_aggregator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type LogsAggregator struct {
	status container_status.ContainerStatus

	// This information will be nil if the logs database container isn't running
	// TOOD: consider changing these names
	//this information should specify ip and port log collectors should forward logs too
	maybePrivateIpAddr net.IP
	privateHttpPort    *port_spec.PortSpec
}

func NewLogsAggregator(
	status container_status.ContainerStatus,
	maybePrivateIpAddr net.IP,
	privateHttpPort *port_spec.PortSpec,
) *LogsAggregator {
	return &LogsAggregator{status: status, maybePrivateIpAddr: maybePrivateIpAddr, privateHttpPort: privateHttpPort}
}

func (logsAggregator *LogsAggregator) GetStatus() container_status.ContainerStatus {
	return logsAggregator.status
}

func (logsAggregator *LogsAggregator) GetMaybePrivateIpAddr() net.IP {
	return logsAggregator.maybePrivateIpAddr
}

func (logsAggregator *LogsAggregator) GetPrivateHttpPort() *port_spec.PortSpec {
	return logsAggregator.privateHttpPort
}
