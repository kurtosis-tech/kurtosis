package networking_sidecar

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"net"
)

type NetworkingSidecar struct {
	serviceGuid service.ServiceGUID
	privateIpAddr net.IP // TODO Delete this when we solve the static IP problem
	enclaveId enclave.EnclaveID
	status container_status.ContainerStatus
}

func NewNetworkingSidecar(serviceGuid service.ServiceGUID, privateIpAddr net.IP, enclaveId enclave.EnclaveID, status container_status.ContainerStatus) *NetworkingSidecar {
	return &NetworkingSidecar{serviceGuid: serviceGuid, privateIpAddr: privateIpAddr, enclaveId: enclaveId, status: status}
}

func (sidecar *NetworkingSidecar) GetServiceGUID() service.ServiceGUID {
	return sidecar.serviceGuid
}

func (sidecar *NetworkingSidecar) GetPrivateIpAddr() net.IP {
	return sidecar.privateIpAddr
}

func (sidecar *NetworkingSidecar) GetEnclaveID() enclave.EnclaveID {
	return sidecar.enclaveId
}

func (sidecar *NetworkingSidecar) GetStatus() container_status.ContainerStatus {
	return sidecar.status
}
