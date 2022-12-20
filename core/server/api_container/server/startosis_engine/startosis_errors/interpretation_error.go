package startosis_errors

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
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
//
// The `cause` field is the underlying error that caused the InterpretationError to be created. The InterpretationError
// message will contain the message of both the InterpretationError and the cause error if present. Adding a `cause` to
// an InterpretationError is handy when you want to surface a Go error to the user. For example, when you read a file
// in Go as part of a Startosis execution thread, you may get a Go error which contains valuable information about what
// went wrong. You can create a InterpretationError wrapping the Go error.
// As mentioned above, it is still discouraged to wrap an error what was returned by stacktrace.Propagate(...),
// as this error message will contain the Go stacktrace the user doesn't really care about. Basically, make sure the
// error is valuable enough to the end user before wrapping it.
// It is also doable to wrap an InterpretationError into another InterpretationError. This is useful when you want to
// surface context about the root error as well as where it ended up failing. Since both error messages are
// surfaced to the user, it will have the info from both errors. However, when the root error is explicit enough, it's
// not necessary to wrap it. In case of a doubt, wrapping should be the default choice (explicit over implicit!)
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
		msg:        fmt.Sprintf(msg, args...),
		cause:      nil,
		stacktrace: nil,
	}
}

func WrapWithInterpretationError(err error, msg string, args ...interface{}) *InterpretationError {
	return &InterpretationError{
		msg:        fmt.Sprintf(msg, args...),
		cause:      err,
		stacktrace: nil,
	}
}

func NewInterpretationErrorFromStacktrace(stacktrace []CallFrame) *InterpretationError {
	return &InterpretationError{
		msg:        "",
		cause:      nil,
		stacktrace: stacktrace,
	}
}

func NewInterpretationErrorWithCustomMsg(stacktrace []CallFrame, msg string, args ...interface{}) *InterpretationError {
	return &InterpretationError{
		msg:        fmt.Sprintf(msg, args...),
		cause:      nil,
		stacktrace: stacktrace,
	}
}

func NewInterpretationErrorWithCauseAndCustomMsg(err error, stacktrace []CallFrame, msg string, args ...interface{}) *InterpretationError {
	return &InterpretationError{
		msg:        fmt.Sprintf(msg, args...),
		cause:      err,
		stacktrace: stacktrace,
	}
}

func (err *InterpretationError) ToAPIType() *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError {
	return binding_constructors.NewStarlarkInterpretationError(err.Error())
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
