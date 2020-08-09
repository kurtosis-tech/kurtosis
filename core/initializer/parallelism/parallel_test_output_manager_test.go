package parallelism

import (
	"github.com/palantir/stacktrace"
	"gotest.tools/assert"
	"testing"
)


func TestLogTestResult(t *testing.T) {
	assert.Equal(t, getTestStatusFromResult(nil, true), PASSED, "Expected passed test")
	assert.Equal(t, getTestStatusFromResult(nil, false), FAILED, "Expected failed test")
	assert.Equal(t, getTestStatusFromResult(stacktrace.NewError("Test"), false), ERRORED, "Expected errored test")
	assert.Equal(t, getTestStatusFromResult(stacktrace.NewError("Test"), true), ERRORED, "Expected errored test")
}
