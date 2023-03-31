package out

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveFilePathFromErrorMessage(t *testing.T) {
	stacktraceErr := createDummyStackTraceWithNonEmptyMsg()
	errorClean := removeFilePathFromErrorMessage(stacktraceErr.Error())
	expectedValue := "this is propagated error\n  Caused by: Error: this is base error"
	require.Equal(t, expectedValue, errorClean.Error())

	stacktraceErrEmpty := createDummyStackTraceWithEmptyMsg()
	errorClean = removeFilePathFromErrorMessage(stacktraceErrEmpty.Error())
	expectedValue = "  Caused by: Error: this is base error"
	require.Equal(t, expectedValue, errorClean.Error())
}

func createDummyStackTraceWithNonEmptyMsg() error {
	baseError := stacktrace.NewError("Error: this is base error")
	propagatedError := stacktrace.Propagate(baseError, "this is propagated error")
	return propagatedError
}

func createDummyStackTraceWithEmptyMsg() error {
	baseError := stacktrace.NewError("Error: this is base error")
	propagatedError := stacktrace.Propagate(baseError, "")
	return propagatedError
}
