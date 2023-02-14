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
	privateTcpPort  *port_spec.PortSpec
	privateHttpPort *port_spec.PortSpec

	// bridge network ip address
	maybeEnclaveNetworkIpAddress net.IP
	maybeBridgeNetworkIpAddress  net.IP
}

func NewLogsCollector(
	status container_status.ContainerStatus,
	maybeEnclaveNetworkIpAddress net.IP,
	maybeBridgeNetworkIpAddress net.IP,
	privateTcpPort *port_spec.PortSpec,
	privateHttpPort *port_spec.PortSpec,
) *LogsCollector {
	return &LogsCollector{
		status:                       status,
		maybeEnclaveNetworkIpAddress: maybeEnclaveNetworkIpAddress,
		maybeBridgeNetworkIpAddress:  maybeBridgeNetworkIpAddress,
		privateTcpPort:               privateTcpPort,
		privateHttpPort:              privateHttpPort,
	}
}

func (logsCollector *LogsCollector) GetStatus() container_status.ContainerStatus {
	return logsCollector.status
}

func (logsCollector *LogsCollector) GetEnclaveNetworkIpAddress() net.IP {
	return logsCollector.maybeEnclaveNetworkIpAddress
}

func (LogsCollector *LogsCollector) GetBridgeNetworkIpAddress() net.IP {
	return LogsCollector.maybeBridgeNetworkIpAddress
}

func (logsCollector *LogsCollector) GetPrivateTcpPort() *port_spec.PortSpec {
	return logsCollector.privateTcpPort
}

func (logsCollector *LogsCollector) GetPrivateHttpPort() *port_spec.PortSpec {
	return logsCollector.privateHttpPort
}

func (logsCollector *LogsCollector) GetEnclaveNetworkTcpAddress() (LogsCollectorAddress, error) {
	if logsCollector.maybeEnclaveNetworkIpAddress == nil {
		return "", stacktrace.NewError("It is impossible to return the logs collector private TCP address because the value of its private IP address is nil")
	}

	if logsCollector.privateTcpPort == nil {
		return "", stacktrace.NewError("It is impossible to return the logs collector private TCP address because the value of its private TCP port spec is nil")
	}

	logsCollectorAddressStr := fmt.Sprintf("%v%v%v", logsCollector.maybeEnclaveNetworkIpAddress, ipAndPortSeparator, logsCollector.privateTcpPort.GetNumber())
	logsCollectorAddress := LogsCollectorAddress(logsCollectorAddressStr)

	return logsCollectorAddress, nil
}
