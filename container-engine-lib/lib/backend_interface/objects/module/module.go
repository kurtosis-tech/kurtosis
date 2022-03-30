package module

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
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
	publicIp net.IP
	publicPort *port_spec.PortSpec
}

func NewModule(enclaveId enclave.EnclaveID, id ModuleID, guid ModuleGUID, status container_status.ContainerStatus, privateIp net.IP, privatePort *port_spec.PortSpec, publicIp net.IP, publicPort *port_spec.PortSpec) *Module {
	return &Module{enclaveId: enclaveId, id: id, guid: guid, status: status, privateIp: privateIp, privatePort: privatePort, publicIp: publicIp, publicPort: publicPort}
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

func (module *Module) GetPrivateIp() net.IP {
	return module.privateIp
}

func (module *Module) GetPrivatePort() *port_spec.PortSpec {
	return module.privatePort
}

func (module *Module) GetPublicIp() net.IP {
	return module.publicIp
}

func (module *Module) GetPublicPort() *port_spec.PortSpec {
	return module.publicPort
}
