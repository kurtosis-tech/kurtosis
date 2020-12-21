/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_runner

import (
	"github.com/palantir/stacktrace"
	"testing"
)

const (
	testInterfaceIp = "127.0.0.1"
	testProtocol = "tcp"
)

func TestFailingOnInvalidPortRanges(t *testing.T) {
	if _, err := NewFreeHostPortBindingSupplier(testInterfaceIp, testProtocol, 443, 444); err == nil {
		t.Fatal(stacktrace.NewError("FreeHostPortBindingSupplier should fail if port range overlaps with special ports"))
	}
	if _, err := NewFreeHostPortBindingSupplier(testInterfaceIp, testProtocol, 9651, 9650); err == nil {
		t.Fatal(stacktrace.NewError("FreeHostPortBindingSupplier should fail if end is less than beginning."))
	}
}