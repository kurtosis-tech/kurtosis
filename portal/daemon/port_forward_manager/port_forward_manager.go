package port_forward_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_gateway/port_utils"
	"github.com/kurtosis-tech/stacktrace"
	"strconv"
)

const (
	localhostIpString = "127.0.0.1"
)

type PortForwardManager struct {
	// gets relevant service information from Kurtosis
	serviceEnumerator *ServiceEnumerator

	// tracks active port forwarding tunnels and manages their lifecycle
	tunnelSessionTracker *TunnelSessionTracker
}

func NewPortForwardManager(serviceEnumerator *ServiceEnumerator, tracker *TunnelSessionTracker) *PortForwardManager {
	return &PortForwardManager{
		serviceEnumerator:    serviceEnumerator,
		tunnelSessionTracker: tracker,
	}
}

func (manager *PortForwardManager) Ping(ctx context.Context) error {
	return manager.serviceEnumerator.checkHealth(ctx)
}

// CreateUserServicePortForward
// This can run in two manners:
// 1. requestedLocalPort is specified: this will target only one (enclaveId, serviceId, portId), so all must be specified
// 2. requestedLocalPort is unspecified (0): we will bind all services to ephemeral local ports.  The list of services depends
// upon what's specified:
//   - (enclaveId, ): finds all services and ports within the enclave and binds them
//   - (enclaveId, serviceId): finds all ports in the given service and binds them
//   - (enclaveId, serviceId, portId): finds a specific service/port and binds that (similar to case 1 but ephemeral)
func (manager *PortForwardManager) CreateUserServicePortForward(ctx context.Context, enclaveServicePort EnclaveServicePort, requestedLocalPort uint16) (map[EnclaveServicePort]uint16, error) {
	if err := validateCreateUserServicePortForwardArgs(enclaveServicePort, requestedLocalPort); err != nil {
		return nil, stacktrace.Propagate(err, "Validation failed for arguments")
	}

	serviceInterfaceDetails, err := manager.serviceEnumerator.CollectServiceInformation(ctx, enclaveServicePort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to enumerate service information for (enclave, service, port), %v", enclaveServicePort)
	}

	allBoundPorts := map[EnclaveServicePort]uint16{}

	if requestedLocalPort != 0 {
		if len(serviceInterfaceDetails) != 1 {
			return nil, stacktrace.NewError("Creating a static binding for service %v, to local port %v. "+
				"Expected to find a single matching service but instead found %d", enclaveServicePort, requestedLocalPort, len(serviceInterfaceDetails))
		}

		sid := serviceInterfaceDetails[0]
		boundPortNumber, err := manager.tunnelSessionTracker.CreateAndOpenPortForward(sid, requestedLocalPort)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to open static port %d for service %v", requestedLocalPort, sid)
		}

		// use the enclave/server/port stored in service details, as this will be fully populated
		allBoundPorts[sid.enclaveServicePort] = boundPortNumber
	} else {
		boundPorts, err := manager.createAndOpenEphemeralPortForwardsToUserServices(serviceInterfaceDetails)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to open ephemeral ports for requested services %v", serviceInterfaceDetails)
		}

		for esp, port := range boundPorts {
			allBoundPorts[esp] = port
		}
	}

	return allBoundPorts, nil
}

// RemoveUserServicePortForward
// here we only stop a single session at a time, so require all of enclaveId, serviceId, portId, to be specified
func (manager *PortForwardManager) RemoveUserServicePortForward(ctx context.Context, enclaveServicePort EnclaveServicePort) error {
	if err := validateRemoveUserServicePortForwardArgs(enclaveServicePort); err != nil {
		return stacktrace.Propagate(err, "Validation failed for arguments")
	}

	manager.tunnelSessionTracker.StopForwardingPort(enclaveServicePort)

	return nil
}

func (manager *PortForwardManager) createAndOpenEphemeralPortForwardsToUserServices(serviceInterfaceDetails []*ServiceInterfaceDetail) (map[EnclaveServicePort]uint16, error) {
	allBoundPorts := map[EnclaveServicePort]uint16{}

	for _, sid := range serviceInterfaceDetails {
		ephemeralLocalPortSpec, err := port_utils.GetFreeTcpPort(localhostIpString)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not allocate a local port for the tunnel")
		}

		ephemeralLocalPort := ephemeralLocalPortSpec.GetNumber()
		boundPortNumber, err := manager.tunnelSessionTracker.CreateAndOpenPortForward(sid, ephemeralLocalPort)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to open ephemeral port %d for service %v", ephemeralLocalPort, sid)
		}

		allBoundPorts[sid.enclaveServicePort] = boundPortNumber
	}

	return allBoundPorts, nil
}

// Check for two modes of operation:
// 1. where a local port is requested, we need all of (enclaveId, serviceId, portId) to be specified; this has to target one service
// 2. if no local port is requested, we need at least enclaveId, and will target as many services as possible within the given context
func validateCreateUserServicePortForwardArgs(enclaveServicePort EnclaveServicePort, requestedLocalPort uint16) error {
	if enclaveServicePort.EnclaveId() == "" {
		return stacktrace.NewError("EnclaveId is always required but we received an empty string")
	}

	if enclaveServicePort.ServiceId() == "" && enclaveServicePort.PortId() != "" {
		return stacktrace.NewError("PortId (%v) was specified without a corresponding ServiceId", enclaveServicePort.PortId())
	}

	if requestedLocalPort != 0 {
		if enclaveServicePort.ServiceId() == "" || enclaveServicePort.PortId() == "" {
			return stacktrace.NewError("A static port '%d' was requested, but enclaveId, serviceId, and portId were not all specified: %v", requestedLocalPort, enclaveServicePort)
		}
	}

	return nil
}

// Removal only works for specific service/ports, so make sure all fields are populated
func validateRemoveUserServicePortForwardArgs(enclaveServicePort EnclaveServicePort) error {
	if enclaveServicePort.EnclaveId() == "" || enclaveServicePort.ServiceId() == "" || enclaveServicePort.PortId() == "" {
		return stacktrace.NewError("All of enclaveId, serviceId, and portId, must be specified for removal of a port forward: %v", enclaveServicePort)
	}
	return nil
}

func getLocalChiselServerUri(localPortToChiselServer uint16) string {
	return "localhost:" + strconv.Itoa(int(localPortToChiselServer))
}
