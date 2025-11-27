package logs_collector

import (
	"fmt"
	"net"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	ipAndPortSeparator = ":"
)

type LogsCollectorGuid string

type FilterParam struct {
	Key   string `yaml:"key,omitempty" json:"key,omitempty"`
	Value string `yaml:"value,omitempty" json:"value,omitempty"`
}

type Filter struct {
	Name   string        `yaml:"name,omitempty" json:"name,omitempty"`
	Match  string        `yaml:"match,omitempty" json:"match,omitempty"`
	Params []FilterParam `yaml:"params,omitempty" json:"params,omitempty"`
}

type Parser map[string]string

// This component is responsible for:
// 1. collecting logs from all services within an enclave
// 2. forwarding these logs to the logs aggregator
type LogsCollector struct {
	status container.ContainerStatus

	// This information will be nil if the logs collector container isn't running
	privateTcpPort  *port_spec.PortSpec
	privateHttpPort *port_spec.PortSpec

	// bridge network ip address
	maybeEnclaveNetworkIpAddress net.IP
	maybeBridgeNetworkIpAddress  net.IP
}

func NewLogsCollector(
	status container.ContainerStatus,
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

func (logsCollector *LogsCollector) GetStatus() container.ContainerStatus {
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

// Returns an address string with format <ip address>:<port> of the logs collector's address within the enclave network
// This address can be used by containers in the enclave to forward logs to the logs collector
func (logsCollector *LogsCollector) GetEnclaveNetworkAddressString() (string, error) {
	if logsCollector.maybeEnclaveNetworkIpAddress == nil {
		return "", stacktrace.NewError("It is impossible to return the logs collector private TCP address because the value of its private IP address is nil")
	}

	if logsCollector.privateTcpPort == nil {
		return "", stacktrace.NewError("It is impossible to return the logs collector private TCP address because the value of its private TCP port spec is nil")
	}

	logsCollectorAddressStr := fmt.Sprintf("%v%v%v", logsCollector.maybeEnclaveNetworkIpAddress, ipAndPortSeparator, logsCollector.privateTcpPort.GetNumber())
	return logsCollectorAddressStr, nil
}
