package networking_sidecar

import (
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/service"
)

type NetworkingSidecar struct {
	serviceUuid service.ServiceUUID
	enclaveUuid enclave.EnclaveUUID
	status      container_status.ContainerStatus
}

func NewNetworkingSidecar(serviceUuid service.ServiceUUID, enclaveUuid enclave.EnclaveUUID, status container_status.ContainerStatus) *NetworkingSidecar {
	return &NetworkingSidecar{serviceUuid: serviceUuid, enclaveUuid: enclaveUuid, status: status}
}

func (sidecar *NetworkingSidecar) GetServiceUUID() service.ServiceUUID {
	return sidecar.serviceUuid
}

func (sidecar *NetworkingSidecar) GetEnclaveUUID() enclave.EnclaveUUID {
	return sidecar.enclaveUuid
}

func (sidecar *NetworkingSidecar) GetStatus() container_status.ContainerStatus {
	return sidecar.status
}
