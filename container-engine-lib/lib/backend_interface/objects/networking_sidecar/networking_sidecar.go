package networking_sidecar

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"net"
)

type NetworkingSidecar struct {
	serviceGuid service.ServiceGUID
	privateIpAddr net.IP // TODO Delete this when we solve the static IP problem
	enclaveId enclave.EnclaveID
}

func NewNetworkingSidecar(serviceGuid service.ServiceGUID, privateIpAddr net.IP, enclaveId enclave.EnclaveID) *NetworkingSidecar {
	return &NetworkingSidecar{serviceGuid: serviceGuid, privateIpAddr: privateIpAddr, enclaveId: enclaveId}
}

func (sidecar *NetworkingSidecar) GetServiceGuid() service.ServiceGUID {
	return sidecar.serviceGuid
}

func (sidecar *NetworkingSidecar) GetPrivateIpAddr() net.IP {
	return sidecar.privateIpAddr
}

func (sidecar *NetworkingSidecar) GetEnclaveId() enclave.EnclaveID {
	return sidecar.enclaveId
}
