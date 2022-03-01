package engine

// Object that represents POINT-IN-TIME information about an engine server
// Store this object and continue to reference it at your own risk!!!
type Engine struct {
	// Will always be filled out
	id string

	status EngineStatus

	// Private (i.e. internal-to-the-enclave) information about the engine
	privateIpAddr string  // TODO make this a net.IP object?
	privateGrpcPortNum uint16
	privateGrpcProxyPortNum uint16

	// Public (i.e. external to Kurtosis) information about the engine
	publicIpAddr string  // TODO make this a net.IP object?
	publicGrpcPortNum uint16
	publicGrpcProxyPortNum uint16
}
func NewEngine(id string) *Engine {
	return &Engine{id: id}
}
func (engine *Engine) GetID() string {
	return engine.id
}
func (engine *Engine) GetStatus() EngineStatus {
	return engine.status
}
func (engine *Engine) GetPrivateIPAddress() string {
	return engine.privateIpAddr
}
func (engine *Engine) GetPrivateGRPCPortNum() uint16 {
	return engine.privateGrpcPortNum
}
func (engine *Engine) GetPrivateGRPCProxyPortNum() uint16 {
	return engine.privateGrpcProxyPortNum
}
func (engine *Engine) GetPublicIPAddress() string {
	return engine.publicIpAddr
}
func (engine *Engine) GetPublicGRPCPortNum() uint16 {
	return engine.publicGrpcPortNum
}
func (engine *Engine) GetPublicGRPCProxyPortNum() uint16 {
	return engine.publicGrpcProxyPortNum
}
