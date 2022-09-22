package logs_collector

import (
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

const (
	ipAndPortSeparator = ":"
)

type LogsCollectorAddress string

type LogsCollector struct {
	status container_status.ContainerStatus

	// This information will be nil if the logs database container isn't running
	privateIpAddr   net.IP
	privateTcpPort *port_spec.PortSpec
	privateHttpPort *port_spec.PortSpec
}

func NewLogsCollector(status container_status.ContainerStatus, privateIpAddr net.IP, privateTcpPort *port_spec.PortSpec, privateHttpPort *port_spec.PortSpec) *LogsCollector {
	return &LogsCollector{status: status, privateIpAddr: privateIpAddr, privateTcpPort: privateTcpPort, privateHttpPort: privateHttpPort}
}

func (logsCollector *LogsCollector) GetStatus() container_status.ContainerStatus {
	return logsCollector.status
}

func (logsCollector *LogsCollector) GetPrivateIpAddr() net.IP {
	return logsCollector.privateIpAddr
}

func (logsCollector *LogsCollector) GetPrivateTcpPort() *port_spec.PortSpec {
	return logsCollector.privateTcpPort
}

func (logsCollector *LogsCollector) GetPrivateHttpPort() *port_spec.PortSpec {
	return logsCollector.privateHttpPort
}

func (logsCollector *LogsCollector) GetPrivateTcpAddress() LogsCollectorAddress {
	logsCollectorAddressStr := fmt.Sprintf("%v%v%v", logsCollector.GetPrivateIpAddr(), ipAndPortSeparator, logsCollector.GetPrivateTcpPort().GetNumber())
	logsCollectorAddress := LogsCollectorAddress(logsCollectorAddressStr)

	return logsCollectorAddress
}
