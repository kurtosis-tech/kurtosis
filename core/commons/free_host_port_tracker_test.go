package commons

import (
	"github.com/palantir/stacktrace"
	"testing"
)

func TestFailingOnInvalidPortRanges(t *testing.T) {
	if _, err := NewFreeHostPortTracker(443, 444); err == nil {
		t.Fatal(stacktrace.NewError("FreeHostPortTracker should fail if port range overlaps with special ports"))
	}
	if _, err := NewFreeHostPortTracker(9651, 9650); err == nil {
		t.Fatal(stacktrace.NewError("FreeHostPortTracker should fail if end is less than beginning."))
	}
}