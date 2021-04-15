/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_runner

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	validPortRangeStart = 1024
	validPortRangeEnd   = 65535

	// This is the domain name of the host machine from inside the Docker container
	// See: https://stackoverflow.com/questions/31324981/how-to-access-host-port-from-docker-container
	hostMachineDomainInsideDocker = "host.docker.internal"

	// How long we'll wait in trying to dial the host machine's port to see if it's available (though
	// the timeout shouldn't usually be hit unless there are firewall rules in place - instead, we should
	// get an immediate "connection refused")
	dialTimeout = 100 * time.Millisecond

	connectionRefusedErrorSubstring = "connection refused"
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
that is really bound on another interface


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
		if _, found := tracker.takenPorts[portInt]; found {
			continue
		}

		// NOTE: We'll need to change this to support UDP
		hostAndPort := fmt.Sprintf("%v:%v", hostMachineDomainInsideDocker, portInt)
		conn, err := net.DialTimeout("tcp", hostAndPort, dialTimeout)
		if err == nil {
			// NOTE: When we detect a port that's taken but not one of ours, this will PERMANENTLY blacklist that port
			// We *could* check it every time (in case the third party releases it)... but then we'd need to check it every time
			tracker.takenPorts[portInt] = true
			if conn != nil {
				conn.Close()
			}
			continue
		}

		// This is janky, but I tried a bunch of different ways to detect if the error is connection-refused, and this is the only
		// one that actually worked ~ ktoday, 2021-04-15
		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(connectionRefusedErrorSubstring)) {
			logrus.Debugf("Unrecognized error '%v' when dialing %v", err, hostAndPort)
			continue
		}

		binding := nat.PortBinding{
			HostIP:   tracker.interfaceIpAddr,
			HostPort: strconv.Itoa(portInt),
		}
		tracker.takenPorts[portInt] = true
		return binding, nil
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
