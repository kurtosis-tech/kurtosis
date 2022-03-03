package engine

// Represents the state of an engine
//go:generate go run github.com/dmarkham/enumer -trimprefix=EngineStatus_ -transform=snake-upper -type=EngineStatus
type EngineStatus int
const (
	// The engine has been stopped (and cannot be restarted, as engines are single-use)
	EngineStatus_Stopped EngineStatus = iota

	// The engine is running
	EngineStatus_Running
)