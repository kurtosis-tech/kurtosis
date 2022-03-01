package engine

import "net"

// Object that represents POINT-IN-TIME information about an engine server
// Store this object and continue to reference it at your own risk!!!
type Engine struct {
	// Will always be filled out
	id string

	status EngineStatus

	// Public (i.e. external to Kurtosis) information about the engine
	publicIpAddr net.IP
	publicGrpcPortNum uint16
	publicGrpcProxyPortNum uint16
}

func NewEngine(id string, status EngineStatus, publicIpAddr net.IP, publicGrpcPortNum uint16, publicGrpcProxyPortNum uint16) *Engine {
	return &Engine{id: id, status: status, publicIpAddr: publicIpAddr, publicGrpcPortNum: publicGrpcPortNum, publicGrpcProxyPortNum: publicGrpcProxyPortNum}
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
func (engine *Engine) GetPublicGRPCPortNum() uint16 {
	return engine.publicGrpcPortNum
}
func (engine *Engine) GetPublicGRPCProxyPortNum() uint16 {
	return engine.publicGrpcProxyPortNum
}
