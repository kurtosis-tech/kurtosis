package port_forward_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
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

func (enumerator *ServiceEnumerator) CollectServiceInformation(ctx context.Context, enclaveServicePort EnclaveServicePort) ([]*ServiceInterfaceDetail, error) {
	enclave, err := enumerator.kurtosis.GetEnclave(ctx, enclaveServicePort.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to lookup enclave '%v' from Kurtosis Engine", enclaveServicePort.enclaveId)
	}

	enclaveContext, err := enumerator.kurtosis.GetEnclaveContext(ctx, enclaveServicePort.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get enclave context for enclave '%v'", enclaveServicePort.enclaveId)
	}

	if enclaveServicePort.ServiceId() == "" {
		// enumerate the entire enclave
		return enumerateServices(enclave, enclaveContext)
	} else if enclaveServicePort.PortId() == "" {
		// enumerate all ports of the service
		return enumeratePortsForService(enclave, enclaveContext, enclaveServicePort.ServiceId())
	} else {
		// "enumerate" a specific service/port only
		serviceContext, err := enclaveContext.GetServiceContext(enclaveServicePort.ServiceId())
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get service context for service '%v' in enclave '%v'", enclaveServicePort.serviceId, enclaveServicePort.enclaveId)
		}

		serviceDetail, err := getServicePortDetail(enclave, serviceContext, enclaveServicePort.ServiceId(), enclaveServicePort.PortId())
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get service details for service %v", enclaveServicePort)
		}
		return []*ServiceInterfaceDetail{serviceDetail}, nil
	}
}

func enumerateServices(enclave *kurtosis_engine_rpc_api_bindings.EnclaveInfo, enclaveContext *enclaves.EnclaveContext) ([]*ServiceInterfaceDetail, error) {
	var allServicePorts []*ServiceInterfaceDetail

	serviceNames, err := listServiceNamesInEnclave(enclaveContext)
	if err != nil {
		return nil, err
	}

	for _, serviceName := range serviceNames {
		serviceDetails, err := enumeratePortsForService(enclave, enclaveContext, serviceName)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to fetch details for a port of service %v", serviceName)
		}

		allServicePorts = append(allServicePorts, serviceDetails...)
	}

	return allServicePorts, nil
}

func enumeratePortsForService(enclave *kurtosis_engine_rpc_api_bindings.EnclaveInfo, enclaveContext *enclaves.EnclaveContext, serviceId string) ([]*ServiceInterfaceDetail, error) {
	serviceContext, err := enclaveContext.GetServiceContext(serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service context for service '%v' in enclave '%v'", serviceId, enclave.GetName())
	}

	var allServiceDetails []*ServiceInterfaceDetail
	for portId, _ := range serviceContext.GetPrivatePorts() {
		serviceDetail, err := getServicePortDetail(enclave, serviceContext, serviceId, portId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to fetch service detail for (%v, %v, %v)", enclave.GetName(), serviceId, portId)
		}
		allServiceDetails = append(allServiceDetails, serviceDetail)
	}
	return allServiceDetails, nil
}

func getServicePortDetail(enclave *kurtosis_engine_rpc_api_bindings.EnclaveInfo, serviceContext *services.ServiceContext, serviceId string, portId string) (*ServiceInterfaceDetail, error) {
	serviceIpAddress := serviceContext.GetPrivateIPAddress()
	servicePortSpec, exists := serviceContext.GetPrivatePorts()[portId]
	if !exists {
		return nil, stacktrace.NewError("Failed to find requested port id specified %v.  Available ports are: %v", portId, serviceContext.GetPrivatePorts())
	}
	logrus.Debugf("Found service information for %v-%v: service running at %v:%d in enclave: %v", serviceId, portId, serviceIpAddress, servicePortSpec.GetNumber(), enclave.String())

	localPortToChiselServer := uint16(enclave.GetApiContainerHostMachineInfo().GetTunnelPortOnHostMachine())
	chiselServerUri := getLocalChiselServerUri(localPortToChiselServer)

	enclaveServicePort := NewEnclaveServicePort(enclave.GetName(), serviceId, portId)
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

func listServiceNamesInEnclave(enclaveContext *enclaves.EnclaveContext) ([]string, error) {
	services, err := enclaveContext.GetServices()
	if err != nil {
		return []string{}, stacktrace.Propagate(err, "Failed to list services in enclave '%v'", enclaveContext.GetEnclaveName())
	}
	serviceNames := []string{}
	for name := range services {
		serviceNames = append(serviceNames, string(name))
	}
	return serviceNames, nil
}
