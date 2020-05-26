package docker

import (
	"github.com/palantir/stacktrace"
	"net"
	"strconv"
)

const VALID_PORT_RANGE_START = 1024
const VALID_PORT_RANGE_END = 65535

type FreeHostPortTracker struct {
	portRangeStart int
	portRangeEnd   int
	takenPorts     map[int]bool
}

func NewFreeHostPortTracker(portRangeStart int, portRangeEnd int) (freeHostPortTracker *FreeHostPortTracker, err error) {
	portMap := make(map[int]bool)
	if portRangeEnd <= portRangeStart {
		return nil, stacktrace.NewError("FreeHostPortTracker requires end port range greater than start port range.")
	}
	if !isPortValid(portRangeStart) || !isPortValid(portRangeEnd) {
		return nil, stacktrace.NewError("FreeHostPortTracker requires port range between %v and %v, inclusive.", VALID_PORT_RANGE_START, VALID_PORT_RANGE_END)
	}
	return &FreeHostPortTracker{
		portRangeStart: portRangeStart,
		portRangeEnd:   portRangeEnd,
		takenPorts:     portMap,
	}, nil
}

func (hostPortTracker FreeHostPortTracker) GetFreePort() (port int, err error) {
	for port := hostPortTracker.portRangeStart; port < hostPortTracker.portRangeEnd; port++ {
		if _, ok := hostPortTracker.takenPorts[port]; !ok {
			if isPortFree(port) {
				hostPortTracker.takenPorts[port] = true
				return port, nil
			}
		}
	}
	return -1, stacktrace.NewError("There are no more free ports available given the host port range.")
}

func (hostPortTracker FreeHostPortTracker) ReleasePort(port int) {
	delete(hostPortTracker.takenPorts, port)
}

func isPortValid(port int) bool {
	return port >= VALID_PORT_RANGE_START && port <= VALID_PORT_RANGE_END
}

func isPortFree(port int) bool {
	ln, err := net.Listen("tcp", ":" + strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
