package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"net"
)

// A ServiceRegistration is a stub for a soon-to-be-started service
// We have this as an independent object so we can return the container's IP
// address to the user before the container exists
type ServiceRegistration struct {
	id				 ServiceID
	guid             ServiceGUID
	enclaveId        enclave.EnclaveID
	privateIp        net.IP
}

func NewServiceRegistration(id ServiceID, guid ServiceGUID, enclaveId enclave.EnclaveID, privateIp net.IP) *ServiceRegistration {
	return &ServiceRegistration{id: id, guid: guid, enclaveId: enclaveId, privateIp: privateIp}
}

func (registration *ServiceRegistration) GetID() ServiceID {
	return registration.id
}

func (registration *ServiceRegistration) GetGUID() ServiceGUID {
	return registration.guid
}

func (registration *ServiceRegistration) GetEnclaveID() enclave.EnclaveID {
	return registration.enclaveId
}

func (registration *ServiceRegistration) GetPrivateIP() net.IP {
	return registration.privateIp
}
