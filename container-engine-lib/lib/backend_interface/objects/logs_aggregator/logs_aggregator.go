package logs_aggregator

import (
	"net"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
)

// This component is responsible for:
// 1. aggregating logs from all enclaves (by listening for logs from each enclaves logs collector)
// 2. persisting logs to persistent storage so they can be retrieved, filtered, etc.
type LogsAggregator struct {
	status container.ContainerStatus

	// This will be nil if the container is not running
	maybePrivateIpAddr net.IP

	// This information will be nil if the logs aggregator container isn't running
	privateHttpPort *port_spec.PortSpec

	// PortNum that container will listen for logs on
	logsListeningPortNum uint16
}

func NewLogsAggregator(
	status container.ContainerStatus,
	maybePrivateIpAddr net.IP,
	logsListeningPortNum uint16,
	privateHttpPort *port_spec.PortSpec,
) *LogsAggregator {
	return &LogsAggregator{
		status:               status,
		maybePrivateIpAddr:   maybePrivateIpAddr,
		logsListeningPortNum: logsListeningPortNum,
		privateHttpPort:      privateHttpPort,
	}
}

func (logsAggregator *LogsAggregator) GetStatus() container.ContainerStatus {
	return logsAggregator.status
}

func (logsAggregator *LogsAggregator) GetMaybePrivateIpAddr() net.IP {
	return logsAggregator.maybePrivateIpAddr
}

// Returns port number that logs aggregator listens for logs on
func (logsAggregator *LogsAggregator) GetListeningPortNum() uint16 {
	return logsAggregator.logsListeningPortNum
}

func (logsCollector *LogsAggregator) GetPrivateHttpPort() *port_spec.PortSpec {
	return logsCollector.privateHttpPort
}
