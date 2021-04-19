/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package free_host_port_binding_supplier

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
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
	hostIpAddr string
	interfaceIpAddr string
	protocol string
	portRangeStart uint32
	portRangeEnd   uint32
	takenPorts     map[uint32]bool
	mutex		*sync.Mutex
}

// TODO Protocol shouldn't be an argument to the constructor, but to the GetFreePort call. This struct should maintain
//  multiple usedPort mappings, one per protocol
/*
Creates a supplier that will dole out free host port bindings, tracking which ones are beign used

WARNING: The interface should match the interface ports are bound on! If not, the host port tracker will return a "free" port
that is really bound on another interface


Args:
	hostIpAddr: The IP address of the host to attempt to connect to to verify port if a port is free (since this code will
		be running inside a Docker container, and to the container the host machine will just be another IP)
	interfaceIpAddr: IP address of the host interface in the returned PortBinding objects
	protocol: The protocol of the port bindings supplied
	portRangeStart: Start of the range of ports that will be doled out
	portRangeEnd: EXCLUSIVE end of the range of ports that will be doled out
	takenPorts: "Set" of ports which are known to be taken and shouldn't be doled out
 */
func NewFreeHostPortBindingSupplier(
		hostIpAddr string,
		interfaceIpAddr string,
		protocol string,
		portRangeStart uint32,
		portRangeEnd uint32,
		takenPorts map[uint32]bool) (freeHostPortTracker *FreeHostPortBindingSupplier, err error) {
	if portRangeEnd <= portRangeStart {
		return nil, stacktrace.NewError("Port range end '%v' is <= port range start '%v'", portRangeEnd, portRangeStart)
	}
	if !isPortValid(portRangeStart) || !isPortValid(portRangeEnd - 1) {
		return nil, stacktrace.NewError("Port range start and end must be in range [%v, %v)", validPortRangeStart, validPortRangeEnd)
	}
	takenPortsCopy := map[uint32]bool{}
	for takenPortInt := range takenPorts {
		takenPortsCopy[takenPortInt] = true
	}
	return &FreeHostPortBindingSupplier{
		hostIpAddr: hostIpAddr,
		interfaceIpAddr: interfaceIpAddr,
		protocol: protocol,
		portRangeStart: portRangeStart,
		portRangeEnd:   portRangeEnd,
		takenPorts:     takenPortsCopy,
		mutex: &sync.Mutex{},
	}, nil
}

func (supplier *FreeHostPortBindingSupplier) GetFreePortBinding() (portBinding nat.PortBinding, err error) {
	supplier.mutex.Lock()
	defer supplier.mutex.Unlock()

	for portInt := supplier.portRangeStart; portInt < supplier.portRangeEnd; portInt++ {
		if _, found := supplier.takenPorts[portInt]; found {
			continue
		}

		hostAndPort := fmt.Sprintf("%v:%v", supplier.hostIpAddr, portInt)
		conn, err := net.DialTimeout(supplier.protocol, hostAndPort, dialTimeout)
		if err == nil {
			// NOTE: When we detect a port that's taken but not one of ours, this will PERMANENTLY blacklist that port
			// We *could* check it every time (in case the third party releases it)... but then we'd need to check it every time
			supplier.takenPorts[portInt] = true
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
			HostIP:   supplier.interfaceIpAddr,
			HostPort: fmt.Sprint(portInt),
		}
		supplier.takenPorts[portInt] = true
		return binding, nil
	}

	return nat.PortBinding{}, stacktrace.NewError(
		"There are no more free ports available on interface '%v' on protocol '%v' in range [%v, %v)",
		supplier.interfaceIpAddr,
		supplier.protocol,
		supplier.portRangeStart,
		supplier.portRangeEnd)
}

func (supplier FreeHostPortBindingSupplier) GetInterfaceIp() string {
	return supplier.interfaceIpAddr
}

func (supplier FreeHostPortBindingSupplier) GetProtocol() string {
	return supplier.protocol
}

func (supplier FreeHostPortBindingSupplier) GetPortRangeStart() uint32 {
	return supplier.portRangeStart
}

func (supplier FreeHostPortBindingSupplier) GetPortRangeEnd() uint32 {
	return supplier.portRangeEnd
}

func (supplier *FreeHostPortBindingSupplier) GetTakenPorts() map[uint32]bool {
	supplier.mutex.Lock()
	defer supplier.mutex.Unlock()

	// Defensive copy
	result := map[uint32]bool{}
	for portInt, boolVal := range supplier.takenPorts {
		result[portInt] = boolVal
	}

	return result
}

// NOTE: Uncomment and use this if we ever run into an issue with running out of ports
/*
func (hostPortTracker FreeHostPortBindingSupplier) ReleasePort(port int) {
	delete(hostPortTracker.takenPorts, port)
}
 */

func isPortValid(port uint32) bool {
	return port >= validPortRangeStart && port <= validPortRangeEnd
}
