package logs_database

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type LogsDatabase struct {
	status container_status.ContainerStatus

	// This information will be nil if the logs database container isn't running
	maybePrivateIpAddr net.IP
	privateHttpPort    *port_spec.PortSpec
}

func NewLogsDatabase(
	status container_status.ContainerStatus,
	maybePrivateIpAddr net.IP,
	privateHttpPort *port_spec.PortSpec,
) *LogsDatabase {
	return &LogsDatabase{status: status, maybePrivateIpAddr: maybePrivateIpAddr, privateHttpPort: privateHttpPort}
}

func (logsDatabase *LogsDatabase) GetStatus() container_status.ContainerStatus {
	return logsDatabase.status
}

func (logsDatabase *LogsDatabase) GetMaybePrivateIpAddr() net.IP {
	return logsDatabase.maybePrivateIpAddr
}

func (logsDatabase *LogsDatabase) GetPrivateHttpPort() *port_spec.PortSpec {
	return logsDatabase.privateHttpPort
}
