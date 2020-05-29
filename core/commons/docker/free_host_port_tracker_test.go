package docker

import (
	"github.com/palantir/stacktrace"
	"testing"
)

const TEST_INTERFACE_IP = "127.0.0.1"

func TestFailingOnInvalidPortRanges(t *testing.T) {
	if _, err := NewFreeHostPortTracker(TEST_INTERFACE_IP, 443, 444); err == nil {
		t.Fatal(stacktrace.NewError("FreeHostPortTracker should fail if port range overlaps with special ports"))
	}
	if _, err := NewFreeHostPortTracker(TEST_INTERFACE_IP, 9651, 9650); err == nil {
		t.Fatal(stacktrace.NewError("FreeHostPortTracker should fail if end is less than beginning."))
	}
}