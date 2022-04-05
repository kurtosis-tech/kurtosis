package networking_sidecar

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
)

type NetworkingSidecar struct {
	serviceGuid service.ServiceGUID
	enclaveId enclave.EnclaveID
	status container_status.ContainerStatus
}

func NewNetworkingSidecar(serviceGuid service.ServiceGUID, enclaveId enclave.EnclaveID, status container_status.ContainerStatus) *NetworkingSidecar {
	return &NetworkingSidecar{serviceGuid: serviceGuid, enclaveId: enclaveId, status: status}
}

func (sidecar *NetworkingSidecar) GetServiceGUID() service.ServiceGUID {
	return sidecar.serviceGuid
}

func (sidecar *NetworkingSidecar) GetEnclaveID() enclave.EnclaveID {
	return sidecar.enclaveId
}

func (sidecar *NetworkingSidecar) GetStatus() container_status.ContainerStatus {
	return sidecar.status
}
