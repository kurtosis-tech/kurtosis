package docker

import (
	"github.com/palantir/stacktrace"
	"net"
	"strconv"
)

const VALID_PORT_RANGE_START = 1024
const VALID_PORT_RANGE_END = 65535

type FreeHostPortTracker struct {
	interfaceIpAddr string
	portRangeStart int
	portRangeEnd   int
	takenPorts     map[int]bool
}

/*
Creates a new host port tracker that will track which ports are being used, listening on the given IP address
NOTE: The interface should match the interface ports are bound on! If not, the host port tracker will return a "free" port
that is really bound on another interface.
 */
func NewFreeHostPortTracker(interfaceIpAddr string, portRangeStart int, portRangeEnd int) (freeHostPortTracker *FreeHostPortTracker, err error) {
	portMap := make(map[int]bool)
	if portRangeEnd <= portRangeStart {
		return nil, stacktrace.NewError("FreeHostPortTracker requires end port range greater than start port range.")
	}
	if !isPortValid(portRangeStart) || !isPortValid(portRangeEnd) {
		return nil, stacktrace.NewError("FreeHostPortTracker requires port range between %v and %v, inclusive.", VALID_PORT_RANGE_START, VALID_PORT_RANGE_END)
	}
	return &FreeHostPortTracker{
		interfaceIpAddr: interfaceIpAddr,
		portRangeStart: portRangeStart,
		portRangeEnd:   portRangeEnd,
		takenPorts:     portMap,
	}, nil
}

func (hostPortTracker FreeHostPortTracker) GetFreePort() (port int, err error) {
	for port := hostPortTracker.portRangeStart; port < hostPortTracker.portRangeEnd; port++ {
		if _, ok := hostPortTracker.takenPorts[port]; !ok {
			if isPortFree(hostPortTracker.interfaceIpAddr, port) {
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

func isPortFree(interfaceIpAddr string, port int) bool {
	ln, err := net.Listen("tcp", interfaceIpAddr + ":" + strconv.Itoa(port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
