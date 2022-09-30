package engine_manager

import "github.com/kurtosis-tech/stacktrace"

type EngineStatus string
const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// vvvvvvvvvv Whenever you change these, update the Accept function switch statement! vvvvvvvvvvvvv
	EngineStatus_Stopped                                EngineStatus = "STOPPED"
	EngineStatus_ContainerRunningButServerNotResponding EngineStatus = "CONTAINER_RUNNING_BUT_SERVER_NOT_RESPONDING"
	EngineStatus_Running                                EngineStatus = "RUNNING"
	// ^^^^^^^^^^ Whenever you change these, update the Accept function switch statement! ^^^^^^^^^^^^^
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
)
func (status EngineStatus) Accept(visitor EngineStatusVisitor) error {
	switch status {
	case EngineStatus_Stopped:
		return visitor.VisitStopped()
	case EngineStatus_ContainerRunningButServerNotResponding:
		return visitor.VisitContainerRunningButServerNotResponding()
	case EngineStatus_Running:
		return visitor.VisitRunning()
	default:
		return stacktrace.NewError("No engine status -> visitor function mapping for status '%v'; this is a Kurtosis bug", status)
	}
}

// Visitor interface to force us to exhaustively handle engine statuses
type EngineStatusVisitor interface {
	VisitStopped() error
	VisitContainerRunningButServerNotResponding() error
	VisitRunning() error
}

