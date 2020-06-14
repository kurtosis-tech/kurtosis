package testsuite

import (
	"github.com/palantir/stacktrace"
)

// An object that will be passed in to every test, which the user can use to manipulate the results of the test
// NOTE: This object's methods rely on panicking on error because we want to completely abort whatever the test is doing
type TestContext struct {}

func (context TestContext) Fatal(err error) {
	// We rely on panicking here because we want to completely stop whatever the test is doing
	failTest(err)
}

func (context TestContext) AssertTrue(condition bool) {
	if (!condition) {
		failTest(stacktrace.NewError("Assertion failed"))
	}
}

func failTest(err error) {
	panic(err)
}
