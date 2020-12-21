/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_runner

import (
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"net"
	"strconv"
)

const (
	validPortRangeStart = 1024
	validPortRangeEnd   = 65535
)

type FreeHostPortBindingSupplier struct {
	interfaceIpAddr string
	protocol string
	portRangeStart int
	portRangeEnd   int
	takenPorts     map[int]bool
}

/*
Creates a supplier that will dole out free host port bindings, tracking which ones are beign used

WARNING: The interface should match the interface ports are bound on! If not, the host port tracker will return a "free" port
that is really bound on another interface.

Args:
	interfaceIpAddr: IP address of the interface that the ports will be bound on
	protocol: The protocol of the port bindings supplied
	portRangeStart: Start of the range of ports that will be doled out
	portRangeEnd: End of teh range of ports that will be doled out
 */
func NewFreeHostPortBindingSupplier(interfaceIpAddr string, protocol string, portRangeStart int, portRangeEnd int) (freeHostPortTracker *FreeHostPortBindingSupplier, err error) {
	portMap := make(map[int]bool)
	if portRangeEnd <= portRangeStart {
		return nil, stacktrace.NewError("FreeHostPortBindingSupplier requires end port range greater than start port range.")
	}
	if !isPortValid(portRangeStart) || !isPortValid(portRangeEnd) {
		return nil, stacktrace.NewError("FreeHostPortBindingSupplier requires port range between %v and %v, inclusive.", validPortRangeStart, validPortRangeEnd)
	}
	return &FreeHostPortBindingSupplier{
		interfaceIpAddr: interfaceIpAddr,
		protocol: protocol,
		portRangeStart: portRangeStart,
		portRangeEnd:   portRangeEnd,
		takenPorts:     portMap,
	}, nil
}

func (tracker FreeHostPortBindingSupplier) GetFreePortBinding() (portBinding nat.PortBinding, err error) {
	for portInt := tracker.portRangeStart; portInt < tracker.portRangeEnd; portInt++ {
		if _, found := tracker.takenPorts[portInt]; !found {
			if isPortFree(tracker.interfaceIpAddr, tracker.protocol, portInt) {
				binding := nat.PortBinding{
					HostIP:   tracker.interfaceIpAddr,  // I guess TCP is the default???
					HostPort: strconv.Itoa(portInt),
				}
				tracker.takenPorts[portInt] = true
				return binding, nil
			}
		}
	}

	return nat.PortBinding{}, stacktrace.NewError(
		"There are no more free ports available on interface '%v' on protcol '%v' in range %v-%v",
		tracker.interfaceIpAddr,
		tracker.protocol,
		tracker.portRangeStart,
		tracker.portRangeEnd)
}

// NOTE: Uncomment and use this if we ever run into an issue with running out of ports
/*
func (hostPortTracker FreeHostPortBindingSupplier) ReleasePort(port int) {
	delete(hostPortTracker.takenPorts, port)
}
 */

func isPortValid(port int) bool {
	return port >= validPortRangeStart && port <= validPortRangeEnd
}

func isPortFree(interfaceIpAddr string, protocol string, port int) bool {
	ln, err := net.Listen(protocol, interfaceIpAddr + ":" + strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
