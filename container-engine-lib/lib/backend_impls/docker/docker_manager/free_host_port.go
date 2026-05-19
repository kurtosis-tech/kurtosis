package docker_manager

import (
	"fmt"
	"net"

	"github.com/kurtosis-tech/stacktrace"
)

const (
	probeAnyInterface = "0.0.0.0"
	probeAnyPort      = 0
)

// GetFreeTcpHostPort asks the kernel for a free TCP port on the host by binding
// a probe listener to ":0" and reading back the port the kernel assigned.
//
// The probe socket is closed before returning, so the port may be re-claimed by
// another process between this call and any subsequent bind. Callers must be
// prepared to retry. Compared to letting Docker auto-pick the host port at
// container-create time, this narrows but does not eliminate the race window.
func GetFreeTcpHostPort() (uint16, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", probeAnyInterface, probeAnyPort))
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to resolve TCP probe address")
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to open TCP probe listener for free-port discovery")
	}
	defer listener.Close()
	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, stacktrace.NewError("TCP probe listener address was not a *net.TCPAddr; this is a bug")
	}
	return uint16(tcpAddr.Port), nil
}

// GetFreeUdpHostPort asks the kernel for a free UDP port on the host. See
// GetFreeTcpHostPort for race-window caveats.
func GetFreeUdpHostPort() (uint16, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", probeAnyInterface, probeAnyPort))
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to resolve UDP probe address")
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to open UDP probe listener for free-port discovery")
	}
	defer conn.Close()
	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return 0, stacktrace.NewError("UDP probe listener address was not a *net.UDPAddr; this is a bug")
	}
	return uint16(udpAddr.Port), nil
}
