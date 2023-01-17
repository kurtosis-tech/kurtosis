package networking_sidecar

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type NetworkingSidecar struct {
	serviceGuid service.ServiceGUID
	enclaveUuid enclave.EnclaveUUID
	status      container_status.ContainerStatus
}

func NewNetworkingSidecar(serviceGuid service.ServiceGUID, enclaveUuid enclave.EnclaveUUID, status container_status.ContainerStatus) *NetworkingSidecar {
	return &NetworkingSidecar{serviceGuid: serviceGuid, enclaveUuid: enclaveUuid, status: status}
}

func (sidecar *NetworkingSidecar) GetServiceGUID() service.ServiceGUID {
	return sidecar.serviceGuid
}

func (sidecar *NetworkingSidecar) GetEnclaveUUID() enclave.EnclaveUUID {
	return sidecar.enclaveUuid
}

func (sidecar *NetworkingSidecar) GetStatus() container_status.ContainerStatus {
	return sidecar.status
}
