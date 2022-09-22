package logs_database

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type LogsDatabase struct {
	status container_status.ContainerStatus

	// This information will be nil if the logs database container isn't running
	privateIpAddr   net.IP
	privateHttpPort *port_spec.PortSpec
}

func NewLogsDatabase(
	status container_status.ContainerStatus,
	privateIpAddr net.IP,
	privateHttpPort *port_spec.PortSpec,
) *LogsDatabase {
	return &LogsDatabase{status: status, privateIpAddr: privateIpAddr, privateHttpPort: privateHttpPort}
}

func (logsDatabase *LogsDatabase) GetStatus() container_status.ContainerStatus {
	return logsDatabase.status
}

func (logsDatabase *LogsDatabase) GetPrivateIpAddr() net.IP {
	return logsDatabase.privateIpAddr
}

func (logsDatabase *LogsDatabase) GetPrivateHttpPort() *port_spec.PortSpec {
	return logsDatabase.privateHttpPort
}
