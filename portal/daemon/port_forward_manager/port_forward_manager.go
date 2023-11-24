package port_forward_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	chiselPortInApicReplaceWithLookupLater = 9501
)

type PortForwardManager struct {
	kurtosis *kurtosis_context.KurtosisContext
}

func NewPortForwardManager(kurtosisContext *kurtosis_context.KurtosisContext) *PortForwardManager {
	return &PortForwardManager{
		kurtosis: kurtosisContext,
	}
}

// TODO(omar): get enclaves can take a moment so look for a lighter ping that also verifies we've an engine connection
// or consider an alternative health indicator
func (manager *PortForwardManager) Ping(ctx context.Context) error {
	_, err := manager.kurtosis.GetEnclaves(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Port Forward Manager failed to contact Kurtosis Engine")
	}
	return nil
}

// TODO(omar): make a return struct - see what we end up using to represent port forwards
func (manager *PortForwardManager) ForwardUserServiceToEphemeralPort(ctx context.Context, enclaveId string, serviceId string, portId string) (uint16, error) {
	enclave, err := manager.kurtosis.GetEnclave(ctx, enclaveId)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to lookup enclave '%v' from Kurtosis Engine", enclaveId)
	}

	enclaveContext, err := manager.kurtosis.GetEnclaveContext(ctx, enclaveId)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to get enclave context for enclave '%v'", enclaveId)
	}

	serviceContext, err := enclaveContext.GetServiceContext(serviceId)
	if err != nil {
		return 0, stacktrace.Propagate(err, "Failed to get service context for service '%v' in enclave '%v'", serviceId, enclaveId)
	}

	serviceIpAddress := serviceContext.GetPrivateIPAddress()
	privatePortSpec, exists := serviceContext.GetPrivatePorts()[portId]
	if !exists {
		return 0, stacktrace.NewError("Failed to find requested port id '%v' in service '%v' in enclave '%v'.  Available ports are: %v", portId, serviceId, enclaveId, serviceContext.GetPrivatePorts())
	}

	logrus.Debugf("Found service information for (%v, %v, %v): service running at %v:%d in enclave: %v", enclaveId, serviceId, portId, serviceIpAddress, privatePortSpec.GetNumber(), enclave.String())
	return 0, nil
}

func (manager *PortForwardManager) ForwardUserServiceToStaticPort(ctx context.Context, enclaveId string, serviceId string, portId string, localPortNumber uint16) (uint16, error) {
	return 0, nil
}
