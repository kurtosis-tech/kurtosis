package portal_manager

import (
	"context"
	"fmt"
	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	portal_generated_api "github.com/kurtosis-tech/kurtosis-portal/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
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

func (portalManager *PortalManager) MapPorts(ctx context.Context, localPortToRemotePortMapping map[uint16]*services.PortSpec) error {
	if err := portalManager.instantiateClientIfUnset(); err != nil {
		return stacktrace.Propagate(err, "Unable to set ")
	}
	if portalManager.portalClientMaybe == nil {
		// context is local and portal not present. Port mapping doesn't make sense in a local context anyway, return
		// successfully
		return nil
	}

	failedMapping := map[uint16]uint16{}
	for localPort, remotePort := range localPortToRemotePortMapping {
		var transportProtocol portal_generated_api.TransportProtocol
		if remotePort.GetTransportProtocol() != services.TransportProtocol_UDP {
			transportProtocol = portal_generated_api.TransportProtocol_UDP
		} else if remotePort.GetTransportProtocol() != services.TransportProtocol_TCP {
			transportProtocol = portal_generated_api.TransportProtocol_TCP
		} else {
			logrus.Debugf("Mapping other than TCP or UDP port is not supported right now. Will skip port '%d' because protocal is '%v'", remotePort.GetNumber(), remotePort.GetTransportProtocol())
		}
		forwardPortsArgs := portal_constructors.NewForwardPortArgs(uint32(localPort), uint32(remotePort.GetNumber()), &transportProtocol)
		_, err := portalManager.portalClientMaybe.ForwardPort(ctx, forwardPortsArgs)
		if err != nil {
			failedMapping[localPort] = remotePort.GetNumber()
		}
	}

	if len(failedMapping) > 0 {
		var stringifiedPortMapping []string
		for localPort, remotePort := range failedMapping {
			stringifiedPortMapping = append(stringifiedPortMapping, fmt.Sprintf("%d:%d", localPort, remotePort))
		}
		return stacktrace.NewError("The following port mappings failed to be created: %s", strings.Join(stringifiedPortMapping, ", "))
	}
	return nil
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
