package startosis_errors

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/binding_constructors"
	"strings"
)

const (
	validationErrorDefaultMsg = "/!\\ Errors validating Starlark script"
)

// ValidationError is an error thrown by the Starlark validator.
// This is due to errors made by the Starlark script author and should be returned in a nice and intelligible way.
//
// The `stacktrace` field here should be relative to the Startosis script, NOT the Go code interpreting it.
// Using stacktrace.Propagate(...) to generate those startosis_errors is therefore not recommended.
//
// The `cause` field is the underlying error that caused the ValidationError to be created. The ValidationError
// message will contain the message of both the ValidationError and the cause error if present. Adding a `cause` to
// an ValidationError is handy when you want to surface a Go error to the user. For example, when you read a file
// in Go as part of a Startosis execution thread, you may get a Go error which contains valuable information about what
// went wrong. You can create a ValidationError wrapping the Go error.
// As mentioned above, it is still discouraged to wrap an error that was returned by stacktrace.Propagate(...),
// as this error message will contain the Go stacktrace the user doesn't really care about. Basically, make sure the
// error is valuable enough to the end user before wrapping it.
// It is also doable to wrap a ValidationError into another ValidationError. This is useful when you want to
// surface context about the root error as well as where it ended up failing. Since both error messages are
// surfaced to the user, it will have the info from both errors. However, when the root error is explicit enough, it's
// not necessary to wrap it. In case of a doubt, wrapping should be the default choice (explicit over implicit!)
type ValidationError struct {
	// The error message
	msg string

	// Optional cause
	cause error
}

func NewValidationError(msg string, args ...interface{}) *ValidationError {
	return &ValidationError{
		msg:   fmt.Sprintf(msg, args...),
		cause: nil,
	}
}

func WrapWithValidationError(err error, msg string, args ...interface{}) *ValidationError {
	return &ValidationError{
		msg:   fmt.Sprintf(msg, args...),
		cause: err,
	}
}

func (err *ValidationError) ToAPIType() *kurtosis_core_rpc_api_bindings.StarlarkValidationError {
	return binding_constructors.NewStarlarkValidationError(err.Error())
}

func (err *ValidationError) Error() string {
	var serializedError strings.Builder
	if err.msg == "" {
		serializedError.WriteString(validationErrorDefaultMsg)
	} else {
		serializedError.WriteString(err.msg)
	}
	if err.cause != nil {
		serializedError.WriteString(fmt.Sprintf("\n%s%s", causedByPrefix, err.cause.Error()))
	}
	return serializedError.String()
}
