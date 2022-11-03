package startosis_errors

import (
	"fmt"
	"strings"
)

const (
	errorDefaultMsg  = "/!\\ Errors interpreting Startosis script"
	stacktracePrefix = "\tat "
	causedByPrefix   = "\tCaused by: "
)

// InterpretationError is an error thrown by the Startosis interpreter.
// This is due to errors made by the Startosis script author and should be returned in a nice and intelligible way.
//
// The `stacktrace` field here should be relative to the Startosis script, NOT the Go code interpreting it.
// Using stacktrace.Propagate(...) to generate those startosis_errors is therefore not recommended.
type InterpretationError struct {
	// The error message
	msg string

	// Optional cause
	cause error

	// Optional stacktrace
	stacktrace []CallFrame
}

func NewInterpretationError(msg string, args ...interface{}) *InterpretationError {
	return &InterpretationError{
		msg: fmt.Sprintf(msg, args...),
	}
}

func WrapError(err error, msg string, args ...interface{}) *InterpretationError {
	return &InterpretationError{
		msg:   fmt.Sprintf(msg, args...),
		cause: err,
	}
}

func NewInterpretationErrorFromStacktrace(stacktrace []CallFrame) *InterpretationError {
	return &InterpretationError{
		msg:        "",
		stacktrace: stacktrace,
	}
}

func NewInterpretationErrorWithCustomMsg(stacktrace []CallFrame, msg string, args ...interface{}) *InterpretationError {
	return &InterpretationError{
		msg:        fmt.Sprintf(msg, args...),
		stacktrace: stacktrace,
	}
}

func (err *InterpretationError) Error() string {
	var serializedError strings.Builder
	if err.msg == "" {
		serializedError.WriteString(errorDefaultMsg)
	} else {
		serializedError.WriteString(err.msg)
	}
	if err.cause != nil {
		serializedError.WriteString(fmt.Sprintf("\n%s%s", causedByPrefix, err.cause.Error()))
	}
	for _, stacktraceElement := range err.stacktrace {
		serializedError.WriteString(fmt.Sprintf("\n%s%s", stacktracePrefix, stacktraceElement.String()))
	}
	return serializedError.String()
}
