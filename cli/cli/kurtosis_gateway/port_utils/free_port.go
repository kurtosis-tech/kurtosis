package port_utils

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"net"
)

const (
	tcpProtocolStr              = "tcp"
	localHostZeroPortBindingStr = "localhost:0"
)

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreeTcpPort() (resultFreePortSpec *port_spec.PortSpec, err error) {
	localHostPortAddress, err := net.ResolveTCPAddr(tcpProtocolStr, localHostZeroPortBindingStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to resolve a tcp address for '%v', instead a non-nil error was returned", localHostZeroPortBindingStr)
	}

	localHostPortListener, err := net.ListenTCP(tcpProtocolStr, localHostPortAddress)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to open tcp listener on localhost, instead a non-nil error was returned")
	}
	defer localHostPortListener.Close()
	// Get port number from the port listener
	portNumber := localHostPortListener.Addr().(*net.TCPAddr).Port
	portNumberUint16 := uint16(portNumber)

	localHostPortSpec, err := port_spec.NewPortSpec(portNumberUint16, port_spec.PortProtocol_TCP)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create a port spec describing a free open port on localhost, instead a non-nil error was returned")
	}

	return localHostPortSpec, nil
}
