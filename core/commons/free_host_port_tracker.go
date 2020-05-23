package commons

import "github.com/palantir/stacktrace"

type FreeHostPortTracker struct {
	portRangeStart int
	portRangeEnd   int
	takenPorts     map[int]bool
}

func NewFreeHostPortTracker(portRangeStart int, portRangeEnd int) *FreeHostPortTracker {
	portMap := make(map[int]bool)
	return &FreeHostPortTracker{
		portRangeStart: portRangeStart,
		portRangeEnd:   portRangeEnd,
		takenPorts:     portMap,
	}
}

func (hostPortTracker FreeHostPortTracker) GetFreePort() (port int, err error) {
	for port := hostPortTracker.portRangeStart; port < hostPortTracker.portRangeEnd; port++ {
		if _, ok := hostPortTracker.takenPorts[port]; !ok {
			hostPortTracker.takenPorts[port] = true
			return port, nil
		}
	}
	return -1, stacktrace.NewError("There are no more free ports available given the host port range.")
}

func (hostPortTracker FreeHostPortTracker) ReleasePort(port int) (err error) {
	if _, ok := hostPortTracker.takenPorts[port]; ok {
		delete(hostPortTracker.takenPorts, port)
	}
	return nil
}

