package testsuite

import (
	"github.com/palantir/stacktrace"
	"testing"
)

func TestFatalOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("The code did not panic when it should")
		}
	}()
	TestContext{}.Fatal(stacktrace.NewError("Test error"))
}

func TestFatalOnAssertion(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("The code did not panic when it should")
		}
	}()
	TestContext{}.AssertTrue(false)
}
