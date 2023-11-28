package port_forward_manager

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type TunnelSessionTracker struct {
	activePortForwards map[ServiceInterfaceDetail]PortForwardTunnel
}

func NewTunnelSessionTracker() *TunnelSessionTracker {
	return &TunnelSessionTracker{
		map[ServiceInterfaceDetail]PortForwardTunnel{},
	}
}

func (tracker *TunnelSessionTracker) CreateAndOpenPortForward(serviceInterfaceDetail *ServiceInterfaceDetail, localPortToBind uint16) (uint16, error) {
	portForward := NewPortForwardTunnel(localPortToBind, serviceInterfaceDetail)
	logrus.Infof("Opening port forward session on local port %d, to remote service %v", portForward.localPortNumber, serviceInterfaceDetail)
	err := portForward.RunAsync()
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to open a port forward tunnel to remote service %v", serviceInterfaceDetail)
	}
	return portForward.localPortNumber, nil
}
