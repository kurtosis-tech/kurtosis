/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package free_host_port_binding_supplier

import (
	"github.com/palantir/stacktrace"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

const (
	localhostIp = "127.0.0.1"
	testInterfaceIp = localhostIp
	testProtocol = "tcp"
)

func TestFailingOnInvalidPortRanges(t *testing.T) {
	if _, err := NewFreeHostPortBindingSupplier(localhostIp, testInterfaceIp, testProtocol, 443, 444, map[uint32]bool{}); err == nil {
		t.Fatal(stacktrace.NewError("FreeHostPortBindingSupplier should fail if port range overlaps with special ports"))
	}
	if _, err := NewFreeHostPortBindingSupplier(localhostIp, testInterfaceIp, testProtocol, 9651, 9650, map[uint32]bool{}); err == nil {
		t.Fatal(stacktrace.NewError("FreeHostPortBindingSupplier should fail if end is less than beginning."))
	}
}

// NOTE: If these ports are taken on the user's machine, this test will spuriously fail!
func TestPreTakenPortsAreSkipped(t *testing.T) {
	takenPorts := map[uint32]bool{
		9800: true,
		9801: true,
	}
	supplier, err := NewFreeHostPortBindingSupplier(localhostIp, testInterfaceIp, testProtocol, 9800, 9803, takenPorts)
	if err != nil {
		t.Fatal(stacktrace.Propagate(err, "An unexpected error occurred creating the free host port binding supplier"))
	}
	binding, err := supplier.GetFreePortBinding()
	if err != nil {
		t.Fatal(stacktrace.Propagate(err, "An unexpected error occurred getting the binding"))
	}
	assert.Equal(t, testInterfaceIp, binding.HostIP)
	bindingPortUint, err := strconv.ParseUint(binding.HostPort, 10, 32)
	if err != nil {
		t.Fatal(stacktrace.Propagate(err, "The returned host port binding port string '%v' couldn't be converted to an int", binding.HostPort))
	}
	assert.Equal(t, uint32(9802), uint32(bindingPortUint))
}
