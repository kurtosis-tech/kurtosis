package reverse_proxy

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"net"
)

// This component is responsible for routing http traffic to the services
type ReverseProxy struct {
	status container.ContainerStatus

	// This will be nil if the container is not running
	maybePrivateIpAddr net.IP

	// PortNum that container will listen for logs on
	logsListeningPortNum uint16
}

func NewReverseProxy(
	status container.ContainerStatus,
	maybePrivateIpAddr net.IP) *ReverseProxy {
	return &ReverseProxy{
		status:               status,
		maybePrivateIpAddr:   maybePrivateIpAddr}
}

func (reverseProxy *ReverseProxy) GetStatus() container.ContainerStatus {
	return reverseProxy.status
}

func (reverseProxy *ReverseProxy) GetMaybePrivateIpAddr() net.IP {
	return reverseProxy.maybePrivateIpAddr
}

// Returns port number that logs aggregator listens for logs on
func (reverseProxy *ReverseProxy) GetListeningPortNum() uint16 {
	return reverseProxy.logsListeningPortNum
}
