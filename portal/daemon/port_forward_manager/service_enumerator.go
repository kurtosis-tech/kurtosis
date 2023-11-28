package port_forward_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type ServiceEnumerator struct {
	kurtosis *kurtosis_context.KurtosisContext
}

func NewServiceEnumerator(kurtosisContext *kurtosis_context.KurtosisContext) *ServiceEnumerator {
	return &ServiceEnumerator{
		kurtosis: kurtosisContext,
	}
}

func (enumerator *ServiceEnumerator) findSingleService(ctx context.Context, esp EnclaveServicePort) error {
	return nil
}

func (enumerator *ServiceEnumerator) collectServiceInformation(ctx context.Context, enclaveServicePort EnclaveServicePort) (*ServiceInterfaceDetail, error) {
	enclave, err := enumerator.kurtosis.GetEnclave(ctx, enclaveServicePort.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to lookup enclave '%v' from Kurtosis Engine", enclaveServicePort.enclaveId)
	}

	enclaveContext, err := enumerator.kurtosis.GetEnclaveContext(ctx, enclaveServicePort.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get enclave context for enclave '%v'", enclaveServicePort.enclaveId)
	}

	serviceContext, err := enclaveContext.GetServiceContext(enclaveServicePort.serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service context for service '%v' in enclave '%v'", enclaveServicePort.serviceId, enclaveServicePort.enclaveId)
	}

	serviceIpAddress := serviceContext.GetPrivateIPAddress()
	servicePortSpec, exists := serviceContext.GetPrivatePorts()[enclaveServicePort.portId]
	if !exists {
		return nil, stacktrace.NewError("Failed to find requested port id specified %v.  Available ports are: %v", enclaveServicePort, serviceContext.GetPrivatePorts())
	}
	logrus.Debugf("Found service information for %v: service running at %v:%d in enclave: %v", enclaveServicePort, serviceIpAddress, servicePortSpec.GetNumber(), enclave.String())

	localPortToChiselServer := uint16(enclave.GetApiContainerHostMachineInfo().GetTunnelPortOnHostMachine())
	chiselServerUri := getLocalChiselServerUri(localPortToChiselServer)

	return NewServiceDetail(enclaveServicePort, chiselServerUri, serviceIpAddress, servicePortSpec), nil
}

// TODO(omar): get enclaves can take a while so look for a lighter ping that also verifies we've an engine connection
// or consider an alternative health indicator
func (enumerator *ServiceEnumerator) checkHealth(ctx context.Context) error {
	_, err := enumerator.kurtosis.GetEnclaves(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Port Forward Manager failed to contact Kurtosis Engine")
	}
	return nil
}
