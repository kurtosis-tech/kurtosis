package portal_manager

import (
	"context"
	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	portal_generated_api "github.com/kurtosis-tech/kurtosis-portal/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	allowNilPortalClientForLocalContext = true
)

type PortalManager struct {
	// As it's fairly new, the portal daemon might not be running. If the context is local, it's not a problem and
	// therefore this being set to nil is fine. However, if the context is remote, this should be set.
	portalClientMaybe portal_generated_api.KurtosisPortalClientClient
}

func NewPortalManager() *PortalManager {
	return &PortalManager{
		portalClientMaybe: nil,
	}
}

// MapPorts maps a set of remote ports locally according to the mapping provided
// It returns the set of successfully mapped ports, and potential failed ports
// An error will be returned if the set of failed port is not empty
func (portalManager *PortalManager) MapPorts(ctx context.Context, localPortToRemotePortMapping map[uint16]*services.PortSpec) (map[uint16]*services.PortSpec, map[uint16]*services.PortSpec, error) {
	successfullyMappedPorts := map[uint16]*services.PortSpec{}
	failedPorts := map[uint16]*services.PortSpec{}
	if err := portalManager.instantiateClientIfUnset(); err != nil {
		failedPorts = localPortToRemotePortMapping
		return successfullyMappedPorts, failedPorts, stacktrace.Propagate(err, "Unable to instantiate a client to the Kurtosis Portal daemon")
	}
	if portalManager.portalClientMaybe == nil {
		successfullyMappedPorts = localPortToRemotePortMapping
		// context is local and portal not present. Port mapping doesn't make sense in a local context anyway, return
		// successfully
		logrus.Debug("Context is local, no ports to map via the Portal as they are naturally exposed")
		return successfullyMappedPorts, failedPorts, nil
	}

	for localPort, remotePort := range localPortToRemotePortMapping {
		var transportProtocol portal_generated_api.TransportProtocol
		if remotePort.GetTransportProtocol() == services.TransportProtocol_TCP {
			transportProtocol = portal_generated_api.TransportProtocol_TCP
		} else if remotePort.GetTransportProtocol() == services.TransportProtocol_UDP {
			transportProtocol = portal_generated_api.TransportProtocol_UDP
		} else {
			logrus.Warnf("Mapping other than TCP or UDP port is not supported right now. Will skip port '%d' because protocal is '%v'", remotePort.GetNumber(), remotePort.GetTransportProtocol())
		}
		forwardPortsArgs := portal_constructors.NewForwardPortArgs(uint32(localPort), uint32(remotePort.GetNumber()), &transportProtocol)
		if _, err := portalManager.portalClientMaybe.ForwardPort(ctx, forwardPortsArgs); err != nil {
			failedPorts[localPort] = remotePort
		} else {
			successfullyMappedPorts[localPort] = remotePort
		}
	}

	if len(failedPorts) > 0 {
		return successfullyMappedPorts, failedPorts, stacktrace.NewError("Some ports failed to be mapped")
	}
	return successfullyMappedPorts, failedPorts, nil
}

func (portalManager *PortalManager) instantiateClientIfUnset() error {
	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to retrieve current context")
	}

	portalDaemonClientMaybe, err := kurtosis_context.CreatePortalDaemonClient(currentContext, allowNilPortalClientForLocalContext)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to build client to Kurtosis Portal Daemon")
	}
	portalManager.portalClientMaybe = portalDaemonClientMaybe
	return nil
}
