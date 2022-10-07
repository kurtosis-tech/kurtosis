package logs_collector

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

const (
	ipAndPortSeparator = ":"
)

type LogsCollectorAddress string

type LogsCollector struct {
	status container_status.ContainerStatus

	// This information will be nil if the logs database container isn't running
	maybePrivateIpAddr net.IP
	privateTcpPort     *port_spec.PortSpec
	privateHttpPort    *port_spec.PortSpec
}

func NewLogsCollector(status container_status.ContainerStatus, maybePrivateIpAddr net.IP, privateTcpPort *port_spec.PortSpec, privateHttpPort *port_spec.PortSpec) *LogsCollector {
	return &LogsCollector{status: status, maybePrivateIpAddr: maybePrivateIpAddr, privateTcpPort: privateTcpPort, privateHttpPort: privateHttpPort}
}

func (logsCollector *LogsCollector) GetStatus() container_status.ContainerStatus {
	return logsCollector.status
}

func (logsCollector *LogsCollector) GetMaybePrivateIpAddr() net.IP {
	return logsCollector.maybePrivateIpAddr
}

func (logsCollector *LogsCollector) GetPrivateTcpPort() *port_spec.PortSpec {
	return logsCollector.privateTcpPort
}

func (logsCollector *LogsCollector) GetPrivateHttpPort() *port_spec.PortSpec {
	return logsCollector.privateHttpPort
}

func (logsCollector *LogsCollector) GetPrivateTcpAddress() (LogsCollectorAddress, error){
	if logsCollector.maybePrivateIpAddr == nil {
		return "", stacktrace.NewError("It is impossible to returns the logs collector private TCP address due the value of its private IP address is nil")
	}
	privateIpAddress := logsCollector.maybePrivateIpAddr.String()

	logsCollectorAddressStr := fmt.Sprintf("%v%v%v", privateIpAddress, ipAndPortSeparator, logsCollector.privateTcpPort.GetNumber())
	logsCollectorAddress := LogsCollectorAddress(logsCollectorAddressStr)

	return logsCollectorAddress, nil
}
