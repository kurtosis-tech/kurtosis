package engine

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

// Object that represents POINT-IN-TIME information about an engine server
// Store this object and continue to reference it at your own risk!!!
type Engine struct {
	// Will always be filled out
	id string

	status EngineStatus

	// Public (i.e. external to Kurtosis) information about the engine
	// This information will be nil if the engine isn't running
	publicIpAddr net.IP
	publicGrpcPort *port_spec.PortSpec
	publicGrpcProxyPort *port_spec.PortSpec
}

func NewEngine(id string, status EngineStatus, publicIpAddr net.IP, publicGrpcPort *port_spec.PortSpec, publicGrpcProxyPort *port_spec.PortSpec) *Engine {
	return &Engine{id: id, status: status, publicIpAddr: publicIpAddr, publicGrpcPort: publicGrpcPort, publicGrpcProxyPort: publicGrpcProxyPort}
}

func (engine *Engine) GetID() string {
	return engine.id
}
func (engine *Engine) GetStatus() EngineStatus {
	return engine.status
}
func (engine *Engine) GetPublicIPAddress() net.IP {
	return engine.publicIpAddr
}
func (engine *Engine) GetPublicGRPCPort() *port_spec.PortSpec {
	return engine.publicGrpcPort
}
func (engine *Engine) GetPublicGRPCProxyPortNum() *port_spec.PortSpec {
	return engine.publicGrpcProxyPort
}
