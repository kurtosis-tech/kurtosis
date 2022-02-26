package engine

// Represents the state of an engine
//go:generate go run github.com/dmarkham/enumer -type=EngineStatus
type EngineStatus int
const (
	// The engine has been stopped (and cannot be restarted, as engines are single-use)
	EngineStatus_Stopped EngineStatus = iota

	// The engine is running
	EngineStatus_Running
)