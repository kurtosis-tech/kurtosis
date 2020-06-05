package initializer

import (
	"github.com/palantir/stacktrace"
	"gotest.tools/assert"
	"testing"
)

const (
	testName = "TEST"
)

func TestLogTestResult(t *testing.T) {
	assert.Equal(t, logTestResult(testName, nil, true), PASSED, "Expected passed test")
	assert.Equal(t, logTestResult(testName, nil, false), FAILED, "Expected failed test")
	assert.Equal(t, logTestResult(testName, stacktrace.NewError("Test"), false), ERRORED, "Expected errored test")
	assert.Equal(t, logTestResult(testName, stacktrace.NewError("Test"), true), ERRORED, "Expected errored test")
}
