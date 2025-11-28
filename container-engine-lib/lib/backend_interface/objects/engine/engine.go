package engine

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type EngineGUID string

// Object that represents POINT-IN-TIME information about an engine server
// Store this object and continue to reference it at your own risk!!!
type Engine struct {
	// Will always be filled out
	guid EngineGUID

	status container.ContainerStatus

	// Public (i.e. external to Kurtosis) information about the engine
	// This information will be nil if the engine isn't running
	publicIpAddr   net.IP
	publicGrpcPort *port_spec.PortSpec
}

func NewEngine(guid EngineGUID, status container.ContainerStatus, publicIpAddr net.IP, publicGrpcPort *port_spec.PortSpec) *Engine {
	return &Engine{guid: guid, status: status, publicIpAddr: publicIpAddr, publicGrpcPort: publicGrpcPort}
}

func (engine *Engine) GetGUID() EngineGUID {
	return engine.guid
}
func (engine *Engine) GetStatus() container.ContainerStatus {
	return engine.status
}
func (engine *Engine) GetPublicIPAddress() net.IP {
	return engine.publicIpAddr
}
func (engine *Engine) GetPublicGRPCPort() *port_spec.PortSpec {
	return engine.publicGrpcPort
}
