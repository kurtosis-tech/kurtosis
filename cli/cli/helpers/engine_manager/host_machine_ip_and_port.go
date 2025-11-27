package engine_manager

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/lib/kurtosis_context"
	"net"
)

const (
	localHostIpStr = "127.0.0.1"
)

type hostMachineIpAndPort struct {
	ipAddr  net.IP
	portNum uint16
}

// GetURL returns a url you can use to open a connection to the engine
func (host hostMachineIpAndPort) GetURL() string {
	return fmt.Sprintf("%v:%v", host.ipAddr.String(), host.portNum)
}

// getDefaultKurtosisEngineLocalhostMachineIpAndPort is Used to default our engine connections to a server running on
// localhost on an expected port
// TODO Remove this hacky method
func getDefaultKurtosisEngineLocalhostMachineIpAndPort() *hostMachineIpAndPort {
	engineIp := net.ParseIP(localHostIpStr)

	return &hostMachineIpAndPort{
		ipAddr:  engineIp,
		portNum: kurtosis_context.DefaultGrpcEngineServerPortNum,
	}

}
