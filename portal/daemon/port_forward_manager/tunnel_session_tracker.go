package port_forward_manager

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

// TODO(omar): there will be some complexity in cases where ephemeral port binds are upgraded to static

type TunnelSessionTracker struct {
	activePortForwards map[EnclaveServicePort]*PortForwardTunnel
}

func NewTunnelSessionTracker() *TunnelSessionTracker {
	return &TunnelSessionTracker{
		map[EnclaveServicePort]*PortForwardTunnel{},
	}
}

func (tracker *TunnelSessionTracker) CreateAndOpenPortForward(serviceInterfaceDetail *ServiceInterfaceDetail, localPortToBind uint16) (uint16, error) {
	// TODO(omar): what if a port forward already exists? do we need to be aware of static/ephemeral or can we remain oblivious?

	// TODO(omar): prob defer a close on portForward
	portForward, err := NewPortForwardTunnel(localPortToBind, serviceInterfaceDetail)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to initialise new port forward tunnel for service %v to local port %d", serviceInterfaceDetail, localPortToBind)
	}

	logrus.Infof("Opening port forward session on local port %d, to remote service %v", portForward.localPortNumber, serviceInterfaceDetail)
	err = portForward.RunAsync()
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to open a port forward tunnel to remote service %v", serviceInterfaceDetail)
	}
	// TODO(omar): do we need to wait until port is fully open?

	tracker.addPortForward(serviceInterfaceDetail.enclaveServicePort, portForward)
	return portForward.localPortNumber, nil
}

func (tracker *TunnelSessionTracker) StopForwardingPort(enclaveServicePort EnclaveServicePort) {
	// TODO(omar): i don't think we care about stopping sessions that have been removed right now
	// this depends on where we go wrt to monitoring and cleaning up dead sessions, so I'll see how that
	// evolves prior to doing anything here
	portForward, found := tracker.activePortForwards[enclaveServicePort]
	if found {
		portForward.Close()
	}
}

func (tracker *TunnelSessionTracker) addPortForward(enclaveServicePort EnclaveServicePort, portForward *PortForwardTunnel) {
	tracker.activePortForwards[enclaveServicePort] = portForward
}
