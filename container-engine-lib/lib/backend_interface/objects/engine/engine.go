package engine

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

type EngineGUID string

// Object that represents POINT-IN-TIME information about an engine server
// Store this object and continue to reference it at your own risk!!!
type Engine struct {
	// Will always be filled out
	guid EngineGUID

	status container_status.ContainerStatus

	// Public (i.e. external to Kurtosis) information about the engine
	// This information will be nil if the engine isn't running
	publicIpAddr        net.IP
	publicGrpcPort      *port_spec.PortSpec
	publicGrpcProxyPort *port_spec.PortSpec
}

func (engine *Engine) GetGUID() EngineGUID {
	return engine.guid
}
func (engine *Engine) GetStatus() container_status.ContainerStatus {
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
