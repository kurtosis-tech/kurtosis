package module

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type ModuleID string
type ModuleGUID string

// Object that represents POINT-IN-TIME information about a Kurtosis Module
// Store this object and continue to reference it at your own risk!!!
type Module struct {
	enclaveId enclave.EnclaveID
	id ModuleID
	guid ModuleGUID
	status container_status.ContainerStatus
	privateIp net.IP
	privatePort *port_spec.PortSpec

	// Will be nil if the module isn't running, or if the backend is Kubernetes
	maybePublicIp net.IP

	// Will be nil if the module isn't running, or if the backend is Kubernetes
	maybePublicPort *port_spec.PortSpec
}

func NewModule(enclaveId enclave.EnclaveID, id ModuleID, guid ModuleGUID, status container_status.ContainerStatus, privateIp net.IP, privatePort *port_spec.PortSpec, maybePublicIp net.IP, maybePublicPort *port_spec.PortSpec) *Module {
	return &Module{enclaveId: enclaveId, id: id, guid: guid, status: status, privateIp: privateIp, privatePort: privatePort, maybePublicIp: maybePublicIp, maybePublicPort: maybePublicPort}
}

func (module *Module) GetEnclaveID() enclave.EnclaveID {
	return module.enclaveId
}

func (module *Module) GetID() ModuleID {
	return module.id
}

func (module *Module) GetGUID() ModuleGUID {
	return module.guid
}

func (module *Module) GetStatus() container_status.ContainerStatus {
	return module.status
}

func (module *Module) GetPrivateIP() net.IP {
	return module.privateIp
}

func (module *Module) GetPrivatePort() *port_spec.PortSpec {
	return module.privatePort
}

func (module *Module) GetMaybePublicIP() net.IP {
	return module.maybePublicIp
}

func (module *Module) GetMaybePublicPort() *port_spec.PortSpec {
	return module.maybePublicPort
}
