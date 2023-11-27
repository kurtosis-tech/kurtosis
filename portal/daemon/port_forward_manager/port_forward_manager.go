package port_forward_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_gateway/port_utils"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	localhostIpString = "127.0.0.1"
)

type PortForwardManager struct {
	kurtosis *kurtosis_context.KurtosisContext
}

func NewPortForwardManager(kurtosisContext *kurtosis_context.KurtosisContext) *PortForwardManager {
	return &PortForwardManager{
		kurtosis: kurtosisContext,
	}
}

// TODO(omar): get enclaves can take a while so look for a lighter ping that also verifies we've an engine connection
// or consider an alternative health indicator
func (manager *PortForwardManager) Ping(ctx context.Context) error {
	_, err := manager.kurtosis.GetEnclaves(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Port Forward Manager failed to contact Kurtosis Engine")
	}
	return nil
}

// TODO(omar): make a return struct - see what we end up using to represent port forwards
func (manager *PortForwardManager) CreateUserServicePortForward(ctx context.Context, enclaveServicePort EnclaveServicePort, requestedLocalPort uint16) (map[EnclaveServicePort]uint16, error) {
	// TODO(omar): arg validation galore; figure out the scope of the forwarding then execute
	// might be worth separating the single service (with static port option) from the full ephemeral case

	if requestedLocalPort == 0 {
		ephemeralLocalPortSpec, err := port_utils.GetFreeTcpPort(localhostIpString)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Could not allocate a local port for the tunnel")
		}

		requestedLocalPort = ephemeralLocalPortSpec.GetNumber()
	}

	portForward, err := manager.createAndOpenPortForwardToUserService(ctx, enclaveServicePort, requestedLocalPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to set up port forward to (enclave, service, port), %v", enclaveServicePort)
	}

	return map[EnclaveServicePort]uint16{enclaveServicePort: portForward.localPortNumber}, nil
}

func (manager *PortForwardManager) RemoveUserServicePortForward(ctx context.Context, enclaveServicePort EnclaveServicePort) error {
	panic("implement me")
}

func (manager *PortForwardManager) createAndOpenPortForwardToUserService(ctx context.Context, enclaveServicePort EnclaveServicePort, localPortToBind uint16) (*PortForwardTunnel, error) {
	chiselServerUri, serviceIpAddress, servicePortNumber, err := manager.collectServiceInformation(ctx, enclaveServicePort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to enumerate service information for (enclave, service, port), %v", enclaveServicePort)
	}

	logrus.Debugf("Will connect to chisel server at %v, setting up a tunnel to service %v running at %v:%d", chiselServerUri, enclaveServicePort, serviceIpAddress, servicePortNumber)

	portForward := NewPortForwardTunnel(localPortToBind, serviceIpAddress, servicePortNumber, chiselServerUri)

	logrus.Infof("Opening port forward session on local port %d, to remote service %v at %v:%d", portForward.localPortNumber, enclaveServicePort, serviceIpAddress, servicePortNumber)
	err = portForward.RunAsync()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to open a port forward tunnel to chisel server '%v' for remote service at '%v:%d'", chiselServerUri, serviceIpAddress, servicePortNumber)
	}

	return portForward, nil
}

func (manager *PortForwardManager) collectServiceInformation(ctx context.Context, enclaveServicePort EnclaveServicePort) (string, string, uint16, error) {
	enclave, err := manager.kurtosis.GetEnclave(ctx, enclaveServicePort.enclaveId)
	if err != nil {
		return "", "", 0, stacktrace.Propagate(err, "Failed to lookup enclave '%v' from Kurtosis Engine", enclaveServicePort.enclaveId)
	}

	enclaveContext, err := manager.kurtosis.GetEnclaveContext(ctx, enclaveServicePort.enclaveId)
	if err != nil {
		return "", "", 0, stacktrace.Propagate(err, "Failed to get enclave context for enclave '%v'", enclaveServicePort.enclaveId)
	}

	serviceContext, err := enclaveContext.GetServiceContext(enclaveServicePort.serviceId)
	if err != nil {
		return "", "", 0, stacktrace.Propagate(err, "Failed to get service context for service '%v' in enclave '%v'", enclaveServicePort.serviceId, enclaveServicePort.enclaveId)
	}

	serviceIpAddress := serviceContext.GetPrivateIPAddress()
	servicePortSpec, exists := serviceContext.GetPrivatePorts()[enclaveServicePort.portId]
	if !exists {
		return "", "", 0, stacktrace.NewError("Failed to find requested port id specified %v.  Available ports are: %v", enclaveServicePort, serviceContext.GetPrivatePorts())
	}
	logrus.Debugf("Found service information for %v: service running at %v:%d in enclave: %v", enclaveServicePort, serviceIpAddress, servicePortSpec.GetNumber(), enclave.String())

	localPortToChiselServer := uint16(enclave.GetApiContainerHostMachineInfo().GetTunnelPortOnHostMachine())
	chiselServerUri := getLocalChiselServerUri(localPortToChiselServer)
	return chiselServerUri, serviceIpAddress, servicePortSpec.GetNumber(), nil
}

func getLocalChiselServerUri(localPortToChiselServer uint16) string {
	return "localhost:" + strconv.Itoa(int(localPortToChiselServer))
}
