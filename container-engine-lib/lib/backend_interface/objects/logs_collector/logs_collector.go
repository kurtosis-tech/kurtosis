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
	status      container_status.ContainerStatus

	// This information will be nil if the logs database container isn't running
	maybePrivateIpAddr net.IP
	privateTcpPort     *port_spec.PortSpec
	privateHttpPort    *port_spec.PortSpec

	// Public (i.e. external to Kurtosis) information about the logs collector container
	// This information will be nil if the logs collector isn't running
	maybePublicIpAddr net.IP
	publicTcpPort     *port_spec.PortSpec
	publicHttpPort    *port_spec.PortSpec
}

func NewLogsCollector(
	status container_status.ContainerStatus,
	maybePrivateIpAddr net.IP,
	privateTcpPort *port_spec.PortSpec,
	privateHttpPort *port_spec.PortSpec,
	maybePublicIpAddr net.IP,
	publicTcpPort *port_spec.PortSpec,
	publicHttpPort *port_spec.PortSpec,
) *LogsCollector {
	return &LogsCollector{
		status:             status,
		maybePrivateIpAddr: maybePrivateIpAddr,
		privateTcpPort:     privateTcpPort,
		privateHttpPort:    privateHttpPort,
		maybePublicIpAddr:  maybePublicIpAddr,
		publicTcpPort:      publicTcpPort,
		publicHttpPort:     publicHttpPort,
	}
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

func (logsCollector *LogsCollector) GetPublicTcpPort() *port_spec.PortSpec {
	return logsCollector.publicTcpPort
}

func (logsCollector *LogsCollector) GetMaybePublicIpAddr() net.IP {
	return logsCollector.maybePublicIpAddr
}

func (logsCollector *LogsCollector) GetPublicHttpPort() *port_spec.PortSpec {
	return logsCollector.publicHttpPort
}

func (logsCollector *LogsCollector) GetPublicTcpAddress() (LogsCollectorAddress, error) {
	if logsCollector.maybePublicIpAddr == nil {
		return "", stacktrace.NewError("It is impossible to return the logs collector public TCP address because the value of its public IP address is nil")
	}

	if logsCollector.publicTcpPort == nil {
		return "", stacktrace.NewError("It is impossible to return the logs collector public TCP address because the value of its public TCP port spec is nil")
	}

	logsCollectorAddressStr := fmt.Sprintf("%v%v%v", logsCollector.maybePublicIpAddr, ipAndPortSeparator, logsCollector.publicTcpPort.GetNumber())
	logsCollectorAddress := LogsCollectorAddress(logsCollectorAddressStr)

	return logsCollectorAddress, nil
}
