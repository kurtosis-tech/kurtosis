package engine

type Engine struct {
	// Will always be filled out
	id string

	status EngineStatus

	// The public IP address that the engine is reachable at
	publicIpAddr string

	// The public GRPC port the engine
	publicGrpcPort uint16

	publicGrpcProxyPort uint16
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