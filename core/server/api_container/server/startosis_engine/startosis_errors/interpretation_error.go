package startosis_errors

import (
	"fmt"
	"strings"
)

const (
	errorDefaultMsg  = "/!\\ Errors interpreting Startosis script"
	stacktracePrefix = "\tat "
)

// InterpretationError is an error thrown by the Startosis interpreter.
// This is due to errors made by the Startosis script author and should be returned in a nice and intelligible way.
//
// The `stacktrace` field here should be relative to the Startosis script, NOT the Go code interpreting it.
// Using stacktrace.Propagate(...) to generate those startosis_errors is therefore not recommended.
type InterpretationError struct {
	// The error message
	msg string

	// Optional stacktrace
	stacktrace []CallFrame
}

func NewInterpretationError(msg string) *InterpretationError {
	return &InterpretationError{
		msg: msg,
	}
}

func NewInterpretationErrorFromStacktrace(stacktrace []CallFrame) *InterpretationError {
	return &InterpretationError{
		msg:        "",
		stacktrace: stacktrace,
	}
}

func NewInterpretationErrorWithCustomMsg(msg string, stacktrace []CallFrame) *InterpretationError {
	return &InterpretationError{
		msg:        msg,
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
	for _, stacktraceElement := range err.stacktrace {
		serializedError.WriteString(fmt.Sprintf("\n%s%s", stacktracePrefix, stacktraceElement.String()))
	}
	return serializedError.String()
}
