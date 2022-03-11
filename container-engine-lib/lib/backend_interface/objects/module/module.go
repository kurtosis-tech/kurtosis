package module

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type ModuleID string
type ModuleGUID string

// Object that represents POINT-IN-TIME information about a Kurtosis Module
// Store this object and continue to reference it at your own risk!!!
type Module struct {
	id ModuleID
	guid ModuleGUID
	privateIp net.IP
	privatePort *port_spec.PortSpec
	publicIp net.IP
	publicPort *port_spec.PortSpec
}

func NewModule(id ModuleID, guid ModuleGUID, privateIp net.IP, privatePort *port_spec.PortSpec, publicIp net.IP, publicPort *port_spec.PortSpec) *Module {
	return &Module{id: id, guid: guid, privateIp: privateIp, privatePort: privatePort, publicIp: publicIp, publicPort: publicPort}
}

func (module *Module) GetID() ModuleID {
	return module.id
}

func (module *Module) GetGUID() ModuleGUID {
	return module.guid
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
